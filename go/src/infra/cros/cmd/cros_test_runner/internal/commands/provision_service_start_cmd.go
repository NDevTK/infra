package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"

	test_api "go.chromium.org/chromiumos/config/go/test/api"
	lab_api "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
)

// ProvisionServiceStartCmd represents provision service start cmd.
type ProvisionServiceStartCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	ProvisionState   *test_api.ProvisionState
	DutServerAddress *lab_api.IpEndpoint
	PrimaryDut       *lab_api.Dut
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *ProvisionServiceStartCmd) ExtractDependencies(ctx context.Context, ski interfaces.StateKeeperInterface) error {
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

func (cmd *ProvisionServiceStartCmd) extractDepsFromHwTestStateKeeper(ctx context.Context, sk *data.HwTestStateKeeper) error {
	if sk.CftTestRequest.GetPrimaryDut().GetProvisionState() == nil {
		return fmt.Errorf("Cmd %q missing dependency: ProvisionState", cmd.GetCommandType())
	}

	cmd.ProvisionState = sk.CftTestRequest.GetPrimaryDut().GetProvisionState()

	if sk.DutTopology == nil || len(sk.DutTopology.GetDuts()) == 0 || sk.DutTopology.GetDuts()[0] == nil {
		return fmt.Errorf("Cmd %q missing dependency: PrimaryDut", cmd.GetCommandType())
	}

	cmd.PrimaryDut = sk.DutTopology.GetDuts()[0]

	if sk.DutServerAddress == nil {
		return fmt.Errorf("Cmd %q missing dependency: DutServerAddress", cmd.GetCommandType())
	}

	cmd.DutServerAddress = sk.DutServerAddress

	return nil
}

func NewProvisionServiceStartCmd(executor interfaces.ExecutorInterface) *ProvisionServiceStartCmd {
	cmd := &ProvisionServiceStartCmd{SingleCmdByExecutor: interfaces.NewSingleCmdByExecutor(ProvisionServiceStartCmdType, executor)}
	cmd.ConcreteCmd = cmd
	return cmd
}
