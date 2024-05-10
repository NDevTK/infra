// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package buildbucket

import (
	"context"
	"testing"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
)

func TestGetLatestGreenBuild(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &FakeClient{
		buildBucketClient: FakeBuildClient{},
	}

	wantBuild := &buildbucketpb.Build{
		Id:     1234,
		Number: 1234,
	}

	gotBuild, err := client.GetLatestGreenBuild(ctx)
	if err != nil {
		t.Errorf("unexpected error (%s)", err.Error())
	}
	if wantBuild.GetNumber() != gotBuild.GetNumber() {
		t.Errorf("unexpected build number; wanted: %d; got: %d", wantBuild.GetNumber(), gotBuild.GetNumber())
	}
}

var testTriggerRunArgsData = []struct {
	run                     Run
	wantValidationErrString string
}{
	{ // All errors raised
		Run{
			Image: "volteer-release/R125-15850.0.0",
			Board: "volteer",
			// Milestone:   "R125",
			Build:       "15850.0.0",
			Pool:        "DUT_POOL_QUOTA",
			Tests:       []string{"labqual.SerialNumber"},
			Harness:     "tast",
			CFT:         true,
			TimeoutMins: 1200,
			BBClient: FakeClient{
				buildBucketClient: FakeBuildClient{},
			}.buildBucketClient,
		},
		"",
	},
	{ // Both image and milestone specified
		Run{
			Image:       "volteer-release/R125-15850.0.0",
			Board:       "volteer",
			Milestone:   "R125",
			Build:       "15850.0.0",
			Pool:        "DUT_POOL_QUOTA",
			Tests:       []string{"labqual.SerialNumber"},
			Harness:     "tast",
			CFT:         true,
			TimeoutMins: 1200,
			BBClient: FakeClient{
				buildBucketClient: FakeBuildClient{},
			}.buildBucketClient,
		},
		"cannot specify both image and release branch",
	},
	{ // Missing board
		Run{
			Image:       "volteer-release/R125-15850.0.0",
			Pool:        "DUT_POOL_QUOTA",
			Tests:       []string{"labqual.SerialNumber"},
			Harness:     "tast",
			CFT:         true,
			TimeoutMins: 1200,
			BBClient: FakeClient{
				buildBucketClient: FakeBuildClient{},
			}.buildBucketClient,
		},
		"missing board field",
	},
	{ // Missing pool
		Run{
			Image:       "volteer-release/R125-15850.0.0",
			Board:       "volteer",
			Tests:       []string{"labqual.SerialNumber"},
			Harness:     "tast",
			CFT:         true,
			TimeoutMins: 1200,
			BBClient: FakeClient{
				buildBucketClient: FakeBuildClient{},
			}.buildBucketClient,
		},
		"missing pool field",
	},
	{ // CFT and missing harness
		Run{
			Image:       "volteer-release/R125-15850.0.0",
			Board:       "volteer",
			Build:       "15850.0.0",
			Pool:        "DUT_POOL_QUOTA",
			Tests:       []string{"labqual.SerialNumber"},
			Harness:     "",
			CFT:         true,
			TimeoutMins: 1200,
			BBClient: FakeClient{
				buildBucketClient: FakeBuildClient{},
			}.buildBucketClient,
		},
		"missing harness flag",
	},
	{ // No CFT and harness specified
		Run{
			Image:       "volteer-release/R125-15850.0.0",
			Board:       "volteer",
			Build:       "15850.0.0",
			Pool:        "DUT_POOL_QUOTA",
			Tests:       []string{"labqual.SerialNumber"},
			Harness:     "tast",
			CFT:         false,
			TimeoutMins: 1200,
			BBClient: FakeClient{
				buildBucketClient: FakeBuildClient{},
			}.buildBucketClient,
		},
		"harness should only be provided for single cft test case",
	},
}

func TestTriggerRun(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	for _, tt := range testTriggerRunArgsData {
		_, gotValidationErr := tt.run.TriggerRun(ctx)
		gotValidationErrString := ErrToString(gotValidationErr)
		if tt.wantValidationErrString != gotValidationErrString {
			t.Errorf("unexpected error: wanted '%s', got '%s'", tt.wantValidationErrString, gotValidationErrString)
		}
	}
}

func TestScheduleTest(t *testing.T) {
	ctx := context.Background()
	client := &FakeClient{
		buildBucketClient: FakeBuildClient{},
	}
	link, err := ScheduleBuild(ctx, client)
	if err != nil {
		t.Errorf("unexpected error (%s)", err.Error())
	}
	if link == "" {
		t.Errorf("unexpected error empty build link")
	}
}

func ErrToString(e error) string {
	if e == nil {
		return ""
	}
	return e.Error()
}
