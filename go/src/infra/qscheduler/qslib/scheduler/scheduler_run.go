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

package scheduler

import (
	"fmt"
	"math"
	"sort"

	"infra/qscheduler/qslib/metrics"
)

// schedulerRun stores values that are used within a single run of the scheduling algorithm.
// Its fields may be mutated during the run, as requests get assigned to workers.
type schedulerRun struct {
	// idleWorkers is a collection of currently idle workers.
	idleWorkers map[WorkerID]*worker

	// requestsPerPriority is a per-priority linked list of queued requests, sorted by FIFO order within
	// each priority level. It takes into account throttling of any requests whose account was
	// already at the fanout limit when the pass was started, but not newly throttled accounts
	// as a result of newly assigned requests/workers.
	requestsPerPriority [NumPriorities + 1]requestList

	// jobsUntilThrottled is the number of additional requests that each account may run before reaching
	// its fanout limit and becoming throttled (at which point its other requests are demoted to FreeBucket).
	jobsUntilThrottled map[AccountID]int

	scheduler *Scheduler
}

func (run *schedulerRun) Run(m MetricsSink) []*Assignment {
	var output []*Assignment
	// Proceed through multiple passes of the scheduling algorithm, from highest
	// to lowest priority requests (high priority = low p).
	for p := Priority(0); p < NumPriorities; p++ {
		// Step 1: Match any requests to idle workers that have matching
		// provisionable labels.
		output = append(output, run.matchIdleBots(p, provisionAwareMatch, m)...)
		// Step 2: Match request to any remaining idle workers, regardless of
		// provisionable labels.
		output = append(output, run.matchIdleBots(p, basicMatch, m)...)
		// Step 3: Demote (out of this level) or promote (into this level) any
		// already running tasks that qualify.
		run.reprioritizeRunningTasks(p, m)
		// Step 4: Preempt any lower priority running tasks.
		output = append(output, run.preemptRunningTasks(p, m)...)
		// Step 5: Give any requests that were throttled in this pass a chance to be scheduled
		// during the final FreeBucket pass.
		run.moveThrottledRequests(p)
	}

	// A final pass matches free jobs (in the FreeBucket) to any remaining
	// idle workers. The reprioritize and preempt stages do not apply here.
	// TODO(akeshet): Consider a final sorting step here, so that FIFO ordering is respected among
	// FreeBucket jobs, including those that were moved the the throttled list during the pass above.
	output = append(output, run.matchIdleBots(FreeBucket, provisionAwareMatch, m)...)
	output = append(output, run.matchIdleBots(FreeBucket, basicMatch, m)...)

	return output
}

// assignRequestToWorker updates the information in scheduler pass to reflect the fact that the given request
// (from the given priority) was assigned to a worker.
func (run *schedulerRun) assignRequestToWorker(w WorkerID, request requestNode, priority Priority) {
	delete(run.idleWorkers, w)
	run.jobsUntilThrottled[request.Value().AccountID]--
	run.requestsPerPriority[priority].Remove(request.Element)
}

// newRun initializes a scheduler pass.
func (s *Scheduler) newRun() *schedulerRun {
	// Note: We are using len(s.state.workers) as a capacity hint for this map. In reality,
	// that is the upper bound, and in normal workload (in which fleet is highly utilized) most
	// scheduler passes will have only a few idle workers.
	idleWorkers := make(map[WorkerID]*worker, len(s.state.workers))
	remainingBeforeThrottle := make(map[AccountID]int)
	for aid, ac := range s.config.AccountConfigs {
		if ac.MaxFanout == 0 {
			remainingBeforeThrottle[AccountID(aid)] = math.MaxInt32
		} else {
			remainingBeforeThrottle[AccountID(aid)] = int(ac.MaxFanout)
		}
	}

	for wid, w := range s.state.workers {
		if w.isIdle() {
			idleWorkers[wid] = w
		} else {
			aid := w.runningTask.request.AccountID
			if aid != "" {
				remainingBeforeThrottle[aid]--
			}
		}
	}

	return &schedulerRun{
		idleWorkers:         idleWorkers,
		requestsPerPriority: s.prioritizeRequests(remainingBeforeThrottle),
		jobsUntilThrottled:  remainingBeforeThrottle,
		scheduler:           s,
	}
}

