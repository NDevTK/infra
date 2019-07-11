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

	"infra/qscheduler/qslib/protos/metrics"
)

type requestListItem struct {
	req     *TaskRequest
	matched bool
}

// requestList implements sort.Interface, and sorts items in ascending
// examinedTime order.
type requestList []*requestListItem

func (r requestList) Len() int {
	return len(r)
}

func (r requestList) Less(i, j int) bool {
	return r[i].req.examinedTime.Before(r[j].req.examinedTime)
}

func (r requestList) Swap(i, j int) {
	temp := r[i]
	r[i] = r[j]
	r[j] = temp
}

// schedulerRun stores values that are used within a single run of the scheduling algorithm.
// Its fields may be mutated during the run, as requests get assigned to workers.
type schedulerRun struct {
	// idleWorkers is a collection of currently idle workers.
	idleWorkers map[WorkerID]*Worker

	// requestsPerPriority is a per-priority list of queued requests, sorted in
	// ascending examinedTime order.
	//
	// It takes into account throttling of any requests whose account was
	// already at the fanout limit when the pass was started, but not newly throttled accounts
	// as a result of newly assigned requests/workers.
	requestsPerPriority [NumPriorities + 1]requestList

	// fanout tracks the number of additional requests that each fanout class may run before reaching
	// its fanout limit and becoming throttled (at which point its other requests are demoted to FreeBucket).
	fanout *fanoutCounter

	scheduler *Scheduler
}

func (run *schedulerRun) Run(e EventSink) []*Assignment {
	var output []*Assignment
	// Proceed through multiple passes of the scheduling algorithm, from highest
	// to lowest priority requests (high priority = low p).
	for p := Priority(0); p < NumPriorities; p++ {
		workerMatches := run.computeIdleWorkerMatches(p)

		// Step 1: Match any requests to idle workers that have matching
		// provisionable labels.
		output = append(output, run.matchIdleBots(p, workerMatches, true, e)...)
		// Step 2: Match request to any remaining idle workers, regardless of
		// provisionable labels.
		output = append(output, run.matchIdleBots(p, workerMatches, false, e)...)
		// Step 3: Demote (out of this level) or promote (into this level) any
		// already running tasks that qualify.
		run.reprioritizeRunningTasks(p, e)
		// Step 4: Preempt any lower priority running tasks.
		if !run.scheduler.config.DisablePreemption {
			output = append(output, run.preemptRunningTasks(p, e)...)
		}
		// Step 5: Give any requests that were throttled in this pass a chance to be scheduled
		// during the final FreeBucket pass.
		run.moveThrottledRequests(p)
	}

	// A final pass matches free jobs (in the FreeBucket) to any remaining
	// idle workers. The reprioritize and preempt stages do not apply here.
	// TODO(akeshet): Consider a final sorting step here, so that FIFO ordering is respected among
	// FreeBucket jobs, including those that were moved the the throttled list during the pass above.
	workerMatches := run.computeIdleWorkerMatches(FreeBucket)
	output = append(output, run.matchIdleBots(FreeBucket, workerMatches, true, e)...)
	output = append(output, run.matchIdleBots(FreeBucket, workerMatches, false, e)...)

	run.updateExaminedTimes()

	return output
}

// assignRequestToWorker updates the information in scheduler pass to reflect the fact that the given request
// (from the given priority) was assigned to a worker.
func (run *schedulerRun) assignRequestToWorker(w WorkerID, item *requestListItem, priority Priority) {
	delete(run.idleWorkers, w)
	run.fanout.count(item.req)
	item.matched = true
}

func (run *schedulerRun) updateExaminedTimes() {
	for _, req := range run.scheduler.state.queuedRequests {
		// A task request was fully examined unless it was throttled due to
		// account fanout limit, for an account with free tasks disabled.
		account, ok := run.scheduler.config.AccountConfigs[string(req.AccountID)]
		if ok && account.DisableFreeTasks && run.isThrottled(req) {
			continue
		}

		req.examinedTime = run.scheduler.state.lastUpdateTime
	}
}

// newRun initializes a scheduler pass.
func (s *Scheduler) newRun() *schedulerRun {
	// Note: We are using len(s.state.workers) as a capacity hint for this map. In reality,
	// that is the upper bound, and in normal workload (in which fleet is highly utilized) most
	// scheduler passes will have only a few idle workers.
	idleWorkers := make(map[WorkerID]*Worker, len(s.state.workers))
	fanoutCounter := newFanoutCounter(s.config)

	for wid, w := range s.state.workers {
		if w.IsIdle() {
			idleWorkers[wid] = w
		} else {
			fanoutCounter.count(w.runningTask.request)
		}
	}

	return &schedulerRun{
		idleWorkers:         idleWorkers,
		requestsPerPriority: s.prioritizeRequests(fanoutCounter),
		fanout:              fanoutCounter,
		scheduler:           s,
	}
}

