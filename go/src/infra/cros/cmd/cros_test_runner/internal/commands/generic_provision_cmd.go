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

// GenericProvisionCmd represents gcloud auth cmd.
type GenericProvisionCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	ProvisionRequest *skylab_test_runner.ProvisionRequest
	Identifier       string

	// Updates
	ProvisionResp *testapi.InstallResponse
}

// Instantiate extracts initial state info from the state keeper.
func (cmd *GenericProvisionCmd) Instantiate(
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

func (cmd *GenericProvisionCmd) instantiateWithHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) (err error) {

	if err := common_commands.Instantiate_PopFromQueue(sk.ProvisionQueue, func(element any) {
		cmd.ProvisionRequest = element.(*skylab_test_runner.ProvisionRequest)
	}); err != nil {
		return fmt.Errorf("cmd %s missing dependency: ProvisionRequest, %s", cmd.GetCommandType(), err)
	}

	return nil
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *GenericProvisionCmd) ExtractDependencies(ctx context.Context,
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
func (cmd *GenericProvisionCmd) UpdateStateKeeper(
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

func (cmd *GenericProvisionCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.ProvisionRequest == nil {
		return fmt.Errorf("cmd %q missing dependency: ProvisionRequest", cmd.GetCommandType())
	}

	if err := common.InjectDependencies(cmd.ProvisionRequest, sk.Injectables, cmd.ProvisionRequest.DynamicDeps); err != nil {
		return fmt.Errorf("cmd %q failed injecting dependencies, %s", cmd.GetCommandType(), err)
	}

	for _, dep := range cmd.ProvisionRequest.DynamicDeps {
		if dep.Key == "serviceAddress" {
			cmd.Identifier = dep.GetValue()
		}
	}

	return nil
}

func (cmd *GenericProvisionCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.ProvisionResp != nil {
		sk.ProvisionResp = cmd.ProvisionResp
	}

	// TODO(cdelagarza): Update to use targeted DUT instead of defaulting to primaryDeviceMetadata
	// when we start supporting Multi-DUT provisioning.
	if cmd.ProvisionRequest != nil && cmd.ProvisionRequest.GetInstallRequest() != nil {
		deviceMetadata := &skylab_test_runner.CFTTestRequest_Device{}
		if err := common.Inject(deviceMetadata, "", sk.Injectables, "primaryDeviceMetadata"); err != nil {
			logging.Infof(ctx, "Warning: could not retrieve 'primaryDeviceMetadata' from InjectableStorage, %s", err)
		} else {
			deviceMetadata.ProvisionState = &testapi.ProvisionState{
				SystemImage: &testapi.ProvisionState_SystemImage{
					SystemImagePath: cmd.ProvisionRequest.GetInstallRequest().GetImagePath(),
				},
			}
			if err := sk.Injectables.Set("primaryDeviceMetadata", deviceMetadata); err != nil {
				logging.Infof(ctx, "Warning: failed to set 'primaryDeviceMetadata' into the InjectableStorage, %s", err)
			}
		}
	}

	return nil
}

func NewGenericProvisionCmd(executor interfaces.ExecutorInterface) *GenericProvisionCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(GenericProvisionCmdType, executor)
	cmd := &GenericProvisionCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
