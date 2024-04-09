// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/proto"
	protobuf "google.golang.org/protobuf/proto"

	"infra/cros/cmd/common_lib/analytics"
	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/common_lib/schedulers"
	"infra/cros/cmd/ctpv2/data"
)

const (
	TestStepNameTemplate = "request %s-%s.hw.%s-shard-%v"
)

// getBuildFieldMask is the list of buildbucket fields that are needed.
var getBuildFieldMask = []string{
	"id",
	"infra.backend.task.id.id",
	"infra.swarming.task_id",
	// Build details are parsed from the build's output properties.
	"output.properties",
	// Build status is used to determine whether the build is complete.
	"status",
}

// ScheduleTasksCmd represents scheduling task(s) cmd.
type ScheduleTasksCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Deps
	MiddledOutResp *data.MiddleOutResponse
	BuildState     *build.State
	Scheduler      interfaces.SchedulerInterface
	DynamicRun     bool

	// Deps
	InternalTestPlan *api.InternalTestplan

	// Updates
	TestResults map[string]*data.TestResults

	// For logging
	BQClient *bigquery.Client
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *ScheduleTasksCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.FilterStateKeeper:
		err = cmd.extractDepsFromFilterStateKeeper(ctx, sk)

	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during extracting dependencies for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *ScheduleTasksCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.FilterStateKeeper:
		err = cmd.updateScheduleStateKeeper(ctx, sk)
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

func (cmd *ScheduleTasksCmd) extractDepsFromFilterStateKeeper(
	ctx context.Context,
	sk *data.FilterStateKeeper) error {

	if sk.MiddledOutResp == nil {
		return fmt.Errorf("Cmd %q missing dependency: MiddleOutResponse", cmd.GetCommandType())
	}

	if sk.BuildState == nil {
		return fmt.Errorf("Cmd %q missing dependency: Scheduler", cmd.GetCommandType())
	}

	if sk.Scheduler == api.SchedulerInfo_UNSPECIFIED {
		return fmt.Errorf("Cmd %q missing dependency: Scheduler", cmd.GetCommandType())
	}

	if sk.BQClient != nil {
		cmd.BQClient = sk.BQClient
	}
	cmd.InternalTestPlan = proto.Clone(sk.TestPlanStates[len(sk.TestPlanStates)-1]).(*api.InternalTestplan)

	if sk.CtpReq == nil {
		return fmt.Errorf("Cmd %q missing dependency: CtpReq", cmd.GetCommandType())
	}

	cmd.DynamicRun = sk.CtpReq.RunDynamic
	cmd.MiddledOutResp = sk.MiddledOutResp
	cmd.BuildState = sk.BuildState
	// Assign scheduler
	if sk.Scheduler == api.SchedulerInfo_QSCHEDULER {
		cmd.Scheduler = schedulers.NewDirectBBScheduler()
	} else if sk.Scheduler == api.SchedulerInfo_PRINT_REQUEST_ONLY {
		cmd.Scheduler = schedulers.NewLocalScheduler()
	} else if sk.Scheduler == api.SchedulerInfo_SCHEDUKE {
		cmd.Scheduler = schedulers.NewSchedukeScheduler()
	}

	return nil
}

func (cmd *ScheduleTasksCmd) updateScheduleStateKeeper(ctx context.Context, sk *data.FilterStateKeeper) error {
	if cmd.TestResults != nil && len(cmd.TestResults) != 0 {
		sk.SuiteTestResults = cmd.TestResults
	}
	cmd.InternalTestPlan = proto.Clone(sk.TestPlanStates[len(sk.TestPlanStates)-1]).(*api.InternalTestplan)
	return nil
}

