// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"path/filepath"
	"runtime"
)

func Python(path, py string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(path, "bin", py+".exe")
	}
	return filepath.Join(path, "bin", py)
}

func PythonVENV(path, py string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(path, "Scripts", py+".exe")
	}
	return filepath.Join(path, "bin", py)
}

// On Darwin, because vpython is an app bundle, cpython is located at
// vpython.app/Contents/Resources. On other platforms cpython is next to the
// vpython binary.
func DefaultBundleDir(version string) string {
	if runtime.GOOS == "darwin" {
		return filepath.Join("..", "Resources", version)
	}
	return version
}
