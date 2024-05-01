// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/proto"
	protobuf "google.golang.org/protobuf/proto"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

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
	BuildState *build.State
	Scheduler  interfaces.SchedulerInterface
	DynamicRun bool

	// Deps
	InternalTestPlan *api.InternalTestplan
	BuildsMap        map[string]*data.BuildRequest

	// Updates
	TestResults map[string]*data.TestResults

	// For logging
	BQClient              *bigquery.Client
	StartCmdTime          time.Time
	StartTrSchedulingTime time.Time
	StartTrBuildTime      time.Time
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

	if sk.BuildsMap == nil || len(sk.BuildsMap) == 0 {
		return fmt.Errorf("Cmd %q missing dependency: BuildsMap", cmd.GetCommandType())
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
	cmd.BuildsMap = sk.BuildsMap
	cmd.BuildState = sk.BuildState
	// Assign scheduler
	cmd.Scheduler = schedulers.NewLocalScheduler() // Default
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

	cmd.ObserveCmdStart(ctx)
	cmd.TestResults = map[string]*data.TestResults{}

	scheduler := cmd.Scheduler
	pool := pool(cmd.InternalTestPlan.SuiteInfo)
	err = scheduler.Setup(pool)
	if err != nil {
		errmsg := "error while setting up scheduler"
		cmd.ObserveSchedulerSetupFailure(ctx, errmsg)
		logging.Infof(ctx, "%s: %s", errmsg, err)
		return errors.Annotate(err, errmsg).Err()
	}
	cmd.ObserveSchedulerSetupSuccess(ctx)

	// Todo: batch call
	resultsChan := make(chan *data.TestResults)
	wg := &sync.WaitGroup{}
	for k, v := range cmd.BuildsMap {
		wg.Add(1)
		go cmd.ScheduleAndMonitor(ctx, k, v, wg, resultsChan, 0)
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

	cmd.ObserveCmdEndSuccess(ctx)
	common.WriteAnyObjectToStepLog(ctx, step, cmd.TestResults, "consolidated results")
	return nil

}

func (cmd *ScheduleTasksCmd) ScheduleAndMonitor(rootCtx context.Context, key string, buildReq *data.BuildRequest, wg *sync.WaitGroup, resultsChan chan<- *data.TestResults, retryNum int) error {
	defer wg.Done()
	var err error

	suiteName := suiteName(buildReq.SuiteInfo)
	stepName := key
	if retryNum > 0 {
		stepName = fmt.Sprintf("%s-retry-%d", key, retryNum)
	}
	step, ctx := build.StartStep(rootCtx, stepName)
	defer func() { step.End(err) }()

	// Construct test results
	result := &data.TestResults{Key: key, Suite: suiteName, Attempt: retryNum}

	if buildReq.Err != nil {
		err = buildReq.Err
		return setTopLevelError(ctx, step, result, resultsChan, buildReq.Err)
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

	// Spit out requested dims since scheduke doesn't pass this info to swarming
	common.WriteAnyObjectToStepLog(ctx, step, req.GetDimensions(), "requested dimensions")

	builderId := common.TestRunnerBuilderID()

	bbClient, err := newBBClient(ctx)
	if err != nil {
		return err
	}

	cmd.ObserveTrSchedulingStart(ctx, buildReq)

	// BQ TODO log the request is in the scheduling tool (ie log the scheduke ID if possible?)
	scheduledBuild, err := cmd.Scheduler.ScheduleRequest(ctx, req, step)
	if err != nil {
		err = fmt.Errorf("error while scheduling req: %s", err)
		cmd.ObserveTrSchedulingFail(ctx, buildReq, err.Error())
		return setTopLevelError(ctx, step, result, resultsChan, err)
	}

	if scheduledBuild != nil && scheduledBuild.GetId() != 0 {
		result.BuildUrl = common.BBUrl(builderId, scheduledBuild.GetId())
		step.SetSummaryMarkdown(fmt.Sprintf("[latest attempt](%s)", common.BBUrl(builderId, scheduledBuild.GetId())))
	} else {
		errStr := "no bbid found from scheduler"
		err = fmt.Errorf(errStr)
		cmd.ObserveTrSchedulingFail(ctx, buildReq, err.Error())

		return setTopLevelError(ctx, step, result, resultsChan, err)
	}
	// Log the successful start.
	cmd.ObserveTrSchedulingSuccess(ctx, buildReq, fmt.Sprint(scheduledBuild.GetId()))

	// Re-init the data for the run build step. Keep the previously populated data.
	cmd.ObserveTrBuildStart(ctx, buildReq)

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
		cmd.ObserveTrBuildSuccess(ctx, buildReq, buildInfo)

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
			newBuildReq := cmd.RetryReqIfQualifies(ctx, trResult, step, buildReq)
			if newBuildReq != nil && newBuildReq.ScheduleBuildRequest != nil {
				// Schedule retry
				wg.Add(1)
				go cmd.ScheduleAndMonitor(rootCtx, newBuildReq.Key, newBuildReq, wg, resultsChan, retryNum+1)
			}
		}

		// Send the result via channel
		resultsChan <- result
		break
	}

	return nil
}

