// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"testing"

	"infra/cros/internal/assert"

	pb "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	bbpb "go.chromium.org/luci/buildbucket/proto"
)

func TestCollectState_MaxRetries(t *testing.T) {
	t.Parallel()

	collectState := initCollectState(&pb.CollectConfig{
		Rules: []*pb.RetryRule{
			{
				MaxRetries: 3,
			},
		},
	}, nil, nil)
	build := &bbpb.Build{
		Id:     12345,
		Status: bbpb.Status_FAILURE,
		Builder: &bbpb.BuilderID{
			Project: "chromeos",
			Bucket:  "release",
			Builder: "eve-release-main",
		},
	}

	retries := 0
	for collectState.canRetry(build) {
		collectState.recordRetry(build)
		retries += 1
	}
	assert.IntsEqual(t, retries, 3)
}

func TestCollectState_MaxRetriesPerBuild(t *testing.T) {
	t.Parallel()

	collectState := initCollectState(&pb.CollectConfig{
		Rules: []*pb.RetryRule{
			{
				MaxRetries:         3,
				MaxRetriesPerBuild: 2,
			},
		},
	}, nil, nil)
	build := &bbpb.Build{
		Id:     12345,
		Status: bbpb.Status_FAILURE,
		Builder: &bbpb.BuilderID{
			Project: "chromeos",
			Bucket:  "release",
			Builder: "eve-release-main",
		},
	}

	retries := 0
	for collectState.canRetry(build) {
		collectState.recordRetry(build)
		retries += 1
	}
	assert.IntsEqual(t, retries, 2)
}

type fakeClock struct {
	currentTime int64
}

func (f *fakeClock) Now() int64 {
	return f.currentTime
}

func TestCollectState_CutoffSeconds(t *testing.T) {
	t.Parallel()

	fakeClock := &fakeClock{
		currentTime: 100,
	}

	collectState := initCollectStateTest(&pb.CollectConfig{
		Rules: []*pb.RetryRule{
			{
				CutoffSeconds: 300, // Can't retry after time 400.
			},
		},
	}, fakeClock)
	build := &bbpb.Build{
		Id:     12345,
		Status: bbpb.Status_FAILURE,
		Builder: &bbpb.BuilderID{
			Project: "chromeos",
			Bucket:  "release",
			Builder: "eve-release-main",
		},
	}

	assert.Assert(t, collectState.canRetry(build))
	fakeClock.currentTime = 200
	assert.Assert(t, collectState.canRetry(build))
	fakeClock.currentTime = 300
	assert.Assert(t, collectState.canRetry(build))
	fakeClock.currentTime = 500
	assert.Assert(t, !collectState.canRetry(build))
}

func TestCollectState_BuildMatches(t *testing.T) {
	t.Parallel()

	collectState := initCollectState(&pb.CollectConfig{
		Rules: []*pb.RetryRule{
			{
				Status: []int32{
					int32(bbpb.Status_FAILURE),
					int32(bbpb.Status_INFRA_FAILURE),
				},
				BuilderNameRe: []string{
					"coral-.*",
					"eve-.*",
				},
				SummaryMarkdownRe: []string{
					".*source cache.*",
					".*gclient.*",
				},
			},
		},
	}, nil, nil)
	failedBuild := &bbpb.Build{
		Id:     12345,
		Status: bbpb.Status_FAILURE,
		Builder: &bbpb.BuilderID{
			Project: "chromeos",
			Bucket:  "release",
			Builder: "eve-release-main",
		},
		SummaryMarkdown: "wah, I have a bad source cache.",
	}
	assert.Assert(t, collectState.canRetry(failedBuild))
	successfulBuild := &bbpb.Build{
		Id:     12345,
		Status: bbpb.Status_SUCCESS,
		Builder: &bbpb.BuilderID{
			Project: "chromeos",
			Bucket:  "release",
			Builder: "coral-release-main",
		},
		SummaryMarkdown: "gclient error",
	}
	assert.Assert(t, !collectState.canRetry(successfulBuild))
	otherBuild := &bbpb.Build{
		Id:     12345,
		Status: bbpb.Status_FAILURE,
		Builder: &bbpb.BuilderID{
			Project: "chromeos",
			Bucket:  "release",
			Builder: "atlas-release-main",
		},
		SummaryMarkdown: "unknown error",
	}
	assert.Assert(t, !collectState.canRetry(otherBuild))
}
