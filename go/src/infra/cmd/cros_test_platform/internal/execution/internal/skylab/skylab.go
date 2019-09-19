// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package skylab implements logic necessary for Skylab execution of an
// ExecuteRequest.
package skylab

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"

	build_api "go.chromium.org/chromiumos/infra/proto/go/chromite/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/config"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/steps"
	swarming_api "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/swarming/proto/jsonrpc"

	"infra/cmd/cros_test_platform/internal/execution/internal/common"
	"infra/cmd/cros_test_platform/internal/execution/isolate"
	"infra/cmd/cros_test_platform/internal/execution/swarming"
	"infra/libs/skylab/inventory"
	"infra/libs/skylab/inventory/autotest/labels"
	"infra/libs/skylab/request"
	"infra/libs/skylab/worker"
)

// TaskSet encapsulates the running state of a set of tasks, to satisfy
// a Skylab Execution.
type TaskSet struct {
	testRuns []*testRun
	params   *test_platform.Request_Params
	retries  int32
	// complete indicates that the TaskSet ran to completion of all tasks.
	complete bool
	// running indicates that the TaskSet is still running.
	running bool
}

type testRun struct {
	invocation *steps.EnumerationResponse_AutotestInvocation
	Args       request.Args
	// Do not read from or mutate attempts without holding appropriate lock
	// on l.
	attempts []*attempt
	l        sync.RWMutex
}

func newTestRun(ctx context.Context, invocation *steps.EnumerationResponse_AutotestInvocation, params *test_platform.Request_Params, workerConfig *config.Config_SkylabWorker, parentTaskID string) (*testRun, error) {
	t := testRun{invocation: invocation}
	a, err := t.RequestArgs(ctx, params, workerConfig, parentTaskID)
	if err != nil {
		return nil, errors.Annotate(err, "new test run").Err()
	}
	t.Args = a
	return &t, nil
}

func (t *testRun) addAttempt(a *attempt) {
	t.l.Lock()
	t.attempts = append(t.attempts, a)
	t.l.Unlock()
}

func (t *testRun) getAttempts() []*attempt {
	t.l.RLock()
	val := t.attempts
	t.l.RUnlock()
	return val
}

func (t *testRun) RequestArgs(ctx context.Context, params *test_platform.Request_Params, workerConfig *config.Config_SkylabWorker, parentTaskID string) (request.Args, error) {
	isClient, err := t.isClientTest()
	if err != nil {
		return request.Args{}, errors.Annotate(err, "create request args").Err()
	}

	provisionableDimensions, err := toProvisionableDimensions(params.SoftwareDependencies)
	if err != nil {
		return request.Args{}, errors.Annotate(err, "create request args").Err()
	}

	timeout, err := toTimeout(params)
	if err != nil {
		return request.Args{}, errors.Annotate(err, "create request args").Err()
	}

	kv := getKeyvals(params, parentTaskID)
	t.updateWithInvocationKeyvals(kv)
	t.addKeyvalsForDisplayName(ctx, kv, params)

	cmd := &worker.Command{
		TaskName:        t.invocation.Test.Name,
		ClientTest:      isClient,
		OutputToIsolate: true,
		TestArgs:        t.invocation.TestArgs,
		Keyvals:         kv,
	}
	cmd.Config(wrap(workerConfig))

	labels, err := toInventoryLabels(params, t.invocation.Test.Dependencies)
	if err != nil {
		return request.Args{}, errors.Annotate(err, "create request args").Err()
	}

	args := request.Args{
		Cmd:                     *cmd,
		SchedulableLabels:       *labels,
		Dimensions:              params.GetFreeformAttributes().GetSwarmingDimensions(),
		ParentTaskID:            parentTaskID,
		Priority:                params.GetScheduling().GetPriority(),
		ProvisionableDimensions: provisionableDimensions,
		SwarmingTags:            swarmingTags(cmd, workerConfig, params),
		Timeout:                 timeout,
	}

	return args, nil
}

