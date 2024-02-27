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

	kronpb "go.chromium.org/chromiumos/infra/proto/go/test_platform/kron"
	suschpb "go.chromium.org/chromiumos/infra/proto/go/testplans"
)

// GenerateEventMessage builds a metric event.
func GenerateEventMessage(config *suschpb.SchedulerConfig, schedulingDecision *kronpb.SchedulingDecision, bbid int64) (*kronpb.Event, error) {
	if runID == "" {
		return nil, fmt.Errorf("runID cannot be empty")
	}

	return &kronpb.Event{
		RunUuid:    runID,
		EventUuid:  uuid.NewString(),
		ConfigName: config.Name,
		SuiteName:  config.Suite,
		EventTime:  timestamppb.Now(),
		Decision:   schedulingDecision,
		Bbid:       bbid,
	}, nil
}
