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

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/genproto/protobuf/field_mask"
	protobuf "google.golang.org/protobuf/proto"

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

	// Updates
	TestResults map[string]*data.TestResults
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
		err = cmd.updateFilterStateKeeper(ctx, sk)
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

func (cmd *ScheduleTasksCmd) updateFilterStateKeeper(ctx context.Context, sk *data.FilterStateKeeper) error {
	if cmd.TestResults != nil && len(cmd.TestResults) != 0 {
		sk.SuiteTestResults = cmd.TestResults
	}

	return nil
}

// Execute executes the command.
func (cmd *ScheduleTasksCmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Schedule tasks")
	defer func() { step.End(err) }()

	cmd.TestResults = map[string]*data.TestResults{}
	suiteName := suiteName(cmd.MiddledOutResp.SuiteInfo)
	if len(cmd.MiddledOutResp.TrReqs) == 0 {
		logging.Infof(ctx, "no test found in middle-out response")
		step.SetSummaryMarkdown("enumeration error: no test found")
		err = &data.EnumerationError{SuiteName: suiteName}
		cmd.TestResults[common.EnumerationErrKey] = &data.TestResults{Suite: suiteName, Key: common.EnumerationErrKey, TopLevelError: &err}
		common.WriteAnyObjectToStepLog(ctx, step, cmd.TestResults, "consolidated results")
		return err
	}

	// GenerateRequests
	buildMap := GenerateRequests(ctx, cmd.MiddledOutResp, cmd.BuildState)
	if len(buildMap) == 0 {
		step.SetSummaryMarkdown("enumeration error: no valid test found")
		err = &data.EnumerationError{SuiteName: suiteName}
		cmd.TestResults[common.EnumerationErrKey] = &data.TestResults{Suite: suiteName, Key: common.EnumerationErrKey, TopLevelError: &err}
		return err
	}

	scheduler := cmd.Scheduler
	if cmd.Scheduler == nil {
		logging.Infof(ctx, "empty scheduler")
		return fmt.Errorf("empty scheduler")
	}
	err = scheduler.Setup()
	if err != nil {
		logging.Infof(ctx, "error while setting up scheduler: %s", err)
		return errors.Annotate(err, "error while setting up scheduler").Err()
	}

	// Todo: batch call
	resultsChan := make(chan *data.TestResults)
	wg := &sync.WaitGroup{}
	for k, v := range buildMap {
		wg.Add(1)
		go ScheduleAndMonitor(ctx, cmd.Scheduler, v, cmd.BuildState, k, wg, resultsChan, suiteName)
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
	common.WriteAnyObjectToStepLog(ctx, step, cmd.TestResults, "consolidated results")
	return nil

}

type BuildRequest struct {
	ScheduleBuildRequest *buildbucketpb.ScheduleBuildRequest
	err                  error
}

func GenerateRequests(ctx context.Context, moResp *data.MiddleOutResponse, buildState *build.State) map[string]*BuildRequest {
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

		buildReq := &BuildRequest{}
		req, err := GenerateReq(ctx, trReq, modifiedKey, buildState, moResp.SuiteInfo, shardMap[key])
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

func GenerateReq(ctx context.Context, trReq *data.TrRequest, key string, buildState *build.State, suiteInfo *api.SuiteInfo, shardNum int) (*buildbucketpb.ScheduleBuildRequest, error) {
	var err error

	// '0'ed index because we should always have one hw here. It supports multiple
	// MO should reduce it down to 1 always. The len check is done at MO step.
	TrReqhwDef := trReq.Req.GetHwDefinition()[0]
	testCases := trReq.Tcs
	// Input validations
	if len(testCases) == 0 {
		logging.Infof(ctx, "no test is found in req")
		err = fmt.Errorf("no test is found so, rejecting task")
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
		build:      buildState,
		suiteInfo:  suiteInfo,
		shardNum:   shardNum,
	}

	req, err := GenerateTrv2Req(ctx, true, helper)
	if err != nil {
		logging.Infof(ctx, "error while generating req: %s", err)
		return nil, errors.Annotate(err, "error while generating req:").Err()
	}
	return req, nil
}

func ScheduleAndMonitor(ctx context.Context, scheduler interfaces.SchedulerInterface, buildReq *BuildRequest, buildState *build.State, key string, wg *sync.WaitGroup, resultsChan chan<- *data.TestResults, suiteName string) error {
	defer wg.Done()
	var err error
	step, ctx := build.StartStep(ctx, key)
	defer func() { step.End(err) }()

	// Construct test results
	result := &data.TestResults{Key: key, Suite: suiteName}

	if buildReq.err != nil {
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

	scheduledBuild, err := scheduler.ScheduleRequest(ctx, req, step)
	if err != nil {
		return setTopLevelError(ctx, step, result, resultsChan, fmt.Errorf("error while scheduling req: %s", err))
	}

	if scheduledBuild != nil && scheduledBuild.GetId() != 0 {
		result.BuildUrl = common.BBUrl(builderId, scheduledBuild.GetId())
		step.SetSummaryMarkdown(fmt.Sprintf("[latest attempt](%s)", common.BBUrl(builderId, scheduledBuild.Id)))
	} else {
		return setTopLevelError(ctx, step, result, resultsChan, fmt.Errorf("no bbid found from scheduler: %s", err.Error()))
	}

	// Monitor here
	loopSleepInterval := 30 * time.Second

	if scheduledBuild != nil {
		statusReq := &buildbucketpb.GetBuildStatusRequest{
			Id: scheduledBuild.Id,
		}
		for {
			// Check Build status.
			b, err := bbClient.GetBuildStatus(ctx, statusReq)
			if err != nil {
				logging.Infof(ctx, "error while getting build status: %s", err)
			}

			if b != nil && int(b.GetStatus())&int(buildbucketpb.Status_ENDED_MASK) != 0 {
				logging.Infof(ctx, "bb status: %s", b.GetStatus())
				if b.GetStatus() != buildbucketpb.Status_SUCCESS {
					err = fmt.Errorf("test_runner failed")
				}

				// Get more build info
				req := &buildbucketpb.GetBuildRequest{
					Id:   b.Id,
					Mask: &buildbucketpb.BuildMask{Fields: &field_mask.FieldMask{Paths: getBuildFieldMask}},
				}
				buildInfo, err := bbClient.GetBuild(ctx, req)
				if err != nil {
					logging.Infof(ctx, "error while getting build info: %s", err)
				}
				common.WriteAnyObjectToStepLog(ctx, step, buildInfo, "final build info")

				trResult, err := extractResult(buildInfo)
				if err != nil {
					return setTopLevelError(ctx, step, result, resultsChan, fmt.Errorf("error while extracting results from test_runner build %d: %s", buildInfo.Id, err))
				} else {
					result.Results = trResult
				}
				common.WriteAnyObjectToStepLog(ctx, step, result, "extracted result from trv2")

				// Send the result via channel
				resultsChan <- result
				break
			}

			select {
			case <-ctx.Done():
				// A timeout while waiting for tests to complete is reported as
				// aborts when summarizing individual tests' results.
				// The execute step completes without errors.
				return nil
			case <-clock.After(ctx, loopSleepInterval):
			}

		}

	}
	return nil
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
