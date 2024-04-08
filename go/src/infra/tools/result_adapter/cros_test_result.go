// Copyright 2022 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
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
		status, expected := genTestResultStatus(testCaseResult)
		testId := getTestId(testCaseResult)
		if testId == "" {
			return nil, errors.Reason("testId is unspecified due to the missing id in test case result: %v",
				testCaseResult).Err()
		}

		tr := &sinkpb.TestResult{
			TestId:   testId,
			Status:   status,
			Expected: expected,
			// TODO(b/251357069): Move the invocation-level info and
			// result-level info to the new JSON type columns accordingly when
			// the new JSON type columns are ready in place.
			Tags: genTestResultTags(testRun, r.TestResult.GetTestInvocation()),
		}

		if len(testCaseResult.Errors) > 0 &&
			(status == pb.TestStatus_FAIL || status == pb.TestStatus_ABORT || status == pb.TestStatus_CRASH) {
			var rdbErrors []*pb.FailureReason_Error
			var errorsSize int
			for _, e := range testCaseResult.Errors {
				rdbError := &pb.FailureReason_Error{
					Message: truncateString(e.Message, maxErrorMessageBytes),
				}
				errorSize := proto.Size(rdbError)
				if errorsSize+errorSize > maxErrorsBytes {
					// No more errors fit.
					break
				}
				rdbErrors = append(rdbErrors, rdbError)
				errorsSize += errorSize
			}

			tr.FailureReason = &pb.FailureReason{
				PrimaryErrorMessage:  truncateString(testCaseResult.Errors[0].Message, maxErrorMessageBytes),
				Errors:               rdbErrors,
				TruncatedErrorsCount: int32(len(testCaseResult.Errors) - len(rdbErrors)),
			}
		} else if testCaseResult.GetReason() != "" && status != pb.TestStatus_PASS {
			// This path exists to support legacy results until Testhaus UI
			// migration is complete.
			// In future, failure reason should only be set for results considered
			// failed/crashed/aborted by ResultDB; not skipped or passed results.
			// See go/resultdb-failure-reason-integrity-proposal.
			reason := truncateString(
				testCaseResult.GetReason(), maxErrorMessageBytes)
			tr.FailureReason = &pb.FailureReason{
				PrimaryErrorMessage: reason,
				Errors: []*pb.FailureReason_Error{
					{Message: reason},
				},
			}
		}

		if testCaseResult.GetStartTime().CheckValid() == nil {
			tr.StartTime = testCaseResult.GetStartTime()
			if testCaseResult.GetDuration().CheckValid() == nil {
				tr.Duration = testCaseResult.GetDuration()
			}
		}

		if err := PopulateProperties(tr, testRun); err != nil {
			return nil, errors.Annotate(err,
				"failed to unmarshal properties for test result").Err()
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

// genTestResultStatus converts a TestCase Verdict into a ResultSink Status and
// determines the expected field.
func genTestResultStatus(result *apipb.TestCaseResult) (status pb.TestStatus, expected bool) {
	switch result.GetVerdict().(type) {
	case *apipb.TestCaseResult_Pass_:
		return pb.TestStatus_PASS, true
	case *apipb.TestCaseResult_Fail_:
		return pb.TestStatus_FAIL, false
	case *apipb.TestCaseResult_Crash_:
		return pb.TestStatus_CRASH, false
	case *apipb.TestCaseResult_Abort_:
		return pb.TestStatus_ABORT, false
	// Expectedly skipped (TEST_NA in Testhaus).
	case *apipb.TestCaseResult_Skip_:
		return pb.TestStatus_SKIP, true
	// Unexpectedly skipped (NOSTATUS in Testhaus).
	case *apipb.TestCaseResult_NotRun_:
		return pb.TestStatus_SKIP, false
	default:
		return pb.TestStatus_STATUS_UNSPECIFIED, false
	}
}

// PopulateProperties populates the properties of the test result.
func PopulateProperties(testResult *sinkpb.TestResult, testRun *artifactpb.TestRun) error {
	if testRun == nil {
		return nil
	}

	if testResult == nil {
		return errors.Reason("the input test result is nil").Err()
	}

	// Truncates the reason field in advance to reduce the amount of bytes
	// stored in the properties field of test result.
	testCaseInfo := testRun.GetTestCaseInfo()
	testCaseResult := testCaseInfo.GetTestCaseResult()
	if testCaseResult.GetReason() != "" {
		testCaseResult.Reason = truncateString(
			testCaseResult.GetReason(), maxErrorMessageBytes)
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

	// For common tags.
	// TODO(b/316624079): Support the "is_cft_run" field in the contract proto
	tags = AppendTags(tags, "is_cft_run", "True")

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

				tags = AppendTags(tags, "cbx", strconv.FormatBool(dutInfo.GetCbx()))
			}

			tags = configEnvInfoTags(tags, primaryExecInfo)

			// For Multi-DUT testing info.
			tags = configMultiDUTTags(tags, primaryExecInfo, testInvocation.GetSecondaryExecutionsInfo())
		}

		schedulingMetadata := testInvocation.GetSchedulingMetadata()
		if schedulingMetadata != nil {
			schedulingArgs := schedulingMetadata.GetSchedulingArgs()
			requestedArgs := map[string]bool{
				"analytics_name":    true,
				"ctp-fwd-task-name": true,
				"qs_account":        true,
			}
			for k, v := range schedulingArgs {
				if _, ok := requestedArgs[k]; ok {
					// Converts tag key properly since some might contain
					// hyphens.
					tags = AppendTags(tags, convertTagKey(k), v)
				}
			}
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

			tags = configTestMetadataTags(tags, testCaseInfo.GetTestCaseResult().GetTestCaseMetadata())
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

	newTags := make([]*pb.StringPair, 0, len(tags))
	newTags = append(newTags, tags...)

	firmware := buildMetadata.GetFirmware()
	if firmware != nil {
		newTags = AppendTags(newTags, "ro_fwid", firmware.GetRoVersion())
		newTags = AppendTags(newTags, "rw_fwid", firmware.GetRwVersion())
	}

	chipset := buildMetadata.GetChipset()
	if chipset != nil {
		newTags = AppendTags(newTags, "wifi_chip", chipset.GetWifiChip())
		newTags = AppendTags(newTags, "wifi_router_models", chipset.GetWifiRouterModels())
	}

	kernel := buildMetadata.GetKernel()
	if kernel != nil {
		newTags = AppendTags(newTags, "kernel_version", kernel.GetVersion())
	}

	sku := buildMetadata.GetSku()
	if sku != nil {
		newTags = AppendTags(newTags, "hwid_sku", sku.GetHwidSku())
	}

	cellular := buildMetadata.GetCellular()
	if cellular != nil {
		newTags = AppendTags(newTags, "carrier", cellular.GetCarrier())
	}

	lacros := buildMetadata.GetLacros()
	if lacros != nil {
		newTags = AppendTags(newTags, "ash_version", lacros.GetAshVersion())
		newTags = AppendTags(newTags, "lacros_version", lacros.GetLacrosVersion())
	}

	chameleon := buildMetadata.GetChameleon()
	if chameleon != nil {
		chameleonTypes := []string{}
		for _, t := range chameleon.GetTypes() {
			chameleonTypes = append(chameleonTypes, t.Enum().String())
		}
		if len(chameleonTypes) != 0 {
			newTags = AppendTags(newTags, "chameleon_type", strings.Join(chameleonTypes, ","))
		}
	}

	modemInfo := buildMetadata.GetModemInfo()
	if modemInfo != nil {
		newTags = AppendTags(newTags, "modem_type", modemInfo.GetType().Enum().String())
	}

	return newTags
}

// configEnvInfoTags configs test result tags based on the test environment
// information.
func configEnvInfoTags(tags []*pb.StringPair, execInfo *artifactpb.ExecutionInfo) []*pb.StringPair {
	envInfo := execInfo.GetEnvInfo()
	if envInfo == nil {
		return tags
	}

	newTags := make([]*pb.StringPair, 0, len(tags))
	newTags = append(newTags, tags...)

	switch envInfo.(type) {
	case *artifactpb.ExecutionInfo_SkylabInfo:
		skylabInfo := execInfo.GetSkylabInfo()
		if skylabInfo != nil {
			newTags = configDroneTags(newTags, skylabInfo.GetDroneInfo())
			newTags = configSwarmingTags(newTags, skylabInfo.GetSwarmingInfo())
			newTags = configBuildbucketTags(newTags, skylabInfo.GetBuildbucketInfo())
		}
	case *artifactpb.ExecutionInfo_SatlabInfo:
		satlabInfo := execInfo.GetSatlabInfo()
		if satlabInfo != nil {
			newTags = configSwarmingTags(newTags, satlabInfo.GetSwarmingInfo())
			newTags = configBuildbucketTags(newTags, satlabInfo.GetBuildbucketInfo())
		}
	}
	return newTags
}

// configDroneTags configs test result tags based on the Drone information.
func configDroneTags(tags []*pb.StringPair, droneInfo *artifactpb.DroneInfo) []*pb.StringPair {
	if droneInfo == nil {
		return tags
	}

	newTags := make([]*pb.StringPair, 0, len(tags))
	newTags = append(newTags, tags...)

	newTags = AppendTags(newTags, "drone", droneInfo.GetDrone())
	newTags = AppendTags(newTags, "drone_server", droneInfo.GetDroneServer())
	return newTags
}

// configSwarmingTags configs test result tags based on the swarming
// information.
func configSwarmingTags(tags []*pb.StringPair, swarmingInfo *artifactpb.SwarmingInfo) []*pb.StringPair {
	if swarmingInfo == nil {
		return tags
	}

	newTags := make([]*pb.StringPair, 0, len(tags))
	newTags = append(newTags, tags...)

	newTags = AppendTags(newTags, "task_id", swarmingInfo.GetTaskId())
	newTags = AppendTags(newTags, "suite_task_id", swarmingInfo.GetSuiteTaskId())
	newTags = AppendTags(newTags, "job_name", swarmingInfo.GetTaskName())
	newTags = AppendTags(newTags, "pool", swarmingInfo.GetPool())
	newTags = AppendTags(newTags, "label_pool", swarmingInfo.GetLabelPool())
	return newTags
}

// configBuildbucketTags configs test result tags based on the buildbucket
// information.
func configBuildbucketTags(tags []*pb.StringPair, buildbucketInfo *artifactpb.BuildbucketInfo) []*pb.StringPair {
	if buildbucketInfo == nil {
		return tags
	}

	newTags := make([]*pb.StringPair, 0, len(tags))
	newTags = append(newTags, tags...)

	buildbucketBuilder := buildbucketInfo.GetBuilder()
	if buildbucketBuilder != nil {
		newTags = AppendTags(
			newTags, "buildbucket_builder", buildbucketBuilder.GetBuilder())
	}

	return AppendTags(
		newTags,
		"ancestor_buildbucket_ids",
		strings.Trim(
			strings.Join(strings.Fields(
				fmt.Sprint(buildbucketInfo.GetAncestorIds())), ","),
			"[]"))
}

// configMultiDUTTags configs test result tags based on the multi-DUT testing.
func configMultiDUTTags(tags []*pb.StringPair, primaryExecInfo *artifactpb.ExecutionInfo, secondaryExecInfos []*artifactpb.ExecutionInfo) []*pb.StringPair {
	// PrimaryExecInfo must be set for Multi-DUT testing.
	if primaryExecInfo == nil {
		return tags
	}

	newTags := make([]*pb.StringPair, 0, len(tags))
	newTags = append(newTags, tags...)

	if len(secondaryExecInfos) == 0 {
		return AppendTags(newTags, "multiduts", "False")
	}

	newTags = AppendTags(newTags, "multiduts", "True")
	newTags = AppendTags(newTags, "primary_board", primaryExecInfo.GetBuildInfo().GetBoard())
	newTags = AppendTags(newTags, "primary_model", primaryExecInfo.GetDutInfo().GetDut().GetChromeos().GetDutModel().GetModelName())

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
	newTags = AppendTags(newTags, "secondary_boards", strings.Join(secondaryBoards, " | "))
	newTags = AppendTags(newTags, "secondary_models", strings.Join(secondaryModels, " | "))
	return newTags
}

// configTestMetadataTags configs test result tags based on the test case metadata.
func configTestMetadataTags(tags []*pb.StringPair, testMetadata *apipb.TestCaseMetadata) []*pb.StringPair {
	metadataTags := metadataToTags(testMetadata)
	if len(metadataTags) == 0 {
		return tags
	}

	newTags := make([]*pb.StringPair, 0, len(tags)+len(metadataTags))
	newTags = append(newTags, tags...)
	newTags = append(newTags, metadataTags...)
	return newTags
}
