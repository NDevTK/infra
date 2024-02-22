// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"testing"

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
			want := ToIndicatorType(tt.input)
			if diff := typed.Got(got).Want(want).Diff(); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}
