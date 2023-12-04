// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	buildbucket_pb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/luciexe/exe"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	bb "infra/chromium/compilator_watcher/internal/bb"
)

const fakeTagName = "fake_tag_name"
const fakeTaggedStep = "fake tagged step"

type stepNameStatusTags struct {
	stepName string
	status   buildbucket_pb.Status
	tags     []*buildbucket_pb.StringPair
}

func getSteps(stepInfo []stepNameStatusTags) []*buildbucket_pb.Step {
	steps := make([]*buildbucket_pb.Step, len(stepInfo))

	for i, info := range stepInfo {
		steps[i] = &buildbucket_pb.Step{Name: info.stepName, Status: info.status, Tags: info.tags}
	}
	return steps
}

var genericGitilesCommit = &buildbucket_pb.GitilesCommit{
	Host:     "chromium.googlesource.com",
	Project:  "chromium/src",
	Id:       "ad975cfcd476867068e8c613ac26c64b8cab2567",
	Ref:      "refs/heads/main",
	Position: 1100487,
}

func getBuildsWithSteps(
	stepInfo []stepNameStatusTags,
	outputFields map[string]*structpb.Value,
	buildStatus buildbucket_pb.Status,
) *buildbucket_pb.Build {
	return &buildbucket_pb.Build{
		Status:          buildStatus,
		Id:              12345,
		SummaryMarkdown: "",
		Steps:           getSteps(stepInfo),
		Output: &buildbucket_pb.Build_Output{
			Properties: &structpb.Struct{
				Fields: outputFields,
			},
			GitilesCommit: genericGitilesCommit,
		},
	}
}

