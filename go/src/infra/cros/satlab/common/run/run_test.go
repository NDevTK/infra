// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package run

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/googleapis/gax-go/v2"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/platform/dev-util/src/chromiumos/ctp/builder"
	"go.chromium.org/luci/auth"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	moblabpb "google.golang.org/genproto/googleapis/chromeos/moblab/v1beta1"

	"infra/cros/satlab/common/google.golang.org/google/chromeos/moblab"
	"infra/cros/satlab/common/site"
)

// FakeMoblabClient is a mock Moblab API client that returns hardcoded data
type FakeMoblabClient struct{}

func (f *FakeMoblabClient) StageBuild(ctx context.Context, req *moblabpb.StageBuildRequest, opts ...gax.CallOption) (*moblab.StageBuildOperation, error) {

	return &moblab.StageBuildOperation{}, nil
}

func (f *FakeMoblabClient) CheckBuildStageStatus(context.Context, *moblabpb.CheckBuildStageStatusRequest, ...gax.CallOption) (*moblabpb.CheckBuildStageStatusResponse, error) {

	name := "buildTargets/octopus/models/bobba/builds/1234.0.0/artifacts/chromeos-moblab-peng-staging"
	return &moblabpb.CheckBuildStageStatusResponse{
		IsBuildStaged:       true,
		StagedBuildArtifact: &moblabpb.BuildArtifact{Name: name},
	}, nil
}

// FakeBuildbucketClient is a mock Buildbucket client that returns hardcoded data
type FakeBuildbucketClient struct {
	badData bool
}

func (f *FakeBuildbucketClient) ScheduleCTPBuild(context.Context) (*buildbucketpb.Build, error) {
	if f.badData {
		return &buildbucketpb.Build{}, nil
	}

	return &buildbucketpb.Build{
		Id: 0000,
		Builder: &buildbucketpb.BuilderID{
			Project: "project",
			Bucket:  "bucket",
			Builder: "builder",
		},
	}, nil
}

// TestRun tests the innerRun function of our command with fake Moblab and Buildbucket clients
func TestRun(t *testing.T) {
	t.Parallel()

	expectedLink := "https://ci.chromium.org/ui/b/0"

	fakeMoblabClient := FakeMoblabClient{}
	fakeBuildbucketClient := FakeBuildbucketClient{badData: false}
	bucket := "chromeos-distributed-fleet-s4p"
	buildLink, err := (&Run{}).triggerRunWithClients(context.Background(), &fakeMoblabClient, &fakeBuildbucketClient, bucket)
	if err != nil {
		t.Errorf("Unexpected err: %v", err)
	}
	if buildLink != expectedLink {
		t.Errorf("Unexpected build link, expected: %v, got: %v", expectedLink, buildLink)
	}
}

var ignoreOpts = cmpopts.IgnoreUnexported(auth.Options{}, buildbucketpb.BuilderID{}, test_platform.Request_TestPlan{}, test_platform.Request_Suite{}, test_platform.Request_Test{}, test_platform.Request_Test_Autotest{})

