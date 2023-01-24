package common

import (
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	configpb "go.chromium.org/chromiumos/config/go"
	apipb "go.chromium.org/chromiumos/config/go/test/api"
	artifactpb "go.chromium.org/chromiumos/config/go/test/artifact"
	labpb "go.chromium.org/chromiumos/config/go/test/lab/api"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetMockedTestResultProto returns a mock result proto that can be used for testing.
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
