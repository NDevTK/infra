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

package app

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/sync/parallel"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
)

const (
	// luciferToolsDeploymentPath is the well known path to infra tools deployed on the drone.
	luciferToolsDeploymentPath = "/opt/infra-tools/usr/bin"
	// skylabSwarmingWrokerPath is the path to the binary on the drone
	// that is the entry point of all tasks.
	skylabSwarmingWorkerPath = luciferToolsDeploymentPath + "/skylab_swarming_worker"
)

// taskerServerImpl implements the fleet.TaskerServer interface.
type taskerServerImpl struct {
	swarmingClientFactory
}

func (*taskerServerImpl) TriggerRepairOnIdle(context.Context, *fleet.TriggerRepairOnIdleRequest) (resp *fleet.TaskerTasksResponse, err error) {
	defer func() {
		err = grpcfyRawErrors(err)
	}()

	return nil, status.Errorf(codes.Unimplemented, "Not Implemented")
}

func (*taskerServerImpl) TriggerRepairOnRepairFailed(context.Context, *fleet.TriggerRepairOnRepairFailedRequest) (resp *fleet.TaskerTasksResponse, err error) {
	defer func() {
		err = grpcfyRawErrors(err)
	}()

	return nil, status.Errorf(codes.Unimplemented, "Not Implemented")
}

func (tsi *taskerServerImpl) EnsureBackgroundTasks(c context.Context, req *fleet.EnsureBackgroundTasksRequest) (resp *fleet.TaskerTasksResponse, err error) {
	defer func() {
		err = grpcfyRawErrors(err)
	}()

	bses, err := getBotSummariesFromDatastore(c, req.Selectors)
	if err != nil {
		return nil, errors.Annotate(err, "failed to obtain requested bots from datastore").Err()
	}

	sc, err := tsi.swarmingClient(c, swarmingInstance)
	if err != nil {
		return nil, errors.Annotate(err, "failed to obtain Swarming client").Err()
	}

	// Protects access to botTasks
	m := &sync.Mutex{}
	botTasks := make([]*fleet.TaskerBotTasks, 0, len(bses))
	err = parallel.WorkPool(maxConcurrentSwarmingCalls, func(workC chan<- func() error) {
		for i := range bses {
			// In-scope variable for goroutine closure.
			bse := bses[i]
			workC <- func() error {
				bt, err := ensureBackgroundTasksForBot(c, sc, req, bse.DutID)
				if bt != nil {
					m.Lock()
					defer m.Unlock()
					botTasks = append(botTasks, bt)
				}
				return err
			}
			if c.Err() != nil {
				return
			}
		}
	})

	resp = &fleet.TaskerTasksResponse{
		BotTasks: botTasks,
	}
	return resp, err
}

var dutStateForTask = map[fleet.TaskType]string{
	fleet.TaskType_Cleanup: "needs_cleanup",
	fleet.TaskType_Repair:  "needs_repair",
	fleet.TaskType_Reset:   "needs_reset",
}

func ensureBackgroundTasksForBot(c context.Context, sc SwarmingClient, req *fleet.EnsureBackgroundTasksRequest, dutID string) (*fleet.TaskerBotTasks, error) {
	ts := make([]*fleet.TaskerTask, 0, req.TaskCount)
	tags := []string{luciProjectTag, backgroundTaskTag(req.Type, dutID)}
	oldTasks, err := sc.ListPendingTasks(c, tags)
	if err != nil {
		return nil, errors.Annotate(err, "Failed to list existing tasks of type %s for dut %s",
			req.Type.String(), dutID).Err()
	}
	for _, ot := range oldTasks {
		ts = append(ts, &fleet.TaskerTask{
			TaskUrl: swarmingURLForTask(ot.TaskId),
			Type:    req.Type,
		})
	}

	newTaskCount := int(req.TaskCount) - len(ts)
	for i := 0; i < newTaskCount; i++ {
		tid, err := sc.CreateTask(c, &SwarmingCreateTaskArgs{
			Cmd:                  luciferAdminTaskCmd(req.Type),
			DutID:                dutID,
			DutState:             dutStateForTask[req.Type],
			ExecutionTimeoutSecs: backgroundTaskExecutionTimeoutSecs,
			ExpirationSecs:       backgroundTaskExpirationSecs,
			Pool:                 swarmingBotPool,
			Priority:             req.Priority,
			Tags:                 tags,
		})
		if err != nil {
			return nil, errors.Annotate(err, "Error when creating %dth task for dut %q", i+1, dutID).Err()
		}
		ts = append(ts, &fleet.TaskerTask{
			TaskUrl: swarmingURLForTask(tid),
			Type:    req.Type,
		})
	}
	return &fleet.TaskerBotTasks{
		DutId: dutID,
		Tasks: ts,
	}, nil
}

func backgroundTaskTag(ttype fleet.TaskType, dutID string) string {
	return fmt.Sprintf("background_task:%s_%s", ttype.String(), dutID)
}

func luciferAdminTaskCmd(ttype fleet.TaskType) []string {
	return []string{
		skylabSwarmingWorkerPath,
		"-task-name", fmt.Sprintf("admin_%s", strings.ToLower(ttype.String())),
	}
}

func swarmingURLForTask(tid string) string {
	u := url.URL{
		Scheme: "https",
		Host:   swarmingInstance,
		Path:   "task",
	}
	q := u.Query()
	q.Set("id", tid)
	u.RawQuery = q.Encode()
	return u.String()
}
