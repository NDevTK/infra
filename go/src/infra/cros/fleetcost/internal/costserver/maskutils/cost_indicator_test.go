// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package maskutils

import (
	"testing"

	"go.chromium.org/luci/common/testing/typed"

	fleetcostpb "infra/cros/fleetcost/api"
)

func TestUpdateCostIndicatorProto(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		src       *fleetcostpb.CostIndicator
		dst       *fleetcostpb.CostIndicator
		fieldmask []string
		output    *fleetcostpb.CostIndicator
	}{
		{
			name:      "empty",
			src:       nil,
			dst:       nil,
			fieldmask: nil,
			output:    nil,
		},
		{
			name:      "empty with nontrivial fieldmask",
			src:       nil,
			dst:       nil,
			fieldmask: []string{"name"},
			output:    nil,
		},
		{
			name: "name only",
			src: &fleetcostpb.CostIndicator{
				Name: "platypus",
			},
			dst: &fleetcostpb.CostIndicator{
				Name: "warthog",
			},
			fieldmask: []string{"name"},
			output: &fleetcostpb.CostIndicator{
				Name: "platypus",
			},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			UpdateCostIndicatorProto(tt.dst, tt.src, tt.fieldmask)
			got := tt.dst
			want := tt.output
			if diff := typed.Got(got).Want(want).Diff(); diff != "" {
				t.Errorf("unexpected diff: %s", diff)
			}
		})
	}
}