// matchLevel describes whether a request matches a worker and how good of a match it is.
type matchLevel struct {
	// canMatch indicates if the request can run on the worker.
	canMatch bool

	// quality is a heuristic for the quality a match, used to break ties between multiple
	// requests that can match a worker.
	//
	// A higher number is a better quality.
	quality int
}

// matchListItem is an item in a quality-sorted list of request to worker matches.
type matchListItem struct {
	matchLevel

	// request is the request for this item.
	request *TaskRequest

	// node is the node into the original linked list of requests that this
	// item corresponds to.
	node requestNode
}

// matcher is the type for functions that evaluates request to worker matching.
type matcher func(*worker, *TaskRequest) matchLevel

// basicMatch is a matcher function that considers only whether all of the base labels of the
// given request are satisfied by the worker.
//
// The quality heuristic is the number of the base labels in the request (the more, the better).
// This heuristic allows requests that have higher specificity to be preferentially matched to the
// workers that can support them.
func basicMatch(w *worker, r *TaskRequest) matchLevel {
	if w.labels.Contains(r.BaseLabels) {
		quality := len(r.BaseLabels)
		return matchLevel{true, quality}
	}
	return matchLevel{false, 0}
}

// provisionAwareMatch is a matcher function that requires both the base labels and the provisionable
// labels of the request to be satisfied by the worker.
func provisionAwareMatch(w *worker, r *TaskRequest) matchLevel {
	if !w.labels.Contains(r.ProvisionableLabels) {
		return matchLevel{false, 0}
	}
	return basicMatch(w, r)
}

// computeWorkerMatch computes the match level for all given requests against a single worker,
// and returns the matchable requests sorted by match quality.
func computeWorkerMatch(w *worker, requests requestList, mf matcher) []matchListItem {
	matches := make([]matchListItem, 0, requests.Len())
	for current := requests.Head(); current.Element != nil; current = current.Next() {
		m := mf(w, current.Value())
		if m.canMatch {
			matches = append(matches, matchListItem{matchLevel: m, request: current.Value(), node: current})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].quality > matches[j].quality
	})
	return matches
}

// matchIdleBots matches requests with idle workers.
func (run *schedulerRun) matchIdleBots(priority Priority, mf matcher, mSink MetricsSink) []*Assignment {
	var output []*Assignment
	for wid, w := range run.idleWorkers {
		// Try to match.
		candidates := run.requestsPerPriority[priority]
		matches := computeWorkerMatch(w, candidates, mf)
		// select first non-throttled match
		for _, match := range matches {
			// Enforce fanout (except for Freebucket).
			if run.jobsUntilThrottled[match.request.AccountID] <= 0 && priority != FreeBucket {
				continue
			}

			m := &Assignment{
				Type:      AssignmentIdleWorker,
				WorkerID:  wid,
				RequestID: match.request.ID,
				Priority:  priority,
				Time:      run.scheduler.state.lastUpdateTime,
			}
			run.assignRequestToWorker(wid, match.node, priority)
			run.scheduler.state.applyAssignment(m)
			output = append(output, m)
			mSink.AddEvent(
				eventAssigned(match.request, w, run.scheduler.state, run.scheduler.state.lastUpdateTime,
					&metrics.TaskEvent_AssignedDetails{
						Preempting:        false,
						Priority:          int32(priority),
						ProvisionRequired: !w.labels.Contains(match.request.ProvisionableLabels),
					}))
			break
		}

	}
	return output
}

