// Copyright 2022 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	apipb "go.chromium.org/chromiumos/config/go/test/api"
	artifactpb "go.chromium.org/chromiumos/config/go/test/artifact"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/resultdb/pbutil"
	pb "go.chromium.org/luci/resultdb/proto/v1"
	sinkpb "go.chromium.org/luci/resultdb/sink/proto/v1"
	protojson "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
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
	for _, testRun := range r.TestResult.GetTestRuns() {
		testCaseInfo := testRun.GetTestCaseInfo()
		testCaseResult := testCaseInfo.GetTestCaseResult()
		status := genTestResultStatus(testCaseResult)
		testId := getTestId(testCaseResult)
		if testId == "" {
			return nil, errors.Reason("TestId is unspecified due to the missing id in test case result: %v",
				testCaseResult).Err()
		}

		tr := &sinkpb.TestResult{
			TestId: testId,
			Status: status,
			// The status is expected if the test passed or was skipped
			// expectedly.
			Expected: status == pb.TestStatus_PASS || testCaseResult.GetSkip() != nil,
			// TODO(b/251357069): Move the invocation-level info and
			// result-level info to the new JSON type columns accordingly when
			// the new JSON type columns are ready in place.
			Tags: genTestResultTags(testRun, r.TestResult.GetTestInvocation()),
		}

		if testCaseResult.GetReason() != "" {
			tr.FailureReason = &pb.FailureReason{
				PrimaryErrorMessage: truncateString(testCaseResult.GetReason(), maxPrimaryErrorBytes),
			}
		}

		if testCaseResult.GetStartTime().CheckValid() == nil {
			tr.StartTime = testCaseResult.GetStartTime()
			if testCaseResult.GetDuration().CheckValid() == nil {
				tr.Duration = testCaseResult.GetDuration()
			}
		}

		if err := PopulateProperties(tr, testRun); err != nil {
			return nil, errors.Annotate(err, "Failed to unmarshal properties for test result").Err()
		}

		ret = append(ret, tr)
	}
	return ret, nil
}

// getTestId gets the test id based on the test case result.
func getTestId(testCaseResult *apipb.TestCaseResult) string {
	if testCaseResult == nil || testCaseResult.GetTestCaseId() == nil {
		return ""
	}
	return testCaseResult.GetTestCaseId().GetValue()
}

// Converts a TestCase Verdict into a ResultSink Status.
func genTestResultStatus(result *apipb.TestCaseResult) pb.TestStatus {
	switch result.GetVerdict().(type) {
	case *apipb.TestCaseResult_Pass_:
		return pb.TestStatus_PASS
	case *apipb.TestCaseResult_Fail_:
		return pb.TestStatus_FAIL
	case *apipb.TestCaseResult_Crash_:
		return pb.TestStatus_CRASH
	case *apipb.TestCaseResult_Abort_:
		return pb.TestStatus_ABORT
	// TODO(b/240893570): Split Skip (aka TestNa) and NotRun status once
	// ResultDB can support rich statuses for ChromeOS test results.
	case *apipb.TestCaseResult_Skip_, *apipb.TestCaseResult_NotRun_:
		return pb.TestStatus_SKIP
	default:
		return pb.TestStatus_STATUS_UNSPECIFIED
	}
}

// PopulateProperties populates the properties of the test result.
func PopulateProperties(testResult *sinkpb.TestResult, testRun *artifactpb.TestRun) error {
	if testRun == nil {
		return nil
	}

	if testResult == nil {
		return errors.Reason("The input test result is nil").Err()
	}

	data, err := protojson.Marshal(testRun)
	if err != nil {
		return err
	}

	testResult.Properties = &structpb.Struct{}
	return protojson.Unmarshal(data, testResult.Properties)
}

