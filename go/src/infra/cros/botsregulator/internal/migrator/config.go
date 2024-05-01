// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package migrator

import (
	"context"

	"go.chromium.org/luci/common/logging"

	"infra/cros/botsregulator/protos"
)

type boardToModelsToAmount map[string]map[string]int32

type configSearchable struct {
	generalAmount    int32
	lowRiskAmount    int32
	overrideBoards   map[string]int32
	overrideDUTs     map[string]struct{}
	overrideLowRisks map[string]struct{}
	overrideModels   boardToModelsToAmount
}

// NewConfigSearchable returns an easily searchable struct composed of maps instead of slices.
func NewConfigSearchable(ctx context.Context, config *protos.Config) *configSearchable {
	bma := make(boardToModelsToAmount)
	ba := make(map[string]int32)
	for _, override := range config.Overrides {
		// Check if this is a board Override.
		if override.Model == "*" {
			if _, ok := ba[override.Board]; !ok {
				ba[override.Board] = override.Amount
			} else {
				logging.Errorf(ctx, "board %s has already been processed. Check for duplicate in %s", override.Board, migrationFile)
			}
		} else {
			if _, ok := bma[override.Board]; !ok {
				bma[override.Board] = make(map[string]int32)
			}
			if _, ok := bma[override.Board][override.Model]; !ok {
				bma[override.Board][override.Model] = override.Amount
			} else {
				logging.Errorf(ctx, "model %s has already been processed. Check for duplicate in %s", override.Model, migrationFile)
			}
		}
	}
	// Low risk models.
	lr := make(map[string]struct{})
	for _, m := range config.LowRiskModels {
		if _, ok := lr[m]; !ok {
			lr[m] = struct{}{}
		} else {
			logging.Errorf(ctx, "model %s has already been processed. Check for duplicate in %s", m, migrationFile)
		}
	}
	// Exclude DUTs.
	duts := make(map[string]struct{})
	for _, dut := range config.ExcludeDuts {
		if _, ok := duts[dut]; !ok {
			duts[dut] = struct{}{}
		} else {
			logging.Errorf(ctx, "model %s has already been processed. Check for duplicate in %s", dut, migrationFile)
		}
	}
	searchable := &configSearchable{
		generalAmount:    config.MinCloudbotsPercentage,
		lowRiskAmount:    config.MinLowRiskModelsPercentage,
		overrideModels:   bma,
		overrideBoards:   ba,
		overrideLowRisks: lr,
		overrideDUTs:     duts,
	}
	return searchable
}
