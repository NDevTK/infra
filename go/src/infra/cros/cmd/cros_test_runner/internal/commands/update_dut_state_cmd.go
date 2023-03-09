// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	"infra/cros/cmd/cros_test_runner/internal/tools/ufs"
	"infra/cros/dutstate"

	"infra/cros/cmd/cros_test_runner/common"

	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
)

// UpdateDutStateCmd represents update dut state command.
type UpdateDutStateCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Deps
	HostName      string
	TestResponses *testapi.CrosTestResponse // optional
	ProvisionResp *testapi.InstallResponse  // optional

	// Updates
	CurrentDutState dutstate.State
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *UpdateDutStateCmd) ExtractDependencies(ctx context.Context, ski interfaces.StateKeeperInterface) error {
	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.extractDepsFromHwTestStateKeeper(ctx, sk)

	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during extracting dependencies for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *UpdateDutStateCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.updateHwTestStateKeeper(ctx, sk)
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// Execute executes the command.
func (cmd *UpdateDutStateCmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Update dut state if required")
	defer func() { step.End(err) }()

	triedToUpdateState := false
	cmd.CurrentDutState, err = ufs.GetDutStateFromUFS(ctx, cmd.HostName)
	if err != nil {
		logging.Infof(ctx, "error while getting current dut state: %s", err.Error())
	}
	logging.Infof(ctx, "Dut state before any kind of update: %s", cmd.CurrentDutState)

	// Update on provision/test failures
	if cmd.ProvisionResp != nil && cmd.ProvisionResp.GetStatus() != api.InstallResponse_STATUS_SUCCESS {
		triedToUpdateState = updateDutState(ctx, cmd.HostName, dutstate.NeedsRepair, "provision")
	} else if cmd.TestResponses != nil && len(cmd.TestResponses.GetTestCaseResults()) > 0 && common.IsAnyTestFailure(cmd.TestResponses.GetTestCaseResults()) {
		triedToUpdateState = updateDutState(ctx, cmd.HostName, dutstate.NeedsRepair, "test(s)")
	}

	if triedToUpdateState {
		cmd.CurrentDutState, err = ufs.GetDutStateFromUFS(ctx, cmd.HostName)
		if err != nil {
			logging.Infof(ctx, "error while getting current dut state: %s", err.Error())
		}
		logging.Infof(ctx, "Dut state after update: %s", cmd.CurrentDutState)
	}

	step.SetSummaryMarkdown(fmt.Sprintf("dut state: %s", cmd.CurrentDutState.String()))
	step.AddTagValue("dut_state", cmd.CurrentDutState.String())
	return nil
}

func (cmd *UpdateDutStateCmd) extractDepsFromHwTestStateKeeper(ctx context.Context, sk *data.HwTestStateKeeper) error {
	if sk.ProvisionResp == nil {
		logging.Infof(ctx, "Warning: cmd %q missing non-critical dependency: ProvisionResp", cmd.GetCommandType())
	}
	if sk.TestResponses == nil {
		logging.Infof(ctx, "Warning: cmd %q missing non-critical dependency: TestResponses", cmd.GetCommandType())
	}
	if sk.HostName == "" {
		return fmt.Errorf("Cmd %q missing dependency: HostName", cmd.GetCommandType())
	}
	cmd.ProvisionResp = sk.ProvisionResp
	cmd.TestResponses = sk.TestResponses
	cmd.HostName = sk.HostName

	return nil
}

func (cmd *UpdateDutStateCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.CurrentDutState != "" {
		sk.CurrentDutState = cmd.CurrentDutState
	}

	return nil
}

// updateDutState tries to update dut state
func updateDutState(ctx context.Context, hostName string, dutState dutstate.State, failureType string) bool {
	logging.Infof(ctx, "Trying to update dut state to %s due to %s failure.", dutstate.NeedsRepair)
	err := ufs.SafeUpdateUFSDUTState(ctx, hostName, dutState)
	if err != nil {
		logging.Infof(ctx, "Error while updating dut state: %s", err)
		return false
	}
	return true
}

func NewUpdateDutStateCmd() *UpdateDutStateCmd {
	abstractCmd := interfaces.NewAbstractCmd(UpdateDutStateCmdType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &UpdateDutStateCmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}
