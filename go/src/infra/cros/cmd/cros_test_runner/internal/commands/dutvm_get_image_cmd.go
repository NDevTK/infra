// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	vmlabapi "infra/libs/vmlab/api"
)

// DutVmGetImageCmd defines the step I/O of get the GCE image of Dut VM.
type DutVmGetImageCmd struct {
	*interfaces.SingleCmdByExecutor
	// Deps
	CftTestRequest *skylab_test_runner.CFTTestRequest

	// Updates
	DutVmGceImage *vmlabapi.GceImage
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *DutVmGetImageCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		if sk.CftTestRequest == nil {
			return fmt.Errorf("cmd %q missing dependency: CftTestRequest", cmd.GetCommandType())
		}
		buildName := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, "build")
		if buildName == "" {
			return fmt.Errorf("cmd %q missing dependency: CftTestRequest.AutotestKeyvals['build']", cmd.GetCommandType())
		}
		cmd.CftTestRequest = sk.CftTestRequest
	default:
		return fmt.Errorf("stateKeeper '%T' is not supported by cmd type %s", sk, cmd.GetCommandType())
	}
	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *DutVmGetImageCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		sk.DutVmGceImage = cmd.DutVmGceImage
	default:
		return fmt.Errorf("stateKeeper '%T' is not supported by cmd type %s", sk, cmd.GetCommandType())
	}
	return nil
}

func NewDutVmGetImageCmd(executor interfaces.ExecutorInterface) *DutVmGetImageCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(DutVmGetImageCmdType, executor)
	cmd := &DutVmGetImageCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
