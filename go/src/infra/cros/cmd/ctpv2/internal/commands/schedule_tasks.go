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
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/common_lib/schedulers"
	"infra/cros/cmd/ctpv2/data"
)

const (
	TestStepNameTemplate = "request %s-%s.hw.%s-shard-%v"
)

// ScheduleTasksCmd represents scheduling task(s) cmd.
type ScheduleTasksCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Deps
	MiddledOutResp *data.MiddleOutResponse
	BuildState     *build.State
	Scheduler      interfaces.SchedulerInterface

	// Updates
	// TODO (azrahman): Consider adding TRv2 output props if needed.
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
	// Set TRv2 output props

	return nil
}

// Execute executes the command.
func (cmd *ScheduleTasksCmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Schedule tasks")
	defer func() { step.End(err) }()

	if len(cmd.MiddledOutResp.TrReqs) == 0 {
		logging.Infof(ctx, "no test found in middle-out response")
		step.SetSummaryMarkdown("enumeration error: no test found")
		err = fmt.Errorf("enumeration error: no test found")
		return err
	}

	wg := &sync.WaitGroup{}
	for i, trReq := range cmd.MiddledOutResp.TrReqs {
		wg.Add(1)
		logging.Infof(ctx, "scheduling task %d...", i)
		go ScheduleTask(ctx, trReq, cmd.BuildState, cmd.MiddledOutResp.SuiteInfo, wg, cmd.Scheduler, i)
	}
	wg.Wait()

	return nil
}

func ScheduleTask(ctx context.Context, trReq *data.TrRequest, buildState *build.State, suiteInfo *api.SuiteInfo, wg *sync.WaitGroup, scheduler interfaces.SchedulerInterface, shardNum int) (*buildbucketpb.ScheduleBuildRequest, error) {
	defer wg.Done()

	if trReq.Req == nil {
		return nil, nil
	}

	if len(trReq.Req.GetHwDefinition()) == 0 {
		logging.Infof(ctx, "no hw def is found in req")
		return nil, fmt.Errorf("no hw def is found so, rejecting task")
	}

	// '0'ed index because we should always have one hw here. It supports multiple
	// MO should reduce it down to 1 always. The len check is done at MO step.
	TrReqhwDef := trReq.Req.GetHwDefinition()[0]
	testCases := trReq.Tcs
	board := strings.ToLower(getBuildTargetfromHwDef(TrReqhwDef))
	variant := strings.ToLower(getVariantFromHwDef(TrReqhwDef))
	model := strings.ToLower(getModelTargetfromHwDef(TrReqhwDef))
	suiteName := suiteName(suiteInfo)

	builderString := board
	if model != "" {
		builderString = fmt.Sprintf("%s-%s", builderString, model)
	}
	if variant != "" {
		builderString = fmt.Sprintf("%s-%s", builderString, variant)
	}

	var err error
	step, ctx := build.StartStep(ctx, fmt.Sprintf(TestStepNameTemplate, builderString, FindBuildName(suiteInfo, board), suiteName, shardNum))
	defer func() { step.End(err) }()

	// Input validations
	if len(trReq.Tcs) == 0 {
		logging.Infof(ctx, "no test is found in req")
		step.SetSummaryMarkdown("no test-cases found to run")
		err = fmt.Errorf("no test is found so, rejecting task")
		return nil, err
	}

	if trReq.LabDevices == 0 {
		logging.Infof(ctx, "no suitable device found to run tests so, rejecting task")
		step.SetSummaryMarkdown("bot params rejected")
		err = fmt.Errorf("bot params rejected")
		return nil, err
	}

	builderId := common.TestRunnerBuilderID()

	bbClient, err := newBBClient(ctx)
	if err != nil {
		return nil, nil
	}

	// Generate req
	var scheduledBuild *buildbucketpb.Build

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
	} else {
		if scheduler == nil {
			logging.Infof(ctx, "empty scheduler")
			return nil, fmt.Errorf("empty scheduler")
		}
		err = scheduler.Setup(ctx)
		if err != nil {
			logging.Infof(ctx, "error while setting up scheduler: %s", err)
			return nil, errors.Annotate(err, "error while setting up scheduler").Err()
		}
		// Spit out the request
		requestData, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			logging.Infof(
				ctx,
				"error during writing request data to log: %s",
				err.Error())
		}
		step.Log("BB Request").Write(requestData)
		scheduledBuild, err = scheduler.ScheduleRequest(ctx, req, step)
		if err != nil {
			logging.Infof(ctx, "error while scheduling req: %s", err)
			return nil, errors.Annotate(err, "error while generating req:").Err()
		}
		if scheduledBuild != nil {
			step.SetSummaryMarkdown(fmt.Sprintf("[latest attempt](%s)", common.BBUrl(builderId, scheduledBuild.Id)))
		}
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

			if int(b.GetStatus())&int(buildbucketpb.Status_ENDED_MASK) != 0 {
				if b.GetStatus() != buildbucketpb.Status_SUCCESS {
					err = fmt.Errorf("test_runner failed")
				}
				break
			}

			select {
			case <-ctx.Done():
				// A timeout while waiting for tests to complete is reported as
				// aborts when summarizing individual tests' results.
				// The execute step completes without errors.
				return nil, nil
			case <-clock.After(ctx, loopSleepInterval):
			}
		}
	}

	return nil, nil
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
