// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package regulator

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	ufspb "infra/unifiedfleet/api/v1/models"
)

func TestConsolidateAvailableDUTs(t *testing.T) {
	t.Parallel()
	r := &regulator{}
	t.Run("Fail", func(t *testing.T) {
		t.Parallel()
		t.Run("CutPrefix cannot parse machineLSE", func(t *testing.T) {
			t.Parallel()
			lses := []*ufspb.MachineLSE{
				{
					Name: "dut-1",
				},
			}
			_, err := r.ConsolidateAvailableDUTs(context.Background(), lses, nil)
			if err == nil {
				t.Errorf("CutPrefix should not be able to parse machineLSE name")
			}
		})
		t.Run("CutPrefix cannot parse schedulingUnit", func(t *testing.T) {
			t.Parallel()
			sus := []*ufspb.SchedulingUnit{
				{
					Name:        "su-1",
					MachineLSEs: []string{"dut-1"},
				},
			}
			lses := []*ufspb.MachineLSE{
				{
					Name: "machineLSEs/dut-1",
				},
			}
			_, err := r.ConsolidateAvailableDUTs(context.Background(), lses, sus)
			if err == nil {
				t.Errorf("CutPrefix should not be able to parse schedulingUnit name")
			}
		})
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()
			sus := []*ufspb.SchedulingUnit{
				{
					Name:        "schedulingunits/su-1",
					MachineLSEs: []string{"dut-1"},
				},
				{
					Name:        "schedulingunits/su-2",
					MachineLSEs: []string{"dut-2", "dut-3"},
				},
				{
					Name:        "schedulingunits/su-3",
					MachineLSEs: []string{"dut-6", "dut-7"},
				},
			}
			lses := []*ufspb.MachineLSE{
				{
					Name: "machineLSEs/dut-1",
				},
				{
					Name: "machineLSEs/dut-2",
				},
				{
					Name: "machineLSEs/dut-3",
				},
				{
					Name: "machineLSEs/dut-4",
				},
				{
					Name: "machineLSEs/dut-5",
				},
			}
			got, err := r.ConsolidateAvailableDUTs(context.Background(), lses, sus)
			if err != nil {
				t.Fatalf("should not error: %v", err)
			}
			want := []string{
				"su-1",
				"su-2",
				"dut-4",
				"dut-5",
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
		t.Run("SchedulingUnit with at least 1 correct lse should be considered", func(t *testing.T) {
			t.Parallel()
			sus := []*ufspb.SchedulingUnit{
				{
					Name:        "schedulingunits/su-1",
					MachineLSEs: []string{"dut-1", "dut-2", "dut-3"},
				},
			}
			lses := []*ufspb.MachineLSE{
				{
					Name: "machineLSEs/dut-1",
				},
			}
			got, err := r.ConsolidateAvailableDUTs(context.Background(), lses, sus)
			if err != nil {
				t.Fatalf("should not error: %v", err)
			}
			want := []string{
				"su-1",
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
		t.Run("No schedulingUnits", func(t *testing.T) {
			t.Parallel()
		})
		lses := []*ufspb.MachineLSE{
			{
				Name: "machineLSEs/dut-1",
			},
		}
		got, err := r.ConsolidateAvailableDUTs(context.Background(), lses, nil)
		if err != nil {
			t.Fatalf("should not error: %v", err)
		}
		want := []string{
			"dut-1",
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})
}
