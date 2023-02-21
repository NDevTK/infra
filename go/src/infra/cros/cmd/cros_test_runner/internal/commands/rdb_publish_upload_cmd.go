// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	"os"

	_go "go.chromium.org/chromiumos/config/go"
	artifactpb "go.chromium.org/chromiumos/config/go/test/artifact"
	"go.chromium.org/luci/common/errors"
)

// RdbPublishUploadCmd represents rdb publish upload cmd.
type RdbPublishUploadCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	CurrentInvocationId string
	TestResultForRdb    *artifactpb.TestResult
	StainlessUrl        string
	TesthausUrl         string
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *RdbPublishUploadCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

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

func (cmd *RdbPublishUploadCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if sk.CurrentInvocationId == "" {
		return fmt.Errorf("Cmd %q missing dependency: CurrentInvocationId", cmd.GetCommandType())
	}
	if sk.StainlessUrl == "" {
		return fmt.Errorf("Cmd %q missing dependency: StainlessUrl", cmd.GetCommandType())
	}
	if sk.TesthausUrl == "" {
		return fmt.Errorf("Cmd %q missing dependency: TesthausUrl", cmd.GetCommandType())
	}
	testResult, err := cmd.constructTestResultFromStateKeeper(ctx, sk)
	if err != nil {
		return errors.Annotate(err, fmt.Sprintf("Cmd %q missing dependency: TestResultForRdb", cmd.GetCommandType())).Err()
	}

	cmd.CurrentInvocationId = sk.CurrentInvocationId
	cmd.StainlessUrl = sk.StainlessUrl
	cmd.TestResultForRdb = testResult
	cmd.TesthausUrl = sk.TesthausUrl

	return nil
}

func (cmd *RdbPublishUploadCmd) constructTestResultFromStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) (*artifactpb.TestResult, error) {

	// TODO (azrahman): consider moving all these logics to rdb-publish if possible.
	resultProto := &artifactpb.TestResult{}

	// Invocation level info
	testInv := &artifactpb.TestInvocation{}
	resultProto.TestInvocation = testInv
	if sk.DutTopology != nil {
		testInv.DutTopology = sk.DutTopology
	}

	// - Primary execution info
	primaryExecInfo := &artifactpb.ExecutionInfo{}
	resultProto.TestInvocation.PrimaryExecutionInfo = primaryExecInfo

	// -- Build info
	primaryBuildInfo := &artifactpb.BuildInfo{}
	primaryExecInfo.BuildInfo = primaryBuildInfo

	buildName := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, "build")
	if buildName != "" {
		primaryBuildInfo.Name = buildName
	}
	// TODO (azrahman): Even though this says build-target, it's always being
	// set to board upstream (since pre trv2). This should be fixed at some point.
	board := sk.CftTestRequest.GetPrimaryDut().GetDutModel().GetBuildTarget()
	if board != "" {
		primaryBuildInfo.Board = board
	}

	buildTarget := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, "build_target")
	if buildTarget != "" {
		primaryBuildInfo.BuildTarget = buildTarget
	}

	// -- Dut info
	primaryDutInfo := &artifactpb.DutInfo{}
	primaryExecInfo.DutInfo = primaryDutInfo

	testDuts := sk.DutTopology.GetDuts()
	if len(testDuts) > 0 {
		primaryDutInfo.Dut = testDuts[0]
	}

	provisionState := sk.CftTestRequest.GetPrimaryDut().GetProvisionState()
	if provisionState != nil {
		primaryDutInfo.ProvisionState = provisionState
	}

	// -- Env info (skylab/satlab)
	// TODO (azrahman): Is this the best way to decide skylab vs satlab?
	if _, exists := os.LookupEnv("SKYLAB_DUT_ID"); exists {
		skylabInfo := &artifactpb.SkylabInfo{}
		primaryExecInfo.EnvInfo = &artifactpb.ExecutionInfo_SkylabInfo{SkylabInfo: skylabInfo}

		// --- Drone info

		// --- Swarming info
	} else {
		satlabInfo := &artifactpb.SatlabInfo{}
		primaryExecInfo.EnvInfo = &artifactpb.ExecutionInfo_SatlabInfo{SatlabInfo: satlabInfo}
	}

	// --- Buildbucket info

	// - TestRuns

	testRuns := []*artifactpb.TestRun{}
	resultProto.TestRuns = testRuns

	suite := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, "suite")
	for _, testCaseResult := range sk.TestResponses.GetTestCaseResults() {
		testRun := &artifactpb.TestRun{}
		testCaseInfo := &artifactpb.TestCaseInfo{}
		testRun.TestCaseInfo = testCaseInfo

		testCaseInfo.TestCaseResult = testCaseResult
		if suite != "" {
			testCaseInfo.Suite = suite
		}

		testRun.LogsInfo = []*_go.StoragePath{testCaseResult.GetResultDirPath()}

		timeInfo := &artifactpb.TimingInfo{}
		testRun.TimeInfo = timeInfo

		timeInfo.StartedTime = testCaseResult.GetStartTime()
		timeInfo.Duration = testCaseResult.GetDuration()

		testRun.TestHarness = testCaseResult.GetTestHarness()

		testRuns = append(testRuns, testRun)
	}

	resultProto.TestRuns = testRuns

	return resultProto, nil
}

func NewRdbPublishUploadCmd(executor interfaces.ExecutorInterface) *RdbPublishUploadCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(RdbPublishUploadCmdType, executor)
	cmd := &RdbPublishUploadCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
