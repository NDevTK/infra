// Copyright 2022 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	apipb "go.chromium.org/chromiumos/config/go/test/api"
	artifactpb "go.chromium.org/chromiumos/config/go/test/artifact"
	"go.chromium.org/luci/resultdb/pbutil"
	pb "go.chromium.org/luci/resultdb/proto/v1"
	sinkpb "go.chromium.org/luci/resultdb/sink/proto/v1"
	protojson "google.golang.org/protobuf/encoding/protojson"
)

// Follows ChromeOS test result convention. Use TestResult to represent the
// ChromeOS test result contract proto defined in
// https://source.chromium.org/chromiumos/chromiumos/codesearch/+/main:src/config/proto/chromiumos/test/artifact/test_result.proto
type CrosTestResult struct {
	TestResult *artifactpb.TestResult `json:"test_result"`
}

// ConvertFromJSON reads the provided reader into the receiver.
// The TestResult is cleared and overwritten.
func (r *CrosTestResult) ConvertFromJSON(reader io.Reader) error {
	r.TestResult = &artifactpb.TestResult{}
	var rawMessage json.RawMessage
	if err := json.NewDecoder(reader).Decode(&rawMessage); err != nil {
		return err
	}
	if err := protojson.Unmarshal(rawMessage, r.TestResult); err != nil {
		return err
	}
	return nil
}

// ToProtos converts ChromeOS test results in r to []*sinkpb.TestResult.
func (r *CrosTestResult) ToProtos(ctx context.Context) ([]*sinkpb.TestResult, error) {
	var ret []*sinkpb.TestResult
	for _, testRun := range r.TestResult.TestRuns {
		testCaseMatadata := testRun.TestCaseInfo.TestCaseMetadata
		testCaseResult := testRun.TestCaseInfo.TestCaseResult
		status := genTestResultStatus(testCaseResult)
		tr := &sinkpb.TestResult{
			TestId:   testCaseMatadata.TestCase.Name,
			Status:   status,
			Expected: status == pb.TestStatus_PASS,
			Tags:     genTestResultTags(testRun),
		}

		if testCaseResult.GetReason() != "" {
			tr.FailureReason = &pb.FailureReason{
				PrimaryErrorMessage: truncateString(testCaseResult.GetReason(), maxPrimaryErrorBytes),
			}
		}

		if testCaseResult.StartTime.CheckValid() == nil {
			tr.StartTime = testCaseResult.StartTime
			if testCaseResult.Duration.CheckValid() == nil {
				tr.Duration = testCaseResult.Duration
			}
		}

		ret = append(ret, tr)
	}
	return ret, nil
}

// Converts a TestCase Verdict into a ResultSink Status.
func genTestResultStatus(result *apipb.TestCaseResult) pb.TestStatus {
	switch result.Verdict.(type) {
	case *apipb.TestCaseResult_Pass_:
		return pb.TestStatus_PASS
	case *apipb.TestCaseResult_Fail_:
		return pb.TestStatus_FAIL
	case *apipb.TestCaseResult_Crash_:
		return pb.TestStatus_CRASH
	case *apipb.TestCaseResult_Abort_:
		return pb.TestStatus_ABORT
	// TODO(b/240893570): Split SKIP and NOT_RUN status once ResultDB can
	// support rich statuses for ChromeOS test results.
	case *apipb.TestCaseResult_Skip_:
	case *apipb.TestCaseResult_NotRun_:
		return pb.TestStatus_SKIP
	default:
		return pb.TestStatus_STATUS_UNSPECIFIED
	}
	return pb.TestStatus_STATUS_UNSPECIFIED
}

// Generates test result tags based on the structured ChromeOS test run proto.
func genTestResultTags(testRun *artifactpb.TestRun) []*pb.StringPair {
	tags := []*pb.StringPair{}

	primaryExecInfo := testRun.PrimaryExecutionInfo
	primaryBuildInfo := primaryExecInfo.BuildInfo
	tags = append(tags, pbutil.StringPair("image", primaryBuildInfo.Name))
	tags = append(tags, pbutil.StringPair("build", strings.Split(primaryBuildInfo.Name, "/")[1]))
	tags = append(tags, pbutil.StringPair("board", primaryBuildInfo.Board))

	primaryDutInfo := primaryExecInfo.DutInfo
	dut := primaryDutInfo.Dut
	tags = append(tags, pbutil.StringPair("model", dut.GetChromeos().DutModel.ModelName))
	tags = append(tags, pbutil.StringPair("hostname", dut.GetChromeos().Name))

	for _, logInfo := range testRun.LogsInfo {
		tags = append(tags, pbutil.StringPair("logs_url", logInfo.Path))
	}

	return tags
}
