// Copyright 2023 The ChromiumOS Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package gerrit

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"infra/cros/support/internal/shared"

	gerrit2 "go.chromium.org/luci/common/proto/gerrit"

	"go.chromium.org/luci/common/api/gerrit"
)

const (
	shortHostSuffix = "-review.googlesource.com"
)

type Change struct {
	// Full (chromium-review.googlesource.com) or short (chromium) gerrit host.
	Host string `json:"host"`
	// Change number requested.
	Number int `json:"change_number"`
	// Patch set number as requested. If in (-1, 0), fetch "current" patch set.
	PatchSet int `json:"patch_set"`

	// Change info if found.
	Info *gerrit.Change `json:"info"`
	// Patch set info if found.
	PatchSetRevision string               `json:"patch_set_revision"`
	RevisionInfo     *gerrit.RevisionInfo `json:"revision_info"`
}

type Changes []*Change

type Options struct {
	IncludeDetailedLabels bool `json:"include_detailed_labels"`
	IncludeFiles          bool `json:"include_files"`
	IncludeCommitInfo     bool `json:"include_commit_info"`
	IncludeMessages       bool `json:"include_messages"`
}

func changesToQueryParams(changes Changes, options Options) gerrit.ChangeQueryParams {
	var (
		queryOrs        []string
		currentRevision = false
		allRevisions    = false
	)
	for _, change := range changes {
		queryOrs = append(queryOrs, fmt.Sprintf("change:{%d}", change.Number))
		if change.PatchSet == -1 || change.PatchSet == 0 {
			currentRevision = true
		} else {
			allRevisions = true
		}
	}
	queryOpts := []string{}
	if allRevisions {
		queryOpts = append(queryOpts, "ALL_REVISIONS")
		if options.IncludeCommitInfo {
			queryOpts = append(queryOpts, "ALL_COMMITS")
		}
	} else if currentRevision {
		queryOpts = append(queryOpts, "CURRENT_REVISION")
		if options.IncludeCommitInfo {
			queryOpts = append(queryOpts, "CURRENT_COMMIT")
		}
	}
	if options.IncludeFiles {
		queryOpts = append(queryOpts, "ALL_FILES")
	}
	if options.IncludeDetailedLabels {
		queryOpts = append(queryOpts, "DETAILED_LABELS")
	}
	if options.IncludeMessages {
		queryOpts = append(queryOpts, "MESSAGES")
	}
	return gerrit.ChangeQueryParams{
		Query:   strings.Join(queryOrs, " OR "),
		N:       len(changes),
		Options: queryOpts,
	}
}

func (c *Change) updateChangeFromResults(results []*gerrit.Change) {
	for _, candidate := range results {
		if candidate.ChangeNumber == c.Number {
			c.Info = candidate
			break
		}
	}
	if c.Info == nil {
		return
	}

	var foundRev string
	if c.PatchSet == -1 || c.PatchSet == 0 {
		foundRev = c.Info.CurrentRevision
	} else {
		for rev, revInfo := range c.Info.Revisions {
			if revInfo.PatchSetNumber == c.PatchSet {
				foundRev = rev
				break
			}
		}
	}
	if revInfo, ok := c.Info.Revisions[foundRev]; ok {
		c.PatchSetRevision = foundRev
		c.RevisionInfo = &revInfo
	}
	c.Info.Revisions = nil
}

func batchSlice(changes Changes) []Changes {
	batchSize := 10
	batches := make([]Changes, 0, (len(changes)+batchSize-1)/batchSize)
	for batchSize < len(changes) {
		changes, batches = changes[batchSize:], append(batches, changes[0:batchSize:batchSize])
	}
	batches = append(batches, changes)
	return batches
}

func fetchHostChanges(
	ctx context.Context, httpClient *http.Client,
	host string, changes Changes, options Options,
) error {
	client, err := gerrit.NewClient(httpClient, fmt.Sprintf("https://%s", host))
	if err != nil {
		return err
	}

	changesBatches := batchSlice(changes)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	for _, batch := range changesBatches {
		queryParams := changesToQueryParams(batch, options)
		ch := make(chan []*gerrit.Change, 1)
		err = shared.DoWithRetry(ctx, shared.DefaultOpts, func() error {
			results, more, err := client.ChangeQuery(ctx, queryParams)
			if err != nil {
				return err
			}
			if more {
				// Shouldn't happen, but log just in case.
				log.Print("WARNING: more results than expected!")
			}
			ch <- results
			return nil
		})
		if err != nil {
			return fmt.Errorf("DoWithRetry ChangeQuery: %v", err)
		}
		results := <-ch
		for _, c := range batch {
			c.updateChangeFromResults(results)
			// In some cases (e.g. merge commit), GetChange doesn't return a file list.
			// We thus call into the ListFiles endpoint instead.
			if options.IncludeFiles && len(c.RevisionInfo.Files) == 0 {
				c.fetchFileList(ctx, httpClient)
			}
		}
	}
	return nil
}

func (c *Change) fetchFileList(ctx context.Context, httpClient *http.Client) error {
	rest, err := gerrit.NewRESTClient(httpClient, c.Host, true)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	ch := make(chan *gerrit2.ListFilesResponse, 1)
	err = shared.DoWithRetry(ctx, shared.DefaultOpts, func() error {
		// The "Parent: 1" is what makes ListFiles able to get file lists for merge commits.
		// It's a 1-indexed way to reference parent commits, and we always want a value of 1
		// in order to get the target branch ref.
		resp, err := rest.ListFiles(ctx, &gerrit2.ListFilesRequest{Number: int64(c.Number), RevisionId: "current", Parent: 1})
		if err != nil {
			return err
		}
		ch <- resp
		return nil
	})
	if err != nil {
		return fmt.Errorf("DoWithRetry ListFiles: %v", err)
	}
	results := <-ch
	c.RevisionInfo.Files = make(map[string]gerrit.FileInfo)
	for filename := range results.Files {
		c.RevisionInfo.Files[filename] = gerrit.FileInfo{}
	}
	return nil
}

// Fetch changes from the given hosts (will only make one request per host) or die.
func MustFetchChanges(parentCtx context.Context, httpClient *http.Client, changes Changes, options Options) Changes {
	// Group changes by host.
	hostChanges := make(map[string]Changes)
	for _, c := range changes {
		host := c.Host
		if !strings.ContainsRune(host, '.') {
			host = host + shortHostSuffix
		}
		hostChanges[host] = append(hostChanges[host], c)
	}

	// Error management for parallel requests.
	var hostErrors sync.Map
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Minute)
	defer cancel()

	// Parallel request per host.
	for host, changes := range hostChanges {
		// Copy loop variables into scope.
		host, changes := host, changes
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := fetchHostChanges(ctx, httpClient, host, changes, options)
			if err != nil {
				hostErrors.Store(host, err)
			}
		}()
	}
	wg.Wait()

	failed := false
	hostErrors.Range(func(host, err interface{}) bool {
		log.Printf("request to %s failed: %v", host, err)
		failed = true
		return true
	})
	if failed {
		os.Exit(1)
	}

	return changes
}
