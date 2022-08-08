// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

package main

import (
	"os"
	"syscall"
)

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