// Execute executes the command.
func (cmd *ScheduleTasksCmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Schedule tasks")
	defer func() { step.End(err) }()

	key := "scheduleTasks"
	if cmd.BQClient != nil {
		analytics.SoftInsertStepWInternalPlan(ctx, cmd.BQClient, &analytics.BqData{Step: key, Status: analytics.Start}, cmd.InternalTestPlan, cmd.BuildState)
	}
	start := time.Now()

	enumStatus := analytics.Success
	cmd.TestResults = map[string]*data.TestResults{}
	suiteName := suiteName(cmd.MiddledOutResp.SuiteInfo)
	if len(cmd.MiddledOutResp.TrReqs) == 0 {
		logging.Infof(ctx, "no test found in middle-out response")
		enumStatus = analytics.Fail
		step.SetSummaryMarkdown("enumeration error: no test found")
		err = &data.EnumerationError{SuiteName: suiteName}
		cmd.TestResults[common.EnumerationErrKey] = &data.TestResults{Suite: suiteName, Key: common.EnumerationErrKey, TopLevelError: &err}
		common.WriteAnyObjectToStepLog(ctx, step, cmd.TestResults, "consolidated results")
		return err
	}

	// GenerateRequests
	buildMap := GenerateRequests(ctx, cmd.MiddledOutResp, cmd.BuildState, cmd.BQClient, cmd.DynamicRun)
	if len(buildMap) == 0 {
		enumStatus = analytics.Fail
		step.SetSummaryMarkdown("enumeration error: no valid test found")
		err = &data.EnumerationError{SuiteName: suiteName}
		cmd.TestResults[common.EnumerationErrKey] = &data.TestResults{Suite: suiteName, Key: common.EnumerationErrKey, TopLevelError: &err}
		return err
	}

	bqData := &analytics.BqData{Step: "enumeration", Status: enumStatus, Duration: float32(time.Since(start).Seconds())}
	analytics.SoftInsertStepWInternalPlan(ctx, cmd.BQClient, bqData, cmd.InternalTestPlan, cmd.BuildState)

	schedulerStatus := analytics.Success
	schedulerFreeform := ""

	scheduler := cmd.Scheduler
	if cmd.Scheduler == nil {
		errmsg := "empty scheduler"

		schedulerFreeform = errmsg
		schedulerStatus = analytics.Fail

		logging.Infof(ctx, errmsg)
		return fmt.Errorf(errmsg)
	}
	pool := pool(cmd.MiddledOutResp.SuiteInfo)
	err = scheduler.Setup(pool)
	if err != nil {
		errmsg := "error while setting up scheduler"

		schedulerStatus = analytics.Fail
		schedulerFreeform = errmsg

		logging.Infof(ctx, "%s: %s", errmsg, err)
		return errors.Annotate(err, errmsg).Err()
	}

	bqData = &analytics.BqData{Step: "scheduler-setup", Status: schedulerStatus, Duration: float32(time.Since(start).Seconds())}
	if schedulerFreeform != "" {
		bqData.Freeform = schedulerFreeform
	}
	analytics.SoftInsertStepWInternalPlan(ctx, cmd.BQClient, bqData, cmd.InternalTestPlan, cmd.BuildState)

	// Todo: batch call
	resultsChan := make(chan *data.TestResults)
	wg := &sync.WaitGroup{}
	for k, v := range buildMap {
		wg.Add(1)
		go ScheduleAndMonitor(ctx, cmd.Scheduler, v, cmd.BuildState, k, wg, resultsChan, suiteName, 0, cmd.BQClient, cmd.DynamicRun)
	}

	go func() {
		wg.Wait()
		close(resultsChan) // Close the channel when all workers are done
	}()

	// Read results
	for result := range resultsChan {
		mapKey := result.Key
		if result.Attempt != 0 {
			mapKey = fmt.Sprintf("%s-retry-%d", result.Key, result.Attempt)
		}
		cmd.TestResults[mapKey] = result
	}

	analytics.SoftInsertStepWInternalPlan(ctx, cmd.BQClient, &analytics.BqData{Step: key, Status: analytics.Success, Duration: float32(time.Since(start).Seconds())}, cmd.InternalTestPlan, cmd.BuildState)
	common.WriteAnyObjectToStepLog(ctx, step, cmd.TestResults, "consolidated results")
	return nil

}

