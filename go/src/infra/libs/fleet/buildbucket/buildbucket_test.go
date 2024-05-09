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

func TestTriggerRun(t *testing.T) {
	ctx := context.Background()
	client := &FakeClient{
		buildBucketClient: FakeBuildClient{},
	}
	r := &Run{
		Image:       "volteer-release/R125-15850.0.0",
		Board:       "volteer",
		Milestone:   "R125",
		Build:       "15850.0.0",
		Pool:        "DUT_POOL_QUOTA",
		Tests:       []string{"labqual.SerialNumber"},
		Harness:     "tast",
		CFT:         true,
		TimeoutMins: 1200,
		BBClient:    client.buildBucketClient,
	}
	_, err := r.TriggerRun(ctx)
	if err != nil {
		t.Errorf("unexpected error (%s)", err.Error())
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
