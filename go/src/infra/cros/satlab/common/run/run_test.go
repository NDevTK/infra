// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package run

import (
	"context"
	"testing"

	"github.com/googleapis/gax-go/v2"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	moblabpb "google.golang.org/genproto/googleapis/chromeos/moblab/v1beta1"

	"infra/cros/satlab/common/google.golang.org/google/chromeos/moblab"
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
// It tests input entirely, partially, and not at all user given
func TestRun(t *testing.T) {
	t.Parallel()

	type test struct {
		inputCommand *Run
	}

	tests := []test{
		{
			&Run{ // suite run (rlz)
				Suite:     "rlz",
				Board:     "zork",
				Model:     "gumboz",
				Milestone: "111",
				Build:     "15329.6.0",
			},
		},
		{
			&Run{ // test run (local satlab)
				Test:      "rlz_CheckPing.should_send_rlz_ping_missing",
				Harness:   "tauto",
				Board:     "zork",
				Model:     "gumboz",
				Milestone: "111",
				Build:     "15329.6.0",
			},
		},
		{
			&Run{ // test run (remote satlab)
				Test:      "rlz_CheckPing.should_send_rlz_ping_missing",
				Harness:   "tauto",
				Board:     "zork",
				Model:     "gumboz",
				Milestone: "111",
				Build:     "15329.6.0",
				SatlabId:  "satlab-0wgatfqi21118003",
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
		},
	}

	expectedLink := "https://ci.chromium.org/ui/b/0"

	for _, tc := range tests {
		fakeMoblabClient := FakeMoblabClient{}
		fakeBuildbucketClient := FakeBuildbucketClient{badData: false}
		bucket := "chromeos-distributed-fleet-s4p"
		buildLink, err := tc.inputCommand.triggerRunWithClients(context.Background(), &fakeMoblabClient, &fakeBuildbucketClient, bucket)
		if err != nil {
			t.Errorf("Unexpected err: %v", err)
		}
		if buildLink != expectedLink {
			t.Errorf("Unexpected build link, expected: %v, got: %v", expectedLink, buildLink)
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
