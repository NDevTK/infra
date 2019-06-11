// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package skylab_test

import (
	"context"
	"fmt"
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
	createCalls int
	getCalls    int
}

func (f *fakeSwarming) CreateTask(ctx context.Context, req *swarming_api.SwarmingRpcsNewTaskRequest) (*swarming_api.SwarmingRpcsTaskRequestMetadata, error) {
	f.nextID++
	f.createCalls++
	resp := &swarming_api.SwarmingRpcsTaskRequestMetadata{TaskId: fmt.Sprintf("task%d", f.nextID)}
	return resp, nil
}

func (f *fakeSwarming) GetResults(ctx context.Context, IDs []string) ([]*swarming_api.SwarmingRpcsTaskResult, error) {
	f.getCalls++
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

func newFakeSwarming() *fakeSwarming {
	return &fakeSwarming{nextState: "COMPLETED"}
}

func TestLaunchAndWaitSingleTest(t *testing.T) {
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
