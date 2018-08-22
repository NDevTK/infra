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

// Package reconciler implements logic necessary to reconcile api calls
// (Update, Reap, etc) to qslib with a quotascheduler state.
package reconciler

import (
	"time"

	"infra/qscheduler/qslib/scheduler"
	"infra/qscheduler/qslib/types/task"
)

// WorkerQueue represents the queue of qscheduler operations that are pending
// for a given worker.
//
// At present, the queue of operations for a worker can be at most 2 elements
// in length, and consider of either:
// - An Abort Job operation followed by an Assign Job operation.
// - An Assign Job operation.
//
// Therefore, instead of representing this as a list of operations, it is
// convenient to flatten this queue into a single object.
//
// TODO: Turn this into a proto, because it will need to get serialized.
type WorkerQueue struct {
	// TODO: Implement me.

	// Time at which these operations were enqueued.
	Time time.Time

	// TaskToAbort indicates the task request id that should be aborted on this worker.
	//
	// Empty string "" indicates that there is nothing to abort.
	TaskToAbort string

	// TaskToAssign is the task request that should be assigned to this worker.
	TaskToAssign string
}

// Config represents configuration options for a reconciler.
type Config struct {
	// TODO: Implement me.
	// Include things such as:
	// - ACK timeout for worker aborts.
	// - ACK timeout for worker-task assignments.
}

// State represents a quotascheduler state, plus pending operations that are
// in-flight and have not been ACK'ed yet.
//
// TODO: Turn this into a proto, because it will need to get serialized.
type State struct {

	Config *Config

	WorkerQueues map[string]*WorkerQueue
}

// ReapingWorker represents a worker that is idle and would like to reap a task.
type ReapingWorker struct {
	// Id is the Id of the idle worker.
	Id string

	// ProvisionableLabels is the set of provisionable labels of the idle worker.
	ProvisionableLabels task.LabelSet
}

// Assignment represents a scheduler-initated operation to assign a task to a worker.
type Assignment struct {
	// WorkerId is the id the worker that is being assigned a task.
	WorkerId string

	// RequestId is the id of the task request that is being assigned.
	RequestId string
}

// Scheduler is the interface with which reconciler interacts with a scheduler.
// One implementation of this interface (the quotascheduler) is provided
// by qslib/scheduler.Scheduler.
type Scheduler interface {
	// UpdateTime informs the scheduler of the current time.
	UpdateTime(t time.Time) error

	// MarkIdle informs the scheduler that a given worker is idle, with
	// given labels.
	MarkIdle(id string, labels task.LabelSet)

	// RunOnce runs through one round of the scheduling algorith, and determines
	// and returns work assignments.
	RunOnce() []*scheduler.Assignment
}

// Reap accepts a slice of idle workers, and returns tasks to be reaped
// for those workers (if there are tasks available).
func (state *State) Reap(s Scheduler, workers []*ReapingWorker, t time.Time) []*Assignment {
	// For any workers that already have an assignment in their workerqueue, yield
	// that.
	//
	// Otherwise, consider these workers as Idle, and initiate a new round
	// of the quotascheduler algorithm for them.

	var idleWorkers []*ReapingWorker
	var assignments []*Assignment

	// For any workers where there is a queued operation, return the queued operation.
	for _, w := range workers {
		if queue, ok := state.WorkerQueues[w.Id]; ok {
			if t.Before(queue.Time) {
				// TODO: Handle this case. The reap request is from a time before
				// the operations were enqueued.
				//
				// Should we panic here, ignore this worker, or what?
			} else {
				// Reap the queued task-to-assign.
				assignments = append(assignments, &Assignment{w.Id, queue.TaskToAssign})

				// We've reaped the intended task to assign to this worker, so no need
				// to abort anything on it anymore.
				// TODO: log a warning if this wasn't already "", as that indicates that
				// we missed a task Update() that would otherwise have nulled this
				// out.
				queue.TaskToAbort = ""
			}
		} else {
			idleWorkers = append(idleWorkers, w)
		}
	}

	s.UpdateTime(t)

	for _, w := range idleWorkers {
		s.MarkIdle(w.Id, w.ProvisionableLabels)
	}

	new_assignments := s.RunOnce()
	for _, a := range new_assignments {
		switch a.Type {

		}
	}

	return assignments
}

// Cancellation represents a scheduler-initated operation to cancel a task on a worker.
// The worker should be aborted if and only if it is currently running the given task.
//
// TODO: Consider unifying this with Assignment, since it is in fact the same content.
type Cancellation struct {
	// WorkerId is the id the worker where we should cancel a task.
	WorkerId string

	// RequestId is the id of the task that we should request.
	RequestId string
}

// GetCancellations returns the set of workers and tasks that should be cancelled.
func (state *State) GetCancellations(t time.Time) []Cancellation {
	// TODO: implement me
	return nil
}

// TaskUpdate represents a change in the state of an existing task, or the
// creation of a new task.
type TaskUpdate struct {
	// TODO: Implement me.
	// Should specify things like:
	// - task id
	// - task state (New, Assigned, Cancelled)
	// - worker id (if state is Assigned)
	Time time.Time
}

// UpdateTasks is called to inform a quotascheduler about task state changes
// (creation of new tasks, assignment of task to worker, cancellation of a task).
// These updates must be called in order to acknowledge that previously returned
// scheduler operations have been completed (otherwise, future calls to Reap or
func (state *State) UpdateTasks(updates []TaskUpdate) {
	// TODO: Implement me.
}
