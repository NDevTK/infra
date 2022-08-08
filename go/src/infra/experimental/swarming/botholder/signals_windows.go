// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build windows
// +build windows

package main

import (
	"os"
)

// This file exists exclusively to avoid setting up conditional compilation
// for `botholder`. It will never be running on Windows.

func interrupts() []os.Signal {
	return []os.Signal{
		os.Interrupt,
	}
}

func isTermSignal(s os.Signal) bool {
	return s == os.Interrupt
}

func isUserSignal(s os.Signal) bool {
	return false
}