// reprioritizeRunningTasks changes the priority of running tasks by either
// demoting jobs out of the given priority (from level p to level p + 1),
// or by promoting tasks (from any level > p to level p).
//
// Running tasks are demoted if their quota account has too negative a balance
// (Note: a given request may be demoted multiple times, in successive passes,
// from p -> p + 1 -> p + 2 etc if its account has negative balance in multiple
// priority buckets)
//
// Running tasks are promoted if their quota account has a sufficiently positive
// balance and a recharge rate that can sustain them at this level.
func (run *schedulerRun) reprioritizeRunningTasks(priority Priority, mSink MetricsSink) {
	state := run.scheduler.state
	config := run.scheduler.config
	// TODO(akeshet): jobs that are currently running, but have no corresponding account,
	// should be demoted immediately to the FreeBucket (probably their account
	// was deleted while running).
	for accountID, fullBalance := range state.balances {
		// TODO(akeshet): move the body of this loop to own function.
		accountConfig, ok := config.AccountConfigs[string(accountID)]
		if !ok {
			panic(fmt.Sprintf("There was a balance for unknown account %s", accountID))
		}
		balance := fullBalance[priority]
		demote := balance < DemoteThreshold
		promote := balance > PromoteThreshold
		if !demote && !promote {
			continue
		}

		runningAtP := workersAt(state.workers, priority, accountID)

		chargeRate := accountConfig.ChargeRate[priority] - float64(len(runningAtP))

		switch {
		case demote && chargeRate < 0:
			doDemote(state, runningAtP, chargeRate, priority, mSink)
		case promote && chargeRate > 0:
			runningBelowP := workersBelow(state.workers, priority, accountID)
			doPromote(state, runningBelowP, chargeRate, priority, mSink)
		}
	}
}

// TODO(akeshet): Consider unifying doDemote and doPromote somewhat
// to reuse more code.

// doDemote is a helper function used by reprioritizeRunningTasks
// which demotes some jobs (selected from candidates) from priority to priority + 1.
func doDemote(state *state, candidates []*worker, chargeRate float64, priority Priority, mSink MetricsSink) {
	sortAscendingCost(candidates)

	numberToDemote := minInt(len(candidates), int(math.Ceil(-chargeRate)))
	for _, toDemote := range candidates[:numberToDemote] {
		mSink.AddEvent(eventReprioritized(toDemote.runningTask.request, toDemote, state, state.lastUpdateTime,
			&metrics.TaskEvent_ReprioritizedDetails{
				NewPriority: int32(priority) + 1,
				OldPriority: int32(toDemote.runningTask.priority),
			},
		))
		toDemote.runningTask.priority = priority + 1
	}
}

// doPromote is a helper function use by reprioritizeRunningTasks
// which promotes some jobs (selected from candidates) from any level > priority
// to priority.
func doPromote(state *state, candidates []*worker, chargeRate float64, priority Priority, mSink MetricsSink) {
	sortDescendingCost(candidates)

	numberToPromote := minInt(len(candidates), int(math.Ceil(chargeRate)))
	for _, toPromote := range candidates[:numberToPromote] {
		mSink.AddEvent(eventReprioritized(toPromote.runningTask.request, toPromote, state, state.lastUpdateTime,
			&metrics.TaskEvent_ReprioritizedDetails{
				NewPriority: int32(priority) + 1,
				OldPriority: int32(toPromote.runningTask.priority),
			},
		))
		toPromote.runningTask.priority = priority
	}
}

// workersAt is a helper function that returns the workers with a given
// account id and running.
func workersAt(ws map[WorkerID]*worker, priority Priority, accountID AccountID) []*worker {
	ans := make([]*worker, 0, len(ws))
	for _, worker := range ws {
		if !worker.isIdle() &&
			worker.runningTask.request.AccountID == accountID &&
			worker.runningTask.priority == priority {
			ans = append(ans, worker)
		}
	}
	return ans
}

