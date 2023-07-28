// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

package internal

import (
	"os"
)

func OpenForSharedRead(path string) (*os.File, error) {
	// On every OS other than windows, nothing special do to.
	return os.Open(path)
}
