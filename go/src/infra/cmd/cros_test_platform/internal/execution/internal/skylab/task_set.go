// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package skylab

import (
	"context"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/config"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/steps"
	swarming_api "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/swarming/proto/jsonrpc"

	"infra/cmd/cros_test_platform/internal/execution/isolate"
	"infra/cmd/cros_test_platform/internal/execution/swarming"
	"infra/libs/skylab/inventory"
)

// TaskSet encapsulates the running state of a set of tasks, to satisfy
// a Skylab Execution.
type TaskSet struct {
	testRuns         []*testRun
	globalMaxRetries int32
	retries          int32
}

// NewTaskSet creates a new TaskSet.
func NewTaskSet(ctx context.Context, tests []*steps.EnumerationResponse_AutotestInvocation, params *test_platform.Request_Params, workerConfig *config.Config_SkylabWorker, parentTaskID string) (*TaskSet, error) {
	testRuns := make([]*testRun, len(tests))
	for i, test := range tests {
		t, err := newTestRun(ctx, test, params, workerConfig, parentTaskID)
		if err != nil {
			return nil, errors.Annotate(err, "new task set").Err()
		}
		testRuns[i] = t
	}
	return &TaskSet{
		testRuns:         testRuns,
		globalMaxRetries: inferGlobalMaxRetries(params),
	}, nil
}

// Completed returns true if all test runs in this task set have completed.
func (r *TaskSet) Completed() bool {
	for _, t := range r.testRuns {
		if !t.Completed() {
			return false
		}
	}
	return true
}

func inferGlobalMaxRetries(params *test_platform.Request_Params) int32 {
	if !params.GetRetry().GetAllow() {
		return 0
	}
	return maxInt32IfZero(params.GetRetry().GetMax())
}

// LaunchTasks launches initial tasks for all the testRuns in this task set.
func (r *TaskSet) LaunchTasks(ctx context.Context, client swarming.Client) error {
	for _, testRun := range r.testRuns {
		runnable, err := testRun.ValidateDependencies(ctx, client)
		if err != nil {
			return err
		}
		if !runnable {
			testRun.MarkNotRunnable()
			continue
		}
		if err := testRun.LaunchAttempt(ctx, client); err != nil {
			return err
		}
	}
	return nil
}

// CheckTasksAndRetry checks the status of currently running tasks and retries
// failed tasks when allowed.
func (r *TaskSet) CheckTasksAndRetry(ctx context.Context, client swarming.Client, gf isolate.GetterFactory) error {
	for _, testRun := range r.testRuns {
		if testRun.Completed() {
			continue
		}

		latestAttempt := testRun.GetLatestAttempt()
		if err := latestAttempt.FetchResults(ctx, client, gf); err != nil {
			return errors.Annotate(err, "tick for task %s", latestAttempt.taskID).Err()
		}

		if !testRun.Completed() {
			continue
		}

		logging.Debugf(ctx, "Task %s (%s) completed with verdict %s", latestAttempt.taskID, testRun.Name, latestAttempt.Verdict())

		shouldRetry, err := r.shouldRetry(ctx, testRun)
		if err != nil {
			return errors.Annotate(err, "tick for task %s", latestAttempt.taskID).Err()
		}
		if shouldRetry {
			logging.Debugf(ctx, "Retrying %s", testRun.Name)
			if err := testRun.LaunchAttempt(ctx, client); err != nil {
				return errors.Annotate(err, "tick for task %s: retry test", latestAttempt.taskID).Err()
			}
			r.retries++
		}
	}
	return nil
}

// shouldRetry computes if the given testRun should be retried.
func (r *TaskSet) shouldRetry(ctx context.Context, tr *testRun) (bool, error) {
	if !tr.AttemptedAtLeastOnce() {
		return false, errors.Reason("shouldRetry: can't retry a never-tried test").Err()
	}
	if tr.AttemptsRemaining() <= 0 {
		logging.Debugf(ctx, "Not retrying %s. Hit the test retry limit.", tr.Name)
		return false, nil
	}
	if r.globalRetriesRemaining() <= 0 {
		logging.Debugf(ctx, "Not retrying %s. Hit the task set retry limit.", tr.Name)
		return false, nil
	}

	latestAttempt := tr.GetLatestAttempt()
	switch verdict := latestAttempt.Verdict(); verdict {
	case test_platform.TaskState_VERDICT_UNSPECIFIED:
		fallthrough
	case test_platform.TaskState_VERDICT_FAILED:
		return true, nil
	case test_platform.TaskState_VERDICT_NO_VERDICT:
		fallthrough
	case test_platform.TaskState_VERDICT_PASSED:
		fallthrough
	case test_platform.TaskState_VERDICT_PASSED_ON_RETRY:
		return false, nil
	default:
		return false, errors.Reason("shouldRetry: unknown verdict %s", verdict.String()).Err()
	}
}

func (r *TaskSet) globalRetriesRemaining() int32 {
	return r.globalMaxRetries - r.retries
}

func unpackResult(results []*swarming_api.SwarmingRpcsTaskResult, taskID string) (*swarming_api.SwarmingRpcsTaskResult, error) {
	if len(results) != 1 {
		return nil, errors.Reason("expected 1 result for task id %s, got %d", taskID, len(results)).Err()
	}

	result := results[0]
	if result.TaskId != taskID {
		return nil, errors.Reason("expected result for task id %s, got %s", taskID, result.TaskId).Err()
	}

	return result, nil
}

