// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows

package main

import (
	"context"
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

// notifySIGTERM returns a context which is canceled when SIGTERM is
// received.
func notifySIGTERM(ctx context.Context) (_ context.Context, stop context.CancelFunc) {
	return signal.NotifyContext(ctx, os.Interrupt, unix.SIGTERM)
}
