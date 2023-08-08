// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build windows
// +build windows

package osutil

import (
	"os/exec"
)

func getCmdExitStatus(err error) int {
	panic("not implemented on windows")
}

func terminate(cmd *exec.Cmd, exited <-chan struct{}) {
	panic("not implemented on windows")
}
