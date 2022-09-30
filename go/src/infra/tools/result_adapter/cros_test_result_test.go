// Copyright 2022 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"strings"
	"testing"

	"github.com/golang/protobuf/ptypes/duration"
	. "github.com/smartystreets/goconvey/convey"
	configpb "go.chromium.org/chromiumos/config/go"
	apipb "go.chromium.org/chromiumos/config/go/test/api"
	artifactpb "go.chromium.org/chromiumos/config/go/test/artifact"
	labpb "go.chromium.org/chromiumos/config/go/test/lab/api"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/resultdb/pbutil"
	pb "go.chromium.org/luci/resultdb/proto/v1"
	sinkpb "go.chromium.org/luci/resultdb/sink/proto/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCrosTestResultConversions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testResultsJson := `
	{
		"version":1,
		"test_runs":[
			{
				"test_case_info":{
					"test_case_metadata":{
						"test_case":{
							"id":{
								"value":"invocations/build-8803850119519478545/tests/rlz_CheckPing/results/567764de-00001"
							},
							"name":"rlz_CheckPing"
						}
					},
					"test_case_result":{
						"pass":{},
						"start_time":"2022-09-07T18:53:33.983328614Z",
						"duration": "60s"
					}
				},
				"logs_info":[
					{
						"host_type":"GS",
						"path":"gs://chromeos-test-logs/test-runner/prod/2022-09-07/98098abe-da4f-4bfa-bef5-9cbc4936da03"
					}
				],
				"primary_execution_info":{
					"build_info":{
						"name":"hatch-cq/R106-15048.0.0",
						"milestone":106,
						"chrome_os_version":"15048.0.0",
						"source":"hatch-cq",
						"board":"hatch"
					},
					"dut_info":{
						"dut":{
							"chromeos":{
								"name":"chromeos15-row4-rack5-host1",
								"dut_model":{
									"model_name":"nipperkin"
								}
							}
						}
					}
				}
			},
			{
				"test_case_info":{
					"test_case_metadata":{
						"test_case":{
							"id":{
								"value":"invocations/build-8803850119519478545/tests/power_Resume/results/567764de-00002"
							},
							"name":"power_Resume"
						}
					},
					"test_case_result":{
						"fail":{},
						"reason": "Test failed",
						"start_time":"2022-09-07T18:53:34.983328614Z",
						"duration": "120.1s"
					}
				},
				"logs_info":[
					{
						"host_type":"GS",
						"path":"gs://chromeos-test-logs/test-runner/prod/2022-09-07/98098abe-da4f-4bfa-bef5-9cbc4936da04"
					}
				],
				"primary_execution_info":{
					"build_info":{
						"name":"dedede-cq/R106-15054.52.0",
						"milestone":106,
						"chrome_os_version":"15054.52.0",
						"source":"dedede-cq",
						"board":"dedede"
					},
					"dut_info":{
						"dut":{
							"chromeos":{
								"name":"chromeos8-row13-rack20-host28",
								"dut_model":{
									"model_name":"storo"
								}
							}
						}
					}
				}
			}
		]
	}
	`

	testRuns := []*artifactpb.TestRun{
		{
			TestCaseInfo: &artifactpb.TestCaseInfo{
				TestCaseMetadata: &apipb.TestCaseMetadata{
					TestCase: &apipb.TestCase{
						Id: &apipb.TestCase_Id{
							Value: "invocations/build-8803850119519478545/tests/rlz_CheckPing/results/567764de-00001",
						},
						Name: "rlz_CheckPing",
					},
				},
				TestCaseResult: &apipb.TestCaseResult{
					Verdict:   &apipb.TestCaseResult_Pass_{},
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:33.983328614Z")),
					Duration:  &duration.Duration{Seconds: 60},
				},
			},
			LogsInfo: []*configpb.StoragePath{
				{
					HostType: configpb.StoragePath_GS,
					Path:     "gs://chromeos-test-logs/test-runner/prod/2022-09-07/98098abe-da4f-4bfa-bef5-9cbc4936da03",
				},
			},
			PrimaryExecutionInfo: &artifactpb.ExecutionInfo{
				BuildInfo: &artifactpb.BuildInfo{
					Name:            "hatch-cq/R106-15048.0.0",
					Milestone:       106,
					ChromeOsVersion: "15048.0.0",
					Source:          "hatch-cq",
					Board:           "hatch",
				},
				DutInfo: &artifactpb.DutInfo{
					Dut: &labpb.Dut{
						DutType: &labpb.Dut_Chromeos{
							Chromeos: &labpb.Dut_ChromeOS{
								Name: "chromeos15-row4-rack5-host1",
								DutModel: &labpb.DutModel{
									ModelName: "nipperkin",
								},
							},
						},
					},
				},
			},
		},
		{
			TestCaseInfo: &artifactpb.TestCaseInfo{
				TestCaseMetadata: &apipb.TestCaseMetadata{
					TestCase: &apipb.TestCase{
						Id: &apipb.TestCase_Id{
							Value: "invocations/build-8803850119519478545/tests/power_Resume/results/567764de-00002",
						},
						Name: "power_Resume",
					},
				},
				TestCaseResult: &apipb.TestCaseResult{
					Verdict:   &apipb.TestCaseResult_Fail_{},
					Reason:    "Test failed",
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:34.983328614Z")),
					Duration:  &duration.Duration{Seconds: 120, Nanos: 100000000},
				},
			},
			LogsInfo: []*configpb.StoragePath{
				{
					HostType: configpb.StoragePath_GS,
					Path:     "gs://chromeos-test-logs/test-runner/prod/2022-09-07/98098abe-da4f-4bfa-bef5-9cbc4936da04",
				},
			},
			PrimaryExecutionInfo: &artifactpb.ExecutionInfo{
				BuildInfo: &artifactpb.BuildInfo{
					Name:            "dedede-cq/R106-15054.52.0",
					Milestone:       106,
					ChromeOsVersion: "15054.52.0",
					Source:          "dedede-cq",
					Board:           "dedede",
				},
				DutInfo: &artifactpb.DutInfo{
					Dut: &labpb.Dut{
						DutType: &labpb.Dut_Chromeos{
							Chromeos: &labpb.Dut_ChromeOS{
								Name: "chromeos8-row13-rack20-host28",
								DutModel: &labpb.DutModel{
									ModelName: "storo",
								},
							},
						},
					},
				},
			},
		},
	}

	Convey(`From JSON works`, t, func() {
		results := &CrosTestResult{}
		err := results.ConvertFromJSON(strings.NewReader(testResultsJson))
		So(err, ShouldBeNil)
		So(results.TestResult.TestRuns, ShouldResembleProto, testRuns)
	})

	Convey(`ToProtos works`, t, func() {
		Convey("Basic", func() {
			results := &CrosTestResult{}
			results.ConvertFromJSON(strings.NewReader(testResultsJson))
			testResults, err := results.ToProtos(ctx)
			So(err, ShouldBeNil)

			expected := []*sinkpb.TestResult{
				{
					TestId:    "rlz_CheckPing",
					Expected:  true,
					Status:    pb.TestStatus_PASS,
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:33.983328614Z")),
					Duration:  &duration.Duration{Seconds: 60},
					Tags: []*pb.StringPair{
						pbutil.StringPair("image", "hatch-cq/R106-15048.0.0"),
						pbutil.StringPair("build", "R106-15048.0.0"),
						pbutil.StringPair("board", "hatch"),
						pbutil.StringPair("model", "nipperkin"),
						pbutil.StringPair("hostname", "chromeos15-row4-rack5-host1"),
						pbutil.StringPair("logs_url", "gs://chromeos-test-logs/test-runner/prod/2022-09-07/98098abe-da4f-4bfa-bef5-9cbc4936da03"),
					},
				},
				{
					TestId:   "power_Resume",
					Expected: false,
					Status:   pb.TestStatus_FAIL,
					FailureReason: &pb.FailureReason{
						PrimaryErrorMessage: "Test failed",
					},
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:34.983328614Z")),
					Duration:  &duration.Duration{Seconds: 120, Nanos: 100000000},
					Tags: []*pb.StringPair{
						pbutil.StringPair("image", "dedede-cq/R106-15054.52.0"),
						pbutil.StringPair("build", "R106-15054.52.0"),
						pbutil.StringPair("board", "dedede"),
						pbutil.StringPair("model", "storo"),
						pbutil.StringPair("hostname", "chromeos8-row13-rack20-host28"),
						pbutil.StringPair("logs_url", "gs://chromeos-test-logs/test-runner/prod/2022-09-07/98098abe-da4f-4bfa-bef5-9cbc4936da04"),
					},
				},
			}
			So(testResults, ShouldHaveLength, 2)
			So(testResults, ShouldResembleProto, expected)
		})
	})
}
