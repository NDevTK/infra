// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package skylab_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/duration"
	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/chromiumos/infra/proto/go/chromite/api"
	build_api "go.chromium.org/chromiumos/infra/proto/go/chromite/api"
	"go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/config"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/steps"
	swarming_api "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/isolated"
	"go.chromium.org/luci/swarming/proto/jsonrpc"

	"infra/cmd/cros_test_platform/internal/execution/internal/skylab"
	"infra/cmd/cros_test_platform/internal/execution/isolate"
)

// fakeSwarming implements skylab.Swarming
type fakeSwarming struct {
	nextID      int
	nextState   jsonrpc.TaskState
	nextError   error
	callback    func()
	server      string
	createCalls []*swarming_api.SwarmingRpcsNewTaskRequest
	getCalls    [][]string
	hasRef      bool
	botExists   bool
}

func (f *fakeSwarming) CreateTask(ctx context.Context, req *swarming_api.SwarmingRpcsNewTaskRequest) (*swarming_api.SwarmingRpcsTaskRequestMetadata, error) {
	defer f.callback()
	f.nextID++
	f.createCalls = append(f.createCalls, req)
	if f.nextError != nil {
		return nil, f.nextError
	}
	resp := &swarming_api.SwarmingRpcsTaskRequestMetadata{TaskId: fmt.Sprintf("task%d", f.nextID)}
	return resp, nil
}

func (f *fakeSwarming) GetResults(ctx context.Context, IDs []string) ([]*swarming_api.SwarmingRpcsTaskResult, error) {
	defer f.callback()
	f.getCalls = append(f.getCalls, IDs)
	if f.nextError != nil {
		return nil, f.nextError
	}

	var ref *swarming_api.SwarmingRpcsFilesRef
	if f.hasRef {
		ref = &swarming_api.SwarmingRpcsFilesRef{}
	}

	results := make([]*swarming_api.SwarmingRpcsTaskResult, len(IDs))
	for i, taskID := range IDs {
		results[i] = &swarming_api.SwarmingRpcsTaskResult{
			TaskId:     taskID,
			State:      jsonrpc.TaskState_name[int32(f.nextState)],
			OutputsRef: ref,
		}
	}
	return results, nil
}

func (f *fakeSwarming) BotExists(ctx context.Context, dims []*swarming_api.SwarmingRpcsStringPair) (bool, error) {
	return f.botExists, nil
}

func (f *fakeSwarming) SetCannedBotExistsResponse(b bool) {
	f.botExists = b
}

func (f *fakeSwarming) GetTaskURL(taskID string) string {
	// Note: this is not the true swarming task URL schema.
	return f.server + "/task=" + taskID
}

func (f *fakeSwarming) GetTaskOutputs(ctx context.Context, IDs []string) ([]*swarming_api.SwarmingRpcsTaskOutput, error) {
	return nil, nil
}

// setTaskState causes this fake to start returning the given state of all future
func (f *fakeSwarming) setTaskState(state jsonrpc.TaskState) {
	f.nextState = state
}

func (f *fakeSwarming) setHasOutputRef(has bool) {
	f.hasRef = has
}

// setError causes this fake to start returning the given error on all
// future API calls.
func (f *fakeSwarming) setError(err error) {
	f.nextError = err
}

// setCallback causes this fake to call the given callback function, immediately
// prior to the return of every future API call.
func (f *fakeSwarming) setCallback(fn func()) {
	f.callback = fn
}

func newFakeSwarming(server string) *fakeSwarming {
	return &fakeSwarming{
		nextState: jsonrpc.TaskState_COMPLETED,
		callback:  func() {},
		server:    server,
		hasRef:    true,
		botExists: true,
	}
}

type fakeGetter struct {
	autotestResultGenerator autotestResultGenerator
}

func (g *fakeGetter) GetFile(_ context.Context, _ isolated.HexDigest, _ string) ([]byte, error) {
	r := skylab_test_runner.Result{
		Harness: &skylab_test_runner.Result_AutotestResult{AutotestResult: g.autotestResultGenerator()},
	}
	m := &jsonpb.Marshaler{}
	s, err := m.MarshalToString(&r)
	if err != nil {
		panic(fmt.Sprintf("error when marshalling %#v: %s", r, err))
	}
	return []byte(s), nil
}

func (g *fakeGetter) SetAutotestResultGenerator(f autotestResultGenerator) {
	g.autotestResultGenerator = f
}

func newFakeGetter() *fakeGetter {
	f := &fakeGetter{}
	f.SetAutotestResultGenerator(autotestResultAlwaysPass)
	return f
}

func fakeGetterFactory(getter isolate.Getter) isolate.GetterFactory {
	return func(_ context.Context, _ string) (isolate.Getter, error) {
		return getter, nil
	}
}

func invocation(name string, args string, e build_api.AutotestTest_ExecutionEnvironment) *steps.EnumerationResponse_AutotestInvocation {
	return &steps.EnumerationResponse_AutotestInvocation{
		Test:     &build_api.AutotestTest{Name: name, ExecutionEnvironment: e},
		TestArgs: args,
	}
}

func clientTestInvocation(name string, args string) *steps.EnumerationResponse_AutotestInvocation {
	return &steps.EnumerationResponse_AutotestInvocation{
		Test: &build_api.AutotestTest{
			Name:                 name,
			ExecutionEnvironment: build_api.AutotestTest_EXECUTION_ENVIRONMENT_CLIENT,
		},
		TestArgs: args,
	}
}

