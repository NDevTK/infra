// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"testing"

	"google.golang.org/genproto/googleapis/type/money"

	"go.chromium.org/luci/common/testing/typed"

	fleetcostpb "infra/cros/fleetcost/api"
)

// TestToIndicatorType checks the output of the indicator type.
func TestToIndicatorType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		input  string
		output fleetcostpb.IndicatorType
	}{
		{
			name:   "empty",
			input:  "",
			output: fleetcostpb.IndicatorType_INDICATOR_TYPE_UNKNOWN,
		},
		{
			name:   "dut",
			input:  "dut",
			output: fleetcostpb.IndicatorType_INDICATOR_TYPE_DUT,
		},
		{
			name:   "dut uppercase",
			input:  "DUT",
			output: fleetcostpb.IndicatorType_INDICATOR_TYPE_DUT,
		},
		{
			name:   "dut mixed case",
			input:  "DuT",
			output: fleetcostpb.IndicatorType_INDICATOR_TYPE_DUT,
		},
		{
			name:   "dut with prefix",
			input:  "INDICATOR_TYPE_DUT",
			output: fleetcostpb.IndicatorType_INDICATOR_TYPE_DUT,
		},
		{
			name:   "dut with wrong prefix is unknown",
			input:  "TYPE_DUT",
			output: fleetcostpb.IndicatorType_INDICATOR_TYPE_UNKNOWN,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.output
			want, _ := ToIndicatorType(tt.input)
			if diff := typed.Got(got).Want(want).Diff(); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}

// TestToUSD checks converting a command line value to USD.
func TestToUSD(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		input  string
		output *money.Money
	}{
		{
			name:   "empty",
			input:  "",
			output: nil,
		},
		{
			name:  "one dollar",
			input: "1",
			output: &money.Money{
				CurrencyCode: "USD",
				Units:        1,
			},
		},
		{
			name:  "$1.2 as decimal",
			input: "1.2",
			output: &money.Money{
				CurrencyCode: "USD",
				Units:        1,
				Nanos:        20 * (10 * 1000 * 1000),
			},
		},
		{
			name:  "$1.20 as decimal",
			input: "1.20",
			output: &money.Money{
				CurrencyCode: "USD",
				Units:        1,
				Nanos:        20 * (10 * 1000 * 1000),
			},
		},
		{
			name:  "$1.02 as decimal",
			input: "1.02",
			output: &money.Money{
				CurrencyCode: "USD",
				Units:        1,
				Nanos:        2 * (10 * 1000 * 1000),
			},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.output
			want, _ := ToUSD(tt.input)
			if diff := typed.Got(got).Want(want).Diff(); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}

// TestToCostCadence checks the output of the indicator type.
func TestToCostCadence(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		input  string
		output fleetcostpb.CostCadence
	}{
		{
			name:   "empty",
			input:  "",
			output: fleetcostpb.CostCadence_COST_CADENCE_UNKNOWN,
		},
		{
			name:   "daily",
			input:  "daily",
			output: fleetcostpb.CostCadence_COST_CADENCE_DAILY,
		},
		{
			name:   "daily uppercase",
			input:  "daily",
			output: fleetcostpb.CostCadence_COST_CADENCE_DAILY,
		},
		{
			name:   "daily mixed case",
			input:  "DaiLy",
			output: fleetcostpb.CostCadence_COST_CADENCE_DAILY,
		},
		{
			name:   "dut with prefix",
			input:  "COST_CADENCE_DAILY",
			output: fleetcostpb.CostCadence_COST_CADENCE_DAILY,
		},
		{
			name:   "daily with wrong prefix is unknown",
			input:  "TYPE_DAILY",
			output: fleetcostpb.CostCadence_COST_CADENCE_UNKNOWN,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			want := tt.output
			got, _ := ToCostCadence(tt.input)
			if diff := typed.Got(got).Want(want).Diff(); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}

// TestToLocation checks converting a string to a location.
func TestToLocation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		input  string
		output fleetcostpb.Location
	}{
		{
			name:   "empty",
			input:  "",
			output: fleetcostpb.Location_LOCATION_UNKNOWN,
		},
		{
			name:   "all",
			input:  "all",
			output: fleetcostpb.Location_LOCATION_ALL,
		},
		{
			name:   "all uppercase",
			input:  "ALL",
			output: fleetcostpb.Location_LOCATION_ALL,
		},
		{
			name:   "all mixed case",
			input:  "AlL",
			output: fleetcostpb.Location_LOCATION_ALL,
		},
		{
			name:   "all with prefix",
			input:  "LOCATION_ALL",
			output: fleetcostpb.Location_LOCATION_ALL,
		},
		{
			name:   "all with wrong prefix is unknown",
			input:  "TYPE_ALL",
			output: fleetcostpb.Location_LOCATION_UNKNOWN,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			want := tt.output
			got, _ := ToLocation(tt.input)
			if diff := typed.Got(got).Want(want).Diff(); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}