func (t *testRun) addKeyvalsForDisplayName(ctx context.Context, kv map[string]string, params *test_platform.Request_Params) {
	const displayNameKey = "label"

	if t.invocation.DisplayName != "" {
		kv[displayNameKey] = t.invocation.DisplayName
		return
	}

	testName := t.invocation.GetTest().GetName()
	kv[displayNameKey] = constructDisplayNameFromRequestParams(ctx, kv, params, testName)
}

const (
	suiteKey         = "suite"
	defaultSuiteName = "cros_test_platform"
)

// This is a hack to satisfy tko/parse's insistence on parsing the display name
// (aka "label") keyval to obtain semantic information about the request.
// TODO(crbug.com/1003490): Drop this once result reporting is updated to stop
// parsing the "label" keyval.
func constructDisplayNameFromRequestParams(ctx context.Context, kv map[string]string, params *test_platform.Request_Params, testName string) string {
	builds, err := common.ExtractBuilds(params.SoftwareDependencies)
	if err != nil {
		logging.Warningf(ctx,
			"Failed to get build due to error %s\n Defaulting to test name as display name: %s",
			err.Error(), testName)
		return testName
	}

	build := builds.ChromeOS
	if build == "" {
		logging.Warningf(ctx, "Build missing. Defaulting to test name as display name: %s", testName)
		return testName
	}

	suite := kv[suiteKey]
	if suite == "" {
		suite = defaultSuiteName
	}

	return build + "/" + suite + "/" + testName
}

func (t *testRun) updateWithInvocationKeyvals(kv map[string]string) {
	for k, v := range t.invocation.GetResultKeyvals() {
		if _, ok := kv[k]; !ok {
			kv[k] = v
		}
	}
}

func getKeyvals(params *test_platform.Request_Params, parentTaskID string) map[string]string {
	keyvals := params.GetDecorations().GetAutotestKeyvals()
	if keyvals == nil {
		keyvals = make(map[string]string)
	}
	if parentTaskID != "" {
		// This keyval is inspected by some downstream results consumers such as
		// goldeneye and stainless.
		// TODO(akeshet): Consider whether parameter-specified parent_job_id
		// should be respected if it was specified.
		keyvals["parent_job_id"] = parentTaskID
	}
	return keyvals
}

func (t *testRun) isClientTest() (bool, error) {
	isClient, ok := isClientTest[t.invocation.Test.ExecutionEnvironment]
	if !ok {
		return false, errors.Reason("unknown exec environment %s", t.invocation.Test.ExecutionEnvironment).Err()
	}
	return isClient, nil
}

func swarmingTags(cmd *worker.Command, conf *config.Config_SkylabWorker, params *test_platform.Request_Params) []string {
	tags := []string{
		"luci_project:" + conf.LuciProject,
		"log_location:" + cmd.LogDogAnnotationURL,
	}
	if qa := params.GetScheduling().GetQuotaAccount(); qa != "" {
		tags = append(tags, "qs_account:"+qa)
	}
	// TODO(akeshet): Consider whether to ban qs_account, luci_project, log_location,
	// and other "special tags" from being client-specified here.
	tags = append(tags, params.GetDecorations().GetTags()...)
	return tags
}

type attempt struct {
	taskID string
	state  jsonrpc.TaskState
	// Note: If we ever begin supporting other harnesses's result formats
	// then this field will change to a *skylab_test_runner.Result.
	// For now, the autotest-specific variant is more convenient.
	autotestResult *skylab_test_runner.Result_Autotest
}

func (a *attempt) complete() bool {
	return a.autotestResult != nil
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
		testRuns: testRuns,
		params:   params,
		running:  true,
	}, nil
}

// LaunchAndWait launches a skylab execution and waits for it to complete,
// polling for new results periodically, and retrying tests that need retry,
// based on retry policy.
//
// If the supplied context is cancelled prior to completion, or some other error
// is encountered, this method returns whatever partial execution response
// was visible to it prior to that error.
func (r *TaskSet) LaunchAndWait(ctx context.Context, client swarming.Client, gf isolate.GetterFactory) error {
	defer func() { r.running = false }()

	if err := r.launchAll(ctx, client); err != nil {
		return err
	}

	return r.wait(ctx, client, gf)
}