func serverTestInvocation(name string, args string) *steps.EnumerationResponse_AutotestInvocation {
	return &steps.EnumerationResponse_AutotestInvocation{
		Test: &build_api.AutotestTest{
			Name:                 name,
			ExecutionEnvironment: build_api.AutotestTest_EXECUTION_ENVIRONMENT_SERVER,
		},
		TestArgs: args,
	}
}

func addAutotestDependency(inv *steps.EnumerationResponse_AutotestInvocation, dep string) {
	inv.Test.Dependencies = append(inv.Test.Dependencies, &api.AutotestTaskDependency{Label: dep})
}

func basicParams() *test_platform.Request_Params {
	return &test_platform.Request_Params{
		SoftwareAttributes: &test_platform.Request_Params_SoftwareAttributes{
			BuildTarget: &chromiumos.BuildTarget{Name: "foo-board"},
		},
		HardwareAttributes: &test_platform.Request_Params_HardwareAttributes{
			Model: "foo-model",
		},
		FreeformAttributes: &test_platform.Request_Params_FreeformAttributes{
			SwarmingDimensions: []string{"freeform-key:freeform-value"},
		},
		SoftwareDependencies: []*test_platform.Request_Params_SoftwareDependency{
			{
				Dep: &test_platform.Request_Params_SoftwareDependency_ChromeosBuild{ChromeosBuild: "foo-build"},
			},
			{
				Dep: &test_platform.Request_Params_SoftwareDependency_RoFirmwareBuild{RoFirmwareBuild: "foo-ro-firmware"},
			},
			{
				Dep: &test_platform.Request_Params_SoftwareDependency_RwFirmwareBuild{RwFirmwareBuild: "foo-rw-firmware"},
			},
		},
		Scheduling: &test_platform.Request_Params_Scheduling{
			Pool: &test_platform.Request_Params_Scheduling_ManagedPool_{
				ManagedPool: test_platform.Request_Params_Scheduling_MANAGED_POOL_CQ,
			},
			Priority: 79,
		},
		Time: &test_platform.Request_Params_Time{
			MaximumDuration: &duration.Duration{Seconds: 60},
		},
		Decorations: &test_platform.Request_Params_Decorations{
			AutotestKeyvals: map[string]string{"k1": "v1"},
			Tags:            []string{"foo-tag1", "foo-tag2"},
		},
	}
}

func basicConfig() *config.Config_SkylabWorker {
	return &config.Config_SkylabWorker{
		LuciProject: "foo-luci-project",
		LogDogHost:  "foo-logdog-host",
	}
}

func TestLaunchForNonExistentBot(t *testing.T) {
	Convey("Given one test invocation but non existent bots", t, func() {
		ctx := context.Background()

		swarming := newFakeSwarming("")
		swarming.SetCannedBotExistsResponse(false)
		getter := newFakeGetter()
		gf := fakeGetterFactory(getter)

		invs := []*steps.EnumerationResponse_AutotestInvocation{
			clientTestInvocation("", ""),
		}

		Convey("when running a skylab execution", func() {
			run, err := skylab.NewTaskSet(ctx, invs, basicParams(), basicConfig(), "foo-parent-task-id")
			So(err, ShouldBeNil)
			err = run.LaunchAndWait(ctx, swarming, gf)
			So(err, ShouldBeNil)

			resp := run.Response(swarming)
			So(resp, ShouldNotBeNil)

			Convey("then task result is complete with unspecified verdict.", func() {
				So(resp.TaskResults, ShouldHaveLength, 1)
				tr := resp.TaskResults[0]
				So(tr.State.LifeCycle, ShouldEqual, test_platform.TaskState_LIFE_CYCLE_REJECTED)
				So(tr.State.Verdict, ShouldEqual, test_platform.TaskState_VERDICT_UNSPECIFIED)

			})
			Convey("and overall result is complete with failed verdict.", func() {
				So(resp.State.LifeCycle, ShouldEqual, test_platform.TaskState_LIFE_CYCLE_COMPLETED)
				So(resp.State.Verdict, ShouldEqual, test_platform.TaskState_VERDICT_FAILED)
			})
			Convey("and no skylab swarming tasks are created.", func() {
				So(swarming.getCalls, ShouldHaveLength, 0)
				So(swarming.createCalls, ShouldHaveLength, 0)
			})
		})
	})
}

func TestLaunchAndWaitTest(t *testing.T) {
	Convey("Given two enumerated test", t, func() {
		ctx := context.Background()

		swarming := newFakeSwarming("")
		getter := newFakeGetter()
		gf := fakeGetterFactory(getter)

		var invs []*steps.EnumerationResponse_AutotestInvocation
		invs = append(invs, clientTestInvocation("", ""), clientTestInvocation("", ""))

		Convey("when running a skylab execution", func() {
			run, err := skylab.NewTaskSet(ctx, invs, basicParams(), basicConfig(), "foo-parent-task-id")
			So(err, ShouldBeNil)

			err = run.LaunchAndWait(ctx, swarming, gf)
			So(err, ShouldBeNil)

			resp := run.Response(swarming)
			So(resp, ShouldNotBeNil)

			Convey("then results for all tests are reflected.", func() {
				So(resp.TaskResults, ShouldHaveLength, 2)
				for _, tr := range resp.TaskResults {
					So(tr.State.LifeCycle, ShouldEqual, test_platform.TaskState_LIFE_CYCLE_COMPLETED)
				}
			})
			Convey("then the expected number of external swarming calls are made.", func() {
				So(swarming.getCalls, ShouldHaveLength, 2)
				So(swarming.createCalls, ShouldHaveLength, 2)
			})
		})
	})
}

