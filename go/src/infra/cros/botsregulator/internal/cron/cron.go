// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cron defines the service's cron job.
package cron

import (
	"context"

	"go.chromium.org/luci/common/logging"

	"infra/cros/botsregulator/internal/regulator"
	"infra/cros/botsregulator/internal/util"
)

func Regulate(ctx context.Context, opts *regulator.RegulatorOptions) error {
	r := regulator.NewRegulator(opts)
	lses, err := r.FetchDUTsByHive(ctx)
	if err != nil {
		return err
	}
	logging.Infof(ctx, "lses: %v\n", lses)

	hns, err := util.CutHostnames(lses)
	if err != nil {
		return err
	}
	logging.Infof(ctx, "hostnames: %v\n", hns)

	return nil
}