// matchLevel describes whether a request matches a worker and how good of a match it is.
type matchLevel struct {
	// canMatch indicates if the request can run on the worker.
	canMatch bool

	// provisionMatch indicates if the request's provisionable labels are
	// matched by the worker
	provisionMatch bool

	// quality is a heuristic for the quality a match, used to break ties between multiple
	// requests that can match a worker.
	//
	// A higher number is a better quality.
	quality int
}

// matchListItem is an item in a quality-sorted list of request to worker matches.
type matchListItem struct {
	matchLevel

	item *requestListItem
}

// matcher is the type for functions that evaluates request to worker matching.
type matcher func(*Worker, *TaskRequest) matchLevel

// basicMatch determines whether a request can run on a worker, and the quality
// of the match.
func basicMatch(w *Worker, r *TaskRequest) matchLevel {
	if !w.Labels.HasAll(r.BaseLabels...) {
		return matchLevel{canMatch: false}
	}
	provisionMatch := w.Labels.HasAll(r.ProvisionableLabels...)
	quality := len(r.BaseLabels)
	return matchLevel{
		canMatch:       true,
		quality:        quality,
		provisionMatch: provisionMatch,
	}
}

func (run *schedulerRun) computeIdleWorkerMatches(priority Priority) map[WorkerID][]matchListItem {
	matchesPerWorker := make(map[WorkerID][]matchListItem, len(run.idleWorkers))
	type widAndItem struct {
		wid     WorkerID
		matches []matchListItem
	}
	mChan := make(chan widAndItem, len(run.idleWorkers))
	candidates := run.requestsPerPriority[priority]
	for wid, w := range run.idleWorkers {
		go func(wid WorkerID, w *Worker) {
			matches := computeWorkerMatch(w, candidates)
			mChan <- widAndItem{wid: wid, matches: matches}
		}(wid, w)
	}
	for len(matchesPerWorker) < len(run.idleWorkers) {
		item := <-mChan
		matchesPerWorker[item.wid] = item.matches
	}
	return matchesPerWorker
}

// computeWorkerMatch computes the match level for all given requests against a
// single worker, and returns the matchable requests sorted by match quality.
func computeWorkerMatch(w *Worker, items requestList) []matchListItem {
	var matches []matchListItem
	end := sort.Search(len(items), func(i int) bool {
		return items[i].req.examinedTime.After(w.modifiedTime)
	})
	for _, item := range items[:end] {
		m := basicMatch(w, item.req)
		if m.canMatch {
			matches = append(matches, matchListItem{matchLevel: m, item: item})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].quality > matches[j].quality
	})
	return matches
}

// matchIdleBots matches requests with idle workers.
func (run *schedulerRun) matchIdleBots(priority Priority, matchesPerWorker map[WorkerID][]matchListItem, requireProvisionMatch bool, events EventSink) []*Assignment {
	var output []*Assignment

	for wid, w := range run.idleWorkers {
		matches := matchesPerWorker[wid]
		// select first match that is:
		// - non-throttled match
		// - not already matched
		// - matches provision labels, if necessary
		for _, match := range matches {
			if match.item.matched {
				continue
			}
			if run.shouldSkip(match.item.req, priority) {
				continue
			}
			if requireProvisionMatch && !match.provisionMatch {
				continue
			}

			m := &Assignment{
				Type:      AssignmentIdleWorker,
				WorkerID:  wid,
				RequestID: match.item.req.ID,
				Priority:  priority,
				Time:      run.scheduler.state.lastUpdateTime,
			}
			run.assignRequestToWorker(wid, match.item, priority)
			run.scheduler.state.applyAssignment(m)
			output = append(output, m)
			events.AddEvent(
				eventAssigned(match.item.req, w, run.scheduler.state, run.scheduler.state.lastUpdateTime,
					&metrics.TaskEvent_AssignedDetails{
						Preempting:        false,
						Priority:          int32(priority),
						ProvisionRequired: !w.Labels.HasAll(match.item.req.ProvisionableLabels...),
					}))
			break
		}

	}
	return output
}

func (run *schedulerRun) isThrottled(request *TaskRequest) bool {
	return run.fanout.getRemaining(request) <= 0
}

// shouldSkip computes if the given request should be skipped at the given priority.
func (run *schedulerRun) shouldSkip(request *TaskRequest, priority Priority) bool {
	// Enforce fanout (except for Freebucket).
	if priority != FreeBucket {
		return run.isThrottled(request)
	}

	// Enforce DisableFreeTasks (for FreeBucket).
	if account, ok := run.scheduler.config.AccountConfigs[string(request.AccountID)]; ok {
		return account.DisableFreeTasks
	}

	return false
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
func (run *schedulerRun) reprioritizeRunningTasks(priority Priority, events EventSink) {
	state := run.scheduler.state
	config := run.scheduler.config

	type accountPriority struct {
		p Priority
		a AccountID
	}

	workersAt := make(map[accountPriority][]*Worker)
	for _, worker := range run.scheduler.state.workers {
		if worker.IsIdle() {
			continue
		}
		ap := accountPriority{
			worker.runningTask.priority,
			worker.runningTask.request.AccountID,
		}
		workersAt[ap] = append(workersAt[ap], worker)
	}

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

		runningAtP := workersAt[accountPriority{priority, accountID}]

		chargeRate := accountConfig.ChargeRate[priority] - float32(len(runningAtP))

		switch {
		case demote && chargeRate < 0:
			doDemote(state, runningAtP, chargeRate, priority, events)
		case promote && chargeRate > 0:
			runningBelowP := workersBelow(state.workers, priority, accountID)
			doPromote(state, runningBelowP, chargeRate, priority, events)
		}
	}
}

