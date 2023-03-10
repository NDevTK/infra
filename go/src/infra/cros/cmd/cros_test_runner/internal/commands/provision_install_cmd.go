package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"

	_go "go.chromium.org/chromiumos/config/go"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/types/known/anypb"
)

// ProvisionInstallCmd represents provision install cmd.
type ProvisionInstallCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	OsImagePath     *_go.StoragePath
	PreventReboot   bool
	InstallMetadata *anypb.Any

	// Updates
	ProvisionResp *testapi.InstallResponse
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *ProvisionInstallCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.extractDepsFromHwTestStateKeeper(ctx, sk)
	case *data.LocalTestStateKeeper:
		err = cmd.extractDepsFromLocalTestStateKeeper(ctx, sk)

	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during extracting dependencies for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *ProvisionInstallCmd) UpdateStateKeeper(
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

func (cmd *ProvisionInstallCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	var err error
	if sk.CftTestRequest == nil || sk.CftTestRequest.GetPrimaryDut().GetProvisionState().GetSystemImage().GetSystemImagePath() == nil {
		return fmt.Errorf("Cmd %q missing dependency: OsImagePath", cmd.GetCommandType())
	}
	cmd.OsImagePath = sk.CftTestRequest.GetPrimaryDut().GetProvisionState().GetSystemImage().GetSystemImagePath()

	cmd.PreventReboot = false

	if sk.InstallMetadata == nil {
		cmd.InstallMetadata, err = anypb.New(&testapi.CrOSProvisionMetadata{})
		if err != nil {
			return errors.Annotate(err, "error during creating provision metadata: ").Err()
		}
	} else {
		cmd.InstallMetadata = sk.InstallMetadata
	}

	return nil
}

func (cmd *ProvisionInstallCmd) extractDepsFromLocalTestStateKeeper(
	ctx context.Context,
	sk *data.LocalTestStateKeeper) error {

	var err error
	if sk.ImagePath == "" {
		if sk.CftTestRequest == nil || sk.CftTestRequest.GetPrimaryDut().GetProvisionState().GetSystemImage().GetSystemImagePath() == nil {
			return fmt.Errorf("Cmd %q missing dependency: OsImagePath", cmd.GetCommandType())
		}
		cmd.OsImagePath = sk.CftTestRequest.GetPrimaryDut().GetProvisionState().GetSystemImage().GetSystemImagePath()
	} else {
		cmd.OsImagePath = &_go.StoragePath{
			HostType: _go.StoragePath_GS,
			Path:     sk.ImagePath,
		}
	}

	cmd.PreventReboot = false

	if sk.InstallMetadata == nil {
		cmd.InstallMetadata, err = anypb.New(&testapi.CrOSProvisionMetadata{})
		if err != nil {
			return errors.Annotate(err, "error during creating provision metadata: ").Err()
		}
	} else {
		cmd.InstallMetadata = sk.InstallMetadata
	}

	return nil
}

func (cmd *ProvisionInstallCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.ProvisionResp != nil {
		sk.ProvisionResp = cmd.ProvisionResp
	}

	return nil
}

func NewProvisionInstallCmd(executor interfaces.ExecutorInterface) *ProvisionInstallCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(ProvisonInstallCmdType, executor)
	cmd := &ProvisionInstallCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
