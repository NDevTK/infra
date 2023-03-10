package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"

	testapi "go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
)

// TestsExecutionCmd represents test execution cmd.
type TestsExecutionCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	DutServerAddress *labapi.IpEndpoint
	TestSuites       []*testapi.TestSuite
	PrimaryDevice    *testapi.CrosTestRequest_Device
	CompanionDevices []*testapi.CrosTestRequest_Device

	// Updates
	TestResponses    *testapi.CrosTestResponse
	TkoPublishSrcDir string
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *TestsExecutionCmd) ExtractDependencies(
	ctx context.Context,
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
func (cmd *TestsExecutionCmd) UpdateStateKeeper(
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

func (cmd *TestsExecutionCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if sk.CftTestRequest == nil || sk.CftTestRequest.GetTestSuites() == nil || len(sk.CftTestRequest.GetTestSuites()) == 0 {
		return fmt.Errorf("Cmd %q missing dependency: TestSuites", cmd.GetCommandType())
	}
	cmd.TestSuites = sk.CftTestRequest.GetTestSuites()

	if sk.DutTopology == nil || sk.DutTopology.GetDuts() == nil || len(sk.DutTopology.GetDuts()) == 0 {
		return fmt.Errorf("Cmd %q missing dependency: PrimaryDevice", cmd.GetCommandType())
	}

	if sk.DutServerAddress == nil {
		return fmt.Errorf("Cmd %q missing dependency: DutServerAddress", cmd.GetCommandType())
	}

	cmd.PrimaryDevice = &testapi.CrosTestRequest_Device{Dut: sk.DutTopology.GetDuts()[0], DutServer: sk.DutServerAddress}

	cmd.CompanionDevices = []*testapi.CrosTestRequest_Device{}
	if sk.DutTopology != nil && sk.DutTopology.GetDuts() != nil && len(sk.DutTopology.GetDuts()) > 1 {
		for _, eachDut := range sk.DutTopology.GetDuts() {
			// TODO (azrahman): For multi-dut case, do we need dut server address for each companions?
			cmd.CompanionDevices = append(cmd.CompanionDevices, &testapi.CrosTestRequest_Device{Dut: eachDut})
		}
	}

	return nil
}

func (cmd *TestsExecutionCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.TestResponses != nil {
		sk.TestResponses = cmd.TestResponses
	}
	if cmd.TkoPublishSrcDir != "" {
		sk.TkoPublishSrcDir = cmd.TkoPublishSrcDir
	}

	return nil
}

func NewTestsExecutionCmd(executor interfaces.ExecutorInterface) *TestsExecutionCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(TestsExecutionCmdType, executor)
	cmd := &TestsExecutionCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
