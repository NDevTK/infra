// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"strings"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/cros_test_runner/data"

	_go "go.chromium.org/chromiumos/config/go"
	testapipb "go.chromium.org/chromiumos/config/go/test/api"
	testapi_metadata "go.chromium.org/chromiumos/config/go/test/api/metadata"
	artifactpb "go.chromium.org/chromiumos/config/go/test/artifact"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// RdbPublishUploadCmd represents rdb publish upload cmd.
type RdbPublishUploadCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	CurrentInvocationId string
	TesthausUrl         string
	Sources             *testapi_metadata.PublishRdbMetadata_Sources

	// Either constructed TestResultForRdb is required,
	TestResultForRdb *artifactpb.TestResult
	// Or all these are required.
	GcsUrl        string
	DutTopology   *labapi.DutTopology
	TestResponses *testapipb.CrosTestResponse
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
	if sk.TesthausUrl == "" {
		return fmt.Errorf("Cmd %q missing dependency: TesthausUrl", cmd.GetCommandType())
	}
	if sk.CftTestRequest == nil {
		return fmt.Errorf("Cmd %q missing dependency: CftTestRequest", cmd.GetCommandType())
	}

	// If TestResultForRdb is not provided, try to construct it.
	if sk.TestResultForRdb == nil {
		logging.Infof(ctx, "Since TestResultForRdb is not provided, cmd will try to construct it.")

		if sk.BuildState == nil {
			return fmt.Errorf("Cmd %q missing dependency: BuildState", cmd.GetCommandType())
		}
		if sk.GcsUrl == "" {
			return fmt.Errorf("Cmd %q missing dependency: GcsUrl", cmd.GetCommandType())
		}
		if sk.DutTopology == nil {
			return fmt.Errorf("Cmd %q missing dependency: DutTopology", cmd.GetCommandType())
		}
		if sk.TestResponses == nil {
			return fmt.Errorf("Cmd %q missing dependency: TestResponses", cmd.GetCommandType())
		}

		// Construct testResultProto
		var testResultProtoErr error
		sk.TestResultForRdb, testResultProtoErr = constructTestResultFromStateKeeper(ctx, sk)
		if testResultProtoErr != nil {
			return errors.Annotate(testResultProtoErr, "Cmd %q failed to construct dependency: TestResultForRdb", cmd.GetCommandType()).Err()
		}
	}

	cmd.CurrentInvocationId = sk.CurrentInvocationId
	cmd.TestResultForRdb = sk.TestResultForRdb
	cmd.TesthausUrl = sk.TesthausUrl

	var err error
	if sk.CrosTestRunnerRequest != nil {
		cmd.Sources, err = SourcesFromPrimaryDevice(sk)
		if err != nil {
			return errors.Annotate(err, "Cmd %q failed to construct dependency: Sources", cmd.GetCommandType()).Err()
		}
	} else {
		cmd.Sources, err = SourcesFromCFTTestRequest(sk.CftTestRequest)
		if err != nil {
			return errors.Annotate(err, "Cmd %q failed to construct dependency: Sources", cmd.GetCommandType()).Err()
		}
	}

	cmd.GcsUrl = sk.GcsUrl
	cmd.TestResponses = sk.TestResponses

	return nil
}

func constructTestResultFromStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) (*artifactpb.TestResult, error) {

	build := sk.BuildState.Build()
	botDims := build.GetInfra().GetSwarming().GetBotDimensions()

	resultProto := &artifactpb.TestResult{}

	// Invocation level info
	populateTestInvocationInfo(ctx, resultProto, sk)

	// - Primary execution info
	primaryExecInfo := &artifactpb.ExecutionInfo{}
	resultProto.TestInvocation.PrimaryExecutionInfo = primaryExecInfo

	// -- Build info
	populatePrimaryBuildInfo(ctx, primaryExecInfo, sk, botDims)

	// -- Dut info
	isSkylab := populatePrimaryDutInfo(ctx, primaryExecInfo, sk)

	// -- Env info (skylab/satlab)
	populatePrimaryEnvInfo(ctx, primaryExecInfo, botDims, build, isSkylab)

	// - Secondary execution info
	populateSecondaryExecutionInfo(ctx, resultProto, sk)

	// TestRuns
	populateTestRunsInfo(ctx, resultProto, sk, botDims, build)

	return resultProto, nil
}

// populateTestInvocationInfo populates test invocation info.
func populateTestInvocationInfo(
	ctx context.Context,
	resultProto *artifactpb.TestResult,
	sk *data.HwTestStateKeeper) {

	testInv := &artifactpb.TestInvocation{}
	resultProto.TestInvocation = testInv
	if sk.DutTopology != nil {
		testInv.DutTopology = sk.DutTopology
	}
}