// TestCreateCTPBuilder tests the CTPBuilder struct contains the correct fields
// given an input command.
//
// Downstream testing of the actual build is the responsibility of the ctp
// library.
func TestCreateCTPBuilder(t *testing.T) {
	type test struct {
		inputCommand     *Run
		expectedBBClient *builder.CTPBuilder
	}
	ctx := context.Background()
	opt := site.GetAuthOption(ctx)

	tests := []*test{
		{
			&Run{ // suite run (rlz)
				Suite:     "rlz",
				Board:     "zork",
				Model:     "gumboz",
				Milestone: "111",
				Build:     "15329.6.0",
			},
			&builder.CTPBuilder{
				AuthOptions: &opt,
				Board:       "zork",
				BuilderID: &buildbucketpb.BuilderID{
					Project: "chromeos",
					Bucket:  "cros_test_platform",
					Builder: "cros_test_platform",
				},
				Dimensions:          map[string]string{},
				Image:               "zork-release/R111-15329.6.0",
				ImageBucket:         "chromeos-image-archive",
				Model:               "gumboz",
				TestPlan:            builder.TestPlanForSuites([]string{"rlz"}),
				TestRunnerBuildTags: map[string]string{},
				TimeoutMins:         360,
			},
		},
		{
			&Run{ // test run (local satlab)
				Tests:     []string{"rlz_CheckPing.should_send_rlz_ping_missing"},
				Harness:   "tauto",
				Board:     "zork",
				Model:     "gumboz",
				Milestone: "111",
				Build:     "15329.6.0",
			},
			&builder.CTPBuilder{
				AuthOptions: &opt,
				Board:       "zork",
				BuilderID: &buildbucketpb.BuilderID{
					Project: "chromeos",
					Bucket:  "cros_test_platform",
					Builder: "cros_test_platform",
				},
				Dimensions:          map[string]string{},
				Image:               "zork-release/R111-15329.6.0",
				ImageBucket:         "chromeos-image-archive",
				Model:               "gumboz",
				TestPlan:            builder.TestPlanForTests("", "tauto", []string{"rlz_CheckPing.should_send_rlz_ping_missing"}),
				TestRunnerBuildTags: map[string]string{},
				TimeoutMins:         360,
			},
		},
		{
			&Run{ // test run (remote satlab)
				Tests:     []string{"rlz_CheckPing.should_send_rlz_ping_missing"},
				Harness:   "tauto",
				Board:     "zork",
				Model:     "gumboz",
				Milestone: "111",
				Build:     "15329.6.0",
				SatlabId:  "satlab-0wgatfqi21118003",
			},
			&builder.CTPBuilder{
				AuthOptions: &opt,
				Board:       "zork",
				BuilderID: &buildbucketpb.BuilderID{
					Project: "chromeos",
					Bucket:  "cros_test_platform",
					Builder: "cros_test_platform",
				},
				Dimensions:          map[string]string{"drone": "satlab-0wgatfqi21118003"},
				Image:               "zork-release/R111-15329.6.0",
				ImageBucket:         "chromeos-image-archive",
				Model:               "gumboz",
				TestPlan:            builder.TestPlanForTests("", "tauto", []string{"rlz_CheckPing.should_send_rlz_ping_missing"}),
				TestRunnerBuildTags: map[string]string{},
				TimeoutMins:         360,
			},
		},
		{
			&Run{ // suite run with dims
				Suite:     "rlz",
				Board:     "zork",
				Model:     "gumboz",
				Milestone: "111",
				Build:     "15329.6.0",
				AddedDims: map[string]string{"label-dut": "123"},
			},
			&builder.CTPBuilder{
				AuthOptions: &opt,
				Board:       "zork",
				BuilderID: &buildbucketpb.BuilderID{
					Project: "chromeos",
					Bucket:  "cros_test_platform",
					Builder: "cros_test_platform",
				},
				Dimensions:          map[string]string{"label-dut": "123"},
				Image:               "zork-release/R111-15329.6.0",
				ImageBucket:         "chromeos-image-archive",
				Model:               "gumboz",
				TestPlan:            builder.TestPlanForSuites([]string{"rlz"}),
				TestRunnerBuildTags: map[string]string{},
				TimeoutMins:         360,
			},
		},
		{
			&Run{ // suite run with max timeout
				Suite:      "rlz",
				Board:      "zork",
				Model:      "gumboz",
				Milestone:  "111",
				Build:      "15329.6.0",
				AddedDims:  map[string]string{"label-dut": "123"},
				MaxTimeout: true,
			},
			&builder.CTPBuilder{
				AuthOptions: &opt,
				Board:       "zork",
				BuilderID: &buildbucketpb.BuilderID{
					Project: "chromeos",
					Bucket:  "cros_test_platform",
					Builder: "cros_test_platform",
				},
				Dimensions:          map[string]string{"label-dut": "123"},
				Image:               "zork-release/R111-15329.6.0",
				ImageBucket:         "chromeos-image-archive",
				Model:               "gumboz",
				TestPlan:            builder.TestPlanForSuites([]string{"rlz"}),
				TestRunnerBuildTags: map[string]string{},
				TimeoutMins:         2370,
			},
		},
		{
			&Run{ // testplan run with dims
				TestplanLocal: "testplan.json",
				Board:         "zork",
				Model:         "gumboz",
				Milestone:     "111",
				Build:         "15329.6.0",
				AddedDims:     map[string]string{"label-dut": "123"},
			},
			&builder.CTPBuilder{
				AuthOptions: &opt,
				Board:       "zork",
				BuilderID: &buildbucketpb.BuilderID{
					Project: "chromeos",
					Bucket:  "cros_test_platform",
					Builder: "cros_test_platform",
				},
				Dimensions:  map[string]string{"label-dut": "123"},
				Image:       "zork-release/R111-15329.6.0",
				ImageBucket: "chromeos-image-archive",
				Model:       "gumboz",
				TestPlan: &test_platform.Request_TestPlan{
					Test: []*test_platform.Request_Test{
						{Harness: &test_platform.Request_Test_Autotest_{Autotest: &test_platform.Request_Test_Autotest{Name: "audio_CrasGetNodes"}}},
						{Harness: &test_platform.Request_Test_Autotest_{Autotest: &test_platform.Request_Test_Autotest{Name: "audio_CrasStress.input_only"}}},
						{Harness: &test_platform.Request_Test_Autotest_{Autotest: &test_platform.Request_Test_Autotest{Name: "audio_CrasStress.output_only"}}},
					},
				},
				TestRunnerBuildTags: map[string]string{"test-plan-id": "testplan"},
				TimeoutMins:         360,
			},
		},
		{
			&Run{ // testplan run directory in test plan
				TestplanLocal: "test/testplan.json",
				Board:         "zork",
				Model:         "gumboz",
				Milestone:     "111",
				Build:         "15329.6.0",
			},
			&builder.CTPBuilder{
				AuthOptions: &opt,
				Board:       "zork",
				BuilderID: &buildbucketpb.BuilderID{
					Project: "chromeos",
					Bucket:  "cros_test_platform",
					Builder: "cros_test_platform",
				},
				Dimensions:  map[string]string{},
				Image:       "zork-release/R111-15329.6.0",
				ImageBucket: "chromeos-image-archive",
				Model:       "gumboz",
				TestPlan: &test_platform.Request_TestPlan{
					Test: []*test_platform.Request_Test{
						{Harness: &test_platform.Request_Test_Autotest_{Autotest: &test_platform.Request_Test_Autotest{Name: "audio_CrasGetNodes"}}},
						{Harness: &test_platform.Request_Test_Autotest_{Autotest: &test_platform.Request_Test_Autotest{Name: "audio_CrasStress.input_only"}}},
						{Harness: &test_platform.Request_Test_Autotest_{Autotest: &test_platform.Request_Test_Autotest{Name: "audio_CrasStress.output_only"}}},
					},
				},
				TestRunnerBuildTags: map[string]string{"test-plan-id": "testplan"},
				TimeoutMins:         360,
			},
		},
	}

	for _, tc := range tests {
		ctx := context.Background()
		bbClient, err := tc.inputCommand.createCTPBuilder(ctx)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}

		if diff := cmp.Diff(tc.expectedBBClient, bbClient, ignoreOpts); diff != "" {
			t.Errorf("Unexpected diff in CTPBuilder: %s", diff)
		}
	}
}

