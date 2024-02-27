// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/common_commands"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/cros_test_runner/data"

	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// GenericServiceCmd represents gcloud auth cmd.
type GenericServiceCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	GenericRequest *api.GenericTask
	Identifier     string

	// Updates
	StartResp *testapi.GenericStartResponse
	RunResp   *testapi.GenericRunResponse
	StopResp  *testapi.GenericStopResponse
}

// Instantiate extracts initial state info from the state keeper.
func (cmd *GenericServiceCmd) Instantiate(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {
	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.instantiateWithHwTestStateKeeper(ctx, sk)
	case *data.LocalTestStateKeeper:
		err = cmd.instantiateWithHwTestStateKeeper(ctx, &sk.HwTestStateKeeper)
	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error while instantiating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

func (cmd *GenericServiceCmd) instantiateWithHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) (err error) {

	if err := common_commands.Instantiate_PopFromQueue(sk.GenericQueue, func(element any) {
		cmd.GenericRequest = element.(*api.GenericTask)
	}); err != nil {
		return fmt.Errorf("cmd %s missing dependency: GenericRequest, %s", cmd.GetCommandType(), err)
	}

	return nil
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *GenericServiceCmd) ExtractDependencies(ctx context.Context,
	ski interfaces.StateKeeperInterface) error {
	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.extractDepsFromHwTestStateKeeper(ctx, sk)
	case *data.LocalTestStateKeeper:
		err = cmd.extractDepsFromHwTestStateKeeper(ctx, &sk.HwTestStateKeeper)
	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during extracting dependencies for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *GenericServiceCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.updateHwTestStateKeeper(ctx, sk)
	case *data.LocalTestStateKeeper:
		err = cmd.updateHwTestStateKeeper(ctx, &sk.HwTestStateKeeper)
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

func (cmd *GenericServiceCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.GenericRequest == nil {
		return fmt.Errorf("cmd %q missing dependency: GenericRequest", cmd.GetCommandType())
	}

	if err := common.InjectDependencies(cmd.GenericRequest, sk.Injectables, cmd.GenericRequest.DynamicDeps); err != nil {
		return fmt.Errorf("cmd %q failed injecting dependencies, %s", cmd.GetCommandType(), err)
	}

	for _, dep := range cmd.GenericRequest.DynamicDeps {
		if dep.Key == "serviceAddress" {
			cmd.Identifier = dep.GetValue()
		}
	}

	return nil
}

func (cmd *GenericServiceCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.StartResp != nil {
		if err := sk.Injectables.Set(cmd.Identifier+"_start", cmd.StartResp); err != nil {
			logging.Infof(ctx, "Warning: cmd %s failed to set %s in the Injectables Storage, %s", string(cmd.GetCommandType()), cmd.Identifier+"_start")
		}
	}
	if cmd.RunResp != nil {
		if err := sk.Injectables.Set(cmd.Identifier+"_run", cmd.RunResp); err != nil {
			logging.Infof(ctx, "Warning: cmd %s failed to set %s in the Injectables Storage, %s", string(cmd.GetCommandType()), cmd.Identifier+"_run")
		}
	}
	if cmd.StopResp != nil {
		if err := sk.Injectables.Set(cmd.Identifier+"_stop", cmd.StopResp); err != nil {
			logging.Infof(ctx, "Warning: cmd %s failed to set %s in the Injectables Storage, %s", string(cmd.GetCommandType()), cmd.Identifier+"_stop")
		}
	}

	return nil
}

func NewGenericServiceCmd(executor interfaces.ExecutorInterface) *GenericServiceCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(GenericServiceCmdType, executor)
	cmd := &GenericServiceCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