// populatePrimaryBuildInfo populates primary build info.
func populatePrimaryBuildInfo(
	ctx context.Context,
	primaryExecInfo *artifactpb.ExecutionInfo,
	sk *data.HwTestStateKeeper,
	botDims []*buildbucketpb.StringPair) {

	// Build info
	primaryBuildInfo := &artifactpb.BuildInfo{}
	primaryExecInfo.BuildInfo = primaryBuildInfo

	if buildName := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, sk.CrosTestRunnerRequest, "build"); buildName != "" {
		primaryBuildInfo.Name = buildName
	}

	// TODO (azrahman): Even though this says build-target, it's always being
	// set to board upstream (since pre trv2). This should be fixed at some point.
	if board := sk.CftTestRequest.GetPrimaryDut().GetDutModel().GetBuildTarget(); board != "" {
		primaryBuildInfo.Board = board
	} else if board := common.GetValueFromRequestKeyvals(ctx, nil, sk.CrosTestRunnerRequest, "build_target"); board != "" {
		primaryBuildInfo.Board = board
	}

	if buildTarget := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, sk.CrosTestRunnerRequest, "build_target"); buildTarget != "" {
		primaryBuildInfo.BuildTarget = buildTarget
	}

	populatePrimaryBuildMetadata(ctx, primaryBuildInfo, sk, botDims)

}

// populatePrimaryBuildMetadata populates primary build metadata.
func populatePrimaryBuildMetadata(
	ctx context.Context,
	primaryBuildInfo *artifactpb.BuildInfo,
	sk *data.HwTestStateKeeper,
	botDims []*buildbucketpb.StringPair) {

	// Build metadata
	buildMetadata := &artifactpb.BuildMetadata{}
	primaryBuildInfo.BuildMetadata = buildMetadata

	// - Firmware info
	firmwareInfo := &artifactpb.BuildMetadata_Firmware{}
	buildMetadata.Firmware = firmwareInfo

	// TODO (b/270230867): add missing properties when available.
	// ro_fwid, rw_fwid [Dependant on new cft logging service]

	// - Chipset info
	chipsetInfo := &artifactpb.BuildMetadata_Chipset{}
	buildMetadata.Chipset = chipsetInfo

	if wifiChip := getSingleTagValue(botDims, "label-wifi_chip"); wifiChip != "" {
		chipsetInfo.WifiChip = wifiChip
	}

	// - Kernel info
	kernalInfo := &artifactpb.BuildMetadata_Kernel{}
	buildMetadata.Kernel = kernalInfo

	// TODO (b/270230867): add missing properties when available.
	// kernel_version [Dependant on new cft logging service]

	// - Sku info
	skuInfo := &artifactpb.BuildMetadata_Sku{}
	buildMetadata.Sku = skuInfo

	if hwidSku := getSingleTagValue(botDims, "label-hwid_sku"); hwidSku != "" {
		skuInfo.HwidSku = hwidSku
	}

	// - Cellular info
	cellularInfo := &artifactpb.BuildMetadata_Cellular{}
	buildMetadata.Cellular = cellularInfo

	if carrier := getSingleTagValue(botDims, "label-carrier"); carrier != "" {
		cellularInfo.Carrier = carrier
	}

	// - Lacros info
	lacrosInfo := &artifactpb.BuildMetadata_Lacros{}
	buildMetadata.Lacros = lacrosInfo

	if ashVersion := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, sk.CrosTestRunnerRequest, "ash_version"); ashVersion != "" {
		lacrosInfo.AshVersion = ashVersion
	}

	if lacrosVersion := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, sk.CrosTestRunnerRequest, "lacros_version"); lacrosVersion != "" {
		lacrosInfo.LacrosVersion = lacrosVersion
	}
}

// populatePrimaryDutInfo populates primary dut info.
func populatePrimaryDutInfo(
	ctx context.Context,
	primaryExecInfo *artifactpb.ExecutionInfo,
	sk *data.HwTestStateKeeper) bool {

	// Dut info
	primaryDutInfo := &artifactpb.DutInfo{}
	primaryExecInfo.DutInfo = primaryDutInfo

	isSkylab := true
	testDuts := sk.DutTopology.GetDuts()
	if len(testDuts) > 0 {
		primaryDutInfo.Dut = testDuts[0]
		isSkylab = !strings.HasPrefix(testDuts[0].GetId().GetValue(), "satlab-")
	}

	primaryDut := sk.CftTestRequest.GetPrimaryDut()
	if sk.PrimaryDeviceMetadata != nil {
		primaryDut = sk.PrimaryDeviceMetadata
	}
	provisionState := primaryDut.GetProvisionState()
	if provisionState != nil {
		primaryDutInfo.ProvisionState = provisionState
	}

	return isSkylab
}

