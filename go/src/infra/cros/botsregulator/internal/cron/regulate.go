// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cron defines the service's cron job.
package cron

import (
	"context"

	"go.chromium.org/luci/common/logging"

	"infra/cros/botsregulator/internal/regulator"
)

// Regulate is BotsRegulator main flow.
// It fetches available DUTs from UFS based on specific filters
// and sends out the result to a predefined Bots Provider Interface.
func Regulate(ctx context.Context, opts *regulator.RegulatorOptions) error {
	r, err := regulator.NewRegulator(ctx, opts)
	if err != nil {
		return err
	}
	lses, err := r.FetchLSEsByHive(ctx)
	if err != nil {
		return err
	}
	if len(lses) == 0 {
		logging.Infof(ctx, "no lse found, exiting early")
		return nil
	}
	logging.Infof(ctx, "lses: %v\n", lses)
	sus, err := r.FetchAllSchedulingUnits(ctx)
	if err != nil {
		return err
	}
	dbs, err := r.ListDroneBots(ctx)
	if err != nil {
		return err
	}
	ad, err := r.ConsolidateAvailableDUTs(ctx, dbs, lses, sus)
	if err != nil {
		return err
	}
	logging.Infof(ctx, "available DUTs: %v\n", ad)
	err = r.UpdateConfig(ctx, ad)
	if err != nil {
		return err
	}
	return nil
}
