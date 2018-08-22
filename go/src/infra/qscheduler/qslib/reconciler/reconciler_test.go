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

package reconciler

import (
	"fmt"
	"testing"
	"time"

	"infra/qscheduler/qslib/scheduler"
	"infra/qscheduler/qslib/tutils"
	"infra/qscheduler/qslib/types/task"

	"github.com/kylelemons/godebug/pretty"
)

// epoch is an arbitrary time for testing purposes, corresponds to
// 01/01/2018 @ 1:00 am UTC
var epoch = time.Unix(1514768400, 0)

// TestQuotaschedulerInterface is a compile time check that ensures
// that scheduler.Scheduler is a valid implementation of the Scheduler
// interface.
func TestQuotaschedulerInterface(t *testing.T) {
	var _ Scheduler = &scheduler.Scheduler{}
}

// fakeScheduler is an implementation of the Scheduler interface which reaps
// according to whatever assignments are provided by MockSchedule.
type fakeScheduler struct {
	// assignments is a map from worker ID to the scheduler.Assignment that will
	// be reaped for that worker.
	assignments map[string]*scheduler.Assignment

	// idleWorkers is the set of workers that have been marked as idle and have
	// not yet had any assignments scheduled / reaped for them
	idleWorkers map[string]bool
}

// UpdateTime is an implmementation of the Scheduler interface.
func (s *fakeScheduler) UpdateTime(t time.Time) error {
	return nil
}

// MarkIdle is an implementation of the Scheduler interface.
func (s *fakeScheduler) MarkIdle(id string, labels task.LabelSet) {
	s.idleWorkers[id] = true
}

// RunOnce is an implementation of the Scheduler interface.
func (s *fakeScheduler) RunOnce() []*scheduler.Assignment {
	response := make([]*scheduler.Assignment, 0, len(s.idleWorkers))
	for worker := range s.idleWorkers {
		if match, ok := s.assignments[worker]; ok {
			response = append(response, match)
			delete(s.assignments, worker)
			delete(s.idleWorkers, worker)
		}
	}
	return response
}

// fakeSchedule sets the given assignment in a fakeScheduler.
func (s *fakeScheduler) fakeSchedule(a *scheduler.Assignment) {
	s.assignments[a.WorkerId] = a
}

// newFakeScheduler returns a new initialized mock scheduler.
func newFakeScheduler() *fakeScheduler {
	return &fakeScheduler{
		make(map[string]*scheduler.Assignment),
		make(map[string]bool),
	}
}

func assertAssignments(t *testing.T, description string,
	actual []*Assignment, expect []*Assignment) {
	if diff := pretty.Compare(actual, expect); diff != "" {
		t.Errorf(fmt.Sprintf("%s got unexpected assignment diff (-got +want): %s", description, diff))
	}
}

// TestReapOneAssignment tests that a scheduler assignment for a reaping
// worker is correctly reaped, and that subsequent reap calls prior to
// ack return the same assignment.
func TestReapOneAssignment(t *testing.T) {
	fakeSch := newFakeScheduler()
	state := NewState()

	ti := epoch
	fakeSch.fakeSchedule(&scheduler.Assignment{
		RequestId: "r1",
		WorkerId:  "w1",
		Type:      scheduler.Assignment_IDLE_WORKER,
		Time:      tutils.TimestampProto(ti),
	})

	reapingWorkers := []*ReapingWorker{
		&ReapingWorker{"w1", []string{}},
	}

	// Reap once for worker "w1".
	actual := state.Reap(fakeSch, reapingWorkers, ti)
	expect := []*Assignment{&Assignment{"w1", "r1"}}
	assertAssignments(t, "Simple single-worker reap", actual, expect)

	// Subsequent reap should return the same assignment, as it has not been
	// ack'd.
	ti = ti.Add(1)
	actual = state.Reap(fakeSch, reapingWorkers, ti)
	assertAssignments(t, "Second reap", actual, expect)
}

// TestReapQueuedAssignment tests that a scheduler assignment is queued until
// the relevant worker reaps, even if that worker has already been given
// its assignment by the scheduler.
func TestReapQueuedAssignment(t *testing.T) {
	fakeSch := newFakeScheduler()
	state := NewState()

	ti := epoch

	reapingW1 := []*ReapingWorker{
		&ReapingWorker{"w1", []string{}},
	}
	reapingW2 := []*ReapingWorker{
		&ReapingWorker{"w2", []string{}},
	}

	// Mark w1 as idle, prior to any assignment for it.
	actual := state.Reap(fakeSch, reapingW1, ti)
	assertAssignments(t, "Pre-assignment reap", actual, []*Assignment{})

	// Give an assignment to w1, but reap for w2.
	ti = ti.Add(1)
	fakeSch.fakeSchedule(&scheduler.Assignment{
		RequestId: "r1",
		WorkerId:  "w1",
		Type:      scheduler.Assignment_IDLE_WORKER,
		Time:      tutils.TimestampProto(ti),
	})
	actual = state.Reap(fakeSch, reapingW2, ti)
	assertAssignments(t, "Post-assignment reap of w2", actual, []*Assignment{})

	// Reap for w1.
	ti = ti.Add(1)
	actual = state.Reap(fakeSch, reapingW1, ti)
	expect := []*Assignment{&Assignment{"w1", "r1"}}
	assertAssignments(t, "Post-assignment reap of w1", actual, expect)
}
