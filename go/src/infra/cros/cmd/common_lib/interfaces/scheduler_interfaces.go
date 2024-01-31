// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package interfaces

import (
	"context"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
)

// SchedulerType represents scheduler type
type SchedulerType string

// SchedulerInterface defines the contract a scheduler will have to satisfy.
type SchedulerInterface interface {
	// GetSchedulerType returns the scheduler type
	GetSchedulerType() SchedulerType

	// Setup sets up the scheduler
	Setup(context.Context) error

	// ScheduleRequest schedules request
	ScheduleRequest(context.Context, *buildbucketpb.ScheduleBuildRequest) (*buildbucketpb.Build, error)
}
