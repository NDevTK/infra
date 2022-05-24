// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"path/filepath"
	"runtime"
)

func Python3(path string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(path, "bin", "python3.exe")
	}
	return filepath.Join(path, "bin", "python3")
}

func Python3VENV(path string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(path, "Scripts", "python3.exe")
	}
	return filepath.Join(path, "bin", "python3")
}
