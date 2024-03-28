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
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	goconfig "go.chromium.org/chromiumos/config/go"
	"go.chromium.org/chromiumos/config/go/build/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/config"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
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

type TrV2ReqHelper struct {
	// Top Level Variables
	trReqHWDef *testapi.SwarmingDefinition
	testCases  []*testapi.CTPTestCase
	suiteInfo  *testapi.SuiteInfo
	shardNum   int
	build      *build.State

	// Other fields often used several times throughout.
	suiteName        string
	pool             string
	board            string
	variant          string
	boardWVaraint    string
	currBBID         int64
	model            string
	provisionInfo    string
	analyticsName    string
	parentRequestUID string
	currSwarmingID   string
	gcsArtifactPath  string
	builderStr       string
}

// GenerateTrv2Req generates ScheduleBuildRequest.
func GenerateTrv2Req(ctx context.Context, canOutliveParent bool, trHelper *TrV2ReqHelper) (*buildbucketpb.ScheduleBuildRequest, error) {
	populateHelper(ctx, trHelper)
	err := populateHelper(ctx, trHelper)
	if err != nil {
		return nil, errors.Annotate(err, "unable to build up context: ").Err()
	}

	// Create bb request
	reqArgs, err := GenerateArgs(ctx, trHelper)
	if err != nil {
		return nil, errors.Annotate(err, "error while creating req: ").Err()
	}
	req, err := reqArgs.NewBBRequest(common.TestRunnerBuilderID())
	if err != nil {
		return nil, err
	}

	// bbCtx := lucictx.GetBuildbucket(ctx)

	// if bbCtx != nil && bbCtx.GetScheduleBuildToken() != "" && bbCtx.GetScheduleBuildToken() != buildbucket.DummyBuildbucketToken {
	// 	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(buildbucket.BuildbucketTokenHeader, bbCtx.ScheduleBuildToken))

	// 	// // Decide if the child can outlive its parent or not.
	// 	// if canOutliveParent {
	// 	// 	req.CanOutliveParent = buildbucketpb.Trinary_YES
	// 	// }
	// }

	return req, nil
}

func pool(suiteInfo *testapi.SuiteInfo) string {
	if suiteInfo.GetSuiteMetadata().GetPool() != "" {
		return suiteInfo.GetSuiteMetadata().GetPool()
	}
	return DutPoolQuota
}

func populateHelper(ctx context.Context, trHelper *TrV2ReqHelper) error {
	trHelper.suiteName = trHelper.suiteInfo.GetSuiteRequest().GetTestSuite().GetName()
	trHelper.board = strings.ToLower(getBuildTargetfromHwDef(trHelper.trReqHWDef))
	trHelper.model = strings.ToLower(getModelTargetfromHwDef(trHelper.trReqHWDef))
	trHelper.pool = pool(trHelper.suiteInfo)
	logging.Infof(ctx, "POOL FOUND: %s", trHelper.suiteInfo)
	trHelper.variant = trHelper.trReqHWDef.GetVariant()

	trHelper.boardWVaraint = trHelper.board
	if trHelper.variant != "" {
		trHelper.boardWVaraint = fmt.Sprintf("%s-%s", trHelper.board, trHelper.variant)

	}
	trHelper.currBBID = trHelper.build.Build().GetId()

	trHelper.gcsArtifactPath = findGcsPath(trHelper.suiteInfo, trHelper.board, trHelper.variant)
	if trHelper.gcsArtifactPath == "" {
		logging.Infof(ctx, "GcsPath was not found for build target: %s", trHelper.boardWVaraint)
		return fmt.Errorf("GcsPath was not found for build target: %s", trHelper.boardWVaraint)
	}

	trHelper.builderStr = getBuildFromGcsPath(trHelper.gcsArtifactPath)

	trHelper.parentRequestUID = fmt.Sprintf(CtpRequestUIDTemplate, trHelper.currBBID, trHelper.suiteName)
	trHelper.currSwarmingID = os.Getenv("SWARMING_TASK_ID")
	if trHelper.currSwarmingID == "" {
		logging.Infof(ctx, "SWARMING_TASK_ID NOT FOUND")
	}

	trHelper.analyticsName = trHelper.suiteInfo.GetSuiteRequest().GetAnalyticsName()

	return nil
}

