// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// DISCLAIMER: Bunch of TODOs and lots of commented code in this file.
// Keeping this file as helper for now until this whole file/trv2 request
// generator is stable. The todos will be resolved over time while getting data
// from new input fields and introducing functionalities. The commented out
// codes are from CTPv1 that will help with making sure no data/info are
// getting missed while resolving those functions. Then these functions will be
// moved to common_lib.

package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	goconfig "go.chromium.org/chromiumos/config/go"
	"go.chromium.org/chromiumos/config/go/build/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/config"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/buildbucket"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/ctpv2/data"
	"infra/libs/skylab/inventory"
	"infra/libs/skylab/request"
	"infra/libs/skylab/worker"
)

const (
	// This deadline is constructed from various CTP req params that we do not
	// want to depend on for ctpv2. So hardcoding them for now so that they are in
	// one place. And later may move to new input params for ctpv2 or configs if
	// required.
	// TODO (azrahman): revisit this.
	DefaultTimeout = 8 * time.Hour // Intentionally put a large number for now so
	// that reqs don't timeout due to this.

	CtpRequestUIDTemplate = "TestPlanRuns/%d/%s"
	DutPoolQuota          = "DUT_POOL_QUOTA"
)

// GenerateTrv2Req generates ScheduleBuildRequest.
func GenerateTrv2Req(ctx context.Context, hwDef *testapi.SwarmingDefinition, testCases []*testapi.CTPTestCase, build *build.State, suiteInfo *testapi.SuiteInfo, canOutliveParent bool) (*buildbucketpb.ScheduleBuildRequest, error) {
	var err error

	// Create bb request
	reqArgs, _ := GenerateArgs(ctx, hwDef, testCases, build, suiteInfo)
	req, err := reqArgs.NewBBRequest(common.TestRunnerBuilderID())
	if err != nil {
		return nil, err
	}

	bbCtx := lucictx.GetBuildbucket(ctx)

	if bbCtx != nil && bbCtx.GetScheduleBuildToken() != "" && bbCtx.GetScheduleBuildToken() != buildbucket.DummyBuildbucketToken {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(buildbucket.BuildbucketTokenHeader, bbCtx.ScheduleBuildToken))

		// Decide if the child can outlive its parent or not.
		if canOutliveParent {
			req.CanOutliveParent = buildbucketpb.Trinary_YES
		}
	}

	return req, nil
}

// GenerateArgs generates args for the builder request.
func GenerateArgs(ctx context.Context, hwDef *testapi.SwarmingDefinition, testCases []*testapi.CTPTestCase, build *build.State, suiteInfo *testapi.SuiteInfo) (*request.Args, error) {
	suiteName := suiteInfo.GetSuiteRequest().GetTestSuite().GetName()

	cmd := *createCommand(ctx, suiteName)
	labels, err := createLabels(hwDef, suiteInfo)
	if err != nil {
		return nil, errors.Annotate(err, "error while creating labels: ").Err()
	}
	secondaryLabels, err := createSecondaryLabels()
	if err != nil {
		return nil, errors.Annotate(err, "error while creating secondary labels: ").Err()
	}
	freeformDims := createFreeformDims(hwDef)
	currBbid := build.Build().GetId()
	currSwarmingId := os.Getenv("SWARMING_TASK_ID")
	if currSwarmingId == "" {
		logging.Warningf(ctx, "SWARMING_TASK_ID not set. So child builds won't have this.")
	}
	parentRequestUID := fmt.Sprintf(CtpRequestUIDTemplate, currBbid, suiteName)
	// TODO (azrahman): Should we even use provisionable dims for scheduling?
	provisionableDims, _ := createProvisionableDimensions()

	cftTestRequest, _ := createCftTestRequest(ctx, hwDef, testCases, suiteInfo, parentRequestUID, currBbid, currSwarmingId)

	args := &request.Args{
		Cmd:                              cmd,
		SchedulableLabels:                labels,
		SecondaryDevicesLabels:           secondaryLabels,
		Dimensions:                       freeformDims,
		ParentTaskID:                     currSwarmingId,
		ParentRequestUID:                 parentRequestUID,
		Priority:                         10, // TODO (azrahman): Not required for scheduke. Hard coding it for now.
		ProvisionableDimensions:          provisionableDims,
		ProvisionableDimensionExpiration: time.Minute,
		SwarmingTags:                     createSwarmingTags(),
		SwarmingPool:                     suiteInfo.GetSuiteMetadata().GetPool(),
		TestRunnerRequest:                nil, // Always nil for CFT.
		CFTTestRunnerRequest:             cftTestRequest,
		CFTIsEnabled:                     true,           // Always true
		Timeout:                          DefaultTimeout, // TODO (azrahman): Get this from input.
		Experiments:                      build.Build().GetInput().Experiments,
		GerritChanges:                    build.Build().GetInput().GerritChanges,
		ResultsConfig:                    nil, // TODO (azrahman): Investigate if we need this.
	}

	return args, nil
}

