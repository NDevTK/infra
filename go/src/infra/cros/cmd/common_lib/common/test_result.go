// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/google/uuid"
	configpb "go.chromium.org/chromiumos/config/go"
	apipb "go.chromium.org/chromiumos/config/go/test/api"
	artifactpb "go.chromium.org/chromiumos/config/go/test/artifact"
	labpb "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetMockedTestResultProto returns a mock result proto
// that can be used for testing.
func GetMockedTestResultProto() *artifactpb.TestResult {
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
			},
		},
	}

	return testResult
}

func parseTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05.99Z", s)
	return t
}

// GetValueFromRequestKeyvals gets value from provided keyvals based on key.
func GetValueFromRequestKeyvals(ctx context.Context, cftReq *skylab_test_runner.CFTTestRequest, ctrrReq *skylab_test_runner.CrosTestRunnerRequest, key string) string {
	if cftReq == nil && ctrrReq == nil {
		return ""
	}
	keyvals := map[string]string{}
	if ctrrReq != nil {
		keyvals = ctrrReq.GetParams().GetKeyvals()
	} else {
		keyvals = cftReq.GetAutotestKeyvals()
	}

	value, ok := keyvals[key]
	if !ok {
		logging.Infof(ctx, "%s not found in keyvals.", key)
		return ""
	}

	logging.Infof(ctx, "%s found in keyvals with value: %s", key, value)
	return value
}

// GetTesthausUrl gets testhaus log viewer url.
func GetTesthausUrl(gcsUrl string) string {
	return fmt.Sprintf("%s%s", TesthausUrlPrefix, gcsUrl[len("gs://"):])
}

// GetGcsUrl gets gcs url where all the artifacts will be uploaded.
func GetGcsUrl(gsRoot string) string {
	return fmt.Sprintf(
		"%s/%s/%s",
		gsRoot,
		time.Now().Format("2006-01-02"),
		uuid.New().String())
}

// GetGcsClickableLink constructs the gcs cliclable link from provided gs url.
func GetGcsClickableLink(gsUrl string) string {
	if gsUrl == "" {
		return ""
	}
	gsPrefix := "gs://"
	urlSuffix := gsUrl
	if strings.HasPrefix(gsUrl, gsPrefix) {
		urlSuffix = gsUrl[len(gsPrefix):]
	}
	return fmt.Sprintf("%s%s", GcsUrlPrefix, urlSuffix)
}

// IsAnyTestFailure returns if there is any failed tests in test results
func IsAnyTestFailure(testResults []*apipb.TestCaseResult) bool {
	for _, testResult := range testResults {
		switch testResult.Verdict.(type) {
		case *apipb.TestCaseResult_Fail_, *apipb.TestCaseResult_Abort_, *apipb.TestCaseResult_Crash_:
			return true
		default:
			continue
		}
	}

	return false
}
