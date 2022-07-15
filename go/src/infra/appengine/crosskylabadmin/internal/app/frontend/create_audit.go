// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	"infra/appengine/crosskylabadmin/internal/app/clients"
	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/app/frontend/routing"
	"infra/appengine/crosskylabadmin/internal/app/frontend/worker"
	"infra/libs/skylab/common/heuristics"
)

// CreateAuditTask kicks off an audit job.
func CreateAuditTask(ctx context.Context, botID string, taskname string, actions string, randFloat float64) (string, error) {
	// The actions field is a little bit tricky and consists of a comma-delimited list of actions.
	// We're also using Paris in a slightly different way than legacy.
	// Each audit action will correspond to one paris job, always.
	logging.Infof(ctx, "Creating audit task for %q with random input %f and actions %q", botID, randFloat, actions)

	// Log, but do not otherwise use the chosen task result.
	// RouteTask returning an error does NOT imply that the surrounding task
	// should fail.
	taskType, err := RouteTask(
		ctx,
		RouteTaskParams{
			taskType:      taskname,
			botID:         botID,
			expectedState: "",
			pools:         nil,
		},
		randFloat,
	)
	logging.Infof(ctx, "RouteTask picked the taskType %d", int(taskType))
	if err == nil {
		logging.Infof(ctx, "RouteTask succeeded.")
	} else {
		logging.Infof(ctx, "RouteTask failed with error %q.", err.Error())
	}

	return createLegacyAuditTask(ctx, botID, taskname, actions)
}

// createLegacyAuditTask kicks off a legacy audit job.
func createLegacyAuditTask(ctx context.Context, botID string, taskname string, actions string) (string, error) {
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

// routeAuditTaskImpl routes an audit task (storage, rpm, USB) based on the rollout config that's tied to that specific task.
func routeAuditTaskImpl(ctx context.Context, r *config.RolloutConfig) (heuristics.TaskType, routing.Reason) {
	if r == nil {
		return routing.Legacy, routing.ParisNotEnabled
	}
	return routing.Legacy, routing.NotImplemented
}