type BuildRequest struct {
	Key                  string
	shardNum             int
	ScheduleBuildRequest *buildbucketpb.ScheduleBuildRequest
	OriginalTrReq        *data.TrRequest
	SuiteInfo            *api.SuiteInfo
	err                  error
}

func GenerateRequests(ctx context.Context, moResp *data.MiddleOutResponse, buildState *build.State, BQClient *bigquery.Client, dynamicRun bool) map[string]*BuildRequest {
	var err error
	step, ctx := build.StartStep(ctx, "Generate Trv2 Requests")
	defer func() { step.End(err) }()

	// Generate reqs first
	errCount := 0
	buildMap := map[string]*BuildRequest{}
	shardMap := map[string]int{}
	for _, trReq := range moResp.TrReqs {
		key, err := GetBoardModelVariantKey(ctx, trReq)
		if err != nil {
			logging.Infof(ctx, fmt.Sprintf("error while generating board-model-variant key: %s", err))
			errCount++
			continue
		}
		// check if the tcs were sharded
		if _, ok := shardMap[key]; !ok {
			shardMap[key] = 0
		} else {
			shardMap[key] = shardMap[key] + 1
		}

		modifiedKey := fmt.Sprintf("%s-shard-%d", key, shardMap[key])
		buildReq := &BuildRequest{Key: modifiedKey, OriginalTrReq: trReq, shardNum: shardMap[key], SuiteInfo: moResp.SuiteInfo}
		req, err := GenerateReq(ctx, trReq, modifiedKey, buildState, moResp.SuiteInfo, shardMap[key], BQClient, dynamicRun)
		if err != nil {
			buildReq.err = err
			errCount++
			logging.Infof(ctx, "error while generating trv2 req for %s: %s", modifiedKey, err)
		} else {
			buildReq.ScheduleBuildRequest = req
		}

		buildMap[modifiedKey] = buildReq
	}

	if errCount == 0 {
		step.SetSummaryMarkdown("all test requests were generated successfully")
	} else {
		step.SetSummaryMarkdown(fmt.Sprintf("error found in %d out of %d requests", errCount, len(moResp.TrReqs)))
		err = fmt.Errorf("error found in %d out of %d", errCount, len(moResp.TrReqs))
	}

	return buildMap
}

func GetBoardModelVariantKey(ctx context.Context, trReq *data.TrRequest) (string, error) {
	if trReq.Req == nil || len(trReq.Req.GetHwDefinition()) == 0 {
		// This should not happen. If this happens, we should have flagged input
		// error earlier. Still kept this as a sanity check.
		logging.Infof(ctx, "no hw def is found in req")
		return "", fmt.Errorf("no hw def is found so, rejecting task")
	}

	// '0'ed index because we should always have one hw here. It supports multiple
	// MO should reduce it down to 1 always. The len check is done at MO step.
	TrReqhwDef := trReq.Req.GetHwDefinition()[0]
	board := strings.ToLower(getBuildTargetfromHwDef(TrReqhwDef))
	variant := strings.ToLower(TrReqhwDef.GetVariant())
	model := strings.ToLower(getModelTargetfromHwDef(TrReqhwDef))

	builderString := board
	if model != "" {
		builderString = fmt.Sprintf("%s-%s", builderString, model)
	}
	if variant != "" {
		builderString = fmt.Sprintf("%s-%s", builderString, variant)
	}

	return builderString, nil
}

