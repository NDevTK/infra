// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package execution

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/config"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/steps"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/exe"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	trservice "infra/cmd/cros_test_platform/internal/execution/testrunner/service"
	ufsapi "infra/unifiedfleet/api/v1/rpc"
)

// Args bundles together the arguments for an execution.
type Args struct {
	// Used to get inputs from and send updates to a buildbucket Build.
	// See https://godoc.org/go.chromium.org/luci/luciexe
	Build *bbpb.Build
	Send  exe.BuildSender

	Request *steps.ExecuteRequests

	WorkerConfig *config.Config_SkylabWorker
	ParentTaskID string
	Deadline     time.Time
}

// Run runs an execution until success.
//
// Run may be aborted by cancelling the supplied context.
func Run(ctx context.Context, c trservice.Client, args Args) (map[string]*steps.ExecuteResponse, error) {
	// Build may be updated as each of the task sets is Close()ed by a deferred
	// function. Send() one last time to capture those changes.
	defer args.Send()

	ts := make(map[string]*RequestTaskSet)
	for t, r := range args.Request.GetTaggedRequests() {
		var err error
		requestTaskSet, err := NewRequestTaskSet(
			t,
			args.Build,
			args.WorkerConfig,
			&TaskSetConfig{
				ParentTaskID:        args.ParentTaskID,
				ParentBuildID:       args.Request.GetBuild().GetId(),
				RequestUID:          constructRequestUID(args.Request.GetBuild().GetId(), t),
				Deadline:            args.Deadline,
				StatusUpdateChannel: r.GetConfig().GetTestRunner().GetBbStatusUpdateChannel(),
			},
			r.RequestParams,
			r.Enumeration.AutotestInvocations,
		)
		if err != nil {
			return nil, err
		}
		var validTest bool
		for _, iid := range requestTaskSet.invocationIDs {
			ts := requestTaskSet.getInvocationResponse(iid)
			var image string
			for _, dep := range r.RequestParams.SoftwareDependencies {
				switch d := dep.Dep.(type) {
				case *test_platform.Request_Params_SoftwareDependency_ChromeosBuild:
					image = d.ChromeosBuild
				}
			}
			board := ""
			model := ""
			if r.RequestParams.SoftwareAttributes != nil && r.RequestParams.SoftwareAttributes.BuildTarget != nil {
				board = r.RequestParams.SoftwareAttributes.BuildTarget.Name
			}
			if r.RequestParams.HardwareAttributes != nil {
				model = r.RequestParams.HardwareAttributes.Model
			}
			validTest, err = verifyFleetTestsPolicy(ctx, c, board, model, ts.Name, image)
			if !validTest {
				logging.Errorf(ctx, "Fleet Validation failed for test due to error %v for test request %v, failing test run.", requestTaskSet, err)
				return nil, fmt.Errorf("Fleet Validation failed for test %v due to error %v", requestTaskSet, err)
			}
		}

		ts[t] = requestTaskSet
		defer ts[t].Close()

		// A large number of tasks is created in the beginning as a new task is
		// created for each invocation in the request.
		// We update the build more frequently in the beginning to reflect these
		// tasks on the UI sooner.
		args.Send()
	}

	r := runner{
		requestTaskSets: ts,
		send:            args.Send,
	}
	err := r.LaunchAndWait(ctx, c)
	if isFatalError(ctx, err) {
		return nil, err
	}
	return r.Responses(), err
}

func isFatalError(ctx context.Context, err error) bool {
	return err != nil && !IsGlobalTimeoutError(ctx, err)
}

// IsGlobalTimeoutError checks whether an error returned from an execution is
// a result fo hitting the overall timeout.
func IsGlobalTimeoutError(ctx context.Context, err error) bool {
	d, ok := ctx.Deadline()
	if !ok {
		logging.Infof(ctx, "Not a global timeout: no deadline set")
		return false
	}
	if !isDeadlineExceededError(err) {
		logging.Infof(ctx, "Not a global timeout: error is not a deadline exceeded error")
		return false
	}
	now := time.Now()
	if now.Before(d) {
		logging.Infof(ctx, "Not a global timeout: Current time (%s) is before deadline (%s)", now.String(), d.String())
		return false
	}
	return true
}