// Note: the purpose of this test is the test the behavior when a parsed
// autotest result is not available from a task, because the task didn't run
// far enough to output one.
//
// For detailed tests on the handling of autotest test results, see result_test.go.
func TestTaskStates(t *testing.T) {
	Convey("Given a single test", t, func() {
		ctx := context.Background()

		var invs []*steps.EnumerationResponse_AutotestInvocation
		invs = append(invs, clientTestInvocation("", ""))

		cases := []struct {
			description     string
			swarmingState   jsonrpc.TaskState
			hasRef          bool
			expectTaskState *test_platform.TaskState
		}{
			{
				description:   "with expired state",
				swarmingState: jsonrpc.TaskState_EXPIRED,
				hasRef:        false,
				expectTaskState: &test_platform.TaskState{
					LifeCycle: test_platform.TaskState_LIFE_CYCLE_CANCELLED,
					Verdict:   test_platform.TaskState_VERDICT_FAILED,
				},
			},
			{
				description:   "with killed state",
				swarmingState: jsonrpc.TaskState_KILLED,
				hasRef:        false,
				expectTaskState: &test_platform.TaskState{
					LifeCycle: test_platform.TaskState_LIFE_CYCLE_ABORTED,
					Verdict:   test_platform.TaskState_VERDICT_FAILED,
				},
			},
			{
				description:   "with completed state",
				swarmingState: jsonrpc.TaskState_COMPLETED,
				hasRef:        true,
				expectTaskState: &test_platform.TaskState{
					LifeCycle: test_platform.TaskState_LIFE_CYCLE_COMPLETED,
					Verdict:   test_platform.TaskState_VERDICT_NO_VERDICT,
				},
			},
		}
		for _, c := range cases {
			Convey(c.description, func() {
				swarming := newFakeSwarming("")
				swarming.setTaskState(c.swarmingState)
				swarming.setHasOutputRef(c.hasRef)
				getter := newFakeGetter()
				getter.SetAutotestResultGenerator(autotestResultAlwaysEmpty)
				gf := fakeGetterFactory(getter)

				run, err := skylab.NewTaskSet(ctx, invs, basicParams(), basicConfig(), "foo-parent-task-id")
				So(err, ShouldBeNil)
				err = run.LaunchAndWait(ctx, swarming, gf)
				So(err, ShouldBeNil)

				Convey("then the task state is correct.", func() {
					resp := run.Response(swarming)
					So(resp.TaskResults, ShouldHaveLength, 1)
					So(resp.TaskResults[0].State, ShouldResemble, c.expectTaskState)
				})
			})
		}
	})
}

func TestServiceError(t *testing.T) {
	Convey("Given a single enumerated test", t, func() {
		ctx := context.Background()
		swarming := newFakeSwarming("")
		getter := newFakeGetter()
		gf := fakeGetterFactory(getter)

		invs := []*steps.EnumerationResponse_AutotestInvocation{clientTestInvocation("", "")}
		run, err := skylab.NewTaskSet(ctx, invs, basicParams(), basicConfig(), "foo-parent-task-id")
		So(err, ShouldBeNil)

		Convey("when the swarming service immediately returns errors, that error is surfaced as a launch error.", func() {
			swarming.setError(fmt.Errorf("foo error"))
			err := run.LaunchAndWait(ctx, swarming, gf)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "launch attempt")
			So(err.Error(), ShouldContainSubstring, "foo error")
		})

		Convey("when the swarming service starts returning errors after the initial launch calls, that errors is surfaced as a wait error.", func() {
			swarming.setCallback(func() {
				swarming.setError(fmt.Errorf("foo error"))
			})
			err := run.LaunchAndWait(ctx, swarming, gf)
			So(err.Error(), ShouldContainSubstring, "tick for task")
			So(err.Error(), ShouldContainSubstring, "foo error")
		})
	})
}

func TestTaskURL(t *testing.T) {
	Convey("Given a single enumerated test running to completion, its task URL is well formed.", t, func() {
		ctx := context.Background()
		swarming_service := "https://foo.bar.com/"
		swarming := newFakeSwarming(swarming_service)
		getter := newFakeGetter()
		gf := fakeGetterFactory(getter)

		invs := []*steps.EnumerationResponse_AutotestInvocation{clientTestInvocation("", "")}
		run, err := skylab.NewTaskSet(ctx, invs, basicParams(), basicConfig(), "foo-parent-task-id")
		So(err, ShouldBeNil)
		err = run.LaunchAndWait(ctx, swarming, gf)
		So(err, ShouldBeNil)

		resp := run.Response(swarming)
		So(resp.TaskResults, ShouldHaveLength, 1)
		taskURL := resp.TaskResults[0].TaskUrl
		So(taskURL, ShouldStartWith, swarming_service)
		So(taskURL, ShouldEndWith, "1")
	})
}

