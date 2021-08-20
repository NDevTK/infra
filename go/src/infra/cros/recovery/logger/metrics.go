// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package logger

import (
	"context"
	"fmt"
	"time"
)

// ActionStatus is the status of an action.
type ActionStatus string

const (
	// ActionStatusUnspecified is an unknown status.
	ActionStatusUnspecified ActionStatus = ""
	// ActionStatusSuccess represents a successful action.
	ActionStatusSuccess ActionStatus = "success"
	// ActionStatusFail represents a failed action.
	ActionStatusFail ActionStatus = "fail"
)

// A ValueType is the type of an observation, such as a number or a string.
type ValueType string

const (
	// ValueTypeUnspecified is an unknown value type.
	ValueTypeUnspecified ValueType = ""
	// ValueTypeString represents a string-valued measurement.
	ValueTypeString ValueType = "string"
	// ValueTypeNumber represents a real-valued measurement.
	ValueTypeNumber ValueType = "number"
)

const (
	// TimeAscendings orders fields in ascending order based on time.
	TimeAscending = false
	// TimeDescending orders fields in descending order based on time.
	TimeDescending = true
)

// An action is an event performed on a DUT.
type Action struct {
	// Kind is a coarse-grained type of observation e.g. "ssh".
	Kind string
	// SwarmingTaskID is the ID of the associated swarming task.
	SwarmingTaskID string
	// AssetTag is the asset tag of the DUT that the observation is recorded for.
	AssetTag string
	// StartTime is when the event started.
	StartTime time.Time
	// StopTime is when the event ended.
	StopTime time.Time
	// Status is whether the event was successful, failed, or unknown.
	Status ActionStatus
	// FailReason is an error message with information describing the failure.
	FailReason string
	// Observations are the observations associated with the current observation.
	Observations []*Observation
}

// An observation is a measurement associated with an event performed on a DUT.
type Observation struct {
	// MetricKind is the metric kind (e.g. battery percentage).
	MetricKind string
	// ValueType is the type of value (e.g. String).
	ValueType ValueType
	// Value is the value itself.
	Value string
}

// NewFloat64Observation produces a new float-valued observation of the given kind.
func NewFloat64Observation(kind string, value float64) *Observation {
	return &Observation{
		MetricKind: kind,
		ValueType:  ValueTypeNumber,
		Value:      fmt.Sprintf("%f", value),
	}
}

// NewStringObservation produces a new string-valued observation of the given kind.
func NewStringObservation(kind string, value string) *Observation {
	return &Observation{
		MetricKind: kind,
		ValueType:  ValueTypeString,
		Value:      value,
	}
}

// A Query is a collection of time-bounded search criteria for actions on DUTs.
type Query struct {
	// StartTime is the starting time for the query as a unix timestamp.
	StartTime int64
	// StopTime is the ending time for the query as a unix timestamp.
	StopTime int64
	// AssetTag is the asset tag for the DUT in question.
	AssetTag string
	// Kind limits the kinds of actions considered.
	Kind string
	// Limit imposes a limit on the total number of actions returned.
	Limit int
	// PageToken is an opaque blob of data that is used to start the query at a specific point.
	PageToken string
	// OrderDescending controls how the result set should be ordered by time
	OrderDescending bool
}

// NewLastActoinQuery returns a query for the last record of a given kind for the asset in question.
func NewLastActionQuery(assetTag string, kind string) *Query {
	return &Query{
		AssetTag: assetTag,
		Kind:     kind,
		Limit:    1,
	}
}

// NewLastActionBeforeTimeQuery returns a query for the last record before the stop time of a given kind
// for the asset in question.
func NewLastActionBeforeTimeQuery(assetTag string, kind string, stopTime int64) *Query {
	return &Query{
		AssetTag: assetTag,
		Kind:     kind,
		Limit:    1,
		StopTime: stopTime,
	}
}

// NewListActionsInRangeQuery lists the actions for a given asset and given range in order.
//
// Sample usage:
//
//   q := NewListActionsInRangeQuery(..., "token1", 10)
//   res, err := metrics.Search(ctx, q)
//   if err != nil {
//      ...
//   }
//   q = ListActionsInRangeQuery(..., res.PageToken, 10)
//   res, err = metrics.Search(ctx, q)
//   ...
//
func NewListActionsInRangeQuery(assetTag string, kind string, startTime int64, stopTime int64, pageToken string, limit int) *Query {
	return &Query{
		AssetTag:  assetTag,
		Kind:      kind,
		StartTime: startTime,
		StopTime:  stopTime,
		PageToken: pageToken,
	}
}

// A QueryResult is the result of running a query.
type QueryResult struct {
	// ResultSet is the list of actions satisfying the criteria in question.
	ResultSet []*Action
	// PageToken is the token for resuming the query, if such a token exists.
	PageToken string
}

// Metrics is a simple interface for logging
// structured events and metrics.
type Metrics interface {
	// Record records an action with observations.
	Record(ctx context.Context, action *Action) (*Action, error)

	// Search lists all the actions matching a set of constraints, up to
	// a limit on the number of returned actions.
	Search(ctx context.Context, q *Query) (*QueryResult, error)
}
