// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//go:build !windows

package builtins

import (
	"os"
)

func symlink(src, dst string) error {
	return os.Symlink(src, dst)
}