func TestLuciEXEMain(t *testing.T) {
	t.Parallel()

	Convey("luciEXEMain", t, func() {
		now := time.Date(2021, 01, 01, 00, 00, 00, 00, time.UTC)
		ctx, clk := testclock.UseTime(context.Background(), now)

		clk.SetTimerCallback(func(amt time.Duration, timer clock.Timer) {
			tags := testclock.GetTags(timer)
			for _, tag := range tags {
				if tag == clock.ContextDeadlineTag {
					return
				}
			}
			clk.Add(amt)
		})

		input := &buildbucket_pb.Build{
			Output: &buildbucket_pb.Build_Output{
				Properties: &structpb.Struct{},
			},
		}
		sender := exe.BuildSender(func() {})

		genericCompBuildOutputProps := jsonToStruct(`{
			"got_angle_revision": "701d51b101c8ce1a1a840a7b0dbe3f36dfc1eec9",
			"got_revision": "04d2ba64ba046c038f8995982ecde0a7f029da1e",
			"got_revision_cp": "refs/heads/main@{#964359}",
			"affected_files": {
				"first_100": ["src/chrome/browser/extensions/extension_message_bubble_controller_unittest.cc"],
				"total_count": 1}
		}`)

		genericCompBuildOutputPropsNoSwarming := copyPropertiesStruct(genericCompBuildOutputProps)
		genericCompBuildOutputPropsWSwarming := copyPropertiesStruct(genericCompBuildOutputProps)

		genericCompleteSteps := getSteps([]stepNameStatusTags{
			{
				stepName: "builder cache",
				status:   buildbucket_pb.Status_SUCCESS,
			},
			{
				stepName: "lookup GN args",
				status:   buildbucket_pb.Status_SUCCESS,
			},
			{
				stepName: "compile (with patch)",
				status:   buildbucket_pb.Status_SUCCESS,
			},
			{
				stepName: fakeTaggedStep,
				status:   buildbucket_pb.Status_SUCCESS,
				tags: []*buildbucket_pb.StringPair{
					{
						Key:   fakeTagName,
						Value: "dummy text",
					},
				},
			},
			{
				stepName: "test_pre_run (with patch)",
				status:   buildbucket_pb.Status_SUCCESS,
			},
			{
				stepName: "check_network_annotations (with patch)",
				status:   buildbucket_pb.Status_SUCCESS,
			},
			{
				stepName: "gerrit changes",
				status:   buildbucket_pb.Status_SUCCESS,
			},
		})

		userArgs := []string{"-compilator-id", "12345", "-end-step-tag", fakeTagName}

		Convey("fails if userArgs is empty", func() {
			var userArgs []string
			err := luciEXEMain(ctx, input, userArgs, sender)

			expectedErrText := "compilator-id is required"
			So(err, ShouldErrLike, expectedErrText)
			So(
				input.SummaryMarkdown,
				ShouldResemble,
				"Error while running compilator_watcher: "+expectedErrText)
		})
		Convey("fails if userArgs is missing compilator build ID", func() {
			userArgs := []string{""}
			err := luciEXEMain(ctx, input, userArgs, sender)

			expectedErrText := "compilator-id is required"
			So(err, ShouldErrLike, expectedErrText)
			So(
				input.SummaryMarkdown,
				ShouldResemble,
				"Error while running compilator_watcher: "+expectedErrText)
		})
		Convey("copies compilator build failure status and summary", func() {
			compBuild := &buildbucket_pb.Build{
				Status:          buildbucket_pb.Status_FAILURE,
				Id:              12345,
				SummaryMarkdown: "Compile failure",
				Output: &buildbucket_pb.Build_Output{
					Properties: &structpb.Struct{},
				},
			}
			ctx = context.WithValue(
				ctx,
				bb.FakeBuildsContextKey,
				[]bb.FakeGetBuildResponse{{Build: compBuild}})
			err := luciEXEMain(ctx, input, userArgs, sender)

			So(err, ShouldBeNil)
			So(input.Status, ShouldResemble, buildbucket_pb.Status_FAILURE)
			So(input.SummaryMarkdown, ShouldResemble, "Compile failure")

		})

		Convey("copies compilator output properties", func() {
			expectedSubBuildOutputProps := copyPropertiesStruct(genericCompBuildOutputPropsWSwarming)

			compBuild := &buildbucket_pb.Build{
				Status:          buildbucket_pb.Status_STARTED,
				Id:              12345,
				SummaryMarkdown: "",
				Steps:           genericCompleteSteps,
				Output: &buildbucket_pb.Build_Output{
					Properties:    genericCompBuildOutputPropsWSwarming,
					GitilesCommit: genericGitilesCommit,
				},
			}

			ctx = context.WithValue(
				ctx,
				bb.FakeBuildsContextKey,
				[]bb.FakeGetBuildResponse{{Build: compBuild}})

			err := luciEXEMain(ctx, input, userArgs, sender)

			So(err, ShouldBeNil)
			So(input.GetOutput(), ShouldResembleProto, &buildbucket_pb.Build_Output{
				Properties:    expectedSubBuildOutputProps,
				GitilesCommit: genericGitilesCommit,
			})
		})

		// TODO(crbug/1507700): Re-enable when flakiness is fixed.
		// This test fails on ci/infra-continuous-mac-10.14-64
		SkipConvey("cancel context sets status to CANCELED and returns no err", func() {
			compBuild := &buildbucket_pb.Build{
				Status:          buildbucket_pb.Status_CANCELED,
				Id:              12345,
				SummaryMarkdown: "",
				Steps:           genericCompleteSteps,
				Output:          &buildbucket_pb.Build_Output{},
			}

			ctx = context.WithValue(
				ctx,
				bb.FakeBuildsContextKey,
				[]bb.FakeGetBuildResponse{{Build: compBuild}})

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			errC := make(chan error)
			go func() {
				errC <- luciEXEMain(ctx, input, userArgs, sender)
			}()

			cancel()
			err := <-errC

			So(err, ShouldBeNil)
			So(input.Status, ShouldResemble, buildbucket_pb.Status_CANCELED)
		})

		Convey("sets input Status to SUCCESS when compilator build is still running", func() {
			userArgs := []string{"-compilator-id", "12345", "-end-step-tag", fakeTagName}
			expectedSubBuildOutputProps := copyPropertiesStruct(genericCompBuildOutputPropsWSwarming)

			compBuild := &buildbucket_pb.Build{
				Status:          buildbucket_pb.Status_STARTED,
				Id:              12345,
				SummaryMarkdown: "",
				Steps:           genericCompleteSteps,
				Output: &buildbucket_pb.Build_Output{
					Properties:    genericCompBuildOutputPropsWSwarming,
					GitilesCommit: genericGitilesCommit,
				},
			}
			ctx = context.WithValue(
				ctx,
				bb.FakeBuildsContextKey,
				[]bb.FakeGetBuildResponse{{Build: compBuild}})
			err := luciEXEMain(ctx, input, userArgs, sender)

			So(err, ShouldBeNil)
			So(input.Status, ShouldResemble, buildbucket_pb.Status_SUCCESS)
			So(input.GetOutput(), ShouldResembleProto, &buildbucket_pb.Build_Output{
				Properties:    expectedSubBuildOutputProps,
				GitilesCommit: genericGitilesCommit,
			})
		})

		Convey("copies over compilator build status when no end step tag", func() {
			userArgs := []string{"-compilator-id", "12345", "-start-step-tag", fakeTagName}
			compBuild := &buildbucket_pb.Build{
				Status:          buildbucket_pb.Status_FAILURE,
				Id:              12345,
				SummaryMarkdown: "",
				Steps:           genericCompleteSteps,
				Output: &buildbucket_pb.Build_Output{
					Properties:    genericCompBuildOutputPropsWSwarming,
					GitilesCommit: genericGitilesCommit,
				},
			}
			ctx = context.WithValue(
				ctx,
				bb.FakeBuildsContextKey,
				[]bb.FakeGetBuildResponse{{Build: compBuild}})
			err := luciEXEMain(ctx, input, userArgs, sender)

			So(err, ShouldBeNil)
			So(input.Status, ShouldResemble, buildbucket_pb.Status_FAILURE)
		})

		Convey("exits after compilator build successfully ends", func() {
			expectedSubBuildOutputProps := copyPropertiesStruct(genericCompBuildOutputPropsNoSwarming)

			compBuild := &buildbucket_pb.Build{
				Status:          buildbucket_pb.Status_SUCCESS,
				Id:              12345,
				SummaryMarkdown: "",
				Output: &buildbucket_pb.Build_Output{
					Properties:    genericCompBuildOutputPropsNoSwarming,
					GitilesCommit: genericGitilesCommit,
				},
			}
			ctx = context.WithValue(
				ctx,
				bb.FakeBuildsContextKey,
				[]bb.FakeGetBuildResponse{{Build: compBuild}})
			err := luciEXEMain(ctx, input, userArgs, sender)

			So(err, ShouldBeNil)
			So(input.Status, ShouldResemble, buildbucket_pb.Status_SUCCESS)

			So(input.GetOutput(), ShouldResembleProto, &buildbucket_pb.Build_Output{
				Properties:    expectedSubBuildOutputProps,
				GitilesCommit: genericGitilesCommit,
			})
		})

		Convey("updates last step even if step name is the same", func() {
			compBuilds := []bb.FakeGetBuildResponse{
				{Build: getBuildsWithSteps([]stepNameStatusTags{
					{
						stepName: "lookup GN args",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "analyze",
						status:   buildbucket_pb.Status_STARTED,
					},
				}, genericCompBuildOutputProps.GetFields(), buildbucket_pb.Status_STARTED)},
				{Build: getBuildsWithSteps([]stepNameStatusTags{
					{
						stepName: "lookup GN args",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "analyze",
						status:   buildbucket_pb.Status_SUCCESS,
					},
				}, genericCompBuildOutputProps.GetFields(), buildbucket_pb.Status_SUCCESS)},
			}
			ctx = context.WithValue(
				ctx,
				bb.FakeBuildsContextKey,
				compBuilds)
			userArgs := []string{"-compilator-id", "12345"}
			err := luciEXEMain(ctx, input, userArgs, sender)
			So(err, ShouldBeNil)
			expectedSteps := getSteps([]stepNameStatusTags{
				{
					stepName: "lookup GN args",
					status:   buildbucket_pb.Status_SUCCESS,
				},
				{
					stepName: "analyze",
					status:   buildbucket_pb.Status_SUCCESS,
				},
			})
			So(input.GetSteps(), ShouldResembleProto, expectedSteps)
		})

		Convey("copies correct Steps", func() {
			compBuilds := []bb.FakeGetBuildResponse{
				{Build: getBuildsWithSteps([]stepNameStatusTags{
					{
						stepName: "setup_build",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "report builders",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "builder cache",
						status:   buildbucket_pb.Status_SUCCESS,
					},
				}, map[string]*structpb.Value{}, buildbucket_pb.Status_STARTED)},
				{Build: getBuildsWithSteps([]stepNameStatusTags{
					{
						stepName: "setup_build",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "report builders",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "builder cache",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "gclient config",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "lookup GN args",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "compile (with patch)",
						status:   buildbucket_pb.Status_SUCCESS,
					},
				}, map[string]*structpb.Value{}, buildbucket_pb.Status_STARTED)},
				{Build: getBuildsWithSteps([]stepNameStatusTags{
					{
						stepName: "setup_build",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "report builders",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "builder cache",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "gclient config",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "lookup GN args",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "compile (with patch)",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: fakeTaggedStep,
						status:   buildbucket_pb.Status_SUCCESS,
						tags: []*buildbucket_pb.StringPair{
							{
								Key:   fakeTagName,
								Value: "dummy text",
							},
						},
					},
					{
						stepName: "test_pre_run (with patch)",
						status:   buildbucket_pb.Status_SUCCESS,
						tags: []*buildbucket_pb.StringPair{
							{
								Key:   "other_fake_tag_name",
								Value: "dummy text",
							},
						},
					},
					{
						stepName: "check_network_annotations (with patch)",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "gerrit changes",
						status:   buildbucket_pb.Status_SUCCESS,
					},
				}, genericCompBuildOutputProps.GetFields(), buildbucket_pb.Status_SUCCESS)},
			}
			ctx = context.WithValue(
				ctx,
				bb.FakeBuildsContextKey,
				compBuilds)

			Convey("with end-step-tag", func() {
				userArgs := []string{"-compilator-id", "12345", "-end-step-tag", fakeTagName}
				err := luciEXEMain(ctx, input, userArgs, sender)
				So(err, ShouldBeNil)
				So(input.Status, ShouldResemble, buildbucket_pb.Status_SUCCESS)
				expectedSteps := getSteps([]stepNameStatusTags{
					{
						stepName: "setup_build",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "report builders",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "builder cache",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "gclient config",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "lookup GN args",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "compile (with patch)",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: fakeTaggedStep,
						status:   buildbucket_pb.Status_SUCCESS,
						tags: []*buildbucket_pb.StringPair{
							{
								Key:   fakeTagName,
								Value: "dummy text",
							},
						},
					},
				})
				So(input.GetSteps(), ShouldResembleProto, expectedSteps)

			})
			Convey("with start-step-tag", func() {
				userArgs := []string{"-compilator-id", "12345", "-start-step-tag", fakeTagName}
				err := luciEXEMain(ctx, input, userArgs, sender)
				So(err, ShouldBeNil)
				So(input.Status, ShouldResemble, buildbucket_pb.Status_SUCCESS)
				expectedSteps := getSteps([]stepNameStatusTags{
					{
						stepName: fakeTaggedStep,
						status:   buildbucket_pb.Status_SUCCESS,
						tags: []*buildbucket_pb.StringPair{
							{
								Key:   fakeTagName,
								Value: "dummy text",
							},
						},
					},
					{
						stepName: "test_pre_run (with patch)",
						status:   buildbucket_pb.Status_SUCCESS,
						tags: []*buildbucket_pb.StringPair{
							{
								Key:   "other_fake_tag_name",
								Value: "dummy text",
							},
						},
					},
					{
						stepName: "check_network_annotations (with patch)",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "gerrit changes",
						status:   buildbucket_pb.Status_SUCCESS,
					},
				})
				So(input.GetSteps(), ShouldResembleProto, expectedSteps)

			})

			Convey("with both start-step-tag and end-step-tag", func() {
				userArgs := []string{"-compilator-id", "12345", "-start-step-tag", fakeTagName, "-end-step-tag", "other_fake_tag_name"}
				err := luciEXEMain(ctx, input, userArgs, sender)
				So(err, ShouldBeNil)
				So(input.Status, ShouldResemble, buildbucket_pb.Status_SUCCESS)
				expectedSteps := getSteps([]stepNameStatusTags{
					{
						stepName: fakeTaggedStep,
						status:   buildbucket_pb.Status_SUCCESS,
						tags: []*buildbucket_pb.StringPair{
							{
								Key:   fakeTagName,
								Value: "dummy text",
							},
						},
					},
					{
						stepName: "test_pre_run (with patch)",
						status:   buildbucket_pb.Status_SUCCESS,
						tags: []*buildbucket_pb.StringPair{
							{
								Key:   "other_fake_tag_name",
								Value: "dummy text",
							},
						},
					},
				})
				So(input.GetSteps(), ShouldResembleProto, expectedSteps)

			})

			Convey("with neither start-step-tag nor end-step-tag", func() {
				userArgs := []string{"-compilator-id", "12345"}
				err := luciEXEMain(ctx, input, userArgs, sender)
				So(err, ShouldBeNil)
				So(input.Status, ShouldResemble, buildbucket_pb.Status_SUCCESS)
				expectedSteps := getSteps([]stepNameStatusTags{
					{
						stepName: "setup_build",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "report builders",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "builder cache",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "gclient config",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "lookup GN args",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "compile (with patch)",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: fakeTaggedStep,
						status:   buildbucket_pb.Status_SUCCESS,
						tags: []*buildbucket_pb.StringPair{
							{
								Key:   fakeTagName,
								Value: "dummy text",
							},
						},
					},
					{
						stepName: "test_pre_run (with patch)",
						status:   buildbucket_pb.Status_SUCCESS,
						tags: []*buildbucket_pb.StringPair{
							{
								Key:   "other_fake_tag_name",
								Value: "dummy text",
							},
						},
					},
					{
						stepName: "check_network_annotations (with patch)",
						status:   buildbucket_pb.Status_SUCCESS,
					},
					{
						stepName: "gerrit changes",
						status:   buildbucket_pb.Status_SUCCESS,
					},
				})
				So(input.GetSteps(), ShouldResembleProto, expectedSteps)

			})

			Convey("sets InfraFailure with summary for timeout", func() {
				// Force luciexe to timeout right after first build is retrieved in copySteps()
				userArgs = []string{
					"-compilator-id",
					"12345",
					"-compilator-polling-timeout-sec",
					"5",
					"-max-consecutive-get-build-timeouts",
					"3",
				}

				clk.SetTimerCallback(func(amt time.Duration, timer clock.Timer) {
					tags := testclock.GetTags(timer)
					for i := 0; i < len(tags); i++ {
						tag := tags[i]
						if tag == clock.ContextDeadlineTag {
							return
						}
					}
					clk.Add(5*time.Second + time.Millisecond)
				})

				ctx = context.WithValue(
					ctx,
					bb.FakeBuildsContextKey,
					compBuilds)

				err := luciEXEMain(ctx, input, userArgs, sender)
				So(err, ShouldNotBeNil)
				So(exe.InfraErrorTag.In(err), ShouldBeTrue)
				So(input.SummaryMarkdown, ShouldResemble, "Error while running compilator_watcher: Timeout waiting for compilator build")
			})

			Convey("handles timeouts from GetBuild", func() {
				userArgs := []string{"-compilator-id", "12345"}

				Convey("by allowing up to max N consecutive errs", func() {
					compBuilds := []bb.FakeGetBuildResponse{
						{Build: getBuildsWithSteps([]stepNameStatusTags{
							{
								stepName: "report builders",
								status:   buildbucket_pb.Status_STARTED,
							},
						}, map[string]*structpb.Value{}, buildbucket_pb.Status_STARTED)},
						{Err: grpcStatus.Error(codes.DeadlineExceeded, "Gateway Timeout")},
						{Err: grpcStatus.Error(codes.DeadlineExceeded, "Gateway Timeout")},
						{Build: getBuildsWithSteps([]stepNameStatusTags{
							{
								stepName: "report builders",
								status:   buildbucket_pb.Status_FAILURE,
							},
						}, map[string]*structpb.Value{}, buildbucket_pb.Status_FAILURE)},
					}
					ctx = context.WithValue(
						ctx,
						bb.FakeBuildsContextKey,
						compBuilds)
					err := luciEXEMain(ctx, input, userArgs, sender)
					So(err, ShouldBeNil)
					expectedSteps := getSteps([]stepNameStatusTags{
						{
							stepName: "report builders",
							status:   buildbucket_pb.Status_FAILURE,
						},
					})
					So(input.GetSteps(), ShouldResembleProto, expectedSteps)

				})
				Convey("and raising err if the num of consecutive errs exceeds max number", func() {
					compBuilds := []bb.FakeGetBuildResponse{
						{Build: getBuildsWithSteps([]stepNameStatusTags{
							{
								stepName: "report builders",
								status:   buildbucket_pb.Status_STARTED,
							},
						}, map[string]*structpb.Value{}, buildbucket_pb.Status_STARTED)},
						{Err: grpcStatus.Error(codes.DeadlineExceeded, "Gateway Timeout")},
						{Err: grpcStatus.Error(codes.DeadlineExceeded, "Gateway Timeout")},
						{Err: grpcStatus.Error(codes.DeadlineExceeded, "Gateway Timeout")},
						{Err: grpcStatus.Error(codes.DeadlineExceeded, "Gateway Timeout")},
					}
					ctx = context.WithValue(
						ctx,
						bb.FakeBuildsContextKey,
						compBuilds)
					err := luciEXEMain(ctx, input, userArgs, sender)
					So(err, ShouldNotBeNil)
					So(err, ShouldErrLike, "rpc error: code = DeadlineExceeded desc = Gateway Timeout")
				})
				Convey("and errs need to be consecutive", func() {
					compBuilds := []bb.FakeGetBuildResponse{
						{Err: grpcStatus.Error(codes.DeadlineExceeded, "Gateway Timeout")},
						{Err: grpcStatus.Error(codes.DeadlineExceeded, "Gateway Timeout")},
						{Build: getBuildsWithSteps([]stepNameStatusTags{
							{
								stepName: "report builders",
								status:   buildbucket_pb.Status_STARTED,
							},
						}, map[string]*structpb.Value{}, buildbucket_pb.Status_STARTED)},
						{Err: grpcStatus.Error(codes.DeadlineExceeded, "Gateway Timeout")},
						{Err: grpcStatus.Error(codes.DeadlineExceeded, "Gateway Timeout")},
						{Build: getBuildsWithSteps([]stepNameStatusTags{
							{
								stepName: "report builders",
								status:   buildbucket_pb.Status_FAILURE,
							},
						}, map[string]*structpb.Value{}, buildbucket_pb.Status_FAILURE)},
					}
					ctx = context.WithValue(
						ctx,
						bb.FakeBuildsContextKey,
						compBuilds)
					err := luciEXEMain(ctx, input, userArgs, sender)
					So(err, ShouldBeNil)
					expectedSteps := getSteps([]stepNameStatusTags{
						{
							stepName: "report builders",
							status:   buildbucket_pb.Status_FAILURE,
						},
					})
					So(input.GetSteps(), ShouldResembleProto, expectedSteps)
				})
			})
		})
	})
}
