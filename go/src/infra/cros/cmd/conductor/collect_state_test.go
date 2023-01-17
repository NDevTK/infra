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
	})
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
	})
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
