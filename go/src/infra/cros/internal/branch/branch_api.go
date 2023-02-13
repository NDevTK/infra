// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package branch

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"infra/cros/internal/shared"

	gerritapi "github.com/andygrunwald/go-gerrit"
	"go.chromium.org/luci/common/errors"
	"golang.org/x/sync/errgroup"
)

const (
	retriesTimeout = 5 * time.Minute
	testRetries    = 1
	testBaseDelay  = 1
)

// GerritProjectBranch contains all the details for creating a new Gerrit branch
// based on an existing one.
type GerritProjectBranch struct {
	GerritURL string
	Project   string
	Branch    string
	SrcRef    string
}

func qpsToPeriod(qps float64) time.Duration {
	if qps <= 0 {
		// some very generous default duration
		return time.Second * 10
	}
	periodSec := float64(time.Second) / qps
	return time.Duration(int64(periodSec))
}

func (c *Client) createRemoteBranch(authedClient *http.Client, b GerritProjectBranch, dryRun bool, errs chan error) error {
	if dryRun {
		return nil
	}
	agClient, err := gerritapi.NewClient(b.GerritURL, authedClient)
	if err != nil {
		clientError := fmt.Errorf("failed to create Gerrit client: %v", err)
		errs <- clientError
		return clientError
	}
	bi, resp, err := agClient.Projects.CreateBranch(b.Project, b.Branch, &gerritapi.BranchInput{Revision: b.SrcRef})
	defer resp.Body.Close()
	if err != nil {
		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			// shouldn't happen
			errs <- err2
			return err2
		}
		if resp.StatusCode == http.StatusConflict && strings.Contains(string(body), "already exists") {
			// Branch already exists, so there's nothing to do.
			c.LogOut("branch %s already exists for %s/%s, nothing to do here", b.Branch, b.GerritURL, b.Project)
			return nil
		}
		err = errors.Annotate(err, "failed to create branch. Got response %v and branch info %v", string(body), bi).Err()
		errs <- err
		return err
	}
	return nil
}

// CreateRemoteBranchesAPI creates a bunch of branches on remote Gerrit instances
// for the specified inputs using the Gerrit API.
func (c *Client) CreateRemoteBranchesAPI(authedClient *http.Client, branches []GerritProjectBranch, dryRun bool, gerritQPS float64, skipRetries bool, isTest bool) error {
	if c.FakeCreateRemoteBranchesAPI != nil {
		return c.FakeCreateRemoteBranchesAPI(authedClient, branches, dryRun, gerritQPS)
	}

	if dryRun {
		c.LogOut("Dry run (no --push): would create remote branches for %v Gerrit repos", len(branches))
	} else {
		c.LogOut("Creating remote branches for %v Gerrit repos. This will take a few minutes, since otherwise Gerrit would throttle us.", len(branches))
	}
	var g errgroup.Group
	throttle := time.Tick(qpsToPeriod(gerritQPS))

	var logPrefix string
	if dryRun {
		logPrefix = "(Dry run) "
	}

	errs := make(chan error, len(branches))
	var createCount int64
	for _, b := range branches {
		<-throttle
		b := b
		g.Go(func() error {
			if skipRetries {
				err := c.createRemoteBranch(authedClient, b, dryRun, errs)
				if err != nil {
					return err
				}
			} else {
				ctx, cancel := context.WithTimeout(context.Background(), retriesTimeout)
				defer cancel()
				opts := shared.DefaultOpts
				if isTest {
					opts.Retries = testRetries
					opts.BaseDelay = testBaseDelay
				}
				err := shared.DoWithRetry(ctx, opts, func() error {
					err := c.createRemoteBranch(authedClient, b, dryRun, errs)
					return err
				})
				if err != nil {
					return err
				}
			}
			count := atomic.AddInt64(&createCount, 1)
			if count%10 == 0 {
				c.LogOut("%sCreated %v of %v remote branches", logPrefix, count, len(branches))
			}
			return nil
		})
	}
	_ = g.Wait()
	close(errs)
	c.LogOut("%sSuccessfully created %v of %v remote branches", logPrefix, atomic.LoadInt64(&createCount), len(branches))
	errorsArray := errors.NewMultiError()
	for err := range errs {
		errorsArray.MaybeAdd(err)
	}
	if len(errorsArray) == 0 {
		return nil
	}
	return errorsArray
}

// CheckSelfGroupMembership checks if the authenticated user is in the given
// group on the given Gerrit host. It returns a bool indicating whether or
// not that's the case, or an error if the lookup fails.
func CheckSelfGroupMembership(authedClient *http.Client, gerritURL, expectedGroup string) (bool, error) {
	agClient, err := gerritapi.NewClient(gerritURL, authedClient)
	if err != nil {
		return false, fmt.Errorf("failed to create Gerrit client: %v", err)
	}
	groups, resp, err := agClient.Accounts.ListGroups("self")
	defer resp.Body.Close()
	if err != nil {
		return false, errors.Annotate(err, "failed to get list of Gerrit groups for self").Err()
	}
	for _, g := range *groups {
		if g.Name == expectedGroup {
			return true, nil
		}
	}
	return false, nil
}