func GenerateReq(ctx context.Context, trReq *data.TrRequest, key string, buildState *build.State, suiteInfo *api.SuiteInfo, shardNum int, BQClient *bigquery.Client, dynamicRun bool) (*buildbucketpb.ScheduleBuildRequest, error) {
	var err error
	// '0'ed index because we should always have one hw here. It supports multiple
	// MO should reduce it down to 1 always. The len check is done at MO step.
	TrReqhwDef := trReq.Req.GetHwDefinition()[0]
	testCases := trReq.Tcs
	st := time.Now()

	d := logTrAnalyticsStart(ctx, BQClient, trReq, suiteInfo, buildState, "generateReq", key, suiteInfo.GetSuiteRequest().GetAnalyticsName())

	// Input validations
	if len(testCases) == 0 {
		errStr := "no test is found so, rejecting task"
		logging.Infof(ctx, errStr)
		err = fmt.Errorf(errStr)
		logTrAnalyticsFail(ctx, BQClient, d, trReq, suiteInfo, buildState, errStr, st)
		return nil, err

	}

	if trReq.DevicesInfo.LabDevicesCount == 0 {
		logging.Infof(ctx, "no suitable device found to run tests so, rejecting task")

		logTrAnalyticsFail(ctx, BQClient, d, trReq, suiteInfo, buildState, "rejected, no bots found", st)
		err := &data.BotParamsRejectedError{Key: key, RejectedDims: trReq.DevicesInfo.Dims}
		return nil, err
	}

	helper := &TrV2ReqHelper{
		trReqHWDef: TrReqhwDef,
		testCases:  testCases,
		build:      buildState,
		suiteInfo:  suiteInfo,
		shardNum:   shardNum,
		dynamicRun: dynamicRun,
	}

	req, err := GenerateTrv2Req(ctx, true, helper)
	if err != nil {
		logging.Infof(ctx, "error while generating req: %s", err)
		logTrAnalyticsFail(ctx, BQClient, d, trReq, suiteInfo, buildState, "unable to build task", st)

		err := &data.BotParamsRejectedError{Key: key, RejectedDims: trReq.DevicesInfo.Dims}
		return nil, errors.Annotate(err, "error while generating req:").Err()
	}

	logTrAnalyticsSuccess(ctx, BQClient, d, trReq, suiteInfo, buildState, "", st)

	return req, nil
}