// TODO(b/240897202): Remove the tags when a JSON type field is supported in
// ResultDB schema.
// genTestResultTags generates test result tags based on the ChromeOS test
// result contract proto and returns the sorted tags.
func genTestResultTags(testRun *artifactpb.TestRun, testInvocation *artifactpb.TestInvocation) []*pb.StringPair {
	tags := []*pb.StringPair{}

	if testInvocation != nil {
		// For Testhaus MVP parity.
		// Refer to `_generate_resultdb_base_tags` in test_runner recipe:
		// https://source.chromium.org/chromiumos/chromiumos/codesearch/+/main:infra/recipes/recipes/test_platform/test_runner.py;l=472?q=test_platform%2Ftest_runner.py
		primaryExecInfo := testInvocation.GetPrimaryExecutionInfo()
		if primaryExecInfo != nil {
			buildInfo := primaryExecInfo.GetBuildInfo()
			if buildInfo != nil {
				buildName := buildInfo.GetName()
				tags = AppendTags(tags, "image", buildName)
				tags = AppendTags(tags, "build", strings.Split(buildName, "/")[1])
				tags = AppendTags(tags, "board", buildInfo.GetBoard())

				tags = configBuildMetaDataTags(tags, buildInfo.GetBuildMetadata())
			}

			dutInfo := primaryExecInfo.GetDutInfo()
			if dutInfo != nil && dutInfo.GetDut() != nil {
				dut := dutInfo.GetDut()
				chromeOSInfo := dut.GetChromeos()
				if chromeOSInfo != nil {
					tags = AppendTags(tags, "model", chromeOSInfo.GetDutModel().GetModelName())
					tags = AppendTags(tags, "hostname", dut.GetId().GetValue())
				}
			}

			tags = configEnvInfoTags(tags, primaryExecInfo)

			// For Multi-DUT testing info.
			tags = configMultiDUTTags(tags, primaryExecInfo, testInvocation.GetSecondaryExecutionsInfo())
		}
	}

	if testRun != nil {
		// For Testhaus MVP parity.
		// Refer to `_generate_resultdb_base_tags` in test_runner recipe:
		// https://source.chromium.org/chromiumos/chromiumos/codesearch/+/main:infra/recipes/recipes/test_platform/test_runner.py;l=472?q=test_platform%2Ftest_runner.py
		for _, logInfo := range testRun.LogsInfo {
			tags = AppendTags(tags, "logs_url", logInfo.Path)
		}

		testCaseInfo := testRun.GetTestCaseInfo()
		if testCaseInfo != nil {
			tags = AppendTags(tags, "declared_name", testCaseInfo.GetDisplayName())
			tags = AppendTags(tags, "branch", testCaseInfo.GetBranch())
			tags = AppendTags(tags, "main_builder_name", testCaseInfo.GetMainBuilderName())
			tags = AppendTags(tags, "contacts", strings.Join(testCaseInfo.GetContacts(), ","))
			tags = AppendTags(tags, "suite", testCaseInfo.GetSuite())
		}

		timeInfo := testRun.GetTimeInfo()
		if timeInfo != nil {
			if timeInfo.GetQueuedTime().CheckValid() == nil {
				tags = AppendTags(tags, "queued_time", timeInfo.GetQueuedTime().AsTime().UTC().String())
			}
		}
	}

	pbutil.SortStringPairs(tags)
	return tags
}

// configBuildMetaDataTags configs test result tags based on the build metadata.
func configBuildMetaDataTags(tags []*pb.StringPair, buildMetadata *artifactpb.BuildMetadata) []*pb.StringPair {
	if buildMetadata == nil {
		return tags
	}

	firmware := buildMetadata.GetFirmware()
	if firmware != nil {
		tags = AppendTags(tags, "ro_fwid", firmware.GetRoVersion())
		tags = AppendTags(tags, "rw_fwid", firmware.GetRwVersion())
	}

	chipset := buildMetadata.GetChipset()
	if chipset != nil {
		tags = AppendTags(tags, "wifi_chip", chipset.GetWifiChip())
	}

	kernel := buildMetadata.GetKernel()
	if kernel != nil {
		tags = AppendTags(tags, "kernel_version", kernel.GetVersion())
	}

	sku := buildMetadata.GetSku()
	if sku != nil {
		tags = AppendTags(tags, "hwid_sku", sku.GetHwidSku())
	}

	cellular := buildMetadata.GetCellular()
	if cellular != nil {
		tags = AppendTags(tags, "carrier", cellular.GetCarrier())
	}

	lacros := buildMetadata.GetLacros()
	if lacros != nil {
		tags = AppendTags(tags, "ash_version", lacros.GetAshVersion())
		tags = AppendTags(tags, "lacros_version", lacros.GetLacrosVersion())
	}

	return tags
}

