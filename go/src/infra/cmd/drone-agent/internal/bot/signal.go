// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

package bot

import (
	"golang.org/x/sys/unix"
)

// Terminate implements Bot.
func (b realBot) Terminate() error {
	return b.cmd.Process.Signal(unix.SIGTERM)
}
