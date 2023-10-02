// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/cros_test_runner/data"
	"strings"

	testapi "go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
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
	TestArgs         *testapi.AutotestExecutionMetadata
	TastArgs         *testapi.TastExecutionMetadata

	// Updates
	TestResponses      *testapi.CrosTestResponse
	TkoPublishSrcDir   string
	CpconPublishSrcDir string
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
	cmd.TestArgs = sk.TestArgs
	cmd.TastArgs, _ = getTastExecutionMetadata(sk.CftTestRequest)

	if sk.DutTopology == nil || sk.DutTopology.GetDuts() == nil || len(sk.DutTopology.GetDuts()) == 0 {
		return fmt.Errorf("Cmd %q missing dependency: PrimaryDevice", cmd.GetCommandType())
	}

	if sk.DutServerAddress == nil {
		return fmt.Errorf("Cmd %q missing dependency: DutServerAddress", cmd.GetCommandType())
	}

	cmd.PrimaryDevice = &testapi.CrosTestRequest_Device{Dut: sk.DutTopology.GetDuts()[0], DutServer: sk.DutServerAddress}
	cmd.CompanionDevices = sk.CompanionDevices

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
	if cmd.CpconPublishSrcDir != "" {
		sk.CpconPublishSrcDir = cmd.CpconPublishSrcDir
	}
	return nil
}

func getTastExecutionMetadata(cftTestRequest *skylab_test_runner.CFTTestRequest) (*testapi.TastExecutionMetadata, error) {

	if len(cftTestRequest.TestSuites) == 0 {
		return nil, nil
	}

	firstTest := cftTestRequest.TestSuites[0].GetTestCaseIds().TestCaseIds[0].Value
	gcsPath := cftTestRequest.PrimaryDut.ProvisionState.SystemImage.SystemImagePath.Path

	if gcsPath != "" {
		if !strings.HasSuffix(gcsPath, "/") {
			gcsPath = gcsPath + "/"
		}

		if strings.HasPrefix(firstTest, "tast") {
			arg := &testapi.Arg{
				Flag:  "buildartifactsurl",
				Value: gcsPath,
			}
			tastExecutionMetadata := &testapi.TastExecutionMetadata{
				Args: []*testapi.Arg{arg},
			}

			return tastExecutionMetadata, nil
		}
	}

	return nil, nil
}

func NewTestsExecutionCmd(executor interfaces.ExecutorInterface) *TestsExecutionCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(TestsExecutionCmdType, executor)
	cmd := &TestsExecutionCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