// generateReqName generates request name.
func generateReqName(board string, build string, suiteName string) string {
	retVal := board
	if build != "" {
		retVal += fmt.Sprintf("-%s", build)
	}
	if suiteName != "" {
		retVal += fmt.Sprintf(".%s", suiteName)
	}
	return retVal
}

// createCommand creates cmd for the builder request.
func createCommand(ctx context.Context, suiteName string) *worker.Command {
	keyvals := make(map[string]string)
	keyvals["suite"] = suiteName
	keyvals["label"] = suiteName

	cmd := &worker.Command{
		ClientTest:      false,
		Deadline:        time.Now().Add(DefaultTimeout), // Deadline should come from input (add it)
		Keyvals:         keyvals,
		OutputToIsolate: true,
		TaskName:        suiteName, // 111: was test name in ctpv1
		//TestArgs:        "bar", // TODO (azrahman): add test args support when
		// ctpv2 supports it
	}

	// This was retrieved from input in ctpv1 but turns out this was always the same.
	logdogConfig := &config.Config_SkylabWorker{
		LuciProject: "chromeos",
		LogDogHost:  "luci-logdog.appspot.com",
	}
	cmd.Config(data.Wrap(logdogConfig))

	return cmd
}

// createFreeformDims creates free form dims from swarming def.
func createFreeformDims(hwDef *testapi.SwarmingDefinition) []string {
	freeformDims := []string{"dut_state:ready"}
	if hwDef.GetDutInfo().GetChromeos().GetHwid() != "" {
		freeformDims = append(freeformDims, fmt.Sprintf("hwid:%s", hwDef.GetDutInfo().GetChromeos().GetHwid()))
	}

	return freeformDims
}

// findGcsPath finds gcs path for provided board.
func findGcsPath(suiteInfo *testapi.SuiteInfo, board string, id string) string {
	for _, target := range suiteInfo.GetSuiteMetadata().GetTargetRequirements() {
		// This is [0] indexed because MO will always reduce it to 1 item.
		hwDef := target.GetHwRequirements().GetHwDefinition()[0]
		swDef := target.GetSwRequirement().GetVariant()
		if hwDef.GetDutInfo().GetChromeos().GetDutModel().GetBuildTarget() == board && id == swDef {
			return target.GetSwRequirement().GetGcsPath()
		}
	}

	return ""
}

