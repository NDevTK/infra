// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"os/exec"
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

func CIPDCommand(arg ...string) *exec.Cmd {
	cipd, err := exec.LookPath("cipd")
	if err != nil {
		cipd = "cipd"
	}

	// Use cmd to execute batch file on windows.
	if filepath.Ext(cipd) == ".bat" {
		return exec.Command("cmd.exe", append([]string{"/C", cipd}, arg...)...)
	}

	return exec.Command(cipd, arg...)
}
