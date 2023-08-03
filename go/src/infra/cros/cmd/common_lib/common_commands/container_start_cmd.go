// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/cros_test_runner/data"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/common/errors"
)

// ContainerStartCmd represents gcloud auth cmd.
type ContainerStartCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	ContainerRequest *skylab_test_runner.ContainerRequest
	ContainerImage   string

	// Updates
	Endpoint          *labapi.IpEndpoint
	ContainerInstance interfaces.ContainerInterface
}

// Instantiate extracts initial state info from the state keeper.
func (cmd *ContainerStartCmd) Instantiate(
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

func (cmd *ContainerStartCmd) instantiateWithHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) (err error) {
	// Catch panics from bad cast.
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	if sk.ContainerQueue.Len() < 1 {
		return fmt.Errorf("cmd %q missing dependency: ContainerRequest", cmd.GetCommandType())
	}
	cmd.ContainerRequest = sk.ContainerQueue.Remove(sk.ContainerQueue.Front()).(*skylab_test_runner.ContainerRequest)

	return nil
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *ContainerStartCmd) ExtractDependencies(ctx context.Context,
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
func (cmd *ContainerStartCmd) UpdateStateKeeper(
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

func (cmd *ContainerStartCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.ContainerRequest == nil {
		return fmt.Errorf("cmd %q missing dependency: ContainerRequest", cmd.GetCommandType())
	}

	if err := common.InjectDependencies(cmd.ContainerRequest.Container, sk.Injectables, cmd.ContainerRequest.DynamicDeps); err != nil {
		return fmt.Errorf("cmd %q failed injecting dependencies, %s", cmd.GetCommandType(), err)
	}

	containerImage, err := common.GetContainerImageFromMap(cmd.ContainerRequest.ContainerImageKey, sk.ContainerImages)
	if err != nil {
		return fmt.Errorf("cmd %q missing dependency: ContainerImage", cmd.GetCommandType())
	}
	cmd.ContainerImage = containerImage

	return nil
}

func (cmd *ContainerStartCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.Endpoint != nil && cmd.ContainerRequest.DynamicIdentifier != "" {
		sk.Injectables.Set(cmd.ContainerRequest.DynamicIdentifier, cmd.Endpoint)
	}

	if cmd.ContainerInstance != nil && cmd.ContainerRequest.DynamicIdentifier != "" {
		sk.ContainerInstances[cmd.ContainerRequest.ContainerImageKey] = cmd.ContainerInstance
	}

	return nil
}

func NewContainerStartCmd(executor interfaces.ExecutorInterface) *ContainerStartCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(ContainerStartCmdType, executor)
	cmd := &ContainerStartCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