var poolMap = map[test_platform.Request_Params_Scheduling_ManagedPool]inventory.SchedulableLabels_DUTPool{
	test_platform.Request_Params_Scheduling_MANAGED_POOL_ARC_PRESUBMIT: inventory.SchedulableLabels_DUT_POOL_ARC_PRESUBMIT,
	test_platform.Request_Params_Scheduling_MANAGED_POOL_BVT:           inventory.SchedulableLabels_DUT_POOL_BVT,
	test_platform.Request_Params_Scheduling_MANAGED_POOL_CONTINUOUS:    inventory.SchedulableLabels_DUT_POOL_CONTINUOUS,
	test_platform.Request_Params_Scheduling_MANAGED_POOL_CQ:            inventory.SchedulableLabels_DUT_POOL_CQ,
	test_platform.Request_Params_Scheduling_MANAGED_POOL_CTS_PERBUILD:  inventory.SchedulableLabels_DUT_POOL_CTS_PERBUILD,
	test_platform.Request_Params_Scheduling_MANAGED_POOL_CTS:           inventory.SchedulableLabels_DUT_POOL_CTS,
	// TODO(akeshet): This mapping is inexact. Requests that specify a quota account should not
	// specify a pool, and should go routed to the quota pool automatically.
	test_platform.Request_Params_Scheduling_MANAGED_POOL_QUOTA:  inventory.SchedulableLabels_DUT_POOL_QUOTA,
	test_platform.Request_Params_Scheduling_MANAGED_POOL_SUITES: inventory.SchedulableLabels_DUT_POOL_SUITES,
}

var taskStateToLifeCycle = map[jsonrpc.TaskState]test_platform.TaskState_LifeCycle{
	jsonrpc.TaskState_BOT_DIED:  test_platform.TaskState_LIFE_CYCLE_ABORTED,
	jsonrpc.TaskState_CANCELED:  test_platform.TaskState_LIFE_CYCLE_CANCELLED,
	jsonrpc.TaskState_COMPLETED: test_platform.TaskState_LIFE_CYCLE_COMPLETED,
	// TODO(akeshet): This mapping is inexact. Add a lifecycle entry for this.
	jsonrpc.TaskState_EXPIRED:     test_platform.TaskState_LIFE_CYCLE_CANCELLED,
	jsonrpc.TaskState_KILLED:      test_platform.TaskState_LIFE_CYCLE_ABORTED,
	jsonrpc.TaskState_NO_RESOURCE: test_platform.TaskState_LIFE_CYCLE_REJECTED,
	jsonrpc.TaskState_PENDING:     test_platform.TaskState_LIFE_CYCLE_PENDING,
	jsonrpc.TaskState_RUNNING:     test_platform.TaskState_LIFE_CYCLE_RUNNING,
	// TODO(akeshet): This mapping is inexact. Add a lifecycle entry for this.
	jsonrpc.TaskState_TIMED_OUT: test_platform.TaskState_LIFE_CYCLE_ABORTED,
}

func (r *TaskSet) response(urler swarming.URLer, running bool) *steps.ExecuteResponse {
	resp := &steps.ExecuteResponse{
		TaskResults:         r.taskResults(urler),
		ConsolidatedResults: r.results(urler),
		State: &test_platform.TaskState{
			Verdict:   r.verdict(),
			LifeCycle: r.lifecycle(running),
		},
	}
	return resp
}

func (r *TaskSet) lifecycle(running bool) test_platform.TaskState_LifeCycle {
	switch {
	case r.Completed():
		return test_platform.TaskState_LIFE_CYCLE_COMPLETED
	case running:
		return test_platform.TaskState_LIFE_CYCLE_RUNNING
	default:
		// TODO(akeshet): The task set is neither running nor complete, so it
		// was cancelled due to an error while in flight. It's not clear yet
		// if this is the right lifecycle mapping for this state.
		return test_platform.TaskState_LIFE_CYCLE_ABORTED
	}
}

func (r *TaskSet) verdict() test_platform.TaskState_Verdict {
	v := test_platform.TaskState_VERDICT_PASSED
	if !r.Completed() {
		v = test_platform.TaskState_VERDICT_UNSPECIFIED
	}
	for _, t := range r.testRuns {
		if !successfulVerdict(t.Verdict()) {
			v = test_platform.TaskState_VERDICT_FAILED
			break
		}
	}
	return v
}

func successfulVerdict(v test_platform.TaskState_Verdict) bool {
	switch v {
	case test_platform.TaskState_VERDICT_PASSED,
		test_platform.TaskState_VERDICT_PASSED_ON_RETRY,
		test_platform.TaskState_VERDICT_NO_VERDICT:
		return true
	default:
		return false
	}
}

func (r *TaskSet) results(urler swarming.URLer) []*steps.ExecuteResponse_ConsolidatedResult {
	rs := make([]*steps.ExecuteResponse_ConsolidatedResult, len(r.testRuns))
	for i, test := range r.testRuns {
		rs[i] = &steps.ExecuteResponse_ConsolidatedResult{
			Attempts: test.TaskResult(urler),
		}
	}
	return rs
}

func (r *TaskSet) taskResults(urler swarming.URLer) []*steps.ExecuteResponse_TaskResult {
	results := r.results(urler)
	var trs []*steps.ExecuteResponse_TaskResult
	for _, result := range results {
		trs = append(trs, result.Attempts...)
	}
	return trs
}
