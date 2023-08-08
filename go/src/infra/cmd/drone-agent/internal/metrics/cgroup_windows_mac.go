// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// The module 'github.com/containerd/cgroups' can only be built on either
// Windows or Mac.

// Regarding the file name which ends with '_windows_mac', we CANNOT change it
// to either '_mac_windows' or '_darwin_windows' or '_windows_darwin' because go
// build constraints has a rule that uses (and only use the last part
// of) the suffix as a implicit build constraint. So the above three options
// will result in that the file can only be built on one platform, instead of
// both Windows and Mac.

//go:build !linux
// +build !linux

package metrics

import (
	stats "github.com/containerd/cgroups/stats/v1"
)

func cgroupStats() (*stats.Metrics, error) {
	panic("Mac and Windows are not supported")
}
