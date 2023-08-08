// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"google.golang.org/protobuf/testing/protocmp"

	"infra/libs/skylab/common/heuristics"
)

// TestValidatePattern tests that validating accepts the correct strings.
func TestValidatePattern(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		pattern string
		ok      bool
	}{
		{
			"empty string",
			"",
			false,
		},
		{
			"just anchor #1",
			"^",
			false,
		},
		{
			"just anchor #2",
			"$",
			false,
		},
		{
			"good string",
			"^a",
			true,
		},
		{
			"good string #2",
			"a$",
			true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := validatePattern(tt.pattern)
			ok := e == nil
			if diff := cmp.Diff(tt.ok, ok); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}

// TestMatches tests that matches returns whether a string
// matches a pattern.
func TestMatches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		pattern  string
		hostname string
		out      bool
		ok       bool
	}{
		{
			name:     "empty string",
			pattern:  "",
			hostname: "a",
			out:      false,
			ok:       false,
		},
		{
			name:     "trivial match",
			pattern:  "^a",
			hostname: "a",
			out:      true,
			ok:       true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, e := matches(tt.pattern, tt.hostname)
			ok := e == nil
			if diff := cmp.Diff(tt.out, out); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
			if diff := cmp.Diff(tt.ok, ok); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}

// TestComputePermilleData tests ComputePermilleData and getLastMatch.
func TestComputePermilleData(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		config   *RolloutConfig
		hostname string
		out      *PermilleData
		ok       bool
	}{
		{
			name:     "empty config",
			config:   nil,
			hostname: "a",
			out:      &PermilleData{},
			ok:       true,
		},
		{
			name: "fallback",
			config: &RolloutConfig{
				Enable: true,
				Pattern: []*RolloutConfig_Pattern{
					{
						Pattern:        "^a",
						ProdPermille:   100,
						LatestPermille: 200,
					},
				},
			},
			hostname: "a",
			out: &PermilleData{
				Source: "^a",
				Prod:   100,
				Latest: 200,
			},
			ok: true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, e := tt.config.getLastMatch(tt.hostname)
			ok := e == nil
			if diff := cmp.Diff(tt.out, out, protocmp.Transform()); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
			if diff := cmp.Diff(tt.ok, ok); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}

func TestChooseImplementation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("test choose implementation", t, func() {
		Convey("nil permille data", func() {
			var pd *PermilleData
			impl, err := pd.ChooseImplementation(ctx, 0.0)
			So(impl, ShouldEqual, heuristics.ProdTaskType)
			So(err, ShouldBeNil)
		})
		Convey("zeroes", func() {
			pd := &PermilleData{Prod: 0.0, Latest: 0.0}
			impl, err := pd.ChooseImplementation(ctx, 0.0)
			So(impl, ShouldEqual, heuristics.ProdTaskType)
			So(err, ShouldErrLike, "not yet implemented")
		})
	})
}

func TestValidateNoRepairOnlyFields(t *testing.T) {
	t.Parallel()
	Convey("optin_all_duts", t, func() {
		pd := &RolloutConfig{Enable: true, OptinAllDuts: true}
		So(pd.ValidateNoRepairOnlyFields(), ShouldErrLike, "optin_all_duts")
	})
	Convey("optin_all_duts", t, func() {
		pd := &RolloutConfig{Enable: true, OptinDutPool: []string{"aaa"}}
		So(pd.ValidateNoRepairOnlyFields(), ShouldErrLike, "optin_dut_pool")
	})
}