var isClientTest = map[build_api.AutotestTest_ExecutionEnvironment]bool{
	build_api.AutotestTest_EXECUTION_ENVIRONMENT_CLIENT: true,
	build_api.AutotestTest_EXECUTION_ENVIRONMENT_SERVER: false,
}

func (r *TaskSet) launchAll(ctx context.Context, client swarming.Client) error {
	for _, testRun := range r.testRuns {
		if err := r.launchSingle(ctx, client, testRun); err != nil {
			return err
		}
	}
	return nil
}

func (r *TaskSet) launchSingle(ctx context.Context, client swarming.Client, tr *testRun) error {
	req, err := tr.Args.SwarmingNewTaskRequest()
	if err != nil {
		return errors.Annotate(err, "launch test named %s", tr.invocation.Test.Name).Err()
	}

	resp, err := client.CreateTask(ctx, req)
	if err != nil {
		return errors.Annotate(err, "launch test named %s", tr.invocation.Test.Name).Err()
	}

	logging.Infof(ctx, "Launched test named %s as task %s", tr.invocation.Test.Name, client.GetTaskURL(resp.TaskId))

	tr.addAttempt(&attempt{taskID: resp.TaskId})
	return nil
}

func (r *TaskSet) wait(ctx context.Context, swarming swarming.Client, gf isolate.GetterFactory) error {
	for {
		complete, err := r.tick(ctx, swarming, gf)
		if complete || err != nil {
			r.complete = complete
			return err
		}

		select {
		case <-ctx.Done():
			return errors.Annotate(ctx.Err(), "wait for tests").Err()
		case <-clock.After(ctx, 15*time.Second):
		}
	}
}

func (r *TaskSet) tick(ctx context.Context, client swarming.Client, gf isolate.GetterFactory) (complete bool, err error) {
	complete = true

	for _, testRun := range r.testRuns {
		attempts := testRun.getAttempts()
		latestAttempt := attempts[len(attempts)-1]
		if latestAttempt.complete() {
			continue
		}

		if err := r.fetchResults(ctx, latestAttempt, client, gf); err != nil {
			return false, errors.Annotate(err, "tick for task %s", latestAttempt.taskID).Err()
		}

		if !latestAttempt.complete() {
			complete = false
			continue
		}

		logging.Debugf(ctx, "Task %s (%s) completed with verdict %s", latestAttempt.taskID, testRun.invocation.Test.Name, flattenToVerdict(latestAttempt.autotestResult.GetTestCases()))

		shouldRetry, err := r.shouldRetry(testRun)
		if err != nil {
			return false, errors.Annotate(err, "tick for task %s", latestAttempt.taskID).Err()
		}
		if shouldRetry {
			complete = false
			logging.Debugf(ctx, "Retrying %s", testRun.invocation.Test.Name)
			updateLogDogURL(&testRun.Args)
			if err := r.launchSingle(ctx, client, testRun); err != nil {
				return false, errors.Annotate(err, "tick for task %s: retry test", latestAttempt.taskID).Err()
			}
			r.retries++
		} else {
			logging.Debugf(ctx, "Not retrying %s", testRun.invocation.Test.Name)
		}
	}

	return complete, nil
}

var logdogURLPattern = regexp.MustCompile(`logdog\://([\w-_]*)/([\w-_]*)/skylab/[\w-_]*/\+/annotations`)

