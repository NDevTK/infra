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

	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// GenericProvisionCmd represents gcloud auth cmd.
type GenericProvisionCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	ProvisionRequest *api.ProvisionTask
	Identifier       string
	TargetDevice     string

	// Updates
	InstallResp *testapi.InstallResponse
	StartUpResp *testapi.ProvisionStartupResponse
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
		cmd.ProvisionRequest = element.(*api.ProvisionTask)
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
		logging.Infof(ctx, "Warning: cmd %q failed to inject some dependencies, %s", cmd.GetCommandType(), err)
	}

	cmd.Identifier = cmd.ProvisionRequest.GetDynamicIdentifier()
	if cmd.Identifier == "" {
		logging.Infof(ctx, "Warning: cmd %q missing preferred dependency: DynamicIdentifier (required for dynamic referencing)", cmd.GetCommandType())
	}

	cmd.TargetDevice = cmd.ProvisionRequest.GetTarget()
	if cmd.TargetDevice == "" {
		logging.Infof(ctx, "Warning: cmd %q missing preferred dependency: TargetDevice", cmd.GetCommandType())
	}

	return nil
}

func (cmd *GenericProvisionCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	taskIdentifier := common.NewTaskIdentifier(cmd.ProvisionRequest.DynamicIdentifier)
	if cmd.InstallResp != nil {
		responses := sk.ProvisionResponses[cmd.TargetDevice]
		if responses == nil {
			responses = []*testapi.InstallResponse{}
		}
		responses = append(responses, cmd.InstallResp)
		sk.ProvisionResponses[cmd.TargetDevice] = responses
		if err := sk.Injectables.Set(taskIdentifier.GetRpcResponse("install"), cmd.InstallResp); err != nil {
			logging.Infof(ctx, "Warning: cmd %s failed to set %s in the Injectables Storage, %s", string(cmd.GetCommandType()), taskIdentifier.GetRpcResponse("install"))
		}
	}

	if cmd.StartUpResp != nil {
		if err := sk.Injectables.Set(taskIdentifier.GetRpcResponse("startup"), cmd.StartUpResp); err != nil {
			logging.Infof(ctx, "Warning: cmd %s failed to set %s in the Injectables Storage, %s", string(cmd.GetCommandType()), taskIdentifier.GetRpcResponse("startup"))
		}
	}

	if cmd.ProvisionRequest != nil && cmd.ProvisionRequest.GetInstallRequest() != nil {
		key := common.DeviceIdentifierFromString(cmd.TargetDevice).GetDeviceMetadata()
		deviceMetadata := &skylab_test_runner.CFTTestRequest_Device{}
		if err := common.Inject(deviceMetadata, "", sk.Injectables, key); err != nil {
			logging.Infof(ctx, "Warning: could not retrieve '%s' from InjectableStorage, %s", key, err)
		} else if cmd.ProvisionRequest.GetInstallRequest().GetImagePath().GetPath() != "" {
			deviceMetadata.ProvisionState = &testapi.ProvisionState{
				SystemImage: &testapi.ProvisionState_SystemImage{
					SystemImagePath: cmd.ProvisionRequest.GetInstallRequest().GetImagePath(),
				},
			}
			if err := sk.Injectables.Set(key, deviceMetadata); err != nil {
				logging.Infof(ctx, "Warning: failed to set '%s' into the InjectableStorage, %s", key, err)
			}
		}
	}

	// Upload request objects to storage
	if cmd.ProvisionRequest.StartupRequest != nil {
		if err := sk.Injectables.Set(taskIdentifier.GetRpcRequest("startup"), cmd.ProvisionRequest.StartupRequest); err != nil {
			logging.Infof(ctx, "Warning: cmd %s failed to set %s in the Injectables Storage, %s", string(cmd.GetCommandType()), taskIdentifier.GetRpcRequest("startup"))
		}
	}
	if cmd.ProvisionRequest.InstallRequest != nil {
		if err := sk.Injectables.Set(taskIdentifier.GetRpcRequest("install"), cmd.ProvisionRequest.InstallRequest); err != nil {
			logging.Infof(ctx, "Warning: cmd %s failed to set %s in the Injectables Storage, %s", string(cmd.GetCommandType()), taskIdentifier.GetRpcRequest("install"))
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