// GenerateArgs generates args for the builder request.
func GenerateArgs(ctx context.Context, trHelper *TrV2ReqHelper) (*request.Args, error) {
	if trHelper.build == nil {
		return nil, fmt.Errorf("No Build Object set in helper.")
	}
	args := request.Args{
		Cmd:               *createCommand(ctx, trHelper.suiteName),
		SwarmingPool:      trHelper.pool,
		Dimensions:        createFreeformDims(trHelper.trReqHWDef),
		ParentTaskID:      trHelper.currSwarmingID,
		ParentRequestUID:  trHelper.parentRequestUID,
		Priority:          10,
		TestRunnerRequest: nil,            // Always nil for CFT.
		CFTIsEnabled:      true,           // Always true
		Timeout:           DefaultTimeout, // TODO (azrahman): Get this from input.
		Experiments:       trHelper.build.Build().GetInput().Experiments,
		GerritChanges:     trHelper.build.Build().GetInput().GerritChanges,
		ResultsConfig:     nil, // TODO (azrahman): Investigate if we need this.
	}

	labels, err := createLabels(trHelper)
	if err != nil {
		return nil, errors.Annotate(err, "error while creating labels: ").Err()
	}
	args.SchedulableLabels = labels

	secondaryLabels, err := createSecondaryLabels()
	if err != nil {
		return nil, errors.Annotate(err, "error while creating secondary labels: ").Err()
	}
	args.SecondaryDevicesLabels = secondaryLabels

	// TODO (azrahman): Should we even use provisionable dims for scheduling?
	provisionableDims, _ := createProvisionableDimensions()
	args.ProvisionableDimensions = provisionableDims
	args.ProvisionableDimensionExpiration = time.Minute

	provInfo := findProvisionInfo(ctx, trHelper)
	if provInfo == nil {
		return nil, fmt.Errorf("No provision info found!!")
	}

	cftTestRequest, err := createCftTestRequest(ctx, trHelper, provInfo)
	if err != nil {
		return nil, err
	}
	args.CFTTestRunnerRequest = cftTestRequest

	logging.Infof(ctx, "trhelper: %s", trHelper.currBBID)

	tags, err := createSwarmingTags(ctx, trHelper)
	if err != nil {
		return nil, errors.Annotate(err, "error while creating tags: ").Err()
	}
	args.SwarmingTags = tags

	return &args, nil
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
func createFreeformDims(TRRequesthwDef *testapi.SwarmingDefinition) []string {
	freeformDims := []string{"dut_state:ready"}
	if TRRequesthwDef.GetDutInfo().GetChromeos().GetHwid() != "" {
		freeformDims = append(freeformDims, fmt.Sprintf("hwid:%s", TRRequesthwDef.GetDutInfo().GetChromeos().GetHwid()))
	}

	for _, v := range TRRequesthwDef.GetSwarmingLabels() {
		freeformDims = append(freeformDims, formatLabel(v))
	}
	return freeformDims
}

func formatLabel(label string) string {

	if strings.HasPrefix(label, "label") || strings.HasPrefix(label, "dut_name") {
		return label
	} else {
		return fmt.Sprintf("label-%s", label)
	}
}

// findGcsPath finds gcs path for provided board.
// This is based on the given board + id; then looping through the suite metadata to find
// the target which matched these. We then will return the GCS path from there.
func findGcsPath(suiteInfo *testapi.SuiteInfo, board string, variant string) string {
	for _, suiteTarget := range suiteInfo.GetSuiteMetadata().GetTargetRequirements() {
		// This is [0] indexed because we are ignoring multi-dut today.

		suiteDef := suiteTarget.GetHwRequirements().GetHwDefinition()
		if len(suiteDef) == 0 {
			return ""
		}
		suiteHwDef := suiteDef[0]
		if getBuildTargetfromHwDef(suiteHwDef) == board && suiteHwDef.GetVariant() == variant {
			provInfos := suiteHwDef.GetProvisionInfo()
			for _, provInfo := range provInfos {
				if provInfo.GetType() == testapi.ProvisionInfo_CROS {
					return provInfo.GetInstallRequest().GetImagePath().GetPath()
				}
			}
		}
	}
	return ""

}

func findProvisionInfo(ctx context.Context, trHelper *TrV2ReqHelper) []*testapi.ProvisionInfo {
	logging.Infof(ctx, "looking for provision info for board: %s, variant: %s", trHelper.board, trHelper.variant)
	logging.Infof(ctx, "looking for provision info for suiteMD: %s", trHelper.suiteInfo.GetSuiteMetadata())

	for _, suiteTarget := range trHelper.suiteInfo.GetSuiteMetadata().GetTargetRequirements() {
		// This is [0] indexed because we are ignoring multi-dut today.

		suiteDef := suiteTarget.GetHwRequirements().GetHwDefinition()
		if len(suiteDef) == 0 {
			return nil
		}

		suiteHwDef := suiteDef[0]
		logging.Infof(ctx, "looking for provision info for suiteInfo: %s", getBuildTargetfromHwDef(suiteHwDef))

		suiteSwDef := suiteTarget.GetSwRequirement()
		logging.Infof(ctx, "looking for provision info for suiteInfo: %s", suiteSwDef)

		if getBuildTargetfromHwDef(suiteHwDef) == trHelper.board && suiteHwDef.GetVariant() == trHelper.variant {
			return suiteHwDef.GetProvisionInfo()
		}
	}
	return nil

}

func getBuildTargetfromHwDef(TRRequesthwDef *testapi.SwarmingDefinition) string {
	return TRRequesthwDef.GetDutInfo().GetChromeos().GetDutModel().GetBuildTarget()
}

func getModelTargetfromHwDef(TRRequesthwDef *testapi.SwarmingDefinition) string {
	return TRRequesthwDef.GetDutInfo().GetChromeos().GetDutModel().GetModelName()
}
func getBuildTargetWVariantfromHwDef(TRRequesthwDef *testapi.SwarmingDefinition) string {
	if TRRequesthwDef.GetVariant() == "" {
		return getBuildTargetfromHwDef(TRRequesthwDef)
	}
	return fmt.Sprintf("%s-%s", getBuildTargetfromHwDef(TRRequesthwDef), TRRequesthwDef.GetVariant())
}

// createCftTestRequest creates cft test request.
func createCftTestRequest(ctx context.Context, trHelper *TrV2ReqHelper, provInfo []*testapi.ProvisionInfo) (*skylab_test_runner.CFTTestRequest, error) {
	deadline := timestamppb.New(time.Now().Add(19 * time.Hour))

	dutModel := &labapi.DutModel{
		BuildTarget: trHelper.board,
		ModelName:   trHelper.model,
	}

	containerGcsPath := trHelper.gcsArtifactPath + common.ContainerMetadataPath

	tempRootDir := os.Getenv("TEMPDIR")

	// Just here to prevent race conditions of shards fighting over a file.
	rand.Seed(time.Now().UnixNano())
	tempRootDir = path.Join(tempRootDir, strconv.Itoa(rand.Int()))

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
	for _, testCase := range trHelper.testCases {
		testCaseIds = append(testCaseIds, testCase.GetMetadata().GetTestCase().GetId())
	}
	testSuites := []*testapi.TestSuite{
		{
			Name: trHelper.suiteName,
			Spec: &testapi.TestSuite_TestCaseIds{
				TestCaseIds: &testapi.TestCaseIdList{
					TestCaseIds: testCaseIds,
				},
			},
			ExecutionMetadata: trHelper.suiteInfo.GetSuiteMetadata().GetExecutionMetadata(),
		},
	}

	keyvals := make(map[string]string)
	keyvals["suite"] = trHelper.suiteName // suite name

	// TODO (dbeckett) we need the int of the shard # passed into the gofunc.
	keyvals["label"] = fmt.Sprintf("%s-shard-0", trHelper.suiteName) // test name
	keyvals["build"] = trHelper.builderStr                           // Required for rdb-publish
	keyvals["build_target"] = trHelper.board
	keyvals["parent_job_id"] = trHelper.currSwarmingID

	provisionState, err := buildProvisionState(provInfo)
	if err != nil {
		return nil, err
	}

	cftTestRequest := &skylab_test_runner.CFTTestRequest{
		Deadline:         deadline,
		ParentRequestUid: trHelper.parentRequestUID,
		ParentBuildId:    trHelper.currBBID,
		PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
			DutModel:             dutModel,
			ProvisionState:       provisionState,
			ContainerMetadataKey: trHelper.boardWVaraint,
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
func createLabels(trHelper *TrV2ReqHelper) (*inventory.SchedulableLabels, error) {
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
	// inv., inv.Model
	// TODO (azrahman): Handle non chromeos type.

	labels.Board = &trHelper.board
	labels.Model = &trHelper.model

	if trHelper.pool == "" || trHelper.pool == DutPoolQuota {
		labels.CriticalPools = append(labels.CriticalPools, inventory.SchedulableLabels_DUT_POOL_QUOTA)
	} else if trHelper.pool != "" {
		labels.SelfServePools = append(labels.SelfServePools, trHelper.pool)
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
func createSwarmingTags(ctx context.Context, trHelper *TrV2ReqHelper) ([]string, error) {
	tags := []string{}

	qsAccount := trHelper.suiteInfo.GetSuiteMetadata().GetSchedulerInfo().GetQsAccount()
	if qsAccount == "" {
		qsAccount = "unmanaged_p2"
		logging.Infof(ctx, "no qsAccount given, defaulting to unmanaged_p2.")
	}
	tags = append(tags, "qs_account:"+qsAccount)

	tags = append(tags, "label-board:"+trHelper.boardWVaraint)
	if trHelper.model != "" {
		tags = append(tags, "label-model:"+trHelper.model)
	}

	tags = append(tags, "label-pool:"+trHelper.pool)

	if trHelper.suiteName != "" {
		tags = append(tags, "label-suite:"+trHelper.suiteName)
	}

	if trHelper.currSwarmingID != "" {
		tags = append(tags, "parent_task_id:"+trHelper.currSwarmingID)
	}

	// TODO should we un-hardcode this?
	tags = append(tags, "luci_project:"+"chromeos")

	if trHelper.analyticsName != "" {
		tags = append(tags, "analytics_name:"+trHelper.analyticsName)
		tags = append(tags, "ctp-fwd-task-name:"+trHelper.analyticsName)
	}

	tags = append(tags, "build:"+trHelper.builderStr)

	if trHelper.currBBID != 0 {
		tags = append(tags, fmt.Sprintf("parent_buildbucket_id:%v", trHelper.currBBID))
	} else {
		tags = append(tags, "parent_buildbucket_id:0")
	}

	// TODO(dbeckett) THESE BELOW:
	reprName := fmt.Sprintf("shard-%v", trHelper.shardNum)
	tags = append(tags, "display_name:"+makeDisplayName(trHelper.builderStr, trHelper.suiteName, reprName))
	ll := getLogLocation()
	if ll != "" {
		tags = append(tags, "log_location:"+ll)
	}
	// tags = append(tags, removeReservedTags(g.Params.GetDecorations().GetTags())...)
	// // Add primary/secondary DUTs board/model info in swarming tags for
	// // multi-DUTs result reporting purpose.
	// tags = append(tags, g.multiDutsTags()...)

	return tags, nil

}

// TODO (azrahaman)
func getLogLocation() string {
	return ""
}

func makeDisplayName(buildStr string, suite string, TRName string) string {
	return fmt.Sprintf("%s/%s/%s", buildStr, suite, TRName)
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

// This method is currently incomplete, its basically just taking the gcs path from the given info.
// it will migrate to the dynamic TRv2 stuff in the near future.
func buildProvisionState(provInfo []*testapi.ProvisionInfo) (*testapi.ProvisionState, error) {
	if len(provInfo) == 0 {
		return nil, fmt.Errorf("No Provision Info items given")
	}
	gcsPath := provInfo[0].GetInstallRequest().GetImagePath().GetPath()
	if gcsPath == "" {
		return nil, fmt.Errorf("No gcs path found found")
	}

	provisionState := &testapi.ProvisionState{

		SystemImage: &testapi.ProvisionState_SystemImage{
			SystemImagePath: &goconfig.StoragePath{
				HostType: goconfig.StoragePath_GS,
				Path:     gcsPath,
			},
		},
		ProvisionMetadata: nil,
	}
	return provisionState, nil
}

func suiteName(suiteInfo *testapi.SuiteInfo) string {
	return suiteInfo.GetSuiteRequest().GetTestSuite().GetName()
}
