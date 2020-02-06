// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package skylab

import (
	"context"
	"time"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/config"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/steps"
	swarming_api "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/swarming/proto/jsonrpc"
)

// RequestTaskSet encapsulates the running state of the set of tasks for one
// cros_test_platform request.
type RequestTaskSet struct {
	testTaskSets     []*testTaskSet
	globalMaxRetries int32
	retries          int32
}

// NewRequestTaskSet creates a new RequestTaskSet.
func NewRequestTaskSet(tests []*steps.EnumerationResponse_AutotestInvocation, params *test_platform.Request_Params, workerConfig *config.Config_SkylabWorker, parentTaskID string, deadline time.Time) (*RequestTaskSet, error) {
	testTaskSets := make([]*testTaskSet, len(tests))
	for i, test := range tests {
		t, err := newTestTaskSet(test, params, workerConfig, parentTaskID, deadline)
		if err != nil {
			return nil, errors.Annotate(err, "new task set").Err()
		}
		testTaskSets[i] = t
	}
	return &RequestTaskSet{
		testTaskSets:     testTaskSets,
		globalMaxRetries: inferGlobalMaxRetries(params),
	}, nil
}

// Completed returns true if all tasks for this request have completed.
func (r *RequestTaskSet) Completed() bool {
	for _, t := range r.testTaskSets {
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

// LaunchTasks launches initial tasks for all the tests in this request.
func (r *RequestTaskSet) LaunchTasks(ctx context.Context, clients Clients) error {
	for _, testTaskSet := range r.testTaskSets {
		runnable, err := testTaskSet.ValidateDependencies(ctx, clients.Swarming)
		if err != nil {
			return err
		}
		if !runnable {
			testTaskSet.MarkNotRunnable()
			continue
		}
		if err := testTaskSet.LaunchTask(ctx, clients); err != nil {
			return err
		}
	}
	return nil
}

// CheckTasksAndRetry checks the status of currently running tasks for this
// request and retries failed tasks when allowed.
func (r *RequestTaskSet) CheckTasksAndRetry(ctx context.Context, clients Clients) error {
	for _, testTaskSet := range r.testTaskSets {
		if testTaskSet.Completed() {
			continue
		}

		latestTask := testTaskSet.GetLatestTask()
		if err := latestTask.FetchResults(ctx, clients); err != nil {
			return errors.Annotate(err, "tick for task %s", latestTask.ID).Err()
		}

		if !testTaskSet.Completed() {
			continue
		}

		logging.Debugf(ctx, "Task %s (%s) completed with verdict %s", latestTask.ID, testTaskSet.Name, latestTask.Verdict())

		shouldRetry, err := r.shouldRetry(ctx, testTaskSet)
		if err != nil {
			return errors.Annotate(err, "tick for task %s", latestTask.ID).Err()
		}
		if shouldRetry {
			logging.Debugf(ctx, "Retrying %s", testTaskSet.Name)
			if err := testTaskSet.LaunchTask(ctx, clients); err != nil {
				return errors.Annotate(err, "tick for task %s: retry test", latestTask.ID).Err()
			}
			r.retries++
		}
	}
	return nil
}

// shouldRetry computes if the given test should be retried.
func (r *RequestTaskSet) shouldRetry(ctx context.Context, tr *testTaskSet) (bool, error) {
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

	latestTask := tr.GetLatestTask()
	switch verdict := latestTask.Verdict(); verdict {
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

func (r *RequestTaskSet) globalRetriesRemaining() int32 {
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

func (r *RequestTaskSet) response(running bool) *steps.ExecuteResponse {
	resp := &steps.ExecuteResponse{
		TaskResults:         r.taskResults(),
		ConsolidatedResults: r.results(),
		State: &test_platform.TaskState{
			Verdict:   r.verdict(),
			LifeCycle: r.lifecycle(running),
		},
	}
	return resp
}

func (r *RequestTaskSet) lifecycle(running bool) test_platform.TaskState_LifeCycle {
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

func (r *RequestTaskSet) verdict() test_platform.TaskState_Verdict {
	v := test_platform.TaskState_VERDICT_PASSED
	if !r.Completed() {
		v = test_platform.TaskState_VERDICT_UNSPECIFIED
	}
	for _, t := range r.testTaskSets {
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

func (r *RequestTaskSet) results() []*steps.ExecuteResponse_ConsolidatedResult {
	rs := make([]*steps.ExecuteResponse_ConsolidatedResult, len(r.testTaskSets))
	for i, test := range r.testTaskSets {
		rs[i] = &steps.ExecuteResponse_ConsolidatedResult{
			Attempts: test.TaskResult(),
		}
	}
	return rs
}

func (r *RequestTaskSet) taskResults() []*steps.ExecuteResponse_TaskResult {
	results := r.results()
	var trs []*steps.ExecuteResponse_TaskResult
	for _, result := range results {
		trs = append(trs, result.Attempts...)
	}
	return trs
}