// TODO(akeshet): Consider unifying doDemote and doPromote somewhat
// to reuse more code.

// doDemote is a helper function used by reprioritizeRunningTasks
// which demotes some jobs (selected from candidates) from priority to priority + 1.
func doDemote(state *state, candidates []*Worker, chargeRate float32, priority Priority, events EventSink) {
	sortAscendingCost(candidates)

	numberToDemote := minInt(len(candidates), ceil(-chargeRate))
	for _, toDemote := range candidates[:numberToDemote] {
		events.AddEvent(eventReprioritized(toDemote.runningTask.request, toDemote, state, state.lastUpdateTime,
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
func doPromote(state *state, candidates []*Worker, chargeRate float32, priority Priority, events EventSink) {
	sortDescendingCost(candidates)

	numberToPromote := minInt(len(candidates), ceil(chargeRate))
	for _, toPromote := range candidates[:numberToPromote] {
		events.AddEvent(eventReprioritized(toPromote.runningTask.request, toPromote, state, state.lastUpdateTime,
			&metrics.TaskEvent_ReprioritizedDetails{
				NewPriority: int32(priority) + 1,
				OldPriority: int32(toPromote.runningTask.priority),
			},
		))
		toPromote.runningTask.priority = priority
	}
}

// workersBelow is a helper function that returns the workers with a given
// account id and below a given running.
func workersBelow(ws map[WorkerID]*Worker, priority Priority, accountID AccountID) []*Worker {
	ans := make([]*Worker, 0, len(ws))
	for _, worker := range ws {
		if !worker.IsIdle() &&
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
func (run *schedulerRun) preemptRunningTasks(priority Priority, events EventSink) []*Assignment {
	state := run.scheduler.state
	var output []*Assignment
	candidates := make([]*Worker, 0, len(state.workers))
	// Accounts that are already running a lower priority job are not
	// permitted to preempt jobs at this priority. This is to prevent a type
	// of thrashing that may occur if an account is unable to promote jobs to
	// this priority (because that would push it over its charge rate)
	// but still has positive quota at this priority.
	bannedAccounts := make(map[AccountID]bool)
	for _, worker := range state.workers {
		if !worker.IsIdle() && worker.runningTask.priority > priority {
			candidates = append(candidates, worker)
			bannedAccounts[worker.runningTask.request.AccountID] = true
		}
	}

	sortAscendingCost(candidates)

	for _, worker := range candidates {
		candidateRequests := run.requestsPerPriority[priority]
		matches := computeWorkerMatch(worker, candidateRequests)

		// Select first matching request from an account that is:
		// - non-throttled
		// - non-banned
		// - not already matched
		// - has sufficient balance to refund cost of preempted job
		for _, m := range matches {
			if m.item.matched {
				continue
			}

			r := m.item.req
			if bannedAccounts[r.AccountID] {
				continue
			}
			if run.fanout.getRemaining(r) <= 0 {
				continue
			}
			if !worker.runningTask.cost.Less(state.balances[r.AccountID]) {
				continue
			}
			mut := &Assignment{
				Type:        AssignmentPreemptWorker,
				Priority:    priority,
				RequestID:   m.item.req.ID,
				TaskToAbort: worker.runningTask.request.ID,
				WorkerID:    worker.ID,
				Time:        state.lastUpdateTime,
			}
			run.assignRequestToWorker(worker.ID, m.item, priority)
			events.AddEvent(
				eventAssigned(m.item.req, worker, state, state.lastUpdateTime,
					&metrics.TaskEvent_AssignedDetails{
						Preempting:        true,
						PreemptionCost:    worker.runningTask.cost[:],
						PreemptedTaskId:   string(worker.runningTask.request.ID),
						Priority:          int32(priority),
						ProvisionRequired: !worker.Labels.HasAll(r.ProvisionableLabels...),
					}))
			events.AddEvent(
				eventPreempted(worker.runningTask.request, worker, state, state.lastUpdateTime,
					&metrics.TaskEvent_PreemptedDetails{
						PreemptingAccountId: string(m.item.req.AccountID),
						PreemptingPriority:  int32(priority),
						PreemptingTaskId:    string(m.item.req.ID),
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
	for _, item := range run.requestsPerPriority[priority] {
		if item.matched {
			continue
		}
		if run.isThrottled(item.req) {
			run.requestsPerPriority[FreeBucket] = append(run.requestsPerPriority[FreeBucket], item)
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

func ceil(val float32) int {
	return int(math.Ceil(float64(val)))
}
