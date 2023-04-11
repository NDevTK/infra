// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dumper

import (
	"context"

	"go.chromium.org/luci/common/logging"

	"infra/unifiedfleet/app/config"
)

// syncDeviceConfigs fetches devices configs from a file checked into gerrit
// and inserts to UFS datastore
//
// this outer level function creates any clients and calls an inner function to
// execute any work
func syncDeviceConfigs(ctx context.Context) (err error) {
	// get ufs-level config for this cron job
	cronCfg := config.Get(ctx).GetDeviceConfigsPushConfigs()

	if !cronCfg.Enabled {
		logging.Infof(ctx, "ufs.device_config.sync scheduled but is disabled in this env")
		return
	}

	logging.Infof(ctx, "ufs.device_config.sync not implemented")
	return
}
