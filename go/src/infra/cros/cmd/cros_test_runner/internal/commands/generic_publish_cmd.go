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

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/common/errors"
)

// GenericPublishCmd represents gcloud auth cmd.
type GenericPublishCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	PublishRequest *skylab_test_runner.PublishRequest
	Identifier     string

	// Updates
}

// Instantiate extracts initial state info from the state keeper.
func (cmd *GenericPublishCmd) Instantiate(
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

func (cmd *GenericPublishCmd) instantiateWithHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) (err error) {

	if err := common_commands.Instantiate_PopFromQueue(sk.PublishQueue, func(element any) {
		cmd.PublishRequest = element.(*skylab_test_runner.PublishRequest)
	}); err != nil {
		return fmt.Errorf("cmd %s missing dependency: PublishRequest, %s", cmd.GetCommandType(), err)
	}

	return nil
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *GenericPublishCmd) ExtractDependencies(ctx context.Context,
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

func (cmd *GenericPublishCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.PublishRequest == nil {
		return fmt.Errorf("cmd %q missing dependency: PublishRequest", cmd.GetCommandType())
	}

	if err := common.InjectDependencies(cmd.PublishRequest, sk.Injectables, cmd.PublishRequest.DynamicDeps); err != nil {
		return fmt.Errorf("cmd %q failed injecting dependencies, %s", cmd.GetCommandType(), err)
	}

	for _, dep := range cmd.PublishRequest.DynamicDeps {
		if dep.Key == "serviceAddress" {
			cmd.Identifier = dep.GetValue()
		}
	}

	return nil
}

func NewGenericPublishCmd(executor interfaces.ExecutorInterface) *GenericPublishCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(GenericPublishCmdType, executor)
	cmd := &GenericPublishCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
