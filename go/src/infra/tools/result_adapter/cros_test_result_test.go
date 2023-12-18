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

const (
	// Test result JSON file with a subset of common information.
	simpleTestResultFile = "test_data/cros_test_result/simple_test_result.json"

	// Test result JSON file with full information.
	fullTestResultFile = "test_data/cros_test_result/full_test_result.json"

	// Test result JSON file with multi-DUT testing information.
	multiDUTTestResultFile = "test_data/cros_test_result/multi_dut_result.json"

	// Test result JSON file with skipped test results.
	skippedTestResultFile = "test_data/cros_test_result/skipped_test_result.json"

	// Test result JSON file with passed (with warning) test results.
	warnTestResultFile = "test_data/cros_test_result/warn_test_result.json"

	// Test result JSON file with failing test results.
	failedTestResultFile = "test_data/cros_test_result/failed_test_result.json"

	// Test result JSON file with missing test id.
	missingTestIdFile = "test_data/cros_test_result/missing_test_id.json"

	// Test result JSON file with warning test results and a long reason.
	warnTestResultWithLongReasonFile = "test_data/cros_test_result/warn_test_result_with_long_reason.json"
)

func TestCrosTestResultConversions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testResultsJSON := ReadJSONFileToString(simpleTestResultFile)

	testResult := &artifactpb.TestResult{
		Version: 1,
		TestInvocation: &artifactpb.TestInvocation{
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
						Id: &labpb.Dut_Id{
							Value: "chromeos15-row4-rack5-host1",
						},
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
		TestRuns: []*artifactpb.TestRun{
			{
				TestCaseInfo: &artifactpb.TestCaseInfo{
					TestCaseMetadata: &apipb.TestCaseMetadata{
						TestCase: &apipb.TestCase{
							Id: &apipb.TestCase_Id{
								Value: "rlz_CheckPing",
							},
							Name: "rlz_CheckPing",
						},
					},
					TestCaseResult: &apipb.TestCaseResult{
						TestCaseId: &apipb.TestCase_Id{
							Value: "rlz_CheckPing",
						},
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
			},
			{
				TestCaseInfo: &artifactpb.TestCaseInfo{
					TestCaseMetadata: &apipb.TestCaseMetadata{
						TestCase: &apipb.TestCase{
							Id: &apipb.TestCase_Id{
								Value: "power_Resume",
							},
							Name: "power_Resume",
						},
					},
					TestCaseResult: &apipb.TestCaseResult{
						TestCaseId: &apipb.TestCase_Id{
							Value: "power_Resume",
						},
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
			},
		},
	}

	Convey(`From JSON works`, t, func() {
		results := &CrosTestResult{}
		err := results.ConvertFromJSON(strings.NewReader(testResultsJSON))
		So(err, ShouldBeNil)
		So(results.TestResult, ShouldResembleProto, testResult)
	})

	Convey(`ToProtos works`, t, func() {
		Convey("Basic", func() {
			results := &CrosTestResult{}
			err := results.ConvertFromJSON(strings.NewReader(testResultsJSON))
			So(err, ShouldBeNil)
			testResults, err := results.ToProtos(ctx)
			So(err, ShouldBeNil)

			expected := []*sinkpb.TestResult{
				{
					TestId:    "rlz_CheckPing",
					Expected:  true,
					Status:    pb.TestStatus_PASS,
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:33.983328614Z")),
					Duration:  &duration.Duration{Seconds: 60},
					Tags: SortTags([]*pb.StringPair{
						pbutil.StringPair("board", "hatch"),
						pbutil.StringPair("build", "R106-15048.0.0"),
						pbutil.StringPair("cbx", "false"),
						pbutil.StringPair("hostname", "chromeos15-row4-rack5-host1"),
						pbutil.StringPair("image", "hatch-cq/R106-15048.0.0"),
						pbutil.StringPair("is_cft_run", "True"),
						pbutil.StringPair("logs_url", "gs://chromeos-test-logs/test-runner/prod/2022-09-07/98098abe-da4f-4bfa-bef5-9cbc4936da03"),
						pbutil.StringPair("model", "nipperkin"),
						pbutil.StringPair("multiduts", "False"),
					}),
				},
				{
					TestId:   "power_Resume",
					Expected: false,
					Status:   pb.TestStatus_FAIL,
					FailureReason: &pb.FailureReason{
						PrimaryErrorMessage: "Test failed",
						Errors: []*pb.FailureReason_Error{
							{Message: "Test failed"},
						},
					},
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:34.983328614Z")),
					Duration:  &duration.Duration{Seconds: 120, Nanos: 100000000},
					Tags: SortTags([]*pb.StringPair{
						pbutil.StringPair("board", "hatch"),
						pbutil.StringPair("build", "R106-15048.0.0"),
						pbutil.StringPair("cbx", "false"),
						pbutil.StringPair("hostname", "chromeos15-row4-rack5-host1"),
						pbutil.StringPair("image", "hatch-cq/R106-15048.0.0"),
						pbutil.StringPair("is_cft_run", "True"),
						pbutil.StringPair("logs_url", "gs://chromeos-test-logs/test-runner/prod/2022-09-07/98098abe-da4f-4bfa-bef5-9cbc4936da04"),
						pbutil.StringPair("model", "nipperkin"),
						pbutil.StringPair("multiduts", "False"),
					}),
				},
			}

			for i, tr := range expected {
				err := PopulateProperties(tr, results.TestResult.TestRuns[i])
				So(err, ShouldBeNil)
			}

			So(testResults, ShouldHaveLength, 2)
			So(testResults, ShouldResembleProto, expected)
			for _, tr := range testResults {
				So(tr.GetProperties().GetFields(), ShouldNotBeEmpty)
			}
		})

		Convey(`Check expected skip and unexpected skip tests`, func() {
			testResultsJSON := ReadJSONFileToString(skippedTestResultFile)
			results := &CrosTestResult{}
			err := results.ConvertFromJSON(strings.NewReader(testResultsJSON))
			So(err, ShouldBeNil)
			testResults, err := results.ToProtos(ctx)
			So(err, ShouldBeNil)

			expected := []*sinkpb.TestResult{
				{
					TestId:   "rlz_CheckPing",
					Expected: true,
					Status:   pb.TestStatus_SKIP,
					FailureReason: &pb.FailureReason{
						PrimaryErrorMessage: "Test was skipped expectedly",
						Errors: []*pb.FailureReason_Error{
							{Message: "Test was skipped expectedly"},
						},
					},
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:33.983328614Z")),
					Duration:  &duration.Duration{Seconds: 60},
					Tags: SortTags([]*pb.StringPair{
						pbutil.StringPair("is_cft_run", "True"),
					}),
				},
				{
					TestId:   "power_Resume",
					Expected: false,
					Status:   pb.TestStatus_SKIP,
					FailureReason: &pb.FailureReason{
						PrimaryErrorMessage: "Test has not run yet",
						Errors: []*pb.FailureReason_Error{
							{Message: "Test has not run yet"},
						},
					},
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:34.983328614Z")),
					Duration:  &duration.Duration{Seconds: 120, Nanos: 100000000},
					Tags: SortTags([]*pb.StringPair{
						pbutil.StringPair("is_cft_run", "True"),
					}),
				},
			}

			for i, tr := range expected {
				err := PopulateProperties(tr, results.TestResult.TestRuns[i])
				So(err, ShouldBeNil)
			}

			So(testResults, ShouldHaveLength, 2)
			So(testResults, ShouldResembleProto, expected)
			for _, tr := range testResults {
				So(tr.GetProperties().GetFields(), ShouldNotBeEmpty)
			}
		})
		Convey(`Warning results`, func() {
			testResultsJSON := ReadJSONFileToString(warnTestResultFile)
			results := &CrosTestResult{}
			err := results.ConvertFromJSON(strings.NewReader(testResultsJSON))
			So(err, ShouldBeNil)
			testResults, err := results.ToProtos(ctx)
			So(err, ShouldBeNil)

			expected := []*sinkpb.TestResult{
				{
					TestId:   "rlz_CheckPing",
					Expected: true,
					// Warning results are reported as pass, and without
					// the failure reason set. Warning messages are included
					// in the test result properties.
					Status:    pb.TestStatus_PASS,
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:33.983328614Z")),
					Duration:  &duration.Duration{Seconds: 60},
					Tags: SortTags([]*pb.StringPair{
						pbutil.StringPair("is_cft_run", "True"),
					}),
				},
			}

			for i, tr := range expected {
				err := PopulateProperties(tr, results.TestResult.TestRuns[i])
				So(err, ShouldBeNil)
			}

			So(testResults, ShouldResembleProto, expected)
			for _, tr := range testResults {
				So(tr.GetProperties().GetFields(), ShouldNotBeEmpty)
			}
		})
		Convey(`Failed results`, func() {
			testResultsJSON := ReadJSONFileToString(failedTestResultFile)
			results := &CrosTestResult{}
			err := results.ConvertFromJSON(strings.NewReader(testResultsJSON))
			So(err, ShouldBeNil)
			testResults, err := results.ToProtos(ctx)
			So(err, ShouldBeNil)

			expected := []*sinkpb.TestResult{
				{
					TestId:   "rlz_CheckPing",
					Expected: false,
					Status:   pb.TestStatus_ABORT,
					FailureReason: &pb.FailureReason{
						PrimaryErrorMessage: "Failed to start Chrome: login failed: context timeout",
						Errors: []*pb.FailureReason_Error{
							{Message: "Failed to start Chrome: login failed: context timeout"},
						},
					},
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:33.983328614Z")),
					Duration:  &duration.Duration{Seconds: 60},
					Tags: SortTags([]*pb.StringPair{
						pbutil.StringPair("is_cft_run", "True"),
					}),
				},
				{
					TestId:   "power_Resume",
					Expected: false,
					Status:   pb.TestStatus_FAIL,
					FailureReason: &pb.FailureReason{
						PrimaryErrorMessage: "Failed to start Chrome: login failed: OOBE not dismissed, it is on screen \"signin-fatal-error\"",
						Errors: []*pb.FailureReason_Error{
							{Message: "Failed to start Chrome: login failed: OOBE not dismissed, it is on screen \"signin-fatal-error\""},
							{Message: "Failed to clean-up Chrome: some error"},
							{Message: "Error three"},
						},
					},
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:34.983328614Z")),
					Duration:  &duration.Duration{Seconds: 120, Nanos: 100000000},
					Tags: SortTags([]*pb.StringPair{
						pbutil.StringPair("is_cft_run", "True"),
					}),
				},
			}

			for i, tr := range expected {
				err := PopulateProperties(tr, results.TestResult.TestRuns[i])
				So(err, ShouldBeNil)
			}

			So(testResults, ShouldResembleProto, expected)
			for _, tr := range testResults {
				So(tr.GetProperties().GetFields(), ShouldNotBeEmpty)
			}
		})

		Convey(`Check the full list of tags`, func() {
			testResultsJSON := ReadJSONFileToString(fullTestResultFile)
			results := &CrosTestResult{}
			err := results.ConvertFromJSON(strings.NewReader(testResultsJSON))
			So(err, ShouldBeNil)
			testResults, err := results.ToProtos(ctx)
			So(err, ShouldBeNil)

			expected := []*sinkpb.TestResult{
				{
					TestId:    "rlz_CheckPing",
					Expected:  true,
					Status:    pb.TestStatus_PASS,
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:33.983328614Z")),
					Duration:  &duration.Duration{Seconds: 60},
					Tags: SortTags([]*pb.StringPair{
						pbutil.StringPair("ancestor_buildbucket_ids", "8814950840874708945,8814951792758733697"),
						pbutil.StringPair("board", "hatch"),
						pbutil.StringPair("branch", "main"),
						pbutil.StringPair("build", "R106-15048.0.0"),
						pbutil.StringPair("buildbucket_builder", "test_runner"),
						pbutil.StringPair("cbx", "true"),
						pbutil.StringPair("contacts", "user@google.com"),
						pbutil.StringPair("declared_name", "hatch-cq/R102-14632.0.0-62834-8818718496810023809/wificell-cq/tast.wificell-cq"),
						pbutil.StringPair("suite", "arc-cts-vm"),
						pbutil.StringPair("drone", "skylab-drone-deployment-prod-6dc79d4f9-czjlj"),
						pbutil.StringPair("drone_server", "chromeos4-row4-rack1-drone8"),
						pbutil.StringPair("hostname", "chromeos15-row4-rack5-host1"),
						pbutil.StringPair("image", "hatch-cq/R106-15048.0.0"),
						pbutil.StringPair("is_cft_run", "True"),
						pbutil.StringPair("job_name", "bb-8818737803155059937-chromeos/general/Full"),
						pbutil.StringPair("logs_url", "gs://chromeos-test-logs/test-runner/prod/2022-09-07/98098abe-da4f-4bfa-bef5-9cbc4936da03"),
						pbutil.StringPair("main_builder_name", "main-release"),
						pbutil.StringPair("model", "nipperkin"),
						pbutil.StringPair("multiduts", "False"),
						pbutil.StringPair("pool", "ChromeOSSkylab"),
						pbutil.StringPair("queued_time", "2022-06-03 18:53:33.983328614 +0000 UTC"),
						pbutil.StringPair("ro_fwid", "Google_Voema.13672.224.0"),
						pbutil.StringPair("rw_fwid", "Google_Voema.13672.224.0"),
						pbutil.StringPair("suite_task_id", "59ef5e9532bbd611"),
						pbutil.StringPair("task_id", "59f0e13fe7af0710"),
						pbutil.StringPair("label_pool", "DUT_POOL_QUOTA"),
						pbutil.StringPair("wifi_chip", "marvell"),
						pbutil.StringPair("kernel_version", "5.4.151-16902-g93699f4e73de"),
						pbutil.StringPair("hwid_sku", "katsu_MT8183_0B"),
						pbutil.StringPair("carrier", "CARRIER_ESIM"),
						pbutil.StringPair("ash_version", "109.0.5391.0"),
						pbutil.StringPair("lacros_version", "109.0.5391.0"),
					}),
				},
			}
			err = PopulateProperties(expected[0], results.TestResult.TestRuns[0])
			So(err, ShouldBeNil)

			So(testResults, ShouldHaveLength, 1)
			So(testResults, ShouldResembleProto, expected)
			So(testResults[0].GetProperties().GetFields(), ShouldNotBeEmpty)
		})

		Convey(`Check multi DUT testing`, func() {
			testResultsJSON := ReadJSONFileToString(multiDUTTestResultFile)
			results := &CrosTestResult{}
			err := results.ConvertFromJSON(strings.NewReader(testResultsJSON))
			So(err, ShouldBeNil)
			testResults, err := results.ToProtos(ctx)
			So(err, ShouldBeNil)

			expected := []*sinkpb.TestResult{
				{
					TestId:    "rlz_CheckPing",
					Expected:  true,
					Status:    pb.TestStatus_PASS,
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:33.983328614Z")),
					Duration:  &duration.Duration{Seconds: 60},
					Tags: SortTags([]*pb.StringPair{
						pbutil.StringPair("board", "hatch"),
						pbutil.StringPair("build", "R106-15048.0.0"),
						pbutil.StringPair("cbx", "false"),
						pbutil.StringPair("hostname", "chromeos15-row4-rack5-host1"),
						pbutil.StringPair("image", "hatch-cq/R106-15048.0.0"),
						pbutil.StringPair("is_cft_run", "True"),
						pbutil.StringPair("logs_url", "gs://chromeos-test-logs/test-runner/prod/2022-09-07/98098abe-da4f-4bfa-bef5-9cbc4936da03"),
						pbutil.StringPair("model", "nipperkin"),
						pbutil.StringPair("multiduts", "True"),
						pbutil.StringPair("primary_board", "hatch"),
						pbutil.StringPair("primary_model", "nipperkin"),
						pbutil.StringPair("secondary_boards", "brya"),
						pbutil.StringPair("secondary_models", "gimble"),
					}),
				},
			}

			err = PopulateProperties(expected[0], results.TestResult.TestRuns[0])
			So(err, ShouldBeNil)

			So(testResults, ShouldHaveLength, 1)
			So(testResults, ShouldResembleProto, expected)
			So(testResults[0].GetProperties().GetFields(), ShouldNotBeEmpty)
		})

		Convey(`Check with missing test id`, func() {
			// There are 3 test cases in the test file. The first test case has
			// both id and name while the second one only has the name. The
			// third one doesn't have id and name, so it would throw an error
			// to surface the problem explicitly and the first two would be
			// skipped.
			testResultsJSON := ReadJSONFileToString(missingTestIdFile)
			results := &CrosTestResult{}
			err := results.ConvertFromJSON(strings.NewReader(testResultsJSON))
			So(err, ShouldBeNil)
			_, err = results.ToProtos(ctx)
			So(err, ShouldErrLike, "testId is unspecified due to the missing id in test case")
		})

		Convey(`Truncate reason field when stored in the properties of test result`, func() {
			testResultsJSON := ReadJSONFileToString(warnTestResultWithLongReasonFile)
			results := &CrosTestResult{}
			err := results.ConvertFromJSON(strings.NewReader(testResultsJSON))
			So(err, ShouldBeNil)
			testResults, err := results.ToProtos(ctx)
			So(err, ShouldBeNil)

			expected := []*sinkpb.TestResult{
				{
					TestId:   "rlz_CheckPing",
					Expected: true,
					// Warning results are reported as pass, and without
					// the failure reason set. Warning messages are included
					// in the test result properties.
					Status:    pb.TestStatus_PASS,
					StartTime: timestamppb.New(parseTime("2022-09-07T18:53:33.983328614Z")),
					Duration:  &duration.Duration{Seconds: 60},
					Tags: SortTags([]*pb.StringPair{
						pbutil.StringPair("is_cft_run", "True"),
					}),
				},
			}

			for i, tr := range expected {
				err := PopulateProperties(tr, results.TestResult.TestRuns[i])
				So(err, ShouldBeNil)
			}

			So(testResults, ShouldResembleProto, expected)
			for _, tr := range testResults {
				So(tr.GetProperties().
					GetFields()["testCaseInfo"].GetStructValue().
					GetFields()["testCaseResult"].GetStructValue().
					GetFields()["reason"].GetStringValue(),
					ShouldHaveLength, 1024)
			}
		})
	})
}
