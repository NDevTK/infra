// Copyright 2022 The Chromium Authors. All rights reserved.
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
