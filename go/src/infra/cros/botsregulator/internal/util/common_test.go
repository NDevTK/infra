// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	ufspb "infra/unifiedfleet/api/v1/models"
)

func TestCutHostnames(t *testing.T) {
	t.Parallel()
	t.Run("Happy Path", func(t *testing.T) {
		t.Parallel()
		input := []*ufspb.MachineLSE{
			{Name: "machineLSEs/dut-1"},
			{Name: "machineLSEs/dut-2"},
			{Name: "machineLSEs/dut-3"},
		}
		want := []string{
			"dut-1",
			"dut-2",
			"dut-3",
		}
		got, err := CutHostnames(input)
		if err != nil {
			t.Fatalf("err should be nil: %v\n", err)
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("Wrong key", func(t *testing.T) {
		t.Parallel()
		input := []*ufspb.MachineLSE{
			{Name: "machineLSEs/dut-1"},
			{Name: "somethingWeird/dut-2"},
			{Name: "machineLSEs/dut-3"},
		}
		_, err := CutHostnames(input)
		if err == nil {
			t.Errorf("err should NOT be nil: %v\n", err)
		}
	})
	t.Run("Nil", func(t *testing.T) {
		t.Parallel()
		input := []*ufspb.MachineLSE{
			{Name: "machineLSEs/dut-1"},
			nil,
			{Name: "machineLSEs/dut-3"},
		}
		_, err := CutHostnames(input)
		if err == nil {
			t.Errorf("err should NOT be nil: %v\n", err)
		}
	})
}
