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
	TargetDevice     string

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

	cmd.TargetDevice = cmd.ProvisionRequest.GetTarget()
	if cmd.TargetDevice == "" {
		logging.Infof(ctx, "Warning: cmd %q missing preferred dependency: TargetDevice", cmd.GetCommandType())
	}

	return nil
}

func (cmd *GenericProvisionCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.ProvisionResp != nil {
		responses := sk.ProvisionResponses[cmd.TargetDevice]
		if responses == nil {
			responses = []*testapi.InstallResponse{}
		}
		responses = append(responses, cmd.ProvisionResp)
		sk.ProvisionResponses[cmd.TargetDevice] = responses
		if err := sk.Injectables.Set(cmd.TargetDevice+"ProvisionResponses", responses); err != nil {
			logging.Infof(ctx, "Warning: cmd %s failed to set %s in the Injectables Storage, %s", string(cmd.GetCommandType()), cmd.TargetDevice+"ProvisionResponses", err)
		}
	}

	if cmd.ProvisionRequest != nil && cmd.ProvisionRequest.GetInstallRequest() != nil {
		key := cmd.TargetDevice + "Metadata"
		deviceMetadata := &skylab_test_runner.CFTTestRequest_Device{}
		if err := common.Inject(deviceMetadata, "", sk.Injectables, key); err != nil {
			logging.Infof(ctx, "Warning: could not retrieve '%s' from InjectableStorage, %s", key, err)
		} else {
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

	return nil
}

func NewGenericProvisionCmd(executor interfaces.ExecutorInterface) *GenericProvisionCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(GenericProvisionCmdType, executor)
	cmd := &GenericProvisionCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
