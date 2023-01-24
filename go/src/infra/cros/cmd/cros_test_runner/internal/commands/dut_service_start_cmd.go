package commands

import (
	"context"
	"fmt"

	lab_api "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// DutServiceStartCmd represents dut service start cmd.
type DutServiceStartCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	CacheServerAddress *lab_api.IpEndpoint
	DutSshAddress      *lab_api.IpEndpoint

	// Updates
	DutServerAddress *lab_api.IpEndpoint
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *DutServiceStartCmd) ExtractDependencies(ctx context.Context, ski interfaces.StateKeeperInterface) error {
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
func (cmd *DutServiceStartCmd) UpdateStateKeeper(ctx context.Context, ski interfaces.StateKeeperInterface) error {
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

func (cmd *DutServiceStartCmd) extractDepsFromHwTestStateKeeper(ctx context.Context, sk *data.HwTestStateKeeper) error {
	if sk.DutTopology == nil {
		return fmt.Errorf("Cmd %q missing dependency: DutTopology", cmd.GetCommandType())
	}
	if len(sk.DutTopology.GetDuts()) == 0 {
		return fmt.Errorf("Cmd %q missing dependency: PrimaryDut", cmd.GetCommandType())
	}
	primaryDut := sk.DutTopology.GetDuts()[0]

	if primaryDut.GetCacheServer().GetAddress() == nil {
		return fmt.Errorf("Cmd %q missing dependency: CacheServerAddress", cmd.GetCommandType())
	}
	cmd.CacheServerAddress = primaryDut.GetCacheServer().GetAddress()

	if primaryDut.GetChromeos().GetSsh() == nil {
		return fmt.Errorf("Cmd %q missing dependency: DutSshAddress", cmd.GetCommandType())
	}
	cmd.DutSshAddress = primaryDut.GetChromeos().GetSsh()

	return nil
}

func (cmd *DutServiceStartCmd) updateHwTestStateKeeper(ctx context.Context, sk *data.HwTestStateKeeper) error {
	if cmd.DutServerAddress != nil {
		sk.DutServerAddress = cmd.DutServerAddress
	}

	return nil
}

func NewDutServiceStartCmd(executor interfaces.ExecutorInterface) *DutServiceStartCmd {
	cmd := &DutServiceStartCmd{SingleCmdByExecutor: interfaces.NewSingleCmdByExecutor(DutServiceStartCmdType, executor)}
	cmd.ConcreteCmd = cmd
	return cmd
}
