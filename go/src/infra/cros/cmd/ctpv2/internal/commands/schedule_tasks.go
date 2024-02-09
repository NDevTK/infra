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
	TestStepNameTemplate = "request %s-%s.hw.%s"
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

	wg := &sync.WaitGroup{}
	for i, trReq := range cmd.MiddledOutResp.TrReqs {
		wg.Add(1)
		logging.Infof(ctx, "scheduling task %d...", i)
		go ScheduleTask(ctx, trReq, cmd.BuildState, cmd.MiddledOutResp.SuiteInfo, wg, cmd.Scheduler)
	}
	wg.Wait()

	return nil
}

func ScheduleTask(ctx context.Context, trReq *data.TrRequest, buildState *build.State, suiteInfo *api.SuiteInfo, wg *sync.WaitGroup, scheduler interfaces.SchedulerInterface) (*buildbucketpb.ScheduleBuildRequest, error) {
	defer wg.Done()
	// '0'ed index because we should always have one hw here. It supports multiple
	// MO should reduce it down to 1 always. The len check is done at MO step.
	hwDef := trReq.Req.GetHwDefinition()[0]
	testCases := trReq.Tcs
	board := strings.ToLower(hwDef.GetDutInfo().GetChromeos().GetDutModel().GetBuildTarget())
	suiteName := suiteInfo.GetSuiteRequest().GetTestSuite().GetName()
	// TODO (azrhaman): consider using board-model-variant-build rather
	// than board-build in the request step name.
	//model := strings.ToLower(hwDef.GetDutInfo().GetChromeos().GetDutModel().GetModelName())
	//boardModelVariant := fmt.Sprintf("%s-%s", board, model)

	var err error
	step, ctx := build.StartStep(ctx, fmt.Sprintf(TestStepNameTemplate, board, FindBuildName(suiteInfo, board), suiteName))
	defer func() { step.End(err) }()

	if len(trReq.Req.GetHwDefinition()) == 0 {
		return nil, nil
	}
	if len(trReq.Tcs) == 0 {
		return nil, nil
	}

	builderId := common.TestRunnerBuilderID()

	bbClient, err := newBBClient(ctx)
	if err != nil {
		return nil, nil
	}

	// Generate req
	var scheduledBuild *buildbucketpb.Build
	req, err := GenerateTrv2Req(ctx, hwDef, testCases, buildState, suiteInfo, true)
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
		scheduledBuild, err = scheduler.ScheduleRequest(ctx, req, step)
		if err != nil {
			logging.Infof(ctx, "error while scheduling req: %s", err)
			return nil, errors.Annotate(err, "error while generating req:").Err()
		}
		if scheduledBuild != nil {
			step.SetSummaryMarkdown(fmt.Sprintf("[latest attempt](%s)", common.BBUrl(builderId, scheduledBuild.Id)))
		}
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
		hwDef := target.GetHwRequirements().GetHwDefinition()[0]
		if hwDef.GetDutInfo().GetChromeos().GetDutModel().GetBuildTarget() == board {
			return target.GetSwRequirements()[0].GetBuild()
		}
	}

	return ""
}
