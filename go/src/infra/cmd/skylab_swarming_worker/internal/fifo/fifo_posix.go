// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build darwin || linux
// +build darwin linux

package fifo

import (
	"golang.org/x/sys/unix"
)

func makeFIFO(path string) error {
	return unix.Mkfifo(path, 0666)
}