// updateLogDogURL is a terrible hack to assist refactoring this package.
// TODO(crbug.com/1003874, pprabhu) Drop this hack at the top of the refactor
// stack.
func updateLogDogURL(a *request.Args) {
	parts := logdogURLPattern.FindStringSubmatch(a.Cmd.LogDogAnnotationURL)
	if len(parts) != 3 {
		panic(fmt.Sprintf("Malformed logdog URL %s", a.Cmd.LogDogAnnotationURL))
	}
	url := fmt.Sprintf("logdog://%s/%s/skylab/%s/+/annotations", parts[1], parts[2], uuid.New().String())
	(&a.Cmd).LogDogAnnotationURL = url

	// Clone tags slice before modification.
	tags := make([]string, len(a.SwarmingTags))
	copy(tags, a.SwarmingTags)
	a.SwarmingTags = tags
	for i, t := range a.SwarmingTags {
		if strings.HasPrefix(t, "log_location:") {
			a.SwarmingTags[i] = "log_location:" + url
		}
	}
}

// fetchResults fetches the latest swarming and isolate state of the given attempt,
// and updates the attempt accordingly.
func (r *TaskSet) fetchResults(ctx context.Context, a *attempt, client swarming.Client, gf isolate.GetterFactory) error {
	results, err := client.GetResults(ctx, []string{a.taskID})
	if err != nil {
		return errors.Annotate(err, "fetch results").Err()
	}

	result, err := unpackResult(results, a.taskID)
	if err != nil {
		return errors.Annotate(err, "fetch results").Err()
	}

	state, err := swarming.AsTaskState(result.State)
	if err != nil {
		return errors.Annotate(err, "fetch results").Err()
	}
	a.state = state

	switch {
	// Task ran to completion.
	case swarming.CompletedTaskStates[state]:
		r, err := getAutotestResult(ctx, result, gf)
		if err != nil {
			logging.Debugf(ctx, "failed to fetch autotest results for task %s due to error '%s', treating its results as incomplete (failure)", a.taskID, err.Error())
			r = &skylab_test_runner.Result_Autotest{Incomplete: true}
		}
		a.autotestResult = r
	// Task no longer running, but didn't run to completion.
	case !swarming.UnfinishedTaskStates[state]:
		a.autotestResult = &skylab_test_runner.Result_Autotest{Incomplete: true}
	// Task is still running.
	default:
	}

	return nil
}

// shouldRetry computes if the given testRun should be retried.
//
// This will panic if the testRun has never been launched.
func (r *TaskSet) shouldRetry(tr *testRun) (bool, error) {
	if r.params.Retry == nil {
		return false, nil
	}
	if !r.params.Retry.Allow {
		return false, nil
	}
	max := r.params.Retry.Max
	if max != 0 && r.retries >= max {
		return false, nil
	}
	if !tr.invocation.Test.AllowRetries {
		return false, nil
	}
	attempts := tr.getAttempts()
	attemptCount := len(attempts)
	if attemptCount < 1 {
		return false, errors.Reason("should retry: can't retry a never-tried test").Err()
	}
	// Allow up to MaxRetries + 1 total attempts of a given test.
	if attemptCount > int(tr.invocation.Test.MaxRetries) {
		return false, nil
	}
	latestAttempt := attempts[attemptCount-1]
	switch verdict := flattenToVerdict(latestAttempt.autotestResult.GetTestCases()); verdict {
	case test_platform.TaskState_VERDICT_UNSPECIFIED:
		fallthrough
	case test_platform.TaskState_VERDICT_FAILED:
		return true, nil
	case test_platform.TaskState_VERDICT_NO_VERDICT:
		fallthrough
	case test_platform.TaskState_VERDICT_PASSED:
		return false, nil
	default:
		return false, errors.Reason("should retry: unknown verdict %s", verdict.String()).Err()
	}
}

