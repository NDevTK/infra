// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

package main

import (
	"context"
	"os"
	"syscall"
	"time"

	"go.chromium.org/luci/common/logging"
)

var sigTerm = syscall.SIGTERM

func interrupts() []os.Signal {
	return []os.Signal{
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGUSR1,
	}
}

func isTermSignal(s os.Signal) bool {
	return s == os.Interrupt || s == syscall.SIGTERM
}

func isUserSignal(s os.Signal) bool {
	return s == syscall.SIGUSR1
}

func collectZombies(ctx context.Context) {
	for ctx.Err() == nil {
		// It appears wait4 doesn't really hang (even when not passing NOHANG).
		// Instead of trying to figure out why, just collect all pending zombies
		// and then sleep for a second. It is not a big deal to be delayed a bit.
		for {
			var status syscall.WaitStatus
			pid, err := syscall.Wait4(-1, &status, syscall.WNOHANG, nil)
			if err != nil {
				if err != syscall.ECHILD {
					logging.Errorf(ctx, "wait4: %s", err)
				}
				break
			}
			if pid == 0 {
				break
			}
			logging.Infof(ctx, "wait4: collected PID %d", pid)
		}
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
		}
	}
}
