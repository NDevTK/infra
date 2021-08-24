// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package logger

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestNewMetrics verifies that the default metrics logger writes a serialized message to
// the provided logger at the debug level.
func TestNewMetrics(t *testing.T) {
	ctx := context.Background()
	expected := []string{
		(`` +
			`{` + "\n" +
			`    "Kind": "b",` + "\n" +
			`    "SwarmingTaskID": "a",` + "\n" +
			`    "AssetTag": "",` + "\n" +
			`    "StartTime": "0001-01-01T00:00:00Z",` + "\n" +
			`    "StopTime": "0001-01-01T00:00:00Z",` + "\n" +
			`    "Status": "",` + "\n" +
			`    "FailReason": "",` + "\n" +
			`    "Observations": [` + "\n" +
			`        {` + "\n" +
			`            "MetricKind": "c",` + "\n" +
			`            "ValueType": "",` + "\n" +
			`            "Value": ""` + "\n" +
			`        }` + "\n" +
			`    ]` + "\n" +
			`}` + "\n"),
	}
	var actual []string
	l := NewMetrics(newFakeLogger(func(level string, format string, args []interface{}) {
		if level != "debug" {
			t.Errorf("expected log level to be \"debug\" not %q", level)
		}
		actual = append(actual, fmt.Sprintf(format, args...))
	}))

	l.Record(
		ctx,
		&Action{
			SwarmingTaskID: "a",
			Kind:           "b",
			Observations: []*Observation{
				{
					MetricKind: "c",
				},
			},
		},
	)

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected diff: %s", diff)
	}
}
