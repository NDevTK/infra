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
	"strings"

	_go "go.chromium.org/chromiumos/config/go"
	testapipb "go.chromium.org/chromiumos/config/go/test/api"
	artifactpb "go.chromium.org/chromiumos/config/go/test/artifact"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/luciexe/build"
)

// RdbPublishUploadCmd represents rdb publish upload cmd.
type RdbPublishUploadCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	CurrentInvocationId string
	TestResultForRdb    *artifactpb.TestResult
	StainlessUrl        string
	TesthausUrl         string
	BuildState          *build.State
	GcsUrl              string
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
	if sk.BuildState == nil {
		return errors.Annotate(err, fmt.Sprintf("Cmd %q missing dependency: BuildState", cmd.GetCommandType())).Err()
	}
	if sk.GcsUrl == "" {
		return errors.Annotate(err, fmt.Sprintf("Cmd %q missing dependency: GcsUrl", cmd.GetCommandType())).Err()
	}

	cmd.CurrentInvocationId = sk.CurrentInvocationId
	cmd.StainlessUrl = sk.StainlessUrl
	cmd.TestResultForRdb = testResult
	cmd.TesthausUrl = sk.TesthausUrl
	cmd.BuildState = sk.BuildState
	cmd.GcsUrl = sk.GcsUrl

	return nil
}

