// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package autotest

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"go.chromium.org/chromiumos/infra/proto/go/chromite/api"
	"go.chromium.org/luci/common/data/stringset"
)

func TestParseTestControlName(t *testing.T) {
	var cases = []struct {
		Text string
		Want string
	}{
		{`NAME = 'platform'`, "platform"},
		{`NAME = "platform"`, "platform"},
		{`NAME = "platform_23Hours.almost-daily"`, "platform_23Hours.almost-daily"},
		{
			Text: `
				AUTHOR = "kalin"
				NAME = "platform_SuspendResumeTiming"
				PURPOSE = "Servo based suspend-resume timing check test"
				CRITERIA = "This test will fail if time to suspend or resume is too long."
				TIME = "LONG"
				TEST_CATEGORY = "Functional"
				TEST_CLASS = "platform"
				TEST_TYPE = "server"
			`,
			Want: "platform_SuspendResumeTiming",
		},
	}

	for _, c := range cases {
		tm, err := parseTestControl(c.Text)
		if err != nil {
			t.Fatalf("parseTestControl: %s", err)
		}
		if c.Want != tm.Name {
			t.Errorf("Name differs, want: %s, got %s", c.Want, tm.Name)
		}
	}
}

func TestParseTestControlSyncCount(t *testing.T) {
	var cases = []struct {
		Text                  string
		WantNeedsMultipleDuts bool
		WantDutCount          int32
	}{
		{``, false, 0},
		{`SYNC_COUNT = 0`, false, 0},
		{`SYNC_COUNT = 3`, true, 3},
	}

	for _, c := range cases {
		tm, err := parseTestControl(c.Text)
		if err != nil {
			t.Fatalf("parseTestControl: %s", err)
		}
		if c.WantNeedsMultipleDuts != tm.NeedsMultipleDuts {
			t.Errorf("NeedsMultipleDuts differes, want: %t, got %t", c.WantNeedsMultipleDuts, tm.NeedsMultipleDuts)
		}
		if c.WantDutCount != tm.DutCount {
			t.Errorf("NeedsMultipleDuts differes, want: %d, got %d", c.WantDutCount, tm.DutCount)
		}
	}
}

func TestParseTestControlRetries(t *testing.T) {
	var cases = []struct {
		Text             string
		WantAllowRetries bool
		WantMaxRetries   int32
	}{
		{`JOB_RETRIES = 3`, true, 3},
		{`JOB_RETRIES = 0`, false, 0},
		// JOB_RETRIES must be explicitly set to 0 to disallow retries. Default is 1 retry.
		{``, true, 1},
	}

	for _, c := range cases {
		tm, err := parseTestControl(c.Text)
		if err != nil {
			t.Fatalf("parseTestControl: %s", err)
		}
		if c.WantAllowRetries != tm.AllowRetries {
			t.Errorf("AllowRetries differes, want: %t, got %t", c.WantAllowRetries, tm.AllowRetries)
		}
		if c.WantMaxRetries != tm.MaxRetries {
			t.Errorf("MaxRetries differes, want: %d, got %d", c.WantMaxRetries, tm.MaxRetries)
		}
	}
}

func TestParseTestControlDependencies(t *testing.T) {
	var cases = []struct {
		Text string
		Want stringset.Set
	}{
		{``, stringset.NewFromSlice()},
		{`DEPENDENCIES = 'dep1'`, stringset.NewFromSlice("dep")},
		{`DEPENDENCIES = "dep1, dep2"`, stringset.NewFromSlice("dep1", "dep2")},
		{`DEPENDENCIES = "dep1,dep2"`, stringset.NewFromSlice("dep1", "dep2")},
		{`DEPENDENCIES = "dep1,dep2,"`, stringset.NewFromSlice("dep1", "dep2")},
	}

	for _, c := range cases {
		tm, err := parseTestControl(c.Text)
		if err != nil {
			t.Fatalf("parseTestControl: %s", err)
		}
		if diff := pretty.Compare(c.Want, dependencySet(tm.Dependencies)); diff != "" {
			t.Errorf("Dependencies differ, -want, +got, %s", diff)
		}
		_ = tm
	}
}

func dependencySet(deps []*api.AutotestTaskDependency) stringset.Set {
	s := stringset.New(len(deps))
	for _, d := range deps {
		s.Add(d.Label)
	}
	return s
}
