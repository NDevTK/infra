// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package migrator

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"infra/cros/botsregulator/protos"
)

func TestNewConfigSearchable(t *testing.T) {
	t.Parallel()
	t.Run("Happy path", func(t *testing.T) {
		cfg := &protos.Config{
			MinCloudbotsPercentage:     30,
			MinLowRiskModelsPercentage: 60,
			LowRiskModels:              []string{"model-1", "model-2"},
			ExcludeDuts:                []string{"dut-1", "dut-2"},
			Overrides: []*protos.Override{
				{
					Board:      "board-1",
					Model:      "model-1",
					Percentage: 1,
				},
				{
					Board:      "board-2",
					Model:      "*",
					Percentage: 20,
				},
			},
		}
		got := NewConfigSearchable(context.Background(), cfg)
		want := &configSearchable{
			generalPercentage: 30,
			lowRiskPercentage: 60,
			overrideLowRisks: map[string]struct{}{
				"model-1": {},
				"model-2": {},
			},
			overrideDUTs: map[string]struct{}{
				"dut-1": {},
				"dut-2": {},
			},
			overrideBoards: map[string]int32{
				"board-2": 20,
			},
			overrideModels: boardToModelsToPercentage{
				"board-1": map[string]int32{
					"model-1": 1,
				},
			}}
		if diff := cmp.Diff(want, got, cmp.AllowUnexported(configSearchable{})); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})
}
