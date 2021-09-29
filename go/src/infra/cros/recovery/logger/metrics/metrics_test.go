// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package metrics

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// TestNewMetrics verifies that the default metrics logger writes a serialized message to
// the provided logger at the debug level.
//
// TODO(gregorynisbet): Drop this test once the default logger implementation is more substantial.
func TestNewMetrics(t *testing.T) {
	ctx := context.Background()
	expected := []string{
		lines(
			`Create action "b": {`,
			`    "Name": "",`,
			`    "ActionKind": "b",`,
			`    "SwarmingTaskID": "a",`,
			`    "AssetTag": "",`,
			`    "StartTime": "0001-01-01T00:00:00Z",`,
			`    "StopTime": "0001-01-01T00:00:00Z",`,
			`    "Status": "",`,
			`    "FailReason": "",`,
			`    "Observations": [`,
			`        {`,
			`            "MetricKind": "c",`,
			`            "ValueType": "",`,
			`            "Value": ""`,
			`        }`,
			`    ]`,
			`}`,
		),
	}
	l := newFakeLogger().(*fakeLogger)
	m := NewLogMetrics(l)

	m.Create(
		ctx,
		&Action{
			SwarmingTaskID: "a",
			ActionKind:     "b",
			Observations: []*Observation{
				{
					MetricKind: "c",
				},
			},
		},
	)

	actual := l.messages["debug"]

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected diff: %s", diff)
	}
}

// TestCreateNil tests creating a new action using a nil metrics handler.
func TestCreateNil(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	action, closer := Create(ctx, nil, "", "", time.Now())
	if action != nil {
		t.Errorf("expected action to be nil not: %#v", action)
	}
	// Try this and see if it panics.
	closer(ctx)
}

// TestCreate tests creating a new action through the Create convenience function.
func TestCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := newFakeLogger().(*fakeLogger)
	m := NewLogMetrics(l)
	_, closer := Create(ctx, m, "a", "b", time.Now())
	defer closer(ctx)
}

// Join a sequence of lines together to make a string with newlines inserted after
// each element.
func lines(a ...string) string {
	return fmt.Sprintf("%s\n", strings.Join(a, "\n"))
}
