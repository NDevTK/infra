// Copyright (c) 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package system

func shouldIgnoreFstype(fstype string) bool {
	return false
}

func shouldIgnoreMountpoint(mountpoint string) bool {
	return false
}

func shouldIgnoreDevice(device string) bool {
	return false
}

func removeDiskDevices(names []string) []string {
	return names
}