// createCftTestRequest creates cft test request.
func createCftTestRequest(ctx context.Context, hwDef *testapi.SwarmingDefinition, testCases []*testapi.CTPTestCase, suiteInfo *testapi.SuiteInfo, parentRequestUID string, currBbid int64, currSwarmingId string) (*skylab_test_runner.CFTTestRequest, error) {
	deadline := timestamppb.New(time.Now().Add(2 * time.Hour))

	buildTarget := hwDef.GetDutInfo().GetChromeos().GetDutModel().GetBuildTarget()
	modelName := hwDef.GetDutInfo().GetChromeos().GetDutModel().GetBuildTarget()
	dutModel := &labapi.DutModel{
		BuildTarget: buildTarget,
		ModelName:   modelName,
	}

	gcsPath := findGcsPath(suiteInfo, buildTarget, hwDef.GetProvisionInfo()[0].GetIdentifier())
	if gcsPath == "" {
		logging.Infof(ctx, "GcsPath was not found for build target: %s", buildTarget)
		return nil, fmt.Errorf("GcsPath was not found for build target: %s", buildTarget)
	}
	containerGcsPath := gcsPath + common.ContainerMetadataPath

	tempRootDir := os.Getenv("TEMPDIR")
	localFilePath, err := common.DownloadGcsFileToLocal(ctx, containerGcsPath, tempRootDir)
	if err != nil {
		logging.Infof(ctx, "error while downloading gcs file to local: %s", err)
		return nil, err
	}

	containerMetadata := &api.ContainerMetadata{}
	err = common.ReadProtoJSONFile(ctx, localFilePath, containerMetadata)
	if err != nil {
		logging.Infof(ctx, "error while reading proto json file: %s", err)
		return nil, err
	}

	companionDuts := []*skylab_test_runner.CFTTestRequest_Device{}

	testCaseIds := []*testapi.TestCase_Id{}
	for _, testCase := range testCases {
		testCaseIds = append(testCaseIds, testCase.GetMetadata().GetTestCase().GetId())
	}
	testSuites := []*testapi.TestSuite{
		{
			Name: suiteInfo.GetSuiteRequest().GetTestSuite().GetName(),
			Spec: &testapi.TestSuite_TestCaseIds{
				TestCaseIds: &testapi.TestCaseIdList{
					TestCaseIds: testCaseIds,
				},
			},
			ExecutionMetadata: suiteInfo.GetSuiteMetadata().GetExecutionMetadata(),
		},
	}

	suiteName := suiteInfo.GetSuiteRequest().GetTestSuite().GetName()

	keyvals := make(map[string]string)
	// TODO (azrahman): look into how to set these properly
	keyvals["suite"] = suiteName                            // suite name
	keyvals["label"] = fmt.Sprintf("%s-shard-0", suiteName) // test name
	keyvals["build"] = getBuildFromGcsPath(gcsPath)         // Required for rdb-publish
	keyvals["build_target"] = buildTarget
	keyvals["parent_job_id"] = currSwarmingId
	// keyvals["fwrw_build"] = "" // NewInput?
	// keyvals["fwro_build"] = "" // NewInput?

	provisionState := &testapi.ProvisionState{

		SystemImage: &testapi.ProvisionState_SystemImage{
			SystemImagePath: &goconfig.StoragePath{
				HostType: goconfig.StoragePath_GS,
				Path:     gcsPath,
			},
		},
		ProvisionMetadata: nil,
	}

	cftTestRequest := &skylab_test_runner.CFTTestRequest{
		Deadline:         deadline,
		ParentRequestUid: parentRequestUID,
		ParentBuildId:    currBbid,
		PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
			DutModel:             dutModel,
			ProvisionState:       provisionState,
			ContainerMetadataKey: buildTarget,
		},
		CompanionDuts:                companionDuts,
		ContainerMetadata:            containerMetadata,
		TestSuites:                   testSuites,
		DefaultTestExecutionBehavior: test_platform.Request_Params_NON_CRITICAL,
		AutotestKeyvals:              keyvals,
		RunViaTrv2:                   true,
		StepsConfig:                  nil,
	}

	return cftTestRequest, nil
}

// getBuildFromGcsPath gets build from gcs path.
func getBuildFromGcsPath(gcsPath string) string {
	dirNames := strings.Split(gcsPath, "/")
	if len(dirNames) < 2 {
		return ""
	}
	return dirNames[len(dirNames)-2] + "/" + dirNames[len(dirNames)-1]
}

// createProvisionableDimensions creates provisionalbe dims.
func createProvisionableDimensions() ([]string, error) {
	dims := []string{}
	// TODO (azrahman): add support post mvp. Should get these info from provision
	// filter.
	// dimChromeOS      = "provisionable-cros-version"
	// dimFirmwareRO    = "provisionable-fwro-version"
	// dimFirmwareRW    = "provisionable-fwrw-version"
	// dimLacrosGCSPath = "provisionable-lacros-gcs-path"

	// if b := builds.ChromeOS; b != "" {
	// 	dims = append(dims, dimChromeOS+":"+b)
	// }
	// if b := builds.FirmwareRO; b != "" {
	// 	dims = append(dims, dimFirmwareRO+":"+b)
	// }
	// if b := builds.FirmwareRW; b != "" {
	// 	dims = append(dims, dimFirmwareRW+":"+b)
	// }
	// if b := builds.LacrosGCSPath; b != "" {
	// 	dims = append(dims, dimLacrosGCSPath+":"+b)

	return dims, nil
}

// createLabels creates labels.
func createLabels(hwDef *testapi.SwarmingDefinition, suiteInfo *testapi.SuiteInfo) (*inventory.SchedulableLabels, error) {
	labels := &inventory.SchedulableLabels{}

	// TODO (azrahman): Revisit this.
	// Gotta come back to this:
	// https://logs.chromium.org/logs/chromeos/led/azrahman_google.com/5553cf70b91da45971ba1857ead3fa96fed5297323cf7b084f5a07f9722ade50/+/u/ctpv2/u/step/24/log/2

	// 1. Get test.Dependencies and convert the autotest labels to dut labels
	// deps := g.Invocation.Test.Dependencies
	// flatDims := make([]string, len(deps))
	// for i, dep := range deps {
	// 	flatDims[i] = dep.Label
	// }
	// labels.Revert(flatDims)
	// Sol: 1) handle it ctpv2 via test_finder

	// 2. Add buildTarget and model (Possible from middle out response)
	// inv.Board, inv.Model
	// TODO (azrahman): Handle non chromeos type.
	board := strings.ToLower(hwDef.GetDutInfo().GetChromeos().GetDutModel().GetBuildTarget())
	model := strings.ToLower(hwDef.GetDutInfo().GetChromeos().GetDutModel().GetModelName())
	labels.Board = &board
	labels.Model = &model

	if suiteInfo.GetSuiteMetadata().GetPool() == "" || suiteInfo.GetSuiteMetadata().GetPool() == DutPoolQuota {
		labels.CriticalPools = append(labels.CriticalPools, inventory.SchedulableLabels_DUT_POOL_QUOTA)
	} else if suiteInfo.GetSuiteMetadata().GetPool() != "" {
		labels.SelfServePools = append(labels.SelfServePools, suiteInfo.GetSuiteMetadata().GetPool())
	} else {
		return nil, fmt.Errorf("no pool specified")
	}

	// TODO (azrahman): revisit this.
	// 4. Add device stability?
	// if g.Params.GetHardwareAttributes().GetRequireStableDevice() {
	// 	*inv.Stability = true
	// }

	return labels, nil
}