// configEnvInfoTags configs test result tags based on the test environment
// information.
func configEnvInfoTags(tags []*pb.StringPair, execInfo *artifactpb.ExecutionInfo) []*pb.StringPair {
	envInfo := execInfo.GetEnvInfo()
	if envInfo == nil {
		return tags
	}

	switch envInfo.(type) {
	case *artifactpb.ExecutionInfo_SkylabInfo:
		skylabInfo := execInfo.GetSkylabInfo()
		if skylabInfo != nil {
			tags = configDroneTags(tags, skylabInfo.GetDroneInfo())
			tags = configSwarmingTags(tags, skylabInfo.GetSwarmingInfo())
			tags = configBuildbucketTags(tags, skylabInfo.GetBuildbucketInfo())
		}
	case *artifactpb.ExecutionInfo_SatlabInfo:
		satlabInfo := execInfo.GetSatlabInfo()
		if satlabInfo != nil {
			tags = configSwarmingTags(tags, satlabInfo.GetSwarmingInfo())
			tags = configBuildbucketTags(tags, satlabInfo.GetBuildbucketInfo())
		}
	}
	return tags
}

// configDroneTags configs test result tags based on the Drone information.
func configDroneTags(tags []*pb.StringPair, droneInfo *artifactpb.DroneInfo) []*pb.StringPair {
	if droneInfo == nil {
		return tags
	}

	tags = AppendTags(tags, "drone", droneInfo.GetDrone())
	tags = AppendTags(tags, "drone_server", droneInfo.GetDroneServer())
	return tags
}

// configSwarmingTags configs test result tags based on the swarming
// information.
func configSwarmingTags(tags []*pb.StringPair, swarmingInfo *artifactpb.SwarmingInfo) []*pb.StringPair {
	if swarmingInfo == nil {
		return tags
	}

	tags = AppendTags(tags, "task_id", swarmingInfo.GetTaskId())
	tags = AppendTags(tags, "suite_task_id", swarmingInfo.GetSuiteTaskId())
	tags = AppendTags(tags, "job_name", swarmingInfo.GetTaskName())
	tags = AppendTags(tags, "pool", swarmingInfo.GetPool())
	tags = AppendTags(tags, "label_pool", swarmingInfo.GetLabelPool())
	return tags
}

// configBuildbucketTags configs test result tags based on the buildbucket
// information.
func configBuildbucketTags(tags []*pb.StringPair, buildbucketInfo *artifactpb.BuildbucketInfo) []*pb.StringPair {
	if buildbucketInfo == nil {
		return tags
	}

	return AppendTags(
		tags,
		"ancestor_buildbucket_ids",
		strings.Trim(strings.Join(strings.Fields(fmt.Sprint(buildbucketInfo.GetAncestorIds())), ","), "[]"))
}

// configMultiDUTTags configs test result tags based on the multi-DUT testing.
func configMultiDUTTags(tags []*pb.StringPair, primaryExecInfo *artifactpb.ExecutionInfo, secondaryExecInfos []*artifactpb.ExecutionInfo) []*pb.StringPair {
	// PrimaryExecInfo must be set for Multi-DUT testing.
	if primaryExecInfo == nil {
		return tags
	}

	if len(secondaryExecInfos) == 0 {
		return AppendTags(tags, "multiduts", "False")
	}

	tags = AppendTags(tags, "multiduts", "True")
	tags = AppendTags(tags, "primary_board", primaryExecInfo.GetBuildInfo().GetBoard())
	tags = AppendTags(tags, "primary_model", primaryExecInfo.GetDutInfo().GetDut().GetChromeos().GetDutModel().GetModelName())

	secordaryDUTSize := len(secondaryExecInfos)
	secondaryBoards := make([]string, 0, secordaryDUTSize)
	secondaryModels := make([]string, 0, secordaryDUTSize)
	for _, execInfo := range secondaryExecInfos {
		buildInfo := execInfo.GetBuildInfo()
		dutInfo := execInfo.GetDutInfo()
		if buildInfo != nil && dutInfo != nil {
			secondaryBoards = append(secondaryBoards, buildInfo.GetBoard())
			secondaryModels = append(secondaryModels, dutInfo.GetDut().GetChromeos().GetDutModel().GetModelName())
		}
	}

	// Concatenates board names and model names separately.
	tags = AppendTags(tags, "secondary_boards", strings.Join(secondaryBoards, " | "))
	tags = AppendTags(tags, "secondary_models", strings.Join(secondaryModels, " | "))

	return tags
}
