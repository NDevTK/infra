// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// The module of 'github.com/containerd/cgroups' doesn't work on both Mac and
// Windows

//go:build linux
// +build linux

package bot

import (
	"fmt"

	"github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Cgroup is an alias of cgroups.Cgroup, which cannot be built on neither
// Windows or Mac.
type Cgroup cgroups.Cgroup

// addToCgroup adds the bot process to its dedicated cgroup.
func addToCgroup(botID string, pid uint64, resources *specs.LinuxResources) (Cgroup, error) {
	control, err := cgroups.New(cgroups.V1, cgroups.NestedPath(botID), resources)
	if err != nil {
		return nil, fmt.Errorf("add bot to cgroup: %s", err)
	}

	if err := control.AddProc(pid); err != nil {
		control.Delete()
		return nil, fmt.Errorf("add bot to cgroup: %s", err)
	}
	return control, nil
}
