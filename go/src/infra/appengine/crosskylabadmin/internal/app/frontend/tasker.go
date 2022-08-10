// Copyright 2018 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package frontend

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/clients"
	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/app/frontend/util"
)

func runTaskByBotID(ctx context.Context, at util.Task, sc clients.SwarmingClient, botID, expectedState string, expirationSecs, executionTimeoutSecs int64) (string, error) {
	cfg := config.Get(ctx)
	tags := util.AddCommonTags(
		ctx,
		fmt.Sprintf("%s:%s", at.Name, botID),
		fmt.Sprintf("task:%s", at.Name),
	)
	tags = append(tags, at.Tags...)

	a := util.SetCommonTaskArgs(ctx, &clients.SwarmingCreateTaskArgs{
		Cmd:                  at.Cmd,
		BotID:                botID,
		ExecutionTimeoutSecs: executionTimeoutSecs,
		ExpirationSecs:       expirationSecs,
		Priority:             cfg.Cron.FleetAdminTaskPriority,
		Tags:                 tags,
	})
	if expectedState != "" {
		a.DutState = expectedState
	}
	tid, err := sc.CreateTask(ctx, at.Name, a)
	if err != nil {
		return "", errors.Annotate(err, "failed to create task for bot %s", botID).Err()
	}
	logging.Infof(ctx, "successfully kick off task %s for bot %s", tid, botID)
	return util.URLForTask(ctx, tid), nil
}

var dutStateForTask = map[fleet.TaskType]string{
	fleet.TaskType_Cleanup: "needs_cleanup",
	fleet.TaskType_Repair:  "needs_repair",
	fleet.TaskType_Reset:   "needs_reset",
}
