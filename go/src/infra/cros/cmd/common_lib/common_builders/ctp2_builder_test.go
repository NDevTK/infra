// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_builders_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"

	builders "infra/cros/cmd/common_lib/common_builders"
)

func TestCTPv1Tov2Translation(t *testing.T) {
	Convey("Single Translation", t, func() {
		expected := &testapi.CTPv2Request{
			Requests: []*testapi.CTPRequest{
				getCTPv2Request("board", "model", "release", "board-release/R123.0.0", "", "suite", ""),
			},
		}
		requests := map[string]*test_platform.Request{
			"r1": getCTPv1Request("board", "model", "board-release/R123.0.0", "suite", ""),
		}
		result := builders.NewCTPV2FromV1(context.Background(), requests).BuildRequest()

		So(result.GetRequests(), ShouldHaveLength, 1)
		So(IsEqualCTPv2(result, expected), ShouldBeTrue)
	})

	Convey("Multi Translation, no grouping", t, func() {
		expected := &testapi.CTPv2Request{
			Requests: []*testapi.CTPRequest{
				getCTPv2Request("board", "model", "release", "board-release/R123.0.0", "", "suite", ""),
				getCTPv2Request("board", "model", "release", "board-release/R124.0.0", "", "suite", ""),
			},
		}
		requests := map[string]*test_platform.Request{
			"r1": getCTPv1Request("board", "model", "board-release/R123.0.0", "suite", ""),
			"r2": getCTPv1Request("board", "model", "board-release/R124.0.0", "suite", ""),
		}
		result := builders.NewCTPV2FromV1(context.Background(), requests).BuildRequest()

		So(result.GetRequests(), ShouldHaveLength, 2)
		So(IsEqualCTPv2(result, expected), ShouldBeTrue)
	})

	Convey("Multi Translation, grouping", t, func() {
		expected := &testapi.CTPv2Request{
			Requests: builders.GroupV2Requests(context.Background(), []*testapi.CTPRequest{
				getCTPv2Request("board", "model", "release", "board-release/R123.0.0", "", "suite", ""),
				getCTPv2Request("board", "model2", "release", "board-release/R123.0.0", "", "suite", ""),
			}),
		}
		requests := map[string]*test_platform.Request{
			"r1": getCTPv1Request("board", "model", "board-release/R123.0.0", "suite", ""),
			"r2": getCTPv1Request("board", "model2", "board-release/R123.0.0", "suite", ""),
		}
		result := builders.NewCTPV2FromV1(context.Background(), requests).BuildRequest()

		So(result.GetRequests(), ShouldHaveLength, 1)
		So(IsEqualCTPv2(result, expected), ShouldBeTrue)
	})
}

func TestGetBuildType(t *testing.T) {
	Convey("GetBuildType", t, func() {
		buildType := builders.GetBuildType(getChromeosSoftwareDeps("board-release/R123.0.0"))
		So(buildType, ShouldEqual, "release")
	})

	Convey("GetBuildType remove postfix", t, func() {
		buildType := builders.GetBuildType(getChromeosSoftwareDeps("board-release-main/R123.0.0"))
		So(buildType, ShouldEqual, "release")
	})

	Convey("GetBuildType empty string", t, func() {
		buildType := builders.GetBuildType(getChromeosSoftwareDeps(""))
		So(buildType, ShouldEqual, "")
	})
}

func TestGetVariant(t *testing.T) {
	Convey("GetVariant", t, func() {
		variant := builders.GetVariant(getChromeosSoftwareDeps("board-variant-arc-release/R123.0.0"))
		So(variant, ShouldEqual, "variant-arc")
	})

	Convey("GetVariant no variant", t, func() {
		variant := builders.GetVariant(getChromeosSoftwareDeps("board-release/R123.0.0"))
		So(variant, ShouldEqual, "")
	})

	Convey("GetVariant remove postfix", t, func() {
		variant := builders.GetVariant(getChromeosSoftwareDeps("board-variant-arc-release-main/R123.0.0"))
		So(variant, ShouldEqual, "variant-arc")
	})

	Convey("GetVariant remove prefix", t, func() {
		variant := builders.GetVariant(getChromeosSoftwareDeps("staging-board-variant-arc-release/R123.0.0"))
		So(variant, ShouldEqual, "variant-arc")
	})

	Convey("GetVariant remove prefix and postfix", t, func() {
		variant := builders.GetVariant(getChromeosSoftwareDeps("dev-board-variant-arc-release-main/R123.0.0"))
		So(variant, ShouldEqual, "variant-arc")
	})

	Convey("GetVariant empty string", t, func() {
		variant := builders.GetVariant(getChromeosSoftwareDeps(""))
		So(variant, ShouldEqual, "")
	})
}

