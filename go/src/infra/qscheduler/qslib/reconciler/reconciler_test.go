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

// TestQuotaschedulerInterface ensures that scheduler.Scheduler is a valid
// implementation of the Scheduler interface.
func TestQuotaschedulerInterface(t *testing.T) {
	var s interface{} = &scheduler.Scheduler{}
	if _, ok := s.(Scheduler); !ok {
		t.Errorf("Scheduler interface should be implemented by *scheduler.Scheduler")
	}
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
		assignments: make(map[string]*scheduler.Assignment),
		idleWorkers: make(map[string]bool),
	}
}

func assertAssignments(t *testing.T, description string,
	got []Assignment, want []Assignment) {
	t.Helper()
	if diff := pretty.Compare(got, want); diff != "" {
		t.Errorf(fmt.Sprintf("%s got unexpected assignment diff (-got +want): %s", description, diff))
	}
}

// TestReapOneAssignment tests that a scheduler assignment for a reaping
// worker is correctly reaped, and that subsequent reap calls prior to
// ack return the same assignment.
func TestReapOneAssignment(t *testing.T) {
	fs := newFakeScheduler()
	state := New()

	epoch := time.Unix(0, 0)
	ti := epoch
	fs.fakeSchedule(&scheduler.Assignment{
		RequestId: "r1",
		WorkerId:  "w1",
		Type:      scheduler.Assignment_IDLE_WORKER,
		Time:      tutils.TimestampProto(ti),
	})

	reapingWorkers := []*ReapingWorker{
		&ReapingWorker{"w1", []string{}},
	}

	// Reap once for worker "w1".
	got := state.Reap(fs, reapingWorkers, ti)
	want := []Assignment{Assignment{"w1", "r1"}}
	assertAssignments(t, "Simple single-worker reap", got, want)

	// Subsequent reap should return the same assignment, as it has not been
	// ack'd.
	ti = ti.Add(1)
	got = state.Reap(fs, reapingWorkers, ti)
	assertAssignments(t, "Second reap", got, want)
}

// TestReapQueuedAssignment tests that a scheduler assignment is queued until
// the relevant worker reaps, even if that worker has already been given
// its assignment by the scheduler.
func TestReapQueuedAssignment(t *testing.T) {
	fakeSch := newFakeScheduler()
	state := New()

	epoch := time.Unix(0, 0)
	ti := epoch

	w1 := []*ReapingWorker{
		&ReapingWorker{"w1", []string{}},
	}
	w2 := []*ReapingWorker{
		&ReapingWorker{"w2", []string{}},
	}

	// Mark w1 as idle, prior to any assignment for it.
	got := state.Reap(fakeSch, w1, ti)
	assertAssignments(t, "Pre-assignment reap", got, []Assignment{})

	// Give an assignment to w1, but reap for w2.
	ti = ti.Add(1)
	fakeSch.fakeSchedule(&scheduler.Assignment{
		RequestId: "r1",
		WorkerId:  "w1",
		Type:      scheduler.Assignment_IDLE_WORKER,
		Time:      tutils.TimestampProto(ti),
	})
	got = state.Reap(fakeSch, w2, ti)
	assertAssignments(t, "Post-assignment reap of w2", got, []Assignment{})

	// Reap for w1.
	ti = ti.Add(1)
	got = state.Reap(fakeSch, w1, ti)
	want := []Assignment{Assignment{"w1", "r1"}}
	assertAssignments(t, "Post-assignment reap of w1", got, want)
}
