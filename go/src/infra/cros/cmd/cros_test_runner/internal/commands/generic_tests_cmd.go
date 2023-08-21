// Copyright 2023 The Chromium Authors
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

	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// GenericTestsCmd represents gcloud auth cmd.
type GenericTestsCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	TestRequest *skylab_test_runner.TestRequest
	Identifier  string

	// Updates
	TestResponses      *testapi.CrosTestResponse
	CpconPublishSrcDir string
}

// Instantiate extracts initial state info from the state keeper.
func (cmd *GenericTestsCmd) Instantiate(
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

func (cmd *GenericTestsCmd) instantiateWithHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) (err error) {

	if err := common_commands.Instantiate_PopFromQueue(sk.TestQueue, func(element any) {
		cmd.TestRequest = element.(*skylab_test_runner.TestRequest)
	}); err != nil {
		return fmt.Errorf("cmd %s missing dependency: TestRequest, %s", cmd.GetCommandType(), err)
	}

	return nil
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *GenericTestsCmd) ExtractDependencies(ctx context.Context,
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
func (cmd *GenericTestsCmd) UpdateStateKeeper(
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

func (cmd *GenericTestsCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.TestRequest == nil {
		return fmt.Errorf("cmd %q missing dependency: TestRequest", cmd.GetCommandType())
	}

	if err := common.InjectDependencies(cmd.TestRequest, sk.Injectables, cmd.TestRequest.DynamicDeps); err != nil {
		return fmt.Errorf("cmd %q failed injecting dependencies, %s", cmd.GetCommandType(), err)
	}

	for _, dep := range cmd.TestRequest.DynamicDeps {
		if dep.Key == "serviceAddress" {
			cmd.Identifier = dep.GetValue()
		}
	}

	return nil
}

func (cmd *GenericTestsCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.TestResponses != nil {
		sk.TestResponses = cmd.TestResponses
		if err := sk.Injectables.Set("test-response", sk.TestResponses); err != nil {
			logging.Infof(ctx, "Warning: failed to set 'test-response' into the InjectableStorage, %s", err)
		}
		rdbTestResult, err := constructTestResultFromStateKeeper(ctx, sk)
		if err != nil {
			return errors.Annotate(err, "Cmd %q failed to construct update: TestResultForRdb", cmd.GetCommandType()).Err()
		}
		sk.TestResultForRdb = rdbTestResult
		if err := sk.Injectables.Set("rdb-test-result", sk.TestResultForRdb); err != nil {
			logging.Infof(ctx, "Warning: failed to set 'rdb-test-result' into the InjectableStorage, %s", err)
		}
	}
	if cmd.CpconPublishSrcDir != "" {
		sk.CpconPublishSrcDir = cmd.CpconPublishSrcDir
	}

	return nil
}

func NewGenericTestsCmd(executor interfaces.ExecutorInterface) *GenericTestsCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(GenericTestsCmdType, executor)
	cmd := &GenericTestsCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