func isDeadlineExceededError(err error) bool {
	// The original error raised is context.DeadlineExceeded but the prpc client
	// library may transmute that into its own error type.
	return errors.Any(err, func(err error) bool {
		if err == context.DeadlineExceeded {
			return true
		}
		if s, ok := status.FromError(err); ok {
			return s.Code() == codes.DeadlineExceeded
		}
		return false
	})
}

// ctpRequestUIDTemplate is the template to generate the UID of
// a test plan run, a.k.a. CTP request.
const ctpRequestUIDTemplate = "TestPlanRuns/%d/%s"

// runner manages task sets for multiple cros_test_platform requests.
type runner struct {
	requestTaskSets map[string]*RequestTaskSet
	send            exe.BuildSender
}

// LaunchAndWait launches a skylab execution and waits for it to complete,
// polling for new results periodically, and retrying tests that need retry,
// based on retry policy.
//
// If the supplied context is cancelled prior to completion, or some other error
// is encountered, this method returns whatever partial execution response
// was visible to it prior to that error.
func (r *runner) LaunchAndWait(ctx context.Context, c trservice.Client) error {
	if err := r.launchTasks(ctx, c); err != nil {
		return err
	}
	for {
		allDone, err := r.checkTasksAndRetry(ctx, c)

		// Each call to checkTasksAndRetry() potentially updates the Build.
		// We unconditionally send() the updated build so that we reflect the
		// update irrespective of abnormal exits.
		// Since this loop sleeps between iterations, the load generated on
		// the buildbucket service is bounded.
		r.send()

		if err != nil {
			return err
		}
		if allDone {
			return nil
		}

		select {
		case <-ctx.Done():
			// A timeout while waiting for tests to complete is reported as
			// aborts when summarizing individual tests' results.
			// The execute step completes without errors.
			return nil
		case <-clock.After(ctx, 15*time.Second):
		}
	}
}

func (r *runner) launchTasks(ctx context.Context, c trservice.Client) error {
	for t, ts := range r.requestTaskSets {
		if err := ts.LaunchTasks(ctx, c); err != nil {
			return errors.Annotate(err, "launch tasks for %s", t).Err()
		}
	}
	return nil
}

// Returns whether all tasks are complete (so future calls to this function are
// unnecessary)
func (r *runner) checkTasksAndRetry(ctx context.Context, c trservice.Client) (bool, error) {
	allDone := true
	for t, ts := range r.requestTaskSets {
		c, err := ts.CheckTasksAndRetry(ctx, c)
		if err != nil {
			return false, errors.Annotate(err, "check tasks and retry for %s", t).Err()
		}
		allDone = allDone && c
	}
	return allDone, nil
}

// Responses constructs responses for each request managed by the runner.
func (r *runner) Responses() map[string]*steps.ExecuteResponse {
	resps := make(map[string]*steps.ExecuteResponse)
	for t, ts := range r.requestTaskSets {
		resps[t] = ts.Response()
		// The test hasn't completed, but we're not waiting for it to complete
		// anymore.
		if resps[t].GetState().LifeCycle == test_platform.TaskState_LIFE_CYCLE_RUNNING {
			resps[t].State.LifeCycle = test_platform.TaskState_LIFE_CYCLE_ABORTED
		}
	}
	return resps
}

func constructRequestUID(buildID int64, key string) string {
	return fmt.Sprintf(ctpRequestUIDTemplate, buildID, key)
}

// verifyFleetTestsPolicy validate tests based on fleet-side permission check.
//
// This method calls UFS CheckFleetTestsPolicy RPC for a testName, board, image and model combination.
func verifyFleetTestsPolicy(ctx context.Context, client trservice.Client, board string, model string,
	testName string, image string) (bool, error) {
	resp, err := client.CheckFleetTestsPolicy(ctx, &ufsapi.CheckFleetTestsPolicyRequest{
		TestName: testName,
		Board:    board,
		Model:    model,
		Image:    image,
	})
	if err != nil {
		return false, err
	}
	if resp.TestStatus.Code == ufsapi.TestStatus_OK {
		return true, nil
	} else {
		return false, fmt.Errorf("%s - %s", resp.TestStatus.Code.String(), resp.TestStatus.Message)
	}
}
