// Copyright 2021 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"strings"
	"testing"

	pb "go.chromium.org/luci/resultdb/proto/v1"
	sinkpb "go.chromium.org/luci/resultdb/sink/proto/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/golang/protobuf/ptypes/duration"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSkylabTestRunnerConversions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tc := []TestRunnerTestCase{
		{
			Name:      "test1",
			Verdict:   "VERDICT_PASS",
			StartTime: parseTime("2021-07-26T18:53:33.983328614Z"),
			EndTime:   parseTime("2021-07-26T18:53:37.983328614Z"),
		},
		// No EndTime
		{
			Name:      "test2",
			Verdict:   "VERDICT_NO_VERDICT",
			StartTime: parseTime("2021-07-26T18:53:33.983328614Z"),
		},
		// No StartTime and EndTime
		{
			Name:                 "test3",
			Verdict:              "VERDICT_FAIL",
			HumanReadableSummary: "test failure",
		},
		{
			Name:                 "test4",
			Verdict:              "VERDICT_ERROR",
			HumanReadableSummary: "test error",
			StartTime:            parseTime("2021-07-26T18:53:33.983328614Z"),
			EndTime:              parseTime("2021-07-26T18:53:37.983328614Z"),
		},
		{
			Name:                 "test5",
			Verdict:              "VERDICT_ABORT",
			HumanReadableSummary: "test abort",
			StartTime:            parseTime("2021-07-26T18:53:33.983328614Z"),
			EndTime:              parseTime("2021-07-26T18:53:37.983328614Z"),
		},
	}

	results := TestRunnerResult{Autotest: TestRunnerAutotest{
		TestCases: tc,
	}}

	Convey(`From JSON works`, t, func() {
		str := `{
			"autotest_result": {
			  "test_cases": [
				{
					"verdict": "VERDICT_PASS",
					"name": "test1",
					"start_time": "2021-07-26T18:53:33.983328614Z",
					"end_time": "2021-07-26T18:53:37.983328614Z"
				},
				{
					"verdict": "VERDICT_NO_VERDICT",
					"name": "test2",
					"start_time": "2021-07-26T18:53:33.983328614Z"
				},
				{
					"verdict": "VERDICT_FAIL",
					"name": "test3",
					"human_readable_summary": "test failure"
				},
                {
                    "verdict": "VERDICT_ERROR",
                    "name": "test4",
                    "human_readable_summary": "test error",
					"start_time": "2021-07-26T18:53:33.983328614Z",
					"end_time": "2021-07-26T18:53:37.983328614Z"
                },
                {
                    "verdict": "VERDICT_ABORT",
                    "name": "test5",
                    "human_readable_summary": "test abort",
					"start_time": "2021-07-26T18:53:33.983328614Z",
					"end_time": "2021-07-26T18:53:37.983328614Z"
                }
			  ]
			}
		  }`

		results := &TestRunnerResult{}
		err := results.ConvertFromJSON(strings.NewReader(str))
		So(err, ShouldBeNil)
		So(results.Autotest.TestCases, ShouldResemble, tc)
	})

	Convey(`ToProtos`, t, func() {
		Convey("test passes", func() {

			testResults, err := results.ToProtos(ctx)
			So(err, ShouldBeNil)

			expected := []*sinkpb.TestResult{
				{
					TestId:    "test1",
					Expected:  true,
					Status:    pb.TestStatus_PASS,
					StartTime: timestamppb.New(parseTime("2021-07-26T18:53:33.983328614Z")),
					Duration:  &duration.Duration{Seconds: 4},
				},
				{
					TestId:    "test2",
					Expected:  false,
					Status:    pb.TestStatus_SKIP,
					StartTime: timestamppb.New(parseTime("2021-07-26T18:53:33.983328614Z")),
				},
				{
					TestId:      "test3",
					Expected:    false,
					Status:      pb.TestStatus_FAIL,
					SummaryHtml: "<pre>test failure</pre>",
					FailureReason: &pb.FailureReason{
						PrimaryErrorMessage: "test failure",
					},
				},
				{
					TestId:      "test4",
					Expected:    false,
					Status:      pb.TestStatus_CRASH,
					SummaryHtml: "<pre>test error</pre>",
					FailureReason: &pb.FailureReason{
						PrimaryErrorMessage: "test error",
					},
					StartTime: timestamppb.New(parseTime("2021-07-26T18:53:33.983328614Z")),
					Duration:  &duration.Duration{Seconds: 4},
				},
				{
					TestId:      "test5",
					Expected:    false,
					Status:      pb.TestStatus_ABORT,
					SummaryHtml: "<pre>test abort</pre>",
					FailureReason: &pb.FailureReason{
						PrimaryErrorMessage: "test abort",
					},
					StartTime: timestamppb.New(parseTime("2021-07-26T18:53:33.983328614Z")),
					Duration:  &duration.Duration{Seconds: 4},
				},
			}
			So(testResults, ShouldHaveLength, 5)
			So(testResults, ShouldResemble, expected)
		})
	})
}
