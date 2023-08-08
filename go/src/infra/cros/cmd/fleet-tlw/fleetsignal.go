// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

package main

import (
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

func setupSignalHandler() <-chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, unix.SIGINT, unix.SIGHUP, unix.SIGTERM, unix.SIGQUIT)
	return c
}
