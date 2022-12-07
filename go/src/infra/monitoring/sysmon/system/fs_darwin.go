// Copyright (c) 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package system

import (
	"os"
	"path"
)

func shouldIgnoreFstype(fstype string) bool {
	return false
}

func shouldIgnoreMountpoint(mountpoint string) bool {
	return false
}

func shouldIgnoreDevice(device string) bool {
	// gopsutil may return an invalid partition due to
	// https://github.com/shirou/gopsutil/issues/1390 .

	// This restores the logic from before
	// https://github.com/shirou/gopsutil/commit/fb1c75054a14368e1adb1cf447015244cb14ea7c
	// which had the effect of filtering these out.
	// Remove this once we have moved to a new gopsutil release with a
	// fix for the above issue.
	if !path.IsAbs(device) {
		return true
	}
	if _, err := os.Stat(device); err != nil {
		return true
	}

	return false
}

func removeDiskDevices(names []string) []string {
	return names
}
