// Copyright 2023 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package components

import (
	"io/fs"
)

// DefaultFilePermissions is the default file permissions for log files.
// Currently, we allow everyone to read and write and nobody to execute.
const DefaultFilePermissions fs.FileMode = 0666