func (cmd *RdbPublishUploadCmd) constructTestResultFromStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) (*artifactpb.TestResult, error) {

	build := sk.BuildState.Build()
	botDims := build.GetInfra().GetSwarming().GetBotDimensions()

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

	if buildName := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, "build"); buildName != "" {
		primaryBuildInfo.Name = buildName
	}
	// TODO (azrahman): Even though this says build-target, it's always being
	// set to board upstream (since pre trv2). This should be fixed at some point.

	if board := sk.CftTestRequest.GetPrimaryDut().GetDutModel().GetBuildTarget(); board != "" {
		primaryBuildInfo.Board = board
	}

	if buildTarget := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, "build_target"); buildTarget != "" {
		primaryBuildInfo.BuildTarget = buildTarget
	}

	// --- Build metadata
	buildMetadata := &artifactpb.BuildMetadata{}
	primaryBuildInfo.BuildMetadata = buildMetadata

	// ---- Firmware info
	firmwareInfo := &artifactpb.BuildMetadata_Firmware{}
	buildMetadata.Firmware = firmwareInfo

	// Missing
	// ro_fwid, rw_fwid [Dependant on new cft logging service]

	// ---- Chipset info
	chipsetInfo := &artifactpb.BuildMetadata_Chipset{}
	buildMetadata.Chipset = chipsetInfo

	if wifiChip := getSingleTagValue(botDims, "label-wifi_chip"); wifiChip != "" {
		chipsetInfo.WifiChip = wifiChip
	}

	// ---- Kernel info
	kernalInfo := &artifactpb.BuildMetadata_Kernel{}
	buildMetadata.Kernel = kernalInfo

	// Missing
	// kernel_version [Dependant on new cft logging service]

	// ---- Sku info
	skuInfo := &artifactpb.BuildMetadata_Sku{}
	buildMetadata.Sku = skuInfo

	if hwidSku := getSingleTagValue(botDims, "label-hwid_sku"); hwidSku != "" {
		skuInfo.HwidSku = hwidSku
	}

	// ---- Cellular info
	cellularInfo := &artifactpb.BuildMetadata_Cellular{}
	buildMetadata.Cellular = cellularInfo

	if carrier := getSingleTagValue(botDims, "label-carrier"); carrier != "" {
		cellularInfo.Carrier = carrier
	}

	// ---- Lacros info
	lacrosInfo := &artifactpb.BuildMetadata_Lacros{}
	buildMetadata.Lacros = lacrosInfo

	if ashVersion := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, "ash_version"); ashVersion != "" {
		lacrosInfo.AshVersion = ashVersion
	}

	if lacrosVersion := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, "lacros_version"); lacrosVersion != "" {
		lacrosInfo.LacrosVersion = lacrosVersion
	}

	// -- Dut info
	primaryDutInfo := &artifactpb.DutInfo{}
	primaryExecInfo.DutInfo = primaryDutInfo

	isSkylab := true
	testDuts := sk.DutTopology.GetDuts()
	if len(testDuts) > 0 {
		primaryDutInfo.Dut = testDuts[0]
		isSkylab = !strings.HasPrefix(testDuts[0].GetId().GetValue(), "satlab-")
	}

	provisionState := sk.CftTestRequest.GetPrimaryDut().GetProvisionState()
	if provisionState != nil {
		primaryDutInfo.ProvisionState = provisionState
	}

	// -- Env info (skylab/satlab)

	// --- Drone info
	droneInfo := &artifactpb.DroneInfo{}

	if drone := getSingleTagValue(botDims, "drone"); drone != "" {
		droneInfo.Drone = drone
	}
	if droneServer := getSingleTagValue(botDims, "drone_server"); droneServer != "" {
		droneInfo.DroneServer = droneServer
	}

	// --- Swarming info
	swarmingInfo := &artifactpb.SwarmingInfo{}

	if testTaskId := getTaskRequestId(build.GetInfra().GetSwarming().GetTaskId()); testTaskId != "" {
		swarmingInfo.TaskId = testTaskId
	}
	if suiteTaskId := getTaskRequestId(build.GetInfra().GetSwarming().GetParentRunId()); suiteTaskId != "" {
		swarmingInfo.SuiteTaskId = suiteTaskId
	}
	if builder := build.GetBuilder(); builder != nil {
		swarmingInfo.TaskName = fmt.Sprintf("bb-%d-%s/%s/%s", build.GetId(), builder.GetProject(), builder.GetBucket(), builder.GetBuilder())
	}
	if pool := getSingleTagValue(botDims, "pool"); pool != "" {
		swarmingInfo.Pool = pool
	}
	if labelPool := getSingleTagValue(botDims, "label-pool"); labelPool != "" {
		swarmingInfo.LabelPool = labelPool
	}

	// --- BuildBucket info
	bbInfo := &artifactpb.BuildbucketInfo{}

	if len(build.AncestorIds) > 0 {
		bbInfo.AncestorIds = build.AncestorIds
	}

	if isSkylab {
		// Skylab
		skylabInfo := &artifactpb.SkylabInfo{DroneInfo: droneInfo, SwarmingInfo: swarmingInfo, BuildbucketInfo: bbInfo}
		primaryExecInfo.EnvInfo = &artifactpb.ExecutionInfo_SkylabInfo{SkylabInfo: skylabInfo}
	} else {
		// Satlab
		satlabInfo := &artifactpb.SatlabInfo{SwarmingInfo: swarmingInfo, BuildbucketInfo: bbInfo}
		primaryExecInfo.EnvInfo = &artifactpb.ExecutionInfo_SatlabInfo{SatlabInfo: satlabInfo}
	}

	// - Secondary execution info

	// If more than one dut, then it's multi-duts.
	// TODO (azrahman): check if inventory service actually provides these duts info
	// or not for multi-duts. If not, raise this issue to proper channel.
	if len(testDuts) > 1 {
		inputCompDuts := sk.CftTestRequest.GetCompanionDuts()

		secondaryExecInfos := []*artifactpb.ExecutionInfo{}
		for i, dut := range testDuts {
			secondaryExecInfo := &artifactpb.ExecutionInfo{}
			secondaryDutInfo := &artifactpb.DutInfo{}
			secondaryExecInfo.DutInfo = secondaryDutInfo
			secondaryDutInfo.Dut = dut

			secondaryBuildInfo := &artifactpb.BuildInfo{}
			secondaryExecInfo.BuildInfo = secondaryBuildInfo
			if i < len(inputCompDuts) {
				if secondaryBoard := inputCompDuts[i].GetDutModel().GetBuildTarget(); secondaryBoard != "" {
					secondaryBuildInfo.Board = secondaryBoard
				}
			}

			secondaryExecInfos = append(secondaryExecInfos, secondaryExecInfo)
		}
		resultProto.TestInvocation.SecondaryExecutionsInfo = secondaryExecInfos
	}

	// - TestRuns

	testRuns := []*artifactpb.TestRun{}
	resultProto.TestRuns = testRuns

	suite := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, "suite")
	branch := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, "branch")
	main_builder_name := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, "master_build_config")
	for _, testCaseResult := range sk.TestResponses.GetTestCaseResults() {
		// -- TestRun
		testRun := &artifactpb.TestRun{}
		testCaseInfo := &artifactpb.TestCaseInfo{}
		testRun.TestCaseInfo = testCaseInfo

		testRun.LogsInfo = []*_go.StoragePath{{HostType: _go.StoragePath_GS, Path: cmd.GcsUrl}}

		// --- TestCaseMetadata
		// TODO (azrahman): Remove this once result_adapter gets the id from testcaseinfo.
		// Adding this for now to prevent result_adapter to fail upload.
		testCaseMetadata := &testapipb.TestCaseMetadata{}
		testCase := &testapipb.TestCase{Id: testCaseResult.TestCaseId}
		testCaseMetadata.TestCase = testCase

		// --- TestCaseInfo
		testCaseInfo.TestCaseResult = testCaseResult
		testCaseInfo.DisplayName = testCaseResult.GetTestCaseId().GetValue()
		if suite != "" {
			testCaseInfo.Suite = suite
		}
		if branch != "" {
			testCaseInfo.Branch = branch
		}
		if main_builder_name != "" {
			testCaseInfo.MainBuilderName = main_builder_name
		}

		timeInfo := &artifactpb.TimingInfo{}
		testRun.TimeInfo = timeInfo

		timeInfo.StartedTime = testCaseResult.GetStartTime()
		timeInfo.Duration = testCaseResult.GetDuration()
		timeInfo.QueuedTime = build.GetCreateTime()

		testRun.TestHarness = testCaseResult.GetTestHarness()

		testRuns = append(testRuns, testRun)
	}

	resultProto.TestRuns = testRuns

	return resultProto, nil
}

// getTagValues gets tag values from provided string pairs.
func getTagValues(tags []*buildbucketpb.StringPair, key string) []string {
	values := []string{}
	if len(tags) == 0 {
		return values
	}

	for _, tag := range tags {
		if tag.GetKey() == key {
			values = append(values, tag.GetValue())
		}
	}

	return values
}

// getSingleTagValue gets the first value found from provided string pairs.
func getSingleTagValue(tags []*buildbucketpb.StringPair, key string) string {
	values := getTagValues(tags, key)
	if len(values) > 0 {
		return values[0]
	} else {
		return ""
	}
}

// getTaskRequestId constructs the base swarming id from provided id.
func getTaskRequestId(taskId string) string {
	if taskId == "" {
		return ""
	}

	return fmt.Sprintf("%s0", taskId[:len(taskId)-1])
}

func NewRdbPublishUploadCmd(executor interfaces.ExecutorInterface) *RdbPublishUploadCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(RdbPublishUploadCmdType, executor)
	cmd := &RdbPublishUploadCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