func TestReadTestPlan(t *testing.T) {
	t.Parallel()

	path := "testplan.json"
	res, err := readTestPlan(path)
	if err != nil {
		t.Errorf("Unexpected err: %v", err)
	}

	expected := &test_platform.Request_TestPlan{
		Test: []*test_platform.Request_Test{
			{
				Harness: &test_platform.Request_Test_Autotest_{
					Autotest: &test_platform.Request_Test_Autotest{
						Name: "audio_CrasGetNodes",
					},
				},
			},
			{
				Harness: &test_platform.Request_Test_Autotest_{
					Autotest: &test_platform.Request_Test_Autotest{
						Name: "audio_CrasStress.input_only",
					},
				},
			},
			{
				Harness: &test_platform.Request_Test_Autotest_{
					Autotest: &test_platform.Request_Test_Autotest{
						Name: "audio_CrasStress.output_only",
					},
				},
			},
		},
	}
	if expected.String() != res.String() {
		t.Error("readTestPlan Error")
	}

	_, err = readTestPlan("testplan1.json")
	if err == nil {
		t.Errorf("Unexpected err: %v", err)
	}
}

func TestReadTestPlanFail(t *testing.T) {
	t.Parallel()

	_, err := readTestPlan("testplan1.json")
	if err == nil {
		t.Errorf("Unexpected err: %v", err)
	}
}