func TestIncompleteWait(t *testing.T) {
	Convey("Given a run that is cancelled while running, error and response reflect cancellation.", t, func() {
		ctx, cancel := context.WithCancel(context.Background())

		swarming := newFakeSwarming("")
		swarming.setTaskState(jsonrpc.TaskState_RUNNING)
		getter := newFakeGetter()
		gf := fakeGetterFactory(getter)

		invs := []*steps.EnumerationResponse_AutotestInvocation{clientTestInvocation("", "")}
		run, err := skylab.NewTaskSet(ctx, invs, basicParams(), basicConfig(), "foo-parent-task-id")
		So(err, ShouldBeNil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			err = run.LaunchAndWait(ctx, swarming, gf)
			wg.Done()
		}()

		cancel()
		wg.Wait()

		So(err.Error(), ShouldContainSubstring, context.Canceled.Error())

		resp := run.Response(swarming)
		So(resp, ShouldNotBeNil)
		So(resp.TaskResults, ShouldHaveLength, 1)
		So(resp.TaskResults[0].State.LifeCycle, ShouldEqual, test_platform.TaskState_LIFE_CYCLE_RUNNING)
		// TODO(akeshet): Ensure that response either reflects the error or
		// has an incomplete flag, once that part of the response proto is
		// defined.
	})
}

func TestRequestArguments(t *testing.T) {
	Convey("Given a server test with autotest labels", t, func() {
		ctx := context.Background()
		swarming := newFakeSwarming("")
		getter := newFakeGetter()
		gf := fakeGetterFactory(getter)

		inv := serverTestInvocation("name1", "foo-arg1 foo-arg2")
		addAutotestDependency(inv, "cr50:pvt")
		inv.DisplayName = "given_name"
		invs := []*steps.EnumerationResponse_AutotestInvocation{inv}

		run, err := skylab.NewTaskSet(ctx, invs, basicParams(), basicConfig(), "foo-parent-task-id")
		So(err, ShouldBeNil)
		err = run.LaunchAndWait(ctx, swarming, gf)
		So(err, ShouldBeNil)

		Convey("the launched task request should have correct parameters.", func() {
			So(swarming.createCalls, ShouldHaveLength, 1)
			create := swarming.createCalls[0]
			So(create.TaskSlices, ShouldHaveLength, 2)

			So(create.Tags, ShouldContain, "luci_project:foo-luci-project")
			So(create.Tags, ShouldContain, "foo-tag1")
			So(create.Tags, ShouldContain, "foo-tag2")
			So(create.ParentTaskId, ShouldEqual, "foo-parent-task-id")

			So(create.Priority, ShouldEqual, 79)

			prefix := "log_location:"
			var logdogURL string
			matchingTags := 0
			for _, tag := range create.Tags {
				if strings.HasPrefix(tag, prefix) {
					matchingTags++
					So(tag, ShouldEndWith, "+/annotations")

					logdogURL = strings.TrimPrefix(tag, "log_location:")
				}
			}
			So(matchingTags, ShouldEqual, 1)
			So(logdogURL, ShouldStartWith, "logdog://foo-logdog-host/foo-luci-project/skylab/")
			So(logdogURL, ShouldEndWith, "/+/annotations")

			for i, slice := range create.TaskSlices {
				flatCommand := strings.Join(slice.Properties.Command, " ")

				So(flatCommand, ShouldContainSubstring, "-task-name name1")
				So(flatCommand, ShouldNotContainSubstring, "-client-test")

				// Logdog annotation url argument should match the associated tag's url.
				So(flatCommand, ShouldContainSubstring, "-logdog-annotation-url "+logdogURL)

				So(flatCommand, ShouldContainSubstring, "-test-args foo-arg1 foo-arg2")
				So(slice.Properties.Command, ShouldContain, "-test-args")
				So(slice.Properties.Command, ShouldContain, "foo-arg1 foo-arg2")

				keyvals := extractKeyvalsArgument(flatCommand)
				So(keyvals, ShouldNotBeEmpty)
				So(keyvals, ShouldContainSubstring, `"k1":"v1"`)
				So(keyvals, ShouldContainSubstring, `"parent_job_id":"foo-parent-task-id"`)
				So(keyvals, ShouldContainSubstring, `"label":"given_name"`)

				provisionArg := "-provision-labels cros-version:foo-build,fwro-version:foo-ro-firmware,fwrw-version:foo-rw-firmware"

				if i == 0 {
					So(flatCommand, ShouldNotContainSubstring, provisionArg)
				} else {
					So(flatCommand, ShouldContainSubstring, provisionArg)
				}

				flatDimensions := make([]string, len(slice.Properties.Dimensions))
				for i, d := range slice.Properties.Dimensions {
					flatDimensions[i] = d.Key + ":" + d.Value
				}
				So(flatDimensions, ShouldContain, "label-cr50_phase:CR50_PHASE_PVT")
				So(flatDimensions, ShouldContain, "label-model:foo-model")
				So(flatDimensions, ShouldContain, "label-board:foo-board")
				So(flatDimensions, ShouldContain, "label-pool:DUT_POOL_CQ")
				So(flatDimensions, ShouldContain, "freeform-key:freeform-value")
			}
		})
	})
}

var keyvalsPattern = regexp.MustCompile(`\-keyvals\s*\{([\w\s":,-/]+)\}`)