// populatePrimaryEnvInfo populates primary env info.
func populatePrimaryEnvInfo(
	ctx context.Context,
	primaryExecInfo *artifactpb.ExecutionInfo,
	botDims []*buildbucketpb.StringPair,
	build *buildbucketpb.Build,
	isSkylab bool) {

	// - Drone info
	droneInfo := &artifactpb.DroneInfo{}

	if drone := getSingleTagValue(botDims, "drone"); drone != "" {
		droneInfo.Drone = drone
	}
	if droneServer := getSingleTagValue(botDims, "drone_server"); droneServer != "" {
		droneInfo.DroneServer = droneServer
	}

	// - Swarming info
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

	// - BuildBucket info
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
}

// populateSecondaryExecutionInfo populates secondary execution info.
func populateSecondaryExecutionInfo(
	ctx context.Context,
	resultProto *artifactpb.TestResult,
	sk *data.HwTestStateKeeper) {

	// If more than one dut, then it's multi-duts.
	// TODO (azrahman): check if inventory service actually provides these duts info
	// or not for multi-duts. If not, raise this issue to proper channel.
	testDuts := sk.DutTopology.GetDuts()
	if len(testDuts) > 1 {
		companionDevices := testDuts[1 : len(testDuts)-1]
		companionDevicesMetadata := sk.CftTestRequest.GetCompanionDuts()

		if sk.CompanionDevicesMetadata != nil {
			companionDevices = []*labapi.Dut{}
			for _, device := range sk.CompanionDevices {
				companionDevices = append(companionDevices, device.GetDut())
			}
			companionDevicesMetadata = sk.CompanionDevicesMetadata
		}

		secondaryExecInfos := []*artifactpb.ExecutionInfo{}
		for i, device := range companionDevices {
			secondaryExecInfo := &artifactpb.ExecutionInfo{}
			secondaryDutInfo := &artifactpb.DutInfo{}
			secondaryExecInfo.DutInfo = secondaryDutInfo
			secondaryDutInfo.Dut = device

			secondaryBuildInfo := &artifactpb.BuildInfo{}
			secondaryExecInfo.BuildInfo = secondaryBuildInfo
			if i < len(companionDevicesMetadata) {
				if secondaryBoard := companionDevicesMetadata[i].GetDutModel().GetBuildTarget(); secondaryBoard != "" {
					secondaryBuildInfo.Board = secondaryBoard
				}
				if secondaryProvisionState := companionDevicesMetadata[i].GetProvisionState(); secondaryProvisionState != nil {
					secondaryDutInfo.ProvisionState = secondaryProvisionState
				}
			}

			secondaryExecInfos = append(secondaryExecInfos, secondaryExecInfo)
		}
		resultProto.TestInvocation.SecondaryExecutionsInfo = secondaryExecInfos
	}
}

