// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_builders_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"

	builders "infra/cros/cmd/common_lib/common_builders"
)

func MockManifestFetcher(ctx context.Context, gcsPath string) (string, error) {
	if strings.Contains(gcsPath, "public-manifest") {
		return "PUBLIC", nil
	} else {
		return "PRIVATE", nil
	}

}
func TestCTPv1Tov2Translation(t *testing.T) {
	Convey("Single Translation", t, func() {
		requests := map[string]*test_platform.Request{
			"r1": getCTPv1Request("board", "model", "board-release/R123.0.0", "suite", "", "", false),
		}
		result := builders.NewCTPV2FromV1(context.Background(), requests).BuildRequest()

		So(result.GetRequests(), ShouldHaveLength, 1)
		So(result.GetRequests()[0].GetScheduleTargets(), ShouldHaveLength, 1)
		So(result.GetRequests()[0].GetScheduleTargets()[0].GetTargets(), ShouldHaveLength, 1)
		So(result.GetRequests()[0].GetScheduleTargets()[0].GetTargets()[0].GetHwTarget().GetLegacyHw().GetBoard(), ShouldEqual, "board")
		So(result.GetRequests()[0].GetScheduleTargets()[0].GetTargets()[0].GetHwTarget().GetLegacyHw().GetModel(), ShouldEqual, "model")
		So(result.GetRequests()[0].GetScheduleTargets()[0].GetTargets()[0].GetSwTarget().GetLegacySw().GetBuild(), ShouldEqual, "release")
		So(result.GetRequests()[0].GetScheduleTargets()[0].GetTargets()[0].GetSwTarget().GetLegacySw().GetGcsPath(), ShouldEqual, "gs://chromeos-image-archive/board-release/R123.0.0")
		So(result.GetRequests()[0].GetScheduleTargets()[0].GetTargets()[0].GetSwTarget().GetLegacySw().GetVariant(), ShouldEqual, "")
	})

	Convey("Multi Translation, no grouping", t, func() {
		requests := map[string]*test_platform.Request{
			"r1": getCTPv1Request("board", "model", "board-release/R123.0.0", "suite", "", "", true),
			"r2": getCTPv1Request("board", "model", "board-release/R124.0.0", "suite", "", "", false),
		}
		result := builders.NewCTPV2FromV1(context.Background(), requests).BuildRequest()

		So(result.GetRequests(), ShouldHaveLength, 2)
		So(result.GetRequests()[0].GetScheduleTargets(), ShouldHaveLength, 1)
		So(result.GetRequests()[1].GetScheduleTargets(), ShouldHaveLength, 1)
		So(result.GetRequests()[0].GetScheduleTargets()[0].GetTargets(), ShouldHaveLength, 1)
		So(result.GetRequests()[1].GetScheduleTargets()[0].GetTargets(), ShouldHaveLength, 1)
		target1 := result.GetRequests()[0].GetScheduleTargets()[0].GetTargets()[0]
		target2 := result.GetRequests()[1].GetScheduleTargets()[0].GetTargets()[0]
		if target1.GetSwTarget().GetLegacySw().GetGcsPath() != "gs://chromeos-image-archive/board-release/R123.0.0" {
			swap := target1
			target1 = target2
			target2 = swap
		}
		So(target1.GetSwTarget().GetLegacySw().GetGcsPath(), ShouldEqual, "gs://chromeos-image-archive/board-release/R123.0.0")
		So(target2.GetSwTarget().GetLegacySw().GetGcsPath(), ShouldEqual, "gs://chromeos-image-archive/board-release/R124.0.0")
		So(result.GetRequests()[0].GetSchedulerInfo().GetScheduler(), ShouldEqual, testapi.SchedulerInfo_QSCHEDULER)
		So(result.GetRequests()[1].GetSchedulerInfo().GetScheduler(), ShouldEqual, testapi.SchedulerInfo_SCHEDUKE)
	})

	Convey("Multi Translation, grouping", t, func() {
		requests := map[string]*test_platform.Request{
			"r1": getCTPv1Request("board", "model", "board-release/R123.0.0", "suite", "", "", false),
			"r2": getCTPv1Request("board", "model2", "board-release/R123.0.0", "suite", "", "", false),
		}
		result := builders.NewCTPV2FromV1WithCustomManifestFetcher(context.Background(), requests, MockManifestFetcher).BuildRequest()

		So(result.GetRequests(), ShouldHaveLength, 1)
		So(result.GetRequests()[0].GetScheduleTargets(), ShouldHaveLength, 2)
		So(result.GetRequests()[0].GetScheduleTargets()[0].GetTargets(), ShouldHaveLength, 1)
		So(result.GetRequests()[0].GetScheduleTargets()[1].GetTargets(), ShouldHaveLength, 1)
		target1 := result.GetRequests()[0].GetScheduleTargets()[0].GetTargets()[0]
		target2 := result.GetRequests()[0].GetScheduleTargets()[1].GetTargets()[0]
		if target1.GetHwTarget().GetLegacyHw().GetModel() != "model" {
			swap := target1
			target1 = target2
			target2 = swap
		}
		So(target1.GetHwTarget().GetLegacyHw().GetModel(), ShouldEqual, "model")
		So(target2.GetHwTarget().GetLegacyHw().GetModel(), ShouldEqual, "model2")
	})

	Convey("Multi Translation, grouping, including public manifest", t, func() {
		requests := map[string]*test_platform.Request{
			"r1": getCTPv1Request("board", "model", "public-manifest-release/R123.0.0", "suite", "", "", false),
			"r2": getCTPv1Request("board", "model2", "board-release/R123.0.0", "suite", "", "", false),
		}
		result := builders.NewCTPV2FromV1WithCustomManifestFetcher(context.Background(), requests, MockManifestFetcher).BuildRequest()

		So(result.GetRequests(), ShouldHaveLength, 2)
		So(result.GetRequests()[0].GetScheduleTargets(), ShouldHaveLength, 1)
		So(result.GetRequests()[1].GetScheduleTargets(), ShouldHaveLength, 1)
		So(result.GetRequests()[0].GetScheduleTargets()[0].GetTargets(), ShouldHaveLength, 1)
		So(result.GetRequests()[1].GetScheduleTargets()[0].GetTargets(), ShouldHaveLength, 1)
		target1 := result.GetRequests()[0].GetScheduleTargets()[0].GetTargets()[0]
		target2 := result.GetRequests()[1].GetScheduleTargets()[0].GetTargets()[0]
		if target1.GetSwTarget().GetLegacySw().GetGcsPath() != "gs://chromeos-image-archive/public-manifest-release/R123.0.0" {
			swap := target1
			target1 = target2
			target2 = swap
		}
		So(target1.GetSwTarget().GetLegacySw().GetGcsPath(), ShouldEqual, "gs://chromeos-image-archive/public-manifest-release/R123.0.0")
		So(target2.GetSwTarget().GetLegacySw().GetGcsPath(), ShouldEqual, "gs://chromeos-image-archive/board-release/R123.0.0")
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

	// Currently this incorrectly pulls the variant for the example below; but without better variant info,
	// this appears to be difficult to safely solve. Leaving this test in here as an example of something
	// to look into long term.
	// Convey("GetVariant remove prefix and postfix", t, func() {
	// 	variant := builders.GetVariant(getChromeosSoftwareDeps("staging-rex-release-R124-15823.B/R124-15823.9.0-8752476513443194785"))
	// 	So(variant, ShouldEqual, "")
	// })

	Convey("GetVariant empty string", t, func() {
		variant := builders.GetVariant(getChromeosSoftwareDeps(""))
		So(variant, ShouldEqual, "")
	})
}

func TestCTP2Grouping(t *testing.T) {
	Convey("Same build, same suite, diff boards", t, func() {
		groupings := builders.GroupEligibleV2Requests(context.Background(), []*testapi.CTPRequest{
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
		groupings := builders.GroupEligibleV2Requests(context.Background(), []*testapi.CTPRequest{
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
		groupings := builders.GroupEligibleV2Requests(context.Background(), []*testapi.CTPRequest{
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
		groupings := builders.GroupEligibleV2Requests(context.Background(), []*testapi.CTPRequest{
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

func getCTPv1Request(board, model, build, suite, testArgs string, analyticsName string, runWithQs bool) *test_platform.Request {
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

					fmt.Sprintf("analytics_name:%s", analyticsName),
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
			RunCtpv2WithQs: runWithQs,
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

// isDDDSuite will return if the suite is to run in ddd.
// For now, use the ddd prefix, but long term will move to a proper flag.
func TestIsDDDSuite(t *testing.T) {
	requests := map[string]*test_platform.Request{
		"r1": getCTPv1Request("board", "model", "board-release/R123.0.0", "suite", "", "ddd_meme", false),
		"r2": getCTPv1Request("board", "model", "board-release/R123.0.0", "suite", "", "not_ddd_suite", false),
	}
	if builders.IsDDDSuite(requests["r1"]) != true {
		t.Fatalf("Incorrectly determined if a request was for 3d")
	}
	if builders.IsDDDSuite(requests["r2"]) != false {
		t.Fatalf("Incorrectly determined if a request was for 3d")
	}
}