// createSecondaryLabels creates secondary labels.
func createSecondaryLabels() ([]*inventory.SchedulableLabels, error) {

	// TODO (azrahman): populate this for multi-dut use-case.
	// 1. Add secondary board and model

	// sds := g.Params.GetSecondaryDevices()
	// var sInvLabels []*inventory.SchedulableLabels
	// for _, sd := range sds {
	// 	il := inventory.NewSchedulableLabels()
	// 	if sd.GetSoftwareAttributes().GetBuildTarget() != nil {
	// 		*il.Board = sd.SoftwareAttributes.BuildTarget.Name
	// 	}
	// 	if sd.GetHardwareAttributes().GetModel() != "" {
	// 		*il.Model = sd.HardwareAttributes.Model
	// 	}
	// 	sInvLabels = append(sInvLabels, il)
	// }
	// return sInvLabels

	return []*inventory.SchedulableLabels{{}}, nil
}

// createSwarmingTags creates swarming tags.
func createSwarmingTags() []string {

	tags := []string{}
	// TODO(azrahman): remove this hardcoded qs account.
	tags = append(tags, "qs_account:"+"pupr")
	// tags := []string{
	// 	"luci_project:" + g.WorkerConfig.LuciProject,
	// 	"log_location:" + cmd.LogDogAnnotationURL,
	// }
	// // CTP "builds" triggered by `led` don't have a buildbucket ID.
	// if g.ParentBuildID != 0 {
	// 	tags = append(tags, "parent_buildbucket_id:"+strconv.FormatInt(g.ParentBuildID, 10))
	// }
	// tags = append(tags, "display_name:"+g.displayName(ctx, kv))
	// if qa := g.Params.GetScheduling().GetQsAccount(); qa != "" {
	// 	tags = append(tags, "qs_account:"+qa)
	// }

	// var reservedTags = map[string]bool{
	// 	"qs_account":   true,
	// 	"luci_project": true,
	// 	"log_location": true,
	// }

	// tags = append(tags, removeReservedTags(g.Params.GetDecorations().GetTags())...)
	// // Add primary/secondary DUTs board/model info in swarming tags for
	// // multi-DUTs result reporting purpose.
	// tags = append(tags, g.multiDutsTags()...)

	return tags
}

// func (g *Generator) multiDutsTags() []string {
// 	var tags []string
// 	if g.Params.GetSoftwareAttributes().GetBuildTarget() != nil {
// 		tags = append(tags, fmt.Sprintf("primary_board:%s", g.Params.SoftwareAttributes.BuildTarget.Name))
// 	}
// 	if g.Params.GetHardwareAttributes().GetModel() != "" {
// 		tags = append(tags, fmt.Sprintf("primary_model:%s", g.Params.HardwareAttributes.Model))
// 	}
// 	sds := g.Params.GetSecondaryDevices()
// 	var secondary_boards []string
// 	var secondary_models []string
// 	for _, sd := range sds {
// 		if sd.GetSoftwareAttributes().GetBuildTarget() != nil {
// 			secondary_boards = append(secondary_boards, sd.SoftwareAttributes.BuildTarget.Name)
// 		}
// 		if sd.GetHardwareAttributes().GetModel() != "" {
// 			secondary_models = append(secondary_models, sd.HardwareAttributes.Model)
// 		}
// 	}
// 	if len(secondary_boards) > 0 {
// 		boards := strings.Join(secondary_boards, ",")
// 		tags = append(tags, fmt.Sprintf("secondary_boards:%s", boards))
// 	}
// 	if len(secondary_models) > 0 {
// 		models := strings.Join(secondary_models, ",")
// 		tags = append(tags, fmt.Sprintf("secondary_models:%s", models))
// 	}
// 	return tags
// }
