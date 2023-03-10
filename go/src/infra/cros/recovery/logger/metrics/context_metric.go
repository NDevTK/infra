// Copyright 2023 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package metrics

import (
	"context"
)

// defaultMetricActionKeyType is a unique type for a context key.
type defaultMetricActionKeyType string

const (
	deafultMetricActionKey defaultMetricActionKeyType = "default_metric_action"
)

// WithAction sets metrics to the context.
// If Logger is not provided process will be finished with panic.
func WithAction(ctx context.Context, action *Action) context.Context {
	if action != nil {
		return context.WithValue(ctx, deafultMetricActionKey, action)
	}
	return ctx
}

// GetDefaultAction returns default action from context.
func GetDefaultAction(ctx context.Context) *Action {
	if v, ok := ctx.Value(deafultMetricActionKey).(*Action); ok {
		return v
	}
	return nil
}

// DefaultActionAddObservations adds observation to default action in context.
func DefaultActionAddObservations(ctx context.Context, observations ...*Observation) {
	if len(observations) == 0 {
		// Do nothing. Observation is not provided.
		return
	}
	if execMetric := GetDefaultAction(ctx); execMetric != nil {
		execMetric.Observations = append(execMetric.Observations, observations...)
	}
}