func extractKeyvalsArgument(cmd string) string {
	ms := keyvalsPattern.FindAllStringSubmatch(cmd, -1)
	So(ms, ShouldHaveLength, 1)
	m := ms[0]
	// Guaranteed by the constant regexp definition.
	if len(m) != 2 {
		panic(fmt.Sprintf("Match %s of regexp %s has length %d, want 2", m, keyvalsPattern, len(m)))
	}
	return m[1]
}

type autotestResultGenerator func() *skylab_test_runner.Result_Autotest

func autotestResultAlwaysEmpty() *skylab_test_runner.Result_Autotest {
	return &skylab_test_runner.Result_Autotest{}
}

// generateAutotestResultsFromSlice returns a autotestResultGenerator that
// sequentially returns the provided results.
//
// An attempt to generate more results than provided results in panic().
func generateAutotestResultsFromSlice(canned []*skylab_test_runner.Result_Autotest) autotestResultGenerator {
	i := 0
	f := func() *skylab_test_runner.Result_Autotest {
		if i >= len(canned) {
			panic(fmt.Sprintf("requested more results than available (%d)", len(canned)))
		}
		r := canned[i]
		i++
		return r
	}
	return f
}

func autotestResultAlwaysPass() *skylab_test_runner.Result_Autotest {
	return &skylab_test_runner.Result_Autotest{
		Incomplete: false,
		TestCases: []*skylab_test_runner.Result_Autotest_TestCase{
			{Name: "foo", Verdict: skylab_test_runner.Result_Autotest_TestCase_VERDICT_PASS},
		},
	}
}

func autotestResultAlwaysFail() *skylab_test_runner.Result_Autotest {
	return &skylab_test_runner.Result_Autotest{
		Incomplete: false,
		TestCases: []*skylab_test_runner.Result_Autotest_TestCase{
			{Name: "foo", Verdict: skylab_test_runner.Result_Autotest_TestCase_VERDICT_FAIL},
		},
	}
}

func TestInvocationKeyvals(t *testing.T) {
	Convey("Given an enumeration with a suite keyval", t, func() {
		ctx := context.Background()
		swarming := newFakeSwarming("")
		getter := newFakeGetter()
		gf := fakeGetterFactory(getter)

		invs := []*steps.EnumerationResponse_AutotestInvocation{
			{
				Test: &api.AutotestTest{
					Name:                 "someTest",
					ExecutionEnvironment: api.AutotestTest_EXECUTION_ENVIRONMENT_CLIENT,
				},
				ResultKeyvals: map[string]string{
					"suite": "someSuite",
				},
			},
		}

		Convey("and a request without keyvals", func() {
			p := basicParams()
			p.Decorations = nil
			run, err := skylab.NewTaskSet(ctx, invs, p, basicConfig(), "foo-parent-task-id")
			So(err, ShouldBeNil)
			err = run.LaunchAndWait(ctx, swarming, gf)
			So(err, ShouldBeNil)
			Convey("created command includes invocation suite keyval", func() {
				So(swarming.createCalls, ShouldHaveLength, 1)
				create := swarming.createCalls[0]
				So(create.TaskSlices, ShouldHaveLength, 2)
				for _, slice := range create.TaskSlices {
					flatCommand := strings.Join(slice.Properties.Command, " ")
					keyvals := extractKeyvalsArgument(flatCommand)
					So(keyvals, ShouldContainSubstring, `"suite":"someSuite"`)
					So(keyvals, ShouldContainSubstring, `"label":"foo-build/someSuite/someTest"`)
				}
			})
		})

		Convey("and a request with different suite keyvals", func() {
			p := basicParams()
			p.Decorations = &test_platform.Request_Params_Decorations{
				AutotestKeyvals: map[string]string{
					"suite": "someOtherSuite",
				},
			}
			run, err := skylab.NewTaskSet(ctx, invs, p, basicConfig(), "foo-parent-task-id")
			So(err, ShouldBeNil)
			err = run.LaunchAndWait(ctx, swarming, gf)
			So(err, ShouldBeNil)
			Convey("created command includes request suite keyval", func() {
				So(swarming.createCalls, ShouldHaveLength, 1)
				create := swarming.createCalls[0]
				So(create.TaskSlices, ShouldHaveLength, 2)
				for _, slice := range create.TaskSlices {
					flatCommand := strings.Join(slice.Properties.Command, " ")
					keyvals := extractKeyvalsArgument(flatCommand)
					So(keyvals, ShouldContainSubstring, `"suite":"someOtherSuite"`)
					So(keyvals, ShouldContainSubstring, `"label":"foo-build/someOtherSuite/someTest"`)
				}
			})
		})
	})
}

func invocationsWithServerTests(names ...string) []*steps.EnumerationResponse_AutotestInvocation {
	ret := make([]*steps.EnumerationResponse_AutotestInvocation, len(names))
	for i, n := range names {
		ret[i] = serverTestInvocation(n, "")
	}
	return ret
}

