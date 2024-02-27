// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package metrics holds all the schemas and utilities to handle metrics for
// SuSch v1.5.
package metrics

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	suschpb "go.chromium.org/chromiumos/infra/proto/go/test_platform/suite_scheduler/v15"
	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"
)

// GenerateEventMessage builds a metric event.
func GenerateEventMessage(config *infrapb.SchedulerConfig, schedulingDecision *suschpb.SchedulingDecision, bbid int64) (*suschpb.Event, error) {
	if runID == "" {
		return nil, fmt.Errorf("runID cannot be empty")
	}

	return &suschpb.Event{
		RunUuid:    runID,
		EventUuid:  uuid.NewString(),
		ConfigName: config.Name,
		SuiteName:  config.Suite,
		EventTime:  timestamppb.Now(),
		Decision:   schedulingDecision,
		Bbid:       bbid,
	}, nil
}
