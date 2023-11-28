// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package metrics holds all the schemas and utilities to handle metrics for
// SuSch v1.5.
package metrics

import (
	"fmt"

	"github.com/google/uuid"
	suschpb "go.chromium.org/chromiumos/infra/proto/go/test_platform/suite_scheduler/v15"
	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GenerateEventMessage builds a metric event.
func GenerateEventMessage(config *infrapb.SchedulerConfig, runID *suschpb.UID, schedulingDecision *suschpb.SchedulingDecision, bbids []int64) (*suschpb.SchedulingEvent, error) {
	if runID == nil {
		return nil, fmt.Errorf("runID cannot be nil")
	}

	return &suschpb.SchedulingEvent{
		RunUid: runID,
		EventUid: &suschpb.UID{
			Id: uuid.NewString(),
		},
		ConfigName: config.Name,
		SuiteName:  config.Suite,
		EventTime:  timestamppb.Now(),
		Decision:   schedulingDecision,
		Bbids:      bbids,
	}, nil
}