func TestRetries(t *testing.T) {
	Convey("Given a test with", t, func() {
		ctx := context.Background()
		ctx, ts := testclock.UseTime(ctx, time.Now())
		// Setup testclock to immediately advance whenever timer is set; this
		// avoids slowdown due to timer inside of LaunchAndWait.
		ts.SetTimerCallback(func(d time.Duration, t clock.Timer) {
			ts.Add(2 * d)
		})
		swarming := newFakeSwarming("")
		params := basicParams()
		getter := newFakeGetter()
		gf := fakeGetterFactory(getter)

		cases := []struct {
			name        string
			invocations []*steps.EnumerationResponse_AutotestInvocation
			// autotestResult will be returned by all attempts of this test.
			autotestResultGenerator autotestResultGenerator
			retryParams             *test_platform.Request_Params_Retry
			testAllowRetry          bool
			testMaxRetry            int32

			// Total number of expected tasks is this +1
			expectedRetryCount     int
			expectedSummaryVerdict test_platform.TaskState_Verdict
		}{
			{
				name:                    "1 test; no retry configuration in test or request params",
				invocations:             invocationsWithServerTests("name1"),
				autotestResultGenerator: autotestResultAlwaysFail,

				expectedRetryCount:     0,
				expectedSummaryVerdict: test_platform.TaskState_VERDICT_FAILED,
			},
			{
				name:        "1 passing test; retries allowed",
				invocations: invocationsWithServerTests("name1"),
				retryParams: &test_platform.Request_Params_Retry{
					Allow: true,
				},
				testAllowRetry:          true,
				testMaxRetry:            1,
				autotestResultGenerator: autotestResultAlwaysPass,

				expectedRetryCount:     0,
				expectedSummaryVerdict: test_platform.TaskState_VERDICT_PASSED,
			},
			{
				name:        "1 failing test; retries disabled globally",
				invocations: invocationsWithServerTests("name1"),
				retryParams: &test_platform.Request_Params_Retry{
					Allow: false,
				},
				testAllowRetry:          true,
				testMaxRetry:            1,
				autotestResultGenerator: autotestResultAlwaysFail,

				expectedRetryCount:     0,
				expectedSummaryVerdict: test_platform.TaskState_VERDICT_FAILED,
			},
			{
				name:        "1 failing test; retries allowed globally and for test",
				invocations: invocationsWithServerTests("name1"),
				retryParams: &test_platform.Request_Params_Retry{
					Allow: true,
				},
				testAllowRetry:          true,
				testMaxRetry:            1,
				autotestResultGenerator: autotestResultAlwaysFail,

				expectedRetryCount:     1,
				expectedSummaryVerdict: test_platform.TaskState_VERDICT_FAILED,
			},
			{
				name:        "1 failing test; retries allowed globally, disabled for test",
				invocations: invocationsWithServerTests("name1"),
				retryParams: &test_platform.Request_Params_Retry{
					Allow: true,
				},
				testAllowRetry:          false,
				autotestResultGenerator: autotestResultAlwaysFail,

				expectedRetryCount:     0,
				expectedSummaryVerdict: test_platform.TaskState_VERDICT_FAILED,
			},
			{
				name:        "1 failing test; retries allowed globally with test maximum",
				invocations: invocationsWithServerTests("name1"),
				retryParams: &test_platform.Request_Params_Retry{
					Allow: true,
				},
				testAllowRetry:          true,
				testMaxRetry:            10,
				autotestResultGenerator: autotestResultAlwaysFail,

				expectedRetryCount:     10,
				expectedSummaryVerdict: test_platform.TaskState_VERDICT_FAILED,
			},
			{
				name:        "1 failing test; retries allowed globally with global maximum",
				invocations: invocationsWithServerTests("name1"),
				retryParams: &test_platform.Request_Params_Retry{
					Allow: true,
					Max:   5,
				},
				testAllowRetry:          true,
				autotestResultGenerator: autotestResultAlwaysFail,

				expectedRetryCount:     5,
				expectedSummaryVerdict: test_platform.TaskState_VERDICT_FAILED,
			},
			{
				name:        "1 failing test; retries allowed globally with global maximum smaller than test maxium",
				invocations: invocationsWithServerTests("name1"),
				retryParams: &test_platform.Request_Params_Retry{
					Allow: true,
					Max:   5,
				},
				testAllowRetry:          true,
				testMaxRetry:            7,
				autotestResultGenerator: autotestResultAlwaysFail,

				expectedRetryCount:     5,
				expectedSummaryVerdict: test_platform.TaskState_VERDICT_FAILED,
			},
			{
				name:        "1 failing test; retries allowed globally with test maximum smaller than global maximum",
				invocations: invocationsWithServerTests("name1"),
				retryParams: &test_platform.Request_Params_Retry{
					Allow: true,
					Max:   7,
				},
				testAllowRetry:          true,
				testMaxRetry:            5,
				autotestResultGenerator: autotestResultAlwaysFail,

				expectedRetryCount:     5,
				expectedSummaryVerdict: test_platform.TaskState_VERDICT_FAILED,
			},
			{
				name:        "2 failing tests; retries allowed globally with global maximum",
				invocations: invocationsWithServerTests("name1", "name2"),
				retryParams: &test_platform.Request_Params_Retry{
					Allow: true,
					Max:   5,
				},
				testAllowRetry:          true,
				autotestResultGenerator: autotestResultAlwaysFail,

				expectedRetryCount:     5,
				expectedSummaryVerdict: test_platform.TaskState_VERDICT_FAILED,
			},

			{
				name:        "1 test that fails then passes; retries allowed",
				invocations: invocationsWithServerTests("name1"),
				retryParams: &test_platform.Request_Params_Retry{
					Allow: true,
				},
				testAllowRetry: true,
				autotestResultGenerator: generateAutotestResultsFromSlice([]*skylab_test_runner.Result_Autotest{
					{
						TestCases: []*skylab_test_runner.Result_Autotest_TestCase{
							{Name: "foo", Verdict: skylab_test_runner.Result_Autotest_TestCase_VERDICT_FAIL},
						},
					},
					{
						TestCases: []*skylab_test_runner.Result_Autotest_TestCase{
							{Name: "foo", Verdict: skylab_test_runner.Result_Autotest_TestCase_VERDICT_PASS},
						},
					},
				}),

				expectedRetryCount: 1,
				// TODO(crbug.com/1005609) Indicate in *some way* that a test
				// passed only on retry.
				expectedSummaryVerdict: test_platform.TaskState_VERDICT_PASSED,
			},
		}
		for _, c := range cases {
			Convey(c.name, func() {
				getter.SetAutotestResultGenerator(c.autotestResultGenerator)
				params.Retry = c.retryParams
				for _, inv := range c.invocations {
					inv.Test.AllowRetries = c.testAllowRetry
					inv.Test.MaxRetries = c.testMaxRetry
				}
				run, err := skylab.NewTaskSet(ctx, c.invocations, params, basicConfig(), "foo-parent-task-id")
				So(err, ShouldBeNil)
				err = run.LaunchAndWait(ctx, swarming, gf)
				So(err, ShouldBeNil)
				response := run.Response(swarming)

				Convey("each attempt request should have a unique logdog url in the.", func() {
					s := map[string]bool{}
					for _, req := range swarming.createCalls {
						url, ok := extractLogdogUrlFromCommand(req.TaskSlices[0].Properties.Command)
						So(ok, ShouldBeTrue)
						s[url] = true
					}
					So(s, ShouldHaveLength, len(swarming.createCalls))
				})
				// TODO(crbug.com/1003874, pprabhu) This test case is in the wrong place.
				// Once the hack to manipulate logdog URL is removed, this block can also be dropped.
				Convey("the logdog url in the command and in tags should match.", func() {
					for _, req := range swarming.createCalls {
						cmdURL, _ := extractLogdogUrlFromCommand(req.TaskSlices[0].Properties.Command)
						tagURL := extractLogdogUrlFromTags(req.Tags)
						So(cmdURL, ShouldEqual, tagURL)
					}
				})
				Convey("then the launched task count should be correct.", func() {
					// Each test is tried at least once.
					attemptCount := len(c.invocations) + c.expectedRetryCount
					So(response.TaskResults, ShouldHaveLength, attemptCount)
				})
				Convey("then task (name, attempt) should be unique.", func() {
					s := make(stringset.Set)
					for _, res := range response.TaskResults {
						s.Add(fmt.Sprintf("%s__%d", res.Name, res.Attempt))
					}
					So(s, ShouldHaveLength, len(response.TaskResults))
				})

				Convey("then the build verdict should be correct.", func() {
					So(response.State.Verdict, ShouldEqual, c.expectedSummaryVerdict)
				})
			})
		}
	})
}

