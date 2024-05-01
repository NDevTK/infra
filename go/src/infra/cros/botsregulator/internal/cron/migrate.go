// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cron

import (
	"context"

	"go.chromium.org/luci/common/logging"

	"infra/cros/botsregulator/internal/migrator"
	"infra/cros/botsregulator/internal/regulator"
)

func Migrate(ctx context.Context, r *regulator.RegulatorOptions) error {
	m, err := migrator.NewMigrator(ctx, r)
	if err != nil {
		return err
	}
	cfg, err := m.GetMigrationConfig(ctx)
	if err != nil {
		return err
	}
	logging.Infof(ctx, "migration file: %v \n", cfg)
	return nil
}