func ScheduleAndMonitor(rootCtx context.Context, scheduler interfaces.SchedulerInterface, buildReq *BuildRequest, buildState *build.State, key string, wg *sync.WaitGroup, resultsChan chan<- *data.TestResults, suiteName string, retryNum int, BQClient *bigquery.Client, dynamicRun bool) error {
	defer wg.Done()
	var err error

	stepName := key
	if retryNum > 0 {
		stepName = fmt.Sprintf("%s-retry-%d", key, retryNum)
	}
	step, ctx := build.StartStep(rootCtx, stepName)
	defer func() { step.End(err) }()

	// Construct test results
	result := &data.TestResults{Key: key, Suite: suiteName, Attempt: retryNum}

	if buildReq.err != nil {
		err = buildReq.err
		return setTopLevelError(ctx, step, result, resultsChan, buildReq.err)
	}
	req := buildReq.ScheduleBuildRequest

	// Spit out the request
	requestData, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		logging.Infof(
			ctx,
			"error during writing request data to log: %s",
			err.Error())
	}
	step.Log("BB Request").Write(requestData)

	builderId := common.TestRunnerBuilderID()

	bbClient, err := newBBClient(ctx)
	if err != nil {
		return err
	}

	st := time.Now()

	d := logTrAnalyticsStart(ctx, BQClient, buildReq.OriginalTrReq, buildReq.SuiteInfo, buildState, "ScheduleBuild", buildReq.Key, buildReq.SuiteInfo.GetSuiteRequest().GetAnalyticsName())

	// BQ TODO log the request is in the scheduling tool (ie log the scheduke ID if possible?)
	scheduledBuild, err := scheduler.ScheduleRequest(ctx, req, step)
	if err != nil {
		err = fmt.Errorf("error while scheduling req: %s", err)
		logTrAnalyticsFail(ctx, BQClient, d, buildReq.OriginalTrReq, buildReq.SuiteInfo, buildState, "error scheduling", st)
		return setTopLevelError(ctx, step, result, resultsChan, err)
	}

	if scheduledBuild != nil && scheduledBuild.GetId() != 0 {
		result.BuildUrl = common.BBUrl(builderId, scheduledBuild.GetId())
		step.SetSummaryMarkdown(fmt.Sprintf("[latest attempt](%s)", common.BBUrl(builderId, scheduledBuild.GetId())))
	} else {
		errStr := "no bbid found from scheduler"
		err = fmt.Errorf(errStr)
		logTrAnalyticsFail(ctx, BQClient, d, buildReq.OriginalTrReq, buildReq.SuiteInfo, buildState, errStr, st)

		return setTopLevelError(ctx, step, result, resultsChan, err)
	}
	// Log the successful start.
	d.TrTaskID = fmt.Sprint(scheduledBuild.GetId())
	logTrAnalyticsSuccess(ctx, BQClient, d, buildReq.OriginalTrReq, buildReq.SuiteInfo, buildState, "", st)

	// Metrics for the task running
	st = time.Now()

	// Re-init the data for the run build step. Keep the previously populated data.
	d.Step = "Run Build"
	d.Status = analytics.Start
	d.Duration = 0

	analytics.SoftInsertStepWTrReq(ctx, BQClient, d, buildReq.OriginalTrReq, buildReq.SuiteInfo, buildState)

	// Monitor here
	loopSleepInterval := 30 * time.Second
	statusReq := &buildbucketpb.GetBuildStatusRequest{
		Id: scheduledBuild.GetId(),
	}
	for {

		buildInfo, err := CheckBuildInfoIfBuildEnded(ctx, statusReq, bbClient)
		if err != nil || buildInfo == nil {
			// this means the build didn't end
			if err != nil {
				logging.Infof(ctx, "error while checking build status: %s", err)
			}

			if ctx.Err() != nil {
				// A timeout while waiting for tests to complete is reported as
				// aborts when summarizing individual tests' results.
				// The execute step completes without errors.
				return nil
			}

			time.Sleep(loopSleepInterval)

			// we don't wanna fail coz it could be a flake so we continue checking
			continue
		}

		// The build ended so we extract results now
		common.WriteAnyObjectToStepLog(ctx, step, buildInfo, "final build info")

		// Log success as we found a completed build. The status is not of the child build itself.
		logTrAnalyticsSuccess(ctx, BQClient, d, buildReq.OriginalTrReq, buildReq.SuiteInfo, buildState, fmt.Sprint("build status: ", buildInfo.GetStatus()), st)

		logging.Infof(ctx, "bb status: %s", buildInfo.GetStatus())
		if buildInfo.GetStatus() != buildbucketpb.Status_SUCCESS {
			// setting this for the step to fail
			err = fmt.Errorf("test_runner failed")
		}

		trResult, err := extractResult(buildInfo)
		if err != nil {
			err = fmt.Errorf("error while extracting results from test_runner build %d: %s", buildInfo.Id, err)
			return setTopLevelError(ctx, step, result, resultsChan, err)
		} else {
			result.Results = trResult
		}
		common.WriteAnyObjectToStepLog(ctx, step, result, "extracted result from trv2")

		// Retry if qualifies
		if buildReq.SuiteInfo.GetSuiteRequest().GetRetryCount() > int64(retryNum) {
			newBuildReq := RetryReqIfQualifies(ctx, trResult, step, buildReq, buildState, nil, dynamicRun)
			if newBuildReq != nil && newBuildReq.ScheduleBuildRequest != nil {
				// Schedule retry
				wg.Add(1)
				go ScheduleAndMonitor(rootCtx, scheduler, newBuildReq, buildState, newBuildReq.Key, wg, resultsChan, suiteName, retryNum+1, BQClient, dynamicRun)
			}
		}

		// Send the result via channel
		resultsChan <- result
		break
	}

	return nil
}