func TestCTP2Grouping(t *testing.T) {
	Convey("Same build, same suite, diff boards", t, func() {
		groupings := builders.GroupV2Requests(context.Background(), []*testapi.CTPRequest{
			getCTPv2Request("board1", "model1", "release", "board1-release/R123.0.0", "", "suite1", ""),
			getCTPv2Request("board2", "model1", "release", "board1-release/R123.0.0", "", "suite1", ""),
		})

		So(groupings, ShouldHaveLength, 1)
		So(groupings[0].GetScheduleTargets(), ShouldHaveLength, 2)
		So(groupings[0].GetScheduleTargets()[0].GetTargets(), ShouldHaveLength, 1)
		So(groupings[0].GetScheduleTargets()[1].GetTargets(), ShouldHaveLength, 1)
		So(groupings[0].GetScheduleTargets()[0].GetTargets()[0].GetSwTarget().GetLegacySw().GetGcsPath(),
			ShouldEqual,
			groupings[0].GetScheduleTargets()[1].GetTargets()[0].GetSwTarget().GetLegacySw().GetGcsPath())
	})

	Convey("Same build, diff suite", t, func() {
		groupings := builders.GroupV2Requests(context.Background(), []*testapi.CTPRequest{
			getCTPv2Request("board1", "model1", "release", "board1-release/R123.0.0", "", "suite1", ""),
			getCTPv2Request("board1", "model1", "release", "board1-release/R123.0.0", "", "suite2", ""),
		})

		So(groupings, ShouldHaveLength, 2)
		So(groupings[0].GetScheduleTargets(), ShouldHaveLength, 1)
		So(groupings[1].GetScheduleTargets(), ShouldHaveLength, 1)
		So(groupings[0].GetSuiteRequest().GetTestSuite().GetName(),
			ShouldNotEqual,
			groupings[1].GetSuiteRequest().GetTestSuite().GetName())
	})

	Convey("Diff build, same suite", t, func() {
		groupings := builders.GroupV2Requests(context.Background(), []*testapi.CTPRequest{
			getCTPv2Request("board1", "model1", "release", "board1-release/R123.0.0", "", "suite1", ""),
			getCTPv2Request("board1", "model1", "release", "board1-release/R124.0.0", "", "suite1", ""),
		})

		So(groupings, ShouldHaveLength, 2)
		So(groupings[0].GetScheduleTargets(), ShouldHaveLength, 1)
		So(groupings[1].GetScheduleTargets(), ShouldHaveLength, 1)
		So(groupings[0].GetScheduleTargets()[0].GetTargets(), ShouldHaveLength, 1)
		So(groupings[1].GetScheduleTargets()[0].GetTargets(), ShouldHaveLength, 1)
		So(groupings[0].GetScheduleTargets()[0].GetTargets()[0].GetSwTarget().GetLegacySw().GetGcsPath(),
			ShouldNotEqual,
			groupings[1].GetScheduleTargets()[0].GetTargets()[0].GetSwTarget().GetLegacySw().GetGcsPath())
	})

	Convey("Diff build, diff suite", t, func() {
		groupings := builders.GroupV2Requests(context.Background(), []*testapi.CTPRequest{
			getCTPv2Request("board1", "model1", "release", "board1-release/R123.0.0", "", "suite1", ""),
			getCTPv2Request("board1", "model1", "release", "board1-release/R124.0.0", "", "suite2", ""),
		})

		So(groupings, ShouldHaveLength, 2)
		So(groupings[0].GetScheduleTargets(), ShouldHaveLength, 1)
		So(groupings[1].GetScheduleTargets(), ShouldHaveLength, 1)
		So(groupings[0].GetScheduleTargets()[0].GetTargets(), ShouldHaveLength, 1)
		So(groupings[1].GetScheduleTargets()[0].GetTargets(), ShouldHaveLength, 1)
		So(groupings[0].GetScheduleTargets()[0].GetTargets()[0].GetSwTarget().GetLegacySw().GetGcsPath(),
			ShouldNotEqual,
			groupings[1].GetScheduleTargets()[0].GetTargets()[0].GetSwTarget().GetLegacySw().GetGcsPath())
		So(groupings[0].GetSuiteRequest().GetTestSuite().GetName(),
			ShouldNotEqual,
			groupings[1].GetSuiteRequest().GetTestSuite().GetName())
	})
}

