// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package skylab_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/steps"
	swarming_api "go.chromium.org/luci/common/api/swarming/swarming/v1"

	"infra/cmd/cros_test_platform/internal/skylab"
)

// fakeSwarming implements skylab.Swarming
type fakeSwarming struct {
	nextID      int
	nextState   string
	nextError   error
	callback    func()
	createCalls int
	getCalls    int
}

func (f *fakeSwarming) CreateTask(ctx context.Context, req *swarming_api.SwarmingRpcsNewTaskRequest) (*swarming_api.SwarmingRpcsTaskRequestMetadata, error) {
	defer f.callback()
	f.nextID++
	f.createCalls++
	if f.nextError != nil {
		return nil, f.nextError
	}
	resp := &swarming_api.SwarmingRpcsTaskRequestMetadata{TaskId: fmt.Sprintf("task%d", f.nextID)}
	return resp, nil
}

func (f *fakeSwarming) GetResults(ctx context.Context, IDs []string) ([]*swarming_api.SwarmingRpcsTaskResult, error) {
	defer f.callback()
	f.getCalls++
	if f.nextError != nil {
		return nil, f.nextError
	}
	results := make([]*swarming_api.SwarmingRpcsTaskResult, len(IDs))
	for i, taskID := range IDs {
		results[i] = &swarming_api.SwarmingRpcsTaskResult{TaskId: taskID, State: f.nextState}
	}
	return results, nil
}

// setNextState causes this fake to start returning the given state of all future
// GetResults calls.
func (f *fakeSwarming) setNextState(state string) {
	f.nextState = state
}

// setNextError causes this fake to start returning the given error on all
// future API calls.
func (f *fakeSwarming) setNextError(err error) {
	f.nextError = err
}

// setCallback causes this fake to call the given callback function, immediately
// prior to the return of every future API call.
func (f *fakeSwarming) setCallback(fn func()) {
	f.callback = fn
}

func newFakeSwarming() *fakeSwarming {
	return &fakeSwarming{
		nextState: "COMPLETED",
		callback:  func() {},
	}
}

func TestLaunchAndWaitTest(t *testing.T) {
	Convey("Given two enumerated test", t, func() {
		ctx := context.Background()

		swarming := newFakeSwarming()

		var tests []*steps.EnumerationResponse_Test
		tests = append(tests, &steps.EnumerationResponse_Test{}, &steps.EnumerationResponse_Test{})

		Convey("when running a skylab execution", func() {
			run := skylab.NewRun(tests)

			err := run.LaunchAndWait(ctx, swarming)
			So(err, ShouldBeNil)

			resp := run.Response("")
			So(resp, ShouldNotBeNil)

			Convey("then results for all tests are reflected.", func() {
				So(resp.TaskResults, ShouldHaveLength, 2)
				So(resp.Complete, ShouldBeTrue)
			})
			Convey("then the expected number of external swarming calls are made.", func() {
				So(swarming.getCalls, ShouldEqual, 2)
				So(swarming.createCalls, ShouldEqual, 2)
			})
		})
	})
}

func TestIncompleteWait(t *testing.T) {
	Convey("Given a run that is cancelled while running, error and response reflect cancellation.", t, func() {
		ctx, cancel := context.WithCancel(context.Background())

		swarming := newFakeSwarming()
		swarming.setNextState("RUNNING")

		tests := []*steps.EnumerationResponse_Test{{}}
		run := skylab.NewRun(tests)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var err error
		go func() {
			err = run.LaunchAndWait(ctx, swarming)
			wg.Done()
		}()

		cancel()
		wg.Wait()

		So(err.Error(), ShouldContainSubstring, context.Canceled.Error())

		resp := run.Response("")
		So(resp.Complete, ShouldBeFalse)
	})
}

func TestServiceError(t *testing.T) {
	Convey("Given a single enumerated test", t, func() {
		ctx := context.Background()
		swarming := newFakeSwarming()

		tests := []*steps.EnumerationResponse_Test{{}}
		run := skylab.NewRun(tests)

		Convey("when the swarming service immediately returns errors, that error is surfaced as a launch error.", func() {
			swarming.setNextError(fmt.Errorf("foo error"))
			err := run.LaunchAndWait(ctx, swarming)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "launch test")
			So(err.Error(), ShouldContainSubstring, "foo error")
		})

		Convey("when the swarming service starts returning errors after the initial launch calls, that errors is surfaced as a wait error.", func() {
			swarming.setCallback(func() {
				swarming.setNextError(fmt.Errorf("foo error"))
			})
			err := run.LaunchAndWait(ctx, swarming)
			So(err.Error(), ShouldContainSubstring, "wait for tests")
			So(err.Error(), ShouldContainSubstring, "foo error")
		})
	})
}

func TestTaskURL(t *testing.T) {
	Convey("Given a single enumerated test running to completion, its task URL is well formed.", t, func() {
		ctx := context.Background()
		swarming := newFakeSwarming()
		tests := []*steps.EnumerationResponse_Test{{}}
		run := skylab.NewRun(tests)
		run.LaunchAndWait(ctx, swarming)

		swarming_service := "https://foo.bar.com/"
		resp := run.Response(swarming_service)
		So(resp.TaskResults, ShouldHaveLength, 1)
		taskURL := resp.TaskResults[0].TaskUrl
		taskID := resp.TaskResults[0].TaskId
		So(taskURL, ShouldStartWith, swarming_service)
		So(taskURL, ShouldEndWith, taskID)
	})
}