func extractLogdogUrlFromCommand(command []string) (string, bool) {
	for i, s := range command[:len(command)-1] {
		if s == "-logdog-annotation-url" {
			return command[i+1], true
		}
	}
	return "", false
}

func extractLogdogUrlFromTags(tags []string) string {
	for _, s := range tags {
		if strings.HasPrefix(s, "log_location:") {
			return s[len("log_location:"):]
		}
	}
	return ""
}

func TestClientTestArg(t *testing.T) {
	Convey("Given a client test", t, func() {
		ctx := context.Background()
		swarming := newFakeSwarming("")

		invs := []*steps.EnumerationResponse_AutotestInvocation{clientTestInvocation("name1", "")}

		run, err := skylab.NewTaskSet(ctx, invs, basicParams(), basicConfig(), "foo-parent-task-id")
		So(err, ShouldBeNil)
		err = run.LaunchAndWait(ctx, swarming, fakeGetterFactory(newFakeGetter()))
		So(err, ShouldBeNil)

		Convey("the launched task request should have correct parameters.", func() {
			So(swarming.createCalls, ShouldHaveLength, 1)
			create := swarming.createCalls[0]
			So(create.TaskSlices, ShouldHaveLength, 2)
			for _, slice := range create.TaskSlices {
				flatCommand := strings.Join(slice.Properties.Command, " ")
				So(flatCommand, ShouldContainSubstring, "-client-test")
			}
		})
	})
}

func TestQuotaSchedulerAccount(t *testing.T) {
	Convey("Given a client test and a selected quota account", t, func() {
		ctx := context.Background()
		swarming := newFakeSwarming("")
		invs := []*steps.EnumerationResponse_AutotestInvocation{serverTestInvocation("name1", "")}
		params := basicParams()
		params.Scheduling.Pool = &test_platform.Request_Params_Scheduling_QuotaAccount{
			QuotaAccount: "foo-account",
		}

		run, err := skylab.NewTaskSet(ctx, invs, params, basicConfig(), "foo-parent-task-id")
		So(err, ShouldBeNil)
		err = run.LaunchAndWait(ctx, swarming, fakeGetterFactory(newFakeGetter()))
		So(err, ShouldBeNil)

		Convey("the launched task request should have a tag specifying the correct quota account and run in the quota pool.", func() {
			So(swarming.createCalls, ShouldHaveLength, 1)
			create := swarming.createCalls[0]
			So(create.Tags, ShouldContain, "qs_account:foo-account")
			for _, slice := range create.TaskSlices {
				flatDimensions := make([]string, len(slice.Properties.Dimensions))
				for i, d := range slice.Properties.Dimensions {
					flatDimensions[i] = d.Key + ":" + d.Value
				}
				So(flatDimensions, ShouldContain, "label-pool:DUT_POOL_QUOTA")
			}
		})
	})
}

