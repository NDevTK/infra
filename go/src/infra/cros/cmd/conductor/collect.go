// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"infra/cros/cmd/try/try"
	"infra/cros/internal/shared"

	"go.chromium.org/luci/common/errors"

	pb "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	bbpb "go.chromium.org/luci/buildbucket/proto"
)

// retryBuild retries a build using `cros try retry`. It returns the BBID of the
// new build.
func (c *collectRun) retryBuild(build *bbpb.Build) (string, error) {
	opts := &try.RetryRunOpts{
		StdoutLog: c.stdoutLog,
		StderrLog: c.stderrLog,
		CmdRunner: c.cmdRunner,

		BBID:   fmt.Sprintf("%d", build.GetId()),
		Dryrun: c.dryrun,
	}
	return c.tryClient.DoRetry(opts)
}

func (c *collectRun) Collect(ctx context.Context, config *pb.CollectConfig) error {
	state := initCollectState(config)
	watchSet := c.bbids

	pollingDelay := time.Duration(c.pollingIntervalSeconds) * time.Second
	// Retry `bb get` as needed, waiting multiples of the polling interval.
	// Want to include a little backoff in case there's a legimiate outage.
	bbRetryOpts := shared.Options{
		BaseDelay:   pollingDelay,
		BackoffBase: 1.5,
		Retries:     5,
	}

	errs := []error{}
	for len(watchSet) > 0 {
		c.LogOut("Sleeping for %d seconds", c.pollingIntervalSeconds)
		time.Sleep(pollingDelay)

		sort.Strings(watchSet)
		c.LogOut("Waiting for %s", strings.Join(watchSet, ","))

		ch := make(chan []*bbpb.Build, 1)
		err := shared.DoWithRetry(ctx, bbRetryOpts, func() error {
			builds, err := c.bbClient.GetBuilds(context.Background(), watchSet)
			if err != nil {
				return err
			}
			ch <- builds
			return nil
		})
		if err != nil {
			return err
		}
		builds := <-ch

		watchSetMap := map[string]bool{}
		for _, bbid := range watchSet {
			watchSetMap[bbid] = true
		}

		newWatchSet := []string{}
		for _, build := range builds {
			if _, ok := watchSetMap[fmt.Sprintf("%d", build.GetId())]; !ok {
				continue
			}
			if (int(build.GetStatus()) & int(bbpb.Status_ENDED_MASK)) == 0 {
				newWatchSet = append(newWatchSet, fmt.Sprintf("%d", build.GetId()))
			} else {
				c.LogOut("Build %d finished with status %s", build.GetId(), build.GetStatus())
				if build.GetStatus() != bbpb.Status_SUCCESS {
					if state.canRetry(build) {
						if c.dryrun {
							c.LogOut("(Dryrun) Would have retried %d", build.GetId())
						} else {
							if newBBID, err := c.retryBuild(build); err != nil {
								errs = append(errs, err)
								c.LogErr("Failed to retry %d: %v", build.GetId(), err)
								c.LogErr("Continuing with best effort collection")
							} else {
								newWatchSet = append(newWatchSet, newBBID)
								state.recordRetry(build)
							}
						}
					}
				}
			}
		}
		watchSet = newWatchSet
	}
	if len(errs) > 0 {
		return errors.NewMultiError(errs...)
	}
	return nil
}
