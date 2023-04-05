// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/common/errors"
	"infra/cros/cmd/cros_test_runner/common"
	vmlabapi "infra/libs/vmlab/api"

	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// DutVmLeaseCmd defines the step I/O of leasing a DUT VM on GCE.
type DutVmLeaseCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	DutVmGceImage  *vmlabapi.GceImage
	CftTestRequest *skylab_test_runner.CFTTestRequest

	// Updates
	DutVm *vmlabapi.VmInstance
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *DutVmLeaseCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		if sk.DutVmGceImage == nil {
			return fmt.Errorf("cmd %q missing dependency: DutVmGceImage", cmd.GetCommandType())
		}
		if sk.DutVmGceImage.GetName() == "" {
			return fmt.Errorf("cmd %q missing dependency: DutVmGceImage.Name", cmd.GetCommandType())
		}
		if sk.DutVmGceImage.GetProject() == "" {
			return fmt.Errorf("cmd %q missing dependency: DutVmGceImage.Project", cmd.GetCommandType())
		}
		cmd.DutVmGceImage = sk.DutVmGceImage
		if sk.CftTestRequest == nil {
			return fmt.Errorf("cmd %q missing dependency: CftTestRequest", cmd.GetCommandType())
		}
		if sk.CftTestRequest.GetPrimaryDut() == nil {
			return fmt.Errorf("cmd %q missing dependency: CftTestRequest.PrimaryDut", cmd.GetCommandType())
		}
		if sk.CftTestRequest.GetPrimaryDut().GetDutModel() == nil {
			return fmt.Errorf("cmd %q missing dependency: CftTestRequest.PrimaryDut.DutModel", cmd.GetCommandType())
		}
		cmd.CftTestRequest = sk.CftTestRequest
	default:
		return fmt.Errorf("stateKeeper '%T' is not supported by cmd type %s", sk, cmd.GetCommandType())
	}
	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *DutVmLeaseCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.updateVmTestStateKeeper(ctx, sk)
	default:
		return fmt.Errorf("stateKeeper '%T' is not supported by cmd type %s", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// updateVmTestStateKeeper updates
// - DutVm in the state to allow release in a later step.
// - DutTopology in the state to fully mimics the state of hardware tests.
func (cmd *DutVmLeaseCmd) updateVmTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.DutVm == nil {
		return nil
	}

	sk.DutVm = cmd.DutVm

	if cmd.DutVm.GetSsh() != nil {
		duts := []*labapi.Dut{{
			Id: &labapi.Dut_Id{Value: common.VmLabDutHostName},
			DutType: &labapi.Dut_Chromeos{
				Chromeos: &labapi.Dut_ChromeOS{
					Ssh: &labapi.IpEndpoint{
						Address: cmd.DutVm.GetSsh().GetAddress(),
						Port:    cmd.DutVm.GetSsh().GetPort(),
					},
					DutModel: cmd.CftTestRequest.GetPrimaryDut().GetDutModel(),
				},
			}}}
		sk.DutTopology = &labapi.DutTopology{
			Duts: duts,
		}
	}

	return nil
}

func NewDutVmLeaseCmd(executor interfaces.ExecutorInterface) *DutVmLeaseCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(DutVmLeaseCmdType, executor)
	cmd := &DutVmLeaseCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
