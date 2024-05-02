// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package schedulers

import (
	"context"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/luciexe/build"
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

func (sc *LocalScheduler) Setup(_ string) error {
	// no-op
	return nil
}

func (sc *LocalScheduler) ScheduleRequest(_ context.Context, _ *buildbucketpb.ScheduleBuildRequest, _ *build.Step) (*buildbucketpb.Build, string, error) {
	// no-op
	return nil, "", nil
}
