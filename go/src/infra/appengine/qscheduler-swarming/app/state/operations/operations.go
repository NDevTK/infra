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

package operations

import (
	"context"
	"fmt"

	swarming "infra/swarming"

	"github.com/pkg/errors"

	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/data/strpair"
	"go.chromium.org/luci/common/logging"

	"infra/appengine/qscheduler-swarming/app/state/types"
	"infra/qscheduler/qslib/reconciler"
	"infra/qscheduler/qslib/scheduler"
	"infra/qscheduler/qslib/tutils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AccountIDTagKey is the key used in Task tags to specify which quotascheduler
// account the task should be charged to.
const AccountIDTagKey = "qs_account"

// AssignTasks returns an operation that will perform the given Assign request.
//
// The result object will have the operation response stored in it after
// the operation has run.
func AssignTasks(r *swarming.AssignTasksRequest) (types.Operation, *swarming.AssignTasksResponse) {
	var response swarming.AssignTasksResponse
	return func(ctx context.Context, state *types.QScheduler, events scheduler.EventSink) error {
		idles := make([]*reconciler.IdleWorker, len(r.IdleBots))
		for i, v := range r.IdleBots {
			s := stringset.NewFromSlice(v.Dimensions...)
			if !s.HasAll(state.Scheduler.Config().Labels...) {
				return status.Errorf(codes.InvalidArgument, "bot with id %s does not have all scheduler dimensions", v.BotId)
			}
			idles[i] = &reconciler.IdleWorker{
				ID:     scheduler.WorkerID(v.BotId),
				Labels: stringset.NewFromSlice(v.Dimensions...),
			}
		}

		schedulerAssignments := state.Reconciler.AssignTasks(ctx, state.Scheduler, tutils.Timestamp(r.Time), events, idles...)

		assignments := make([]*swarming.TaskAssignment, len(schedulerAssignments))
		for i, v := range schedulerAssignments {
			slice := int32(0)
			if v.ProvisionRequired {
				slice = 1
			}
			assignments[i] = &swarming.TaskAssignment{
				BotId:       string(v.WorkerID),
				TaskId:      string(v.RequestID),
				SliceNumber: slice,
			}
		}

		response = swarming.AssignTasksResponse{Assignments: assignments}
		return nil
	}, &response
}

// NotifyTasks returns an operation that will perform the given Notify request,
// and result object that will get the results after the operation is run.
func NotifyTasks(r *swarming.NotifyTasksRequest) (types.Operation, *swarming.NotifyTasksResponse) {
	var response swarming.NotifyTasksResponse
	return func(ctx context.Context, sp *types.QScheduler, events scheduler.EventSink) error {
		if sp.Scheduler.Config() == nil {
			return errors.Errorf("Scheduler with id %s has nil config.", r.SchedulerId)
		}

		for _, n := range r.Notifications {
			var t taskState
			var ok bool
			if t, ok = translateTaskState(n.Task.State); !ok {
				err := fmt.Sprintf("Invalid notification with unhandled state %s.", n.Task.State)
				logging.Warningf(ctx, err)
				sp.Reconciler.TaskError(scheduler.RequestID(n.Task.Id), err)
				continue
			}

			switch t {
			case taskStateAbsent:
				r := &reconciler.TaskAbsentRequest{RequestID: scheduler.RequestID(n.Task.Id), Time: tutils.Timestamp(n.Time)}
				sp.Reconciler.NotifyTaskAbsent(ctx, sp.Scheduler, events, r)
			case taskStateRunning:
				r := &reconciler.TaskRunningRequest{
					RequestID: scheduler.RequestID(n.Task.Id),
					Time:      tutils.Timestamp(n.Time),
					WorkerID:  scheduler.WorkerID(n.Task.BotId),
				}
				sp.Reconciler.NotifyTaskRunning(ctx, sp.Scheduler, events, r)
			case taskStateWaiting:
				if err := notifyTaskWaiting(ctx, sp, events, n); err != nil {
					sp.Reconciler.TaskError(scheduler.RequestID(n.Task.Id), err.Error())
					logging.Warningf(ctx, err.Error())
				}
			default:
				e := fmt.Sprintf("invalid update type %d", t)
				logging.Warningf(ctx, e)
				sp.Reconciler.TaskError(scheduler.RequestID(n.Task.Id), e)
			}
		}
		response = swarming.NotifyTasksResponse{}
		return nil
	}, &response
}

func notifyTaskWaiting(ctx context.Context, sp *types.QScheduler, events scheduler.EventSink, n *swarming.NotifyTasksItem) error {
	var provisionableLabels []string
	var baseLabels []string
	var accountID string
	labels, err := computeLabels(n)
	if err != nil {
		return err
	}
	provisionableLabels = labels.provisionable
	baseLabels = labels.base

	s := stringset.NewFromSlice(baseLabels...)
	if !s.HasAll(sp.Scheduler.Config().Labels...) {
		return fmt.Errorf("task with base dimensions %s does not contain all of scheduler dimensions %s", baseLabels, sp.Scheduler.Config().Labels)
	}

	if accountID, err = GetAccountID(n); err != nil {
		return err
	}
	r := &reconciler.TaskWaitingRequest{
		AccountID:           scheduler.AccountID(accountID),
		BaseLabels:          stringset.NewFromSlice(baseLabels...),
		EnqueueTime:         tutils.Timestamp(n.Task.EnqueuedTime),
		ProvisionableLabels: stringset.NewFromSlice(provisionableLabels...),
		RequestID:           scheduler.RequestID(n.Task.Id),
		Tags:                n.Task.Tags,
		Time:                tutils.Timestamp(n.Time),
	}
	sp.Reconciler.NotifyTaskWaiting(ctx, sp.Scheduler, events, r)

	return nil
}

// computeLabels determines the labels for a given task.
func computeLabels(n *swarming.NotifyTasksItem) (*labels, error) {
	slices := n.Task.Slices
	switch len(slices) {
	case 1:
		return &labels{base: slices[0].Dimensions}, nil
	case 2:
		s1 := stringset.NewFromSlice(slices[0].Dimensions...)
		s2 := stringset.NewFromSlice(slices[1].Dimensions...)
		// s2 must be a subset of s1 (i.e. the first slice must be more specific
		// about dimensions than the second one).
		if flaws := s2.Difference(s1); flaws.Len() != 0 {
			return nil, errors.Errorf("Invalid slice dimensions; task's 2nd slice dimensions are not a subset of 1st slice dimensions.")
		}

		provisionable := s1.Difference(s2).ToSlice()
		base := slices[1].Dimensions
		return &labels{provisionable: provisionable, base: base}, nil
	default:
		return nil, errors.Errorf("Invalid slice count %d; quotascheduler only supports 1-slice or 2-slice tasks.", len(n.Task.Slices))
	}
}

// GetAccountID determines the account id for a given task, based on its tags.
func GetAccountID(n *swarming.NotifyTasksItem) (string, error) {
	m := strpair.ParseMap(n.Task.Tags)
	accounts := m[AccountIDTagKey]
	switch len(accounts) {
	case 0:
		return "", nil
	case 1:
		return accounts[0], nil
	default:
		return "", errors.Errorf("Too many account tags.")
	}
}

type taskState int

const (
	taskStateUnknown taskState = iota
	taskStateWaiting
	taskStateRunning
	taskStateAbsent
)

func translateTaskState(s swarming.TaskState) (taskState, bool) {
	cInt := int(s) &^ int(swarming.TaskStateCategory_TASK_STATE_MASK)
	category := swarming.TaskStateCategory(cInt)

	// These category cases occur in the same order as they are defined in
	// swarming.proto. Please preserve that when adding new cases.
	switch category {
	case swarming.TaskStateCategory_CATEGORY_PENDING:
		return taskStateWaiting, true
	case swarming.TaskStateCategory_CATEGORY_RUNNING:
		return taskStateRunning, true
	// The following categories all translate to "ABSENT", because they are all
	// equivalent to the task being neither running nor waiting.
	case swarming.TaskStateCategory_CATEGORY_TRANSIENT_DONE,
		swarming.TaskStateCategory_CATEGORY_EXECUTION_DONE,
		swarming.TaskStateCategory_CATEGORY_NEVER_RAN_DONE:
		return taskStateAbsent, true

	// Invalid state.
	default:
		return taskStateUnknown, false
	}
}

// labels represents the computed labels for a task.
type labels struct {
	provisionable []string
	base          []string
}