// workersBelow is a helper function that returns the workers with a given
// account id and below a given running.
func workersBelow(ws map[WorkerID]*worker, priority Priority, accountID AccountID) []*worker {
	ans := make([]*worker, 0, len(ws))
	for _, worker := range ws {
		if !worker.isIdle() &&
			worker.runningTask.request.AccountID == accountID &&
			worker.runningTask.priority > priority {
			ans = append(ans, worker)
		}
	}
	return ans
}

// preemptRunningTasks interrupts lower priority already-running tasks, and
// replaces them with higher priority tasks. When doing so, it also reimburses
// the account that had been charged for the task.
func (run *schedulerRun) preemptRunningTasks(priority Priority, mSink MetricsSink) []*Assignment {
	state := run.scheduler.state
	var output []*Assignment
	candidates := make([]*worker, 0, len(state.workers))
	// Accounts that are already running a lower priority job are not
	// permitted to preempt jobs at this priority. This is to prevent a type
	// of thrashing that may occur if an account is unable to promote jobs to
	// this priority (because that would push it over its charge rate)
	// but still has positive quota at this priority.
	bannedAccounts := make(map[AccountID]bool)
	for _, worker := range state.workers {
		if !worker.isIdle() && worker.runningTask.priority > priority {
			candidates = append(candidates, worker)
			bannedAccounts[worker.runningTask.request.AccountID] = true
		}
	}

	sortAscendingCost(candidates)

	for _, worker := range candidates {
		candidateRequests := run.requestsPerPriority[priority]
		matches := computeWorkerMatch(worker, candidateRequests, basicMatch)

		// Select first matching request from an account that is:
		// - non-throttled
		// - non-banned
		// - has sufficient balance to refund cost of preempted job
		for _, m := range matches {
			r := m.request
			if bannedAccounts[r.AccountID] {
				continue
			}
			if run.jobsUntilThrottled[r.AccountID] <= 0 {
				continue
			}
			if !worker.runningTask.cost.Less(state.balances[r.AccountID]) {
				continue
			}
			mut := &Assignment{
				Type:        AssignmentPreemptWorker,
				Priority:    priority,
				RequestID:   m.request.ID,
				TaskToAbort: worker.runningTask.request.ID,
				WorkerID:    worker.ID,
				Time:        state.lastUpdateTime,
			}
			run.assignRequestToWorker(worker.ID, m.node, priority)
			mSink.AddEvent(
				eventAssigned(m.request, worker, state, state.lastUpdateTime,
					&metrics.TaskEvent_AssignedDetails{
						Preempting:     true,
						PreemptionCost: worker.runningTask.cost[:],
						Priority:       int32(priority),
						// TODO(akeshet): Compute this properly.
						ProvisionRequired: !worker.labels.Contains(r.ProvisionableLabels),
					}))
			mSink.AddEvent(
				eventPreempted(worker.runningTask.request, worker, state, state.lastUpdateTime,
					&metrics.TaskEvent_PreemptedDetails{
						PreemptingAccountId: string(m.request.AccountID),
						PreemptingPriority:  int32(priority),
						PreemptingTaskId:    string(m.request.ID),
						Priority:            int32(worker.runningTask.priority),
					}))
			state.applyAssignment(mut)
			output = append(output, mut)
		}
	}
	return output
}

// moveThrottledRequests moves jobs that got throttled at a given prioty level to the FreeBucket priority level
// in the scheduler pass, to give them a second chance to be scheduled if there are any idle workers left
// once the FreeBucket pass is reached.
func (run *schedulerRun) moveThrottledRequests(priority Priority) {
	for current := run.requestsPerPriority[priority].Head(); current.Element != nil; current = current.Next() {
		if run.jobsUntilThrottled[current.Value().AccountID] <= 0 {
			run.requestsPerPriority[FreeBucket].PushBack(current.Value())
			run.requestsPerPriority[priority].Remove(current.Element)
		}
	}
}

// minInt returns the lesser of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
