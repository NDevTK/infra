// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cron

import (
	"context"
	"time"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/runtime/paniccatcher"
	"go.chromium.org/luci/server"
)

// RegisterCronServer initializes the VM Leaser cron server.
func RegisterCronServer(srv *server.Server) {
	srv.RunInBackground("vm_leaser.cron", func(ctx context.Context) {
		// releaseExpiredVMs every minute
		Run(ctx, 1*time.Minute, releaseExpiredVMs)
	})
}

// Run runs f repeatedly, until the context is cancelled.
//
// This method runs f based on minInterval time interval.
func Run(ctx context.Context, minInterval time.Duration, f func(context.Context) error) {
	defer logging.Warningf(ctx, "Exiting cron")

	// call calls the provided cron method f
	//
	// If call catches a panic, the cron run will stop once the whole context is
	// cancelled.
	call := func(ctx context.Context) error {
		defer paniccatcher.Catch(func(p *paniccatcher.Panic) {
			logging.Errorf(ctx, "Caught panic: %s\n%s", p.Reason, p.Stack)
		})
		return f(ctx)
	}

	for {
		start := clock.Now(ctx)
		if err := call(ctx); err != nil {
			logging.Errorf(ctx, "Iteration failed: %s", err)
		}

		// Ensure minInterval between iterations.
		if sleep := minInterval - clock.Since(ctx, start); sleep > 0 {
			select {
			case <-time.After(sleep):
			case <-ctx.Done():
				return
			}
		}
	}
}

// releaseExpiredVMs releases VMs based on their expiration times.
func releaseExpiredVMs(ctx context.Context) error {
	logging.Debugf(ctx, "Releasing expired VMs")
	return nil
}
