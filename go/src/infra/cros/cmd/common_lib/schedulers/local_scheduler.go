// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package schedulers

import (
	"context"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
)

// LocalScheduler defines scheduler that doesn't schedule request(s) anywhere.
// It is a dummy scheduler that prints out request(s) without scheduling them.
type LocalScheduler struct {
	*AbstractScheduler

	BBClient *buildbucketpb.BuildsClient
}

func NewLocalScheduler() *LocalScheduler {
	absSched := NewAbstractScheduler(LocalSchedulerType)
	return &LocalScheduler{AbstractScheduler: absSched}
}

func (sc *LocalScheduler) Setup(ctx context.Context) error {
	// no-op
	return nil
}

func (sc *LocalScheduler) ScheduleRequest(ctx context.Context, req *buildbucketpb.ScheduleBuildRequest) (*buildbucketpb.Build, error) {
	// no-op
	return nil, nil
}