func getCTPv1Request(board, model, build, suite, testArgs string) *test_platform.Request {
	return &test_platform.Request{
		TestPlan: &test_platform.Request_TestPlan{
			Suite: []*test_platform.Request_Suite{
				{
					Name:     suite,
					TestArgs: testArgs,
				},
			},
		},
		Params: &test_platform.Request_Params{
			Decorations: &test_platform.Request_Params_Decorations{
				Tags: []string{
					"label-pool:schedukeTest",
				},
			},
			Scheduling: &test_platform.Request_Params_Scheduling{
				QsAccount: "qs_account",
			},
			HardwareAttributes: &test_platform.Request_Params_HardwareAttributes{
				Model: model,
			},
			SoftwareAttributes: &test_platform.Request_Params_SoftwareAttributes{
				BuildTarget: &chromiumos.BuildTarget{
					Name: board,
				},
			},
			SoftwareDependencies: []*test_platform.Request_Params_SoftwareDependency{
				{
					Dep: &test_platform.Request_Params_SoftwareDependency_ChromeosBuild{
						ChromeosBuild: build,
					},
				},
			},
		},
	}
}

func getCTPv2Request(board, model, build, gcsPath, variant, suite, testArgs string) *testapi.CTPRequest {
	return &testapi.CTPRequest{
		Pool: "schedukeTest",
		SchedulerInfo: &testapi.SchedulerInfo{
			Scheduler: testapi.SchedulerInfo_PRINT_REQUEST_ONLY,
			QsAccount: "qs_account",
		},
		SuiteRequest: &testapi.SuiteRequest{
			SuiteRequest: &testapi.SuiteRequest_TestSuite{
				TestSuite: &testapi.TestSuite{
					Name: suite,
					Spec: &testapi.TestSuite_TestCaseTagCriteria_{
						TestCaseTagCriteria: &testapi.TestSuite_TestCaseTagCriteria{
							Tags: []string{"suite:" + suite},
						},
					},
				},
			},
			TestArgs: testArgs,
		},
		ScheduleTargets: []*testapi.ScheduleTargets{
			{
				Targets: []*testapi.Targets{
					{
						HwTarget: &testapi.HWTarget{
							Target: &testapi.HWTarget_LegacyHw{
								LegacyHw: &testapi.LegacyHW{
									Board: board,
									Model: model,
								},
							},
						},
						SwTarget: &testapi.SWTarget{
							SwTarget: &testapi.SWTarget_LegacySw{
								LegacySw: &testapi.LegacySW{
									Build:   build,
									GcsPath: "gs://chromeos-image-archive/" + gcsPath,
									Variant: variant,
									KeyValues: []*testapi.KeyValue{
										{
											Key:   builders.ChromeosBuild,
											Value: gcsPath,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func getChromeosSoftwareDeps(chromeosBuild string) []*test_platform.Request_Params_SoftwareDependency {
	return []*test_platform.Request_Params_SoftwareDependency{
		{
			Dep: &test_platform.Request_Params_SoftwareDependency_ChromeosBuild{
				ChromeosBuild: chromeosBuild,
			},
		},
	}
}

func IsEqualCTPv2(req1, req2 *testapi.CTPv2Request) bool {
	// If lists end up in same order, all good.
	if proto.Equal(req1, req2) {
		return true
	}

	if len(req1.GetRequests()) != len(req2.GetRequests()) {
		return false
	}

	matchPool := req2.GetRequests()

	// Perform window search
	for _, r1 := range req1.GetRequests() {
		for i, r2 := range matchPool {
			if proto.Equal(r1, r2) {
				matchPool[i] = matchPool[len(matchPool)-1]
				matchPool = matchPool[:len(matchPool)-1]
				break
			}
		}
	}

	return len(matchPool) == 0
}
