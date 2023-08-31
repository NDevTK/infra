// Copyright 2019 The LUCI Authors.
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

package clients

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go.chromium.org/luci/common/logging"

	"infra/appengine/crosskylabadmin/internal/tq"
	"infra/libs/skylab/common/heuristics"
)

const repairBotsQueue = "repair-bots"
const repairLabstationQueue = "repair-labstations"
const auditBotsQueue = "audit-bots"

// PushRepairLabstations pushes BOT ids to taskqueue repairLabstationQueue for
// upcoming repair jobs.
func PushRepairLabstations(ctx context.Context, botIDs []string) error {
	return pushDUTs(ctx, repairLabstationQueue, createTasks(botIDs, "", "", labstationRepairTask))
}

// PushRepairDUTs pushes BOT ids to taskqueue repairBotsQueue for upcoming repair
// jobs.
func PushRepairDUTs(ctx context.Context, botIDs []string, expectedState string, swarmingPool string) error {
	return pushDUTs(ctx, repairBotsQueue, createTasks(botIDs, expectedState, swarmingPool, crosRepairTask))
}

// PushAuditDUTs pushes BOT ids to taskqueue auditBotsQueue for upcoming audit jobs.
func PushAuditDUTs(ctx context.Context, botIDs, actions []string, taskname string) error {
	actionsCSV := strings.Join(actions, ",")
	actionsStr := strings.Join(actions, "-")
	tasks := make([]*tq.Task, 0, len(botIDs))
	for _, id := range botIDs {
		if heuristics.LooksLikeSatlabDevice(id) {
			logging.Infof(ctx, fmt.Sprintf("Skipping audit for satlab device %q", id))
			continue
		}
		tasks = append(tasks, crosAuditTask(id, taskname, actionsCSV, actionsStr))
	}
	return pushDUTs(ctx, auditBotsQueue, tasks)
}

func crosRepairTask(botID, expectedState string, swarmingPool string) *tq.Task {
	values := url.Values{}
	values.Set("botID", botID)
	values.Set("expectedState", expectedState)
	values.Set("swarmingPool", swarmingPool)
	return tq.NewPOSTTask(fmt.Sprintf("/internal/task/cros_repair/%s", botID), values)
}

func labstationRepairTask(botID, expectedState string, pool string) *tq.Task {
	values := url.Values{}
	values.Set("botID", botID)
	return tq.NewPOSTTask(fmt.Sprintf("/internal/task/labstation_repair/%s", botID), values)
}

func crosAuditTask(botID, taskname, actionsCSV, actionsStr string) *tq.Task {
	values := url.Values{}
	values.Set("botID", botID)
	values.Set("taskname", taskname)
	values.Set("actions", actionsCSV)
	return tq.NewPOSTTask(fmt.Sprintf("/internal/task/audit/%s/%s", botID, actionsStr), values)
}

func createTasks(botIDs []string, expectedState string, swarmingPool string, taskGenerator func(string, string, string) *tq.Task) []*tq.Task {
	tasks := make([]*tq.Task, 0, len(botIDs))
	for _, id := range botIDs {
		tasks = append(tasks, taskGenerator(id, expectedState, swarmingPool))
	}
	return tasks
}

func pushDUTs(ctx context.Context, queueName string, tasks []*tq.Task) error {
	if err := tq.Add(ctx, queueName, tasks...); err != nil {
		return err
	}
	logging.Infof(ctx, "enqueued %d tasks", len(tasks))
	return nil
}
