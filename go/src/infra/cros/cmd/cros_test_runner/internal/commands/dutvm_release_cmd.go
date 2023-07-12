// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"

	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/cros_test_runner/internal/data"
	vmlabapi "infra/libs/vmlab/api"
)

// DutVmReleaseCmd defines the step I/O of releasing a DUT VM on GCE.
type DutVmReleaseCmd struct {
	*interfaces.SingleCmdByExecutor
	// Deps
	DutVm      *vmlabapi.VmInstance
	BuildState *build.State

	// Updates
	// Only reset DutTopology to nil
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *DutVmReleaseCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		// BuildState is used for experiment flags. If missing, will use defaults.
		cmd.BuildState = sk.BuildState
		if sk.DutVm == nil {
			return fmt.Errorf("cmd %q missing dependency: DutVm", cmd.GetCommandType())
		}
		if sk.DutVm.GetName() == "" {
			return fmt.Errorf("cmd %q missing dependency: DutVm.Name", cmd.GetCommandType())
		}
		if sk.DutVm.GetConfig() == nil {
			return fmt.Errorf("cmd %q missing dependency: DutVm.Config", cmd.GetCommandType())
		}

		switch sk.DutVm.GetConfig().GetBackend().(type) {
		case *vmlabapi.Config_GcloudBackend:
			if sk.DutVm.GetConfig().GetGcloudBackend().GetProject() == "" {
				return fmt.Errorf("cmd %q missing dependency: DutVm.Config.GcloudBackend.Project", cmd.GetCommandType())
			}
			if sk.DutVm.GetConfig().GetGcloudBackend().GetZone() == "" {
				return fmt.Errorf("cmd %q missing dependency: DutVm.Config.GcloudBackend.Zone", cmd.GetCommandType())
			}
		case *vmlabapi.Config_VmLeaserBackend_:
			if sk.DutVm.GetConfig().GetVmLeaserBackend().GetVmRequirements() == nil {
				return fmt.Errorf("cmd %q missing dependency: DutVm.Config.VmLeaserBackend.VmRequirements", cmd.GetCommandType())
			}
			if sk.DutVm.GetConfig().GetVmLeaserBackend().GetVmRequirements().GetGceProject() == "" {
				return fmt.Errorf("cmd %q missing dependency: DutVm.Config.VmLeaserBackend.GceProject", cmd.GetCommandType())
			}
			if sk.DutVm.GetGceRegion() == "" {
				return fmt.Errorf("cmd %q missing dependency: DutVm.GceRegion", cmd.GetCommandType())
			}
		default:
			return fmt.Errorf("DutVm config backend type %q is not supported", sk.DutVm.GetConfig().GetBackend())
		}
		cmd.DutVm = sk.DutVm
	default:
		return fmt.Errorf("stateKeeper '%T' is not supported by cmd type %s", sk, cmd.GetCommandType())
	}
	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *DutVmReleaseCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		sk.DutTopology = nil
	default:
		return fmt.Errorf("stateKeeper '%T' is not supported by cmd type %s", sk, cmd.GetCommandType())
	}
	return nil
}

func NewDutVmReleaseCmd(executor interfaces.ExecutorInterface) *DutVmReleaseCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(DutVmReleaseCmdType, executor)
	cmd := &DutVmReleaseCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