func toInventoryLabels(params *test_platform.Request_Params, deps []*build_api.AutotestTaskDependency) (*inventory.SchedulableLabels, error) {
	flatDims := make([]string, len(deps))
	for i, dep := range deps {
		flatDims[i] = dep.Label
	}

	inv := labels.Revert(flatDims)

	if params.GetSoftwareAttributes().GetBuildTarget() != nil {
		*inv.Board = params.SoftwareAttributes.BuildTarget.Name
	}
	if params.GetHardwareAttributes().GetModel() != "" {
		*inv.Model = params.HardwareAttributes.Model
	}

	if p := params.GetScheduling().GetPool(); p != nil {
		switch v := p.(type) {
		case *test_platform.Request_Params_Scheduling_ManagedPool_:
			pool, ok := poolMap[v.ManagedPool]
			if !ok {
				return nil, errors.Reason("unknown managed pool %s", v.ManagedPool.String()).Err()
			}
			inv.CriticalPools = append(inv.CriticalPools, pool)
		case *test_platform.Request_Params_Scheduling_UnmanagedPool:
			inv.SelfServePools = append(inv.SelfServePools, v.UnmanagedPool)
		case *test_platform.Request_Params_Scheduling_QuotaAccount:
			inv.CriticalPools = append(inv.CriticalPools, inventory.SchedulableLabels_DUT_POOL_QUOTA)
		default:
			panic(fmt.Sprintf("unhandled scheduling type %#v", p))
		}
	}

	return inv, nil
}

const (
	// These prefixes are interpreted by autotest's provisioning behavior;
	// they are defined in the autotest repo, at utils/labellib.py
	prefixChromeOS   = "cros-version"
	prefixFirmwareRO = "fwro-version"
	prefixFirmwareRW = "fwrw-version"
)

func toProvisionableDimensions(deps []*test_platform.Request_Params_SoftwareDependency) ([]string, error) {
	builds, err := common.ExtractBuilds(deps)
	if err != nil {
		return nil, errors.Annotate(err, "get provisionable dimensions").Err()
	}

	var dims []string
	if b := builds.ChromeOS; b != "" {
		dims = append(dims, "provisionable-"+prefixChromeOS+":"+b)
	}
	if b := builds.FirmwareRO; b != "" {
		dims = append(dims, "provisionable-"+prefixFirmwareRO+":"+b)
	}
	if b := builds.FirmwareRW; b != "" {
		dims = append(dims, "provisionable-"+prefixFirmwareRW+":"+b)
	}

	return dims, nil
}

func toTimeout(params *test_platform.Request_Params) (time.Duration, error) {
	if params.Time == nil {
		return 0, errors.Reason("get timeout: nil params.time").Err()
	}
	duration, err := ptypes.Duration(params.Time.MaximumDuration)
	if err != nil {
		return 0, errors.Annotate(err, "get timeout").Err()
	}
	return duration, nil
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

// Response constructs a response based on the current state of the
// TaskSet.
func (r *TaskSet) Response(swarming swarming.URLer) *steps.ExecuteResponse {
	resp := &steps.ExecuteResponse{}
	resp.TaskResults = toTaskResults(r.testRuns, swarming)

	var verdict test_platform.TaskState_Verdict
	var lifecycle test_platform.TaskState_LifeCycle

	switch {
	case r.complete:
		// The default verdict for a completed TaskSet is passed; if any tasks
		// failed, they will overwrite this below.
		verdict = test_platform.TaskState_VERDICT_PASSED
		lifecycle = test_platform.TaskState_LIFE_CYCLE_COMPLETED
	case r.running:
		lifecycle = test_platform.TaskState_LIFE_CYCLE_RUNNING
	default:
		// TODO(akeshet): The task set is neither running nor complete, so it
		// was cancelled due to an error while in flight. It's not clear yet
		// if this is the right lifecycle mapping for this state.
		lifecycle = test_platform.TaskState_LIFE_CYCLE_ABORTED
	}

	for _, t := range resp.TaskResults {
		switch t.State.Verdict {
		case test_platform.TaskState_VERDICT_UNSPECIFIED:
			fallthrough
		case test_platform.TaskState_VERDICT_FAILED:
			verdict = test_platform.TaskState_VERDICT_FAILED
			break
		}
	}

	resp.State = &test_platform.TaskState{
		Verdict:   verdict,
		LifeCycle: lifecycle,
	}

	return resp
}