func RetryReqIfQualifies(ctx context.Context, trResult *skylab_test_runner.Result, step *build.Step, buildReq *BuildRequest, buildState *build.State, bqClient *bigquery.Client, dynamicRun bool) *BuildRequest {
	retriableTestsMap := determineRetriablity(trResult)
	if len(retriableTestsMap) == 0 {
		logging.Infof(ctx, "no retriable tests found for: %s", buildReq.Key)
		return nil
	}

	logging.Infof(ctx, "This req qualified for retry...")
	common.WriteAnyObjectToStepLog(ctx, step, retriableTestsMap, "retriable test list")

	newBuildReq := GenerateNewBuildReqForRetry(ctx, buildReq, retriableTestsMap)
	common.WriteAnyObjectToStepLog(ctx, step, newBuildReq, "new build req for retry")

	// Generate a new req
	req, err := GenerateReq(ctx, newBuildReq.OriginalTrReq, newBuildReq.Key, buildState, newBuildReq.SuiteInfo, newBuildReq.shardNum, bqClient, dynamicRun)
	if err != nil {
		newBuildReq.err = err
		logging.Infof(ctx, "no more retry will take place for %s since trv2 req generation failed: %s", newBuildReq.Key, err)
	} else {
		newBuildReq.ScheduleBuildRequest = req
	}

	return newBuildReq
}

func CheckBuildInfoIfBuildEnded(ctx context.Context, statusReq *buildbucketpb.GetBuildStatusRequest, bbClient buildbucketpb.BuildsClient) (*buildbucketpb.Build, error) {
	// Check Build status.
	b, err := bbClient.GetBuildStatus(ctx, statusReq)
	if err != nil {
		logging.Infof(ctx, "error while getting build status: %s", err)
		return nil, err
	}

	if b == nil || int(b.GetStatus())&int(buildbucketpb.Status_ENDED_MASK) == 0 {
		return nil, nil
	}

	// Get more build info
	req := &buildbucketpb.GetBuildRequest{
		Id:   b.Id,
		Mask: &buildbucketpb.BuildMask{Fields: &field_mask.FieldMask{Paths: getBuildFieldMask}},
	}
	buildInfo, err := bbClient.GetBuild(ctx, req)
	if err != nil {
		logging.Infof(ctx, "error while getting build info: %s", err)
		return nil, err
	}

	return buildInfo, nil
}

func GenerateNewBuildReqForRetry(ctx context.Context, buildReq *BuildRequest, retriableTests map[string]bool) *BuildRequest {
	if len(retriableTests) == 0 {
		return nil
	}

	// dereference so that we can make changes
	retryBuildReq := *buildReq
	retryBuildReq.ScheduleBuildRequest = nil
	retryBuildReq.err = nil

	testCases := retryBuildReq.OriginalTrReq.Tcs
	newTcs := []*api.CTPTestCase{}
	for _, tc := range testCases {
		// append the retriable tests and ignore others
		if _, ok := retriableTests[tc.GetName()]; ok {
			newTcs = append(newTcs, tc)
		}
	}

	retryBuildReq.OriginalTrReq.Tcs = newTcs
	return &retryBuildReq
}

func determineRetriablity(trResult *skylab_test_runner.Result) map[string]bool {
	retriableTests := map[string]bool{}
	// First check if there are valid trResult
	resultsMap := trResult.GetAutotestResults()
	if len(resultsMap) > 0 {
		testCases := resultsMap["original_test"].GetTestCases()
		for _, testCase := range testCases {
			if IsTcRetriable(testCase.GetVerdict()) {
				retriableTests[testCase.GetName()] = true
			}
		}
	}

	return retriableTests
}