func (cmd *ScheduleTasksCmd) RetryReqIfQualifies(ctx context.Context, trResult *skylab_test_runner.Result, step *build.Step, buildReq *data.BuildRequest) *data.BuildRequest {
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
	req, err := cmd.GenerateReqForRetry(ctx, newBuildReq)
	if err != nil {
		newBuildReq.Err = err
		logging.Infof(ctx, "no more retry will take place for %s since trv2 req generation failed: %s", newBuildReq.Key, err)
	} else {
		newBuildReq.ScheduleBuildRequest = req
	}

	return newBuildReq
}

func (cmd *ScheduleTasksCmd) GenerateReqForRetry(ctx context.Context, buildReq *data.BuildRequest) (*buildbucketpb.ScheduleBuildRequest, error) {
	// TODO (azrahman/dbeckett): add analytics metrics for this func
	var err error
	trReq := buildReq.OriginalTrReq
	key := buildReq.Key
	shardNum := buildReq.ShardNum
	// '0'ed index because we should always have one hw here. It supports multiple
	// MO should reduce it down to 1 always. The len check is done at MO step.
	TrReqhwDef := trReq.Req.GetHwDefinition()[0]
	testCases := trReq.Tcs

	// Input validations
	if len(testCases) == 0 {
		errStr := "no test is found so, rejecting task"
		logging.Infof(ctx, errStr)
		err = fmt.Errorf(errStr)
		return nil, err

	}

	if trReq.DevicesInfo.LabDevicesCount == 0 {
		logging.Infof(ctx, "no suitable device found to run tests so, rejecting task")
		err := &data.BotParamsRejectedError{Key: key, RejectedDims: trReq.DevicesInfo.Dims}
		return nil, err
	}

	helper := &TrV2ReqHelper{
		trReqHWDef: TrReqhwDef,
		testCases:  testCases,
		build:      cmd.BuildState,
		suiteInfo:  cmd.InternalTestPlan.SuiteInfo,
		shardNum:   shardNum,
		dynamicRun: cmd.DynamicRun,
	}

	req, err := GenerateTrv2Req(ctx, true, helper)
	if err != nil {
		logging.Infof(ctx, "error while generating req: %s", err)
		return nil, errors.Annotate(err, "error while generating req:").Err()
	}

	return req, nil
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

func GenerateNewBuildReqForRetry(ctx context.Context, buildReq *data.BuildRequest, retriableTests map[string]bool) *data.BuildRequest {
	if len(retriableTests) == 0 {
		return nil
	}

	// dereference so that we can make changes
	retryBuildReq := *buildReq
	retryBuildReq.ScheduleBuildRequest = nil
	retryBuildReq.Err = nil

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
	result.TopLevelError = err

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

// ------------ analytics funcs -------

func (cmd *ScheduleTasksCmd) ObserveCmdStart(ctx context.Context) {
	cmd.StartCmdTime = time.Now()
	bqData := &analytics.BqData{Step: string(cmd.GetCommandType()), Status: analytics.Start}
	analytics.SoftInsertStepWInternalPlan(ctx, cmd.BQClient, bqData, cmd.InternalTestPlan, cmd.BuildState)
}

func (cmd *ScheduleTasksCmd) ObserveCmdEndSuccess(ctx context.Context) {
	bqData := &analytics.BqData{Step: string(cmd.GetCommandType()), Status: analytics.Start, Duration: float32(time.Since(cmd.StartCmdTime).Seconds())}
	analytics.SoftInsertStepWInternalPlan(ctx, cmd.BQClient, bqData, cmd.InternalTestPlan, cmd.BuildState)
}

func (cmd *ScheduleTasksCmd) ObserveSchedulerSetupSuccess(ctx context.Context) {
	bqData := &analytics.BqData{Step: "SchedulerSetup", Status: analytics.Success, Duration: float32(time.Since(cmd.StartCmdTime).Seconds())}
	analytics.SoftInsertStepWInternalPlan(ctx, cmd.BQClient, bqData, cmd.InternalTestPlan, cmd.BuildState)
}

func (cmd *ScheduleTasksCmd) ObserveSchedulerSetupFailure(ctx context.Context, err string) {
	bqData := &analytics.BqData{Step: "SchedulerSetup", Status: analytics.Fail, Duration: float32(time.Since(cmd.StartCmdTime).Seconds()), Freeform: err}
	analytics.SoftInsertStepWInternalPlan(ctx, cmd.BQClient, bqData, cmd.InternalTestPlan, cmd.BuildState)
}

func (cmd *ScheduleTasksCmd) ObserveTrSchedulingStart(ctx context.Context, buildReq *data.BuildRequest) {
	cmd.StartTrSchedulingTime = time.Now()
	data := &analytics.TaskData{
		Step:          "ScheduleBuild",
		DisplayName:   buildReq.Key,
		AnalyticsName: buildReq.SuiteInfo.GetSuiteRequest().GetAnalyticsName(),
		Status:        analytics.Start,
	}
	analytics.SoftInsertStepWTrReq(ctx, cmd.BQClient, data, buildReq.OriginalTrReq, buildReq.SuiteInfo, cmd.BuildState)
}

func (cmd *ScheduleTasksCmd) ObserveTrSchedulingSuccess(ctx context.Context, buildReq *data.BuildRequest, taskBBId string) {
	data := &analytics.TaskData{
		Step:          "ScheduleBuild",
		DisplayName:   buildReq.Key,
		AnalyticsName: buildReq.SuiteInfo.GetSuiteRequest().GetAnalyticsName(),
		Status:        analytics.Success,
		TrTaskID:      taskBBId,
		Duration:      float32(time.Since(cmd.StartTrSchedulingTime).Seconds()),
	}
	analytics.SoftInsertStepWTrReq(ctx, cmd.BQClient, data, buildReq.OriginalTrReq, buildReq.SuiteInfo, cmd.BuildState)
}

func (cmd *ScheduleTasksCmd) ObserveTrSchedulingFail(ctx context.Context, buildReq *data.BuildRequest, err string) {
	data := &analytics.TaskData{
		Step:          "ScheduleBuild",
		DisplayName:   buildReq.Key,
		AnalyticsName: buildReq.SuiteInfo.GetSuiteRequest().GetAnalyticsName(),
		Status:        analytics.Fail,
		Freeform:      err,
		Duration:      float32(time.Since(cmd.StartTrSchedulingTime).Seconds()),
	}
	analytics.SoftInsertStepWTrReq(ctx, cmd.BQClient, data, buildReq.OriginalTrReq, buildReq.SuiteInfo, cmd.BuildState)
}

func (cmd *ScheduleTasksCmd) ObserveTrBuildStart(ctx context.Context, buildReq *data.BuildRequest) {
	cmd.StartTrBuildTime = time.Now()
	data := &analytics.TaskData{
		Step:          "RunBuild",
		DisplayName:   buildReq.Key,
		AnalyticsName: buildReq.SuiteInfo.GetSuiteRequest().GetAnalyticsName(),
		Status:        analytics.Start,
	}
	analytics.SoftInsertStepWTrReq(ctx, cmd.BQClient, data, buildReq.OriginalTrReq, buildReq.SuiteInfo, cmd.BuildState)
}

func (cmd *ScheduleTasksCmd) ObserveTrBuildSuccess(ctx context.Context, buildReq *data.BuildRequest, buildInfo *buildbucketpb.Build) {
	data := &analytics.TaskData{
		Step:          "RunBuild",
		DisplayName:   buildReq.Key,
		AnalyticsName: buildReq.SuiteInfo.GetSuiteRequest().GetAnalyticsName(),
		Status:        analytics.Success,
		TrTaskID:      fmt.Sprint(buildInfo.GetId()),
		Freeform:      fmt.Sprint("build status: ", buildInfo.GetStatus()),
		Duration:      float32(time.Since(cmd.StartTrBuildTime).Seconds()),
	}
	analytics.SoftInsertStepWTrReq(ctx, cmd.BQClient, data, buildReq.OriginalTrReq, buildReq.SuiteInfo, cmd.BuildState)
}

func (cmd *ScheduleTasksCmd) ObserveTrBuildFail(ctx context.Context, buildReq *data.BuildRequest, err string) {
	data := &analytics.TaskData{
		Step:          "RunBuild",
		DisplayName:   buildReq.Key,
		AnalyticsName: buildReq.SuiteInfo.GetSuiteRequest().GetAnalyticsName(),
		Status:        analytics.Fail,
		Freeform:      err,
		Duration:      float32(time.Since(cmd.StartTrBuildTime).Seconds()),
	}
	analytics.SoftInsertStepWTrReq(ctx, cmd.BQClient, data, buildReq.OriginalTrReq, buildReq.SuiteInfo, cmd.BuildState)
}

// ----------- end analytics funcs --------

func NewScheduleTasksCmd() *ScheduleTasksCmd {
	abstractCmd := interfaces.NewAbstractCmd(ScheduleTasksCmdType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &ScheduleTasksCmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}