func TestUnmanagedPool(t *testing.T) {
	Convey("Given a client test and an unmanaged pool.", t, func() {
		ctx := context.Background()
		swarming := newFakeSwarming("")
		invs := []*steps.EnumerationResponse_AutotestInvocation{serverTestInvocation("name1", "")}
		params := basicParams()
		params.Scheduling.Pool = &test_platform.Request_Params_Scheduling_UnmanagedPool{
			UnmanagedPool: "foo-pool",
		}

		run, err := skylab.NewTaskSet(ctx, invs, params, basicConfig(), "foo-parent-task-id")
		So(err, ShouldBeNil)
		err = run.LaunchAndWait(ctx, swarming, fakeGetterFactory(newFakeGetter()))
		So(err, ShouldBeNil)

		Convey("the launched task request run in the unmanaged pool.", func() {
			So(swarming.createCalls, ShouldHaveLength, 1)
			create := swarming.createCalls[0]
			for _, slice := range create.TaskSlices {
				flatDimensions := make([]string, len(slice.Properties.Dimensions))
				for i, d := range slice.Properties.Dimensions {
					flatDimensions[i] = d.Key + ":" + d.Value
				}
				So(flatDimensions, ShouldContain, "label-pool:foo-pool")
			}
		})
	})
}

func TestResponseVerdict(t *testing.T) {
	Convey("Given a client test", t, func() {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Setup testclock to immediately advance whenever timer is set; this
		// avoids slowdown due to timer inside of LaunchAndWait.
		ctx, ts := testclock.UseTime(ctx, time.Now())
		ts.SetTimerCallback(func(d time.Duration, t clock.Timer) {
			ts.Add(2 * d)
		})

		swarming := newFakeSwarming("")
		invs := []*steps.EnumerationResponse_AutotestInvocation{serverTestInvocation("name1", "")}
		params := basicParams()
		getter := newFakeGetter()
		gf := fakeGetterFactory(getter)

		run, err := skylab.NewTaskSet(ctx, invs, params, basicConfig(), "foo-parent-task-id")
		So(err, ShouldBeNil)

		// TODO(crbug.com/1001746, akeshet) Fix this test.
		// This test is broken even after adding locks around testRun.attempts because it is possible that the
		// assertions at the end are run before LaunchAndWait() does anything. That is not the intent of this test.
		SkipConvey("when tests are still running, response verdict is correct.", func() {
			swarming.setTaskState(jsonrpc.TaskState_RUNNING)

			wg := sync.WaitGroup{}
			defer wg.Wait()
			defer cancel()
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Can't verify error returned is nil because Convey() doesn't
				// like assertions in goroutines.
				_ = run.LaunchAndWait(ctx, swarming, gf)
			}()

			resp := run.Response(swarming)
			So(resp.State.LifeCycle, ShouldEqual, test_platform.TaskState_LIFE_CYCLE_RUNNING)
			So(resp.State.Verdict, ShouldEqual, test_platform.TaskState_VERDICT_UNSPECIFIED)
		})

		Convey("when the test passed, response verdict is correct.", func() {
			getter.SetAutotestResultGenerator(autotestResultAlwaysPass)

			run.LaunchAndWait(ctx, swarming, gf)
			resp := run.Response(swarming)
			So(resp.State.LifeCycle, ShouldEqual, test_platform.TaskState_LIFE_CYCLE_COMPLETED)
			So(resp.State.Verdict, ShouldEqual, test_platform.TaskState_VERDICT_PASSED)
		})

		Convey("when the test failed, response verdict is correct.", func() {
			getter.SetAutotestResultGenerator(autotestResultAlwaysFail)
			run.LaunchAndWait(ctx, swarming, gf)
			resp := run.Response(swarming)
			So(resp.State.LifeCycle, ShouldEqual, test_platform.TaskState_LIFE_CYCLE_COMPLETED)
			So(resp.State.Verdict, ShouldEqual, test_platform.TaskState_VERDICT_FAILED)
		})

		Convey("when an error cancels the run, response verdict is correct.", func() {
			swarming.setTaskState(jsonrpc.TaskState_RUNNING)

			wg := sync.WaitGroup{}
			wg.Add(1)
			var err error
			go func() {
				err = run.LaunchAndWait(ctx, swarming, gf)
				wg.Done()
			}()

			cancel()
			wg.Wait()
			So(err, ShouldNotBeNil)

			resp := run.Response(swarming)
			So(resp.State.LifeCycle, ShouldEqual, test_platform.TaskState_LIFE_CYCLE_ABORTED)
			So(resp.State.Verdict, ShouldEqual, test_platform.TaskState_VERDICT_FAILED)
		})
	})
}