// populateTestRunsInfo populates test runs info.
func populateTestRunsInfo(
	ctx context.Context,
	resultProto *artifactpb.TestResult,
	sk *data.HwTestStateKeeper,
	botDims []*buildbucketpb.StringPair,
	build *buildbucketpb.Build) {

	testRuns := []*artifactpb.TestRun{}
	resultProto.TestRuns = testRuns

	suite := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, sk.CrosTestRunnerRequest, "suite")
	branch := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, sk.CrosTestRunnerRequest, "branch")
	mainBuilderName := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, sk.CrosTestRunnerRequest, "master_build_config")
	displayName := getSingleTagValue(build.Tags, "display_name")
	testplanId := getSingleTagValue(build.Tags, "test-plan-id")
	for _, testCaseResult := range sk.TestResponses.GetTestCaseResults() {
		// - TestRun
		testRun := &artifactpb.TestRun{}
		testCaseInfo := &artifactpb.TestCaseInfo{}
		testRun.TestCaseInfo = testCaseInfo

		testRun.LogsInfo = []*_go.StoragePath{{HostType: _go.StoragePath_GS, Path: sk.GcsUrl}}

		// -- TestCaseInfo
		testCaseInfo.TestCaseResult = testCaseResult
		if displayName != "" {
			testCaseInfo.DisplayName = displayName
		}
		if testplanId != "" {
			testCaseInfo.TestPlanId = testplanId
		}
		if suite != "" {
			testCaseInfo.Suite = suite
		}
		if branch != "" {
			testCaseInfo.Branch = branch
		}
		if mainBuilderName != "" {
			testCaseInfo.MainBuilderName = mainBuilderName
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

// getTaskRequestId converts the swarming task run id with non "0" suffix to the swarming task
// request id with "0" suffix. Both can be used to point to the same swarming
// task. Swarming supported implicit retry and first task has "1" in suffix and
// retried task has "2" in suffix.
func getTaskRequestId(taskId string) string {
	if taskId == "" {
		return ""
	}

	return fmt.Sprintf("%s0", taskId[:len(taskId)-1])
}

// SourcesFromPrimaryDevice returns the code sources tested in the given
// testing request assuming provision was successful and the version did
// not shcnage during the test.
func SourcesFromPrimaryDevice(sk *data.HwTestStateKeeper) (*testapi_metadata.PublishRdbMetadata_Sources, error) {
	provisionState := sk.PrimaryDeviceMetadata.GetProvisionState()
	if provisionState == nil {
		// Invalid request.
		return nil, errors.Reason("CFTTestRequest: primary_dut.provision_state missing").Err()
	}

	// The path to the system image.
	imagePath := provisionState.SystemImage.GetSystemImagePath()
	if imagePath == nil || imagePath.GetHostType() != _go.StoragePath_GS {
		// For non-GS stored build outputs (e.g. local files),
		// we do not have information about the sources used.
		return nil, nil
	}
	if !strings.HasPrefix(imagePath.GetPath(), "gs://") {
		return nil, errors.Reason("CFTTestRequest: primary_dut.provision_state.system_image.system_image_path.path: must start with gs://").Err()
	}
	if strings.HasSuffix(imagePath.GetPath(), "/") {
		return nil, errors.Reason("CFTTestRequest: primary_dut.provision_state.system_image.system_image_path.path: must not have trailing '/'").Err()
	}
	return &testapi_metadata.PublishRdbMetadata_Sources{
		// Path to the file in Google Cloud Storage that contains
		// information about the code sources built into the build.
		GsPath: imagePath.Path + common.SourceMetadataPath,
		// If custom firmware is used or custom packages are deployed
		// that were not built as part of the Chrome OS image (e.g. Lacros
		// testing or firmware testing), the test is not a pure test
		// of the build sources.
		IsDeploymentDirty: provisionState.GetFirmware() != nil || len(provisionState.GetPackages()) > 0,
	}, nil
}

// SourcesFromCFTTestRequest returns the code sources tested in the given
// CFT testing request.
func SourcesFromCFTTestRequest(request *skylab_test_runner.CFTTestRequest) (*testapi_metadata.PublishRdbMetadata_Sources, error) {
	if request == nil {
		return nil, errors.Reason("CFTTestRequest: missing").Err()
	}
	provisionState := request.GetPrimaryDut().GetProvisionState()
	if provisionState == nil {
		// Invalid request.
		return nil, errors.Reason("CFTTestRequest: primary_dut.provision_state missing").Err()
	}

	// The path to the system image.
	imagePath := provisionState.SystemImage.GetSystemImagePath()
	if imagePath == nil || imagePath.GetHostType() != _go.StoragePath_GS {
		// For non-GS stored build outputs (e.g. local files),
		// we do not have information about the sources used.
		return nil, nil
	}
	if !strings.HasPrefix(imagePath.GetPath(), "gs://") {
		return nil, errors.Reason("CFTTestRequest: primary_dut.provision_state.system_image.system_image_path.path: must start with gs://").Err()
	}
	if strings.HasSuffix(imagePath.GetPath(), "/") {
		return nil, errors.Reason("CFTTestRequest: primary_dut.provision_state.system_image.system_image_path.path: must not have trailing '/'").Err()
	}

	return &testapi_metadata.PublishRdbMetadata_Sources{
		// Path to the file in Google Cloud Storage that contains
		// information about the code sources built into the build.
		GsPath: imagePath.Path + common.SourceMetadataPath,
		// If custom firmware is used or custom packages are deployed
		// that were not built as part of the Chrome OS image (e.g. Lacros
		// testing or firmware testing), the test is not a pure test
		// of the build sources.
		IsDeploymentDirty: provisionState.GetFirmware() != nil || len(provisionState.GetPackages()) > 0,
	}, nil
}

func NewRdbPublishUploadCmd(executor interfaces.ExecutorInterface) *RdbPublishUploadCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(RdbPublishUploadCmdType, executor)
	cmd := &RdbPublishUploadCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
