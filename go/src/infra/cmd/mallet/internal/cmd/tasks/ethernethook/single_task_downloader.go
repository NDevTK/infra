// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ethernethook

import (
	"context"
	"io"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
	"go.chromium.org/luci/common/errors"
)

// patternMap holds the regular expression patterns for different kinds of files.
type patternMap struct {
	resultPat      string
	recoverDUTsPat string
	lspciPat       string
	dmesgPat       string
}

// values returns the patterns in order.
func (p patternMap) Values() []string {
	return []string{
		p.resultPat,
		p.recoverDUTsPat,
		p.lspciPat,
		p.dmesgPat,
	}
}

// patterns are a map from names to regular expressions.
var patterns = patternMap{
	resultPat:      `result_summary.html\z`,
	recoverDUTsPat: `recover_duts.log\z`,
	lspciPat:       `sysinfo/lspci\z`,
	dmesgPat:       `dmesg.gz\z`,
}

// NewSingleTaskDownloader creates an object that manages downloads corresponding to a single swarming task.
func NewSingleTaskDownloader(bucket string, prefix string) (*singleTaskDownloader, error) {
	var out singleTaskDownloader
	if bucket == "" {
		return nil, errors.Reason("new single task downloader: bucket cannot be empty").Err()
	}
	if prefix == "" {
		return nil, errors.Reason("new single task downloader: prefix cannot be empty").Err()
	}
	out.bucket = bucket
	out.prefix = prefix
	query := &storage.Query{
		Prefix: prefix,
	}
	d, err := NewRegexDownloader(bucket, query, patterns)
	if err != nil {
		return nil, errors.Annotate(err, "new single task downloader").Err()
	}
	out.downloader = d
	return &out, nil
}

// Entry is a pair consisting of a GSURL and the contents of the file in Google Storage.
type Entry struct {
	// Name is a human-readable short name for the type of entity, like "results".
	Name string
	// GSURL is the google storage location of the item.
	GSURL string
	// Content is the actual data. It is not the *raw* content. We will, for example,
	// decompress the data if it is compressed.
	Content string
}

// singleTaskDownloader manages the downloads for a single task.
type singleTaskDownloader struct {
	bucket     string
	prefix     string
	downloader *regexDownloader
	// OutputArr is an array of individual files and their contents.
	OutputArr []Entry
	// SwarmingTaskID is the discovered swarming task ID.
	SwarmingTaskID string
}

// ProcessTask reads the contents of the task and populates the output map.
func (s *singleTaskDownloader) ProcessTask(ctx context.Context, e *extendedGSClient) error {
	if err := s.downloader.FindPaths(ctx, e); err != nil {
		return errors.Annotate(err, "process task").Err()
	}
	if err := s.FindResultsSummary(ctx, e); err != nil {
		return errors.Annotate(err, "process task").Err()
	}
	if err := s.FindRecoverLog(ctx, e); err != nil {
		return errors.Annotate(err, "process task").Err()
	}
	return nil
}

// Len gets the number of items scanned.
func (s *singleTaskDownloader) Len() int {
	return s.downloader.Len()
}

// FindResultsSummary finds the results_summary.html and records its output.
func (s *singleTaskDownloader) FindResultsSummary(ctx context.Context, e *extendedGSClient) error {
	var entry Entry
	if ok := s.Len() > 0; !ok {
		return errors.Reason("find results summary: no results were read").Err()
	}
	for _, attrs := range s.downloader.Attrs {
		name := e.ExpandName(s.bucket, attrs)
		if ok := regexp.MustCompile(`result_summary.html\z`).MatchString(name); !ok {
			continue
		}
		entry.Name = "results_summary"
		entry.GSURL = name
		reader, err := e.Bucket(s.bucket).Object(attrs.Name).NewReader(ctx)
		if err != nil {
			// If we didn't abandon the loop earlier, then this error really is unrecoverable.
			// We have to know what's in the file.
			return errors.Reason("find results summary: failed to instantiate reader for %q", name).Err()
		}
		buf := new(strings.Builder)
		if _, err := io.Copy(buf, reader); err != nil {
			return errors.Reason("find results summary: failed to read contents of %q", name).Err()
		}
		entry.Content = buf.String()
		s.OutputArr = append(s.OutputArr, entry)
		return nil
	}
	return errors.Reason("find results summary: no result found").Err()
}

// FindRecoverLog finds and attaches the recovery log.
func (s *singleTaskDownloader) FindRecoverLog(ctx context.Context, e *extendedGSClient) error {
	var entry Entry
	if ok := s.Len() > 0; !ok {
		return errors.Reason("find results summary: no results were read").Err()
	}
	for _, attrs := range s.downloader.Attrs {
		name := e.ExpandName(s.bucket, attrs)
		if ok := regexp.MustCompile(`recover_duts.log\z`).MatchString(name); !ok {
			continue
		}
		entry.GSURL = name
		reader, err := e.Bucket(s.bucket).Object(attrs.Name).NewReader(ctx)
		if err != nil {
			// If we didn't abandon the loop earlier, then this error really is unrecoverable.
			// We have to know what's in the file.
			return errors.Reason("find results summary: failed to instantiate reader for %q", name).Err()
		}
		buf := new(strings.Builder)
		if _, err := io.Copy(buf, reader); err != nil {
			return errors.Reason("find results summary: failed to read contents of %q", name).Err()
		}
		entry.Content = buf.String()
		s.OutputArr = append(s.OutputArr, entry)
		return nil
	}
	return errors.Reason("find results summary: no result found").Err()
}
