// Copyright 2023 The Chromium Authors
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

package bot

import "github.com/opencontainers/runtime-spec/specs-go"

type Cgroup interface {
	Delete() error
}

func addToCgroup(botID string, pid uint64, resources *specs.LinuxResources) (Cgroup, error) {
	panic("Mac and Windows are not supported")
}