// IsTcRetriable determines if a task result indicates that the test needs to
// be retried.
//
// Panics on unknown verdicts.
func IsTcRetriable(verdict skylab_test_runner.Result_Autotest_TestCase_Verdict) bool {
	switch verdict {
	case skylab_test_runner.Result_Autotest_TestCase_VERDICT_FAIL,
		skylab_test_runner.Result_Autotest_TestCase_VERDICT_ERROR,
		skylab_test_runner.Result_Autotest_TestCase_VERDICT_ABORT:
		return true
	case skylab_test_runner.Result_Autotest_TestCase_VERDICT_NO_VERDICT,
		skylab_test_runner.Result_Autotest_TestCase_VERDICT_PASS:
		return false
	default:
		panic(fmt.Sprintf("IsTcRetriable: unknown verdict %s", verdict.String()))
	}
}

func setTopLevelError(ctx context.Context, step *build.Step, result *data.TestResults, resultsChan chan<- *data.TestResults, err error) error {
	logging.Infof(ctx, err.Error())
	step.SetSummaryMarkdown(err.Error())
	result.TopLevelError = &err

	// Send the result via channel
	resultsChan <- result
	return err
}

func extractResult(from *buildbucketpb.Build) (*skylab_test_runner.Result, error) {
	op := from.GetOutput().GetProperties().GetFields()
	if op == nil {
		return nil, fmt.Errorf("output props is empty")
	}
	cr := op["compressed_result"].GetStringValue()
	if cr == "" {
		return nil, fmt.Errorf("compressed_result is empty")
	}
	pb, err := common.Decompress(cr)
	if err != nil {
		return nil, errors.Annotate(err, "extract results from build %d", from.Id).Err()
	}
	var r skylab_test_runner.Result
	if err := protobuf.Unmarshal(pb, &r); err != nil {
		return nil, errors.Annotate(err, "extract results from build %d", from.Id).Err()
	}
	return &r, nil
}

func NewScheduleTasksCmd() *ScheduleTasksCmd {
	abstractCmd := interfaces.NewAbstractCmd(ScheduleTasksCmdType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &ScheduleTasksCmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}

// FindBuildName finds build name from suite info.
func FindBuildName(suiteInfo *api.SuiteInfo, board string) string {
	for _, target := range suiteInfo.GetSuiteMetadata().GetTargetRequirements() {
		// [0] is bc multi-dut
		hwDef := target.GetHwRequirements().GetHwDefinition()[0]
		if hwDef.GetDutInfo().GetChromeos().GetDutModel().GetBuildTarget() == board {
			return target.GetSwRequirement().GetBuild()
		}
	}

	return ""
}

func logTrAnalyticsFail(ctx context.Context, BQClient *bigquery.Client, data *analytics.TaskData, req *data.TrRequest, suiteInfo *api.SuiteInfo, build *build.State, err string, st time.Time) {
	data.Duration = float32(time.Since(st).Seconds())

	data.Status = analytics.Fail
	data.Freeform = err
	analytics.SoftInsertStepWTrReq(ctx, BQClient, data, req, suiteInfo, build)
}

func logTrAnalyticsSuccess(ctx context.Context, BQClient *bigquery.Client, data *analytics.TaskData, req *data.TrRequest, suiteInfo *api.SuiteInfo, build *build.State, freeform string, st time.Time) {
	data.Duration = float32(time.Since(st).Seconds())

	data.Status = analytics.Success
	if freeform != "" {
		data.Freeform = freeform
	}
	analytics.SoftInsertStepWTrReq(ctx, BQClient, data, req, suiteInfo, build)
}

func logTrAnalyticsStart(ctx context.Context, BQClient *bigquery.Client, req *data.TrRequest, suiteInfo *api.SuiteInfo, build *build.State, step string, key string, analyticsName string) *analytics.TaskData {
	data := &analytics.TaskData{
		Step:          step,
		DisplayName:   key,
		AnalyticsName: analyticsName,
		Status:        analytics.Start,
	}
	analytics.SoftInsertStepWTrReq(ctx, BQClient, data, req, suiteInfo, build)
	return data
}
