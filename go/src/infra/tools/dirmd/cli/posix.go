// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

package cli

import "path/filepath"

func canonicalFSPath(path string) (string, error) {
	return filepath.Abs(path)
}
