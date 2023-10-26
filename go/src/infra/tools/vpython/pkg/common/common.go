// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"os/exec"
	"path/filepath"
	"runtime"
)

// Python returns the python path in python installation.
func Python(path, py string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(path, "bin", py+".exe")
	}
	return filepath.Join(path, "bin", py)
}

// PythonVENV returns the python path in venv.
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

// CIPDCommand generates a *exec.Cmd for cipd. It will lookup cipd and its
// wrappers depending on platforms.
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
