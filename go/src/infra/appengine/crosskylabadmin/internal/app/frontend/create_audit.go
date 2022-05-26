// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/appengine/crosskylabadmin/internal/app/clients"
	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/app/frontend/worker"
)

// CreateLegacyAuditTask kicks off a legacy audit job.
func CreateLegacyAuditTask(ctx context.Context, botID, taskname, actions string) (string, error) {
	at := worker.AuditTaskWithActions(ctx, taskname, actions)
	sc, err := clients.NewSwarmingClient(ctx, config.Get(ctx).Swarming.Host)
	if err != nil {
		return "", errors.Annotate(err, "failed to obtain swarming client").Err()
	}
	expSec := int64(24 * 60 * 60)
	execTimeoutSecs := int64(8 * 60 * 60)
	taskURL, err := runTaskByBotID(ctx, at, sc, botID, "", expSec, execTimeoutSecs)
	if err != nil {
		return "", errors.Annotate(err, "fail to create audit task for %s", botID).Err()
	}
	return taskURL, nil
}
