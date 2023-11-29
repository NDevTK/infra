// Copyright 2023 The Chromium Authors
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
)

// Pseudo-immutable package variables.
var (
	runID           *suschpb.UID
	startTime       *timestamppb.Timestamp
	endTime         *timestamppb.Timestamp
	configNames     []string
	scheduledSuites []*suschpb.UID
	rejectedSuites  []*suschpb.UID
)

// SetSuiteSchedulerRunID sets the package variable for the runID. Since we cannot set a
// compile-time constant for a uuid, this setter incorporates logic to make it pseudo-immutable.
func SetSuiteSchedulerRunID() error {
	if runID != nil {
		return fmt.Errorf("suite scheduler runId already set to %s", runID.String())
	}

	uuid := uuid.NewString()
	runID = &suschpb.UID{
		Id: uuid,
	}

	return nil
}

// GetRunID returns the package level runID
func GetRunID() *suschpb.UID {
	return runID
}

// SetStartTime sets the package variable for the startTime. Since we cannot set a
// compile-time constant for startTime, this setter incorporates logic to make it pseudo-immutable.
func SetStartTime() error {
	if startTime != nil {
		return fmt.Errorf("suite scheduler startTime already set to %s", startTime.String())
	}

	startTime = timestamppb.Now()

	return nil
}

// GetStartTime returns the package level startTime
func GetStartTime() *timestamppb.Timestamp {
	return startTime
}

// SetEndTime sets the package variable for the endTime. Since we cannot set a
// compile-time constant for endTime, this setter incorporates logic to make it pseudo-immutable.
func SetEndTime() error {
	if endTime != nil {
		return fmt.Errorf("suite scheduler endTime already set to %s", endTime.String())
	}

	endTime = timestamppb.Now()

	return nil
}

// GetEndTime returns the package level endTime
func GetEndTime() *timestamppb.Timestamp {
	return endTime
}

// addConfigNameToList adds the name of the tracked event to the list of all
// acted on suites.
func addConfigNameToList(name string) {
	if configNames == nil {
		configNames = []string{}
	}

	configNames = append(configNames, name)
}

// RegisterScheduledSuite adds a suite decision event to the tracking list.
func RegisterScheduledSuite(event *suschpb.SchedulingEvent) {
	if scheduledSuites == nil {
		scheduledSuites = []*suschpb.UID{}
	}

	scheduledSuites = append(scheduledSuites, event.EventUid)
	addConfigNameToList(event.ConfigName)
}

// RegisterRejectedSuite adds a suite decision event to the tracking list.
func RegisterRejectedSuite(event *suschpb.SchedulingEvent) {
	if rejectedSuites == nil {
		rejectedSuites = []*suschpb.UID{}
	}

	rejectedSuites = append(rejectedSuites, event.EventUid)
	addConfigNameToList(event.ConfigName)
}

// GenerateRunMessage returns a SchedulingMetric for the current SuiteScheduler
// run.
func GenerateRunMessage() *suschpb.SchedulingRun {
	return &suschpb.SchedulingRun{
		RunUid:          runID,
		StartTime:       startTime,
		EndTime:         endTime,
		ConfigNames:     configNames,
		ScheduledSuites: scheduledSuites,
		RejectedSuites:  rejectedSuites,
	}
}
