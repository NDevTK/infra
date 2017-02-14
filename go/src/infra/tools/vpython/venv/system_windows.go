// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package venv

import (
	"path/filepath"
)

// venvBinPath resolves the path to a VirtualEnv binary.
func venvBinPath(root, name string) string {
	return filepath.Join(root, "Scripts", name)
}
