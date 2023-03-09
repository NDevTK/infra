// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"

	vmlabapi "infra/libs/vmlab/api"

	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// DutVmReleaseCmd defines the step I/O of releasing a DUT VM on GCE.
type DutVmReleaseCmd struct {
	*interfaces.SingleCmdByExecutor
	// Deps
	DutVm *vmlabapi.VmInstance

	// Updates
	// Only reset DutTopology to nil
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *DutVmReleaseCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		if sk.DutVm == nil {
			return fmt.Errorf("cmd %q missing dependency: DutVm", cmd.GetCommandType())
		}
		if sk.DutVm.GetName() == "" {
			return fmt.Errorf("cmd %q missing dependency: DutVm.Name", cmd.GetCommandType())
		}
		if sk.DutVm.GetConfig().GetGcloudBackend() == nil {
			return fmt.Errorf("cmd %q missing dependency: DutVm.Config", cmd.GetCommandType())
		}
		if sk.DutVm.GetConfig().GetGcloudBackend().GetProject() == "" {
			return fmt.Errorf("cmd %q missing dependency: DutVm.Config.Project", cmd.GetCommandType())
		}
		if sk.DutVm.GetConfig().GetGcloudBackend().GetZone() == "" {
			return fmt.Errorf("cmd %q missing dependency: DutVm.Config.Zone", cmd.GetCommandType())
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
