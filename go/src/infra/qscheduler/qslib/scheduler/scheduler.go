package scheduler

import (
	"infra/qscheduler/qslib/priority"
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
	"infra/qscheduler/qslib/types/vector"
	"math"
	"sort"
)

// QuotaSchedule performs a single round of the quota scheduler algorithm
// on a given state and config, and emits state mutations to its output.
func QuotaSchedule(state *types.State, config *types.Config, output chan Mutater) {
	list := priority.PrioritizeRequests(state, config)
	botsMightBeIdle := true
	for priority := 0; priority < vector.NumPriorities; priority++ {
		p := int32(priority)
		jobsAtP := list.ForPriority(p)
		if botsMightBeIdle {
			matchIdleBotsWithLabels(state, jobsAtP, output)
			botsMightBeIdle = matchIdleBots(state, jobsAtP, output)
		}
		reprioritizeRunningTasks(state, config, p, output)
		preemptRunningTasks(state, jobsAtP, p, output)
	}

	if botsMightBeIdle {
		freeJobs := list.ForPriority(account.FreeBucket)
		matchIdleBotsWithLabels(state, freeJobs, output)
		matchIdleBots(state, freeJobs, output)
	}

	close(output)
}

// matchIdleBotsWithLabels matches requests with idle workers that already
// share all of that request's provisionable labels.
func matchIdleBotsWithLabels(state *types.State, jobsAtP priority.List, output chan Mutater) {
	for i, request := range jobsAtP {
		if request.Invalid {
			continue
		}
		for _, worker := range state.Workers {
			if worker.IsIdle() && task.LabelSet(worker.Labels).Equal(request.Request.Labels) {
				m := MutAssignRequestIdleWorker{
					WorkerId:  worker.Id,
					RequestId: request.Request.Id,
					Priority:  request.Priority,
				}
				output <- &m
				m.Mutate(state)
				jobsAtP[i] = priority.Request{Invalid: true}
				break
			}
		}
	}
}

// matchIdleBotsWithLabels matches requests with any idle workers.
// Returns true if any job was matched.
func matchIdleBots(state *types.State, jobsAtP []priority.Request, output chan Mutater) (anyMatched bool) {
	anyMatched = false
	for i, request := range jobsAtP {
		if request.Invalid {
			continue
		}

		// TODO: Replace this O(Jobs) * O(Workers) loop with a single
		// O(Jobs + Workers) loop.
		for _, worker := range state.Workers {
			if worker.IsIdle() {
				m := MutAssignRequestIdleWorker{
					WorkerId:  worker.Id,
					RequestId: request.Request.Id,
					Priority:  request.Priority,
				}
				output <- &m
				m.Mutate(state)
				jobsAtP[i] = priority.Request{Invalid: true}
				anyMatched = true
				break
			}
		}
	}
	return
}

// reprioritizeRunningTasks changes the priority of running tasks by either
// demoting jobs out of the given priority, or promoting them to it. Running
// tasks are demoted if their quota account has too negative a balance, and are
// promoted if their quota account could afford them running at a higher
// priority.
func reprioritizeRunningTasks(state *types.State, config *types.Config, priority int32, output chan Mutater) {
	for accountID, fullBalance := range state.Balances {
		accountConfig, ok := config.AccountConfigs[accountID]
		if !ok {
			// This should not be possible, but guard against it anyway.
			continue
		}
		balance := fullBalance.At(priority)
		demote := balance < account.DemoteThreshold
		promote := balance > account.PromoteThreshold
		if !demote && !promote {
			continue
		}

		runningAtP := make([]*types.Worker, 0, len(state.Workers))
		for _, worker := range state.Workers {
			if !worker.IsIdle() &&
				worker.RunningTask.Request.AccountId == accountID &&
				worker.RunningTask.Priority == priority {
				runningAtP = append(runningAtP, worker)
			}
		}

		chargeRate := accountConfig.ChargeRate.At(priority) - float64(len(runningAtP))
		// TODO: Consider unifying the promote and demote pathways somewhat
		// to reuse more code.
		if demote && chargeRate < 0 {
			candidates := runningAtP
			less := func(i, j int) bool {
				return candidates[i].RunningTask.Cost.Less(*candidates[j].RunningTask.Cost)
			}
			sort.SliceStable(candidates, less)

			numberToDemote := minInt(len(candidates), int(math.Ceil(-chargeRate)))
			for _, toDemote := range candidates[:numberToDemote] {
				mut := &MutChangePriority{
					NewPriority: priority + 1,
					WorkerId:    toDemote.Id,
				}
				output <- mut
				mut.Mutate(state)
			}
		} else if promote && chargeRate > 0 {
			candidates := make([]*types.Worker, 0, len(state.Workers))
			for _, worker := range state.Workers {
				if !worker.IsIdle() &&
					worker.RunningTask.Request.AccountId == accountID &&
					worker.RunningTask.Priority > priority {
					candidates = append(candidates, worker)
				}
			}
			// Note: we sort here in decreasing cost order, so the less function
			// is the inverse of the one for the demote logic above.
			less := func(i, j int) bool {
				return candidates[j].RunningTask.Cost.Less(*candidates[i].RunningTask.Cost)
			}
			sort.SliceStable(candidates, less)

			numberToPromote := minInt(len(candidates), int(math.Ceil(chargeRate)))

			for _, toPromote := range candidates[:numberToPromote] {
				mut := &MutChangePriority{
					NewPriority: priority,
					WorkerId:    toPromote.Id,
				}
				output <- mut
				mut.Mutate(state)
			}
		}
	}
}

// preemptRunningTasks interrupts lower priority already-running tasks, and
// replaces hem with higher priority tasks. When doing so, it also reimburses
// the account that had been charged for the task.
func preemptRunningTasks(state *types.State, jobsAtP []priority.Request, priority int32, output chan Mutater) {
	candidates := make([]*types.Worker, 0, len(state.Workers))
	// Accounts that are already running a lower priority job are not
	// permitted to preempt jobs at this priority. This is to prevent a type
	// of thrashing that may occur if an account is unable to promote jobs to
	// this priority (because that would push it over its charge rate)
	// but still has positive quota at this priority.
	bannedAccounts := make(map[string]bool)
	for _, worker := range state.Workers {
		if !worker.IsIdle() && worker.RunningTask.Priority > priority {
			candidates = append(candidates, worker)
			bannedAccounts[worker.RunningTask.Request.AccountId] = true
		}
	}
	less := func(i, j int) bool {
		return candidates[i].RunningTask.Cost.Less(*candidates[j].RunningTask.Cost)
	}
	sort.SliceStable(candidates, less)

	requestIndex := 0
	candidateIndex := 0
	for requestIndex < len(jobsAtP) && candidateIndex < len(candidates) {
		request := jobsAtP[requestIndex]
		candidate := candidates[candidateIndex]
		if request.Invalid {
			requestIndex++
			continue
		}
		requestAccountID := request.Request.AccountId
		_, ok := bannedAccounts[requestAccountID]
		if ok {
			requestIndex++
			continue
		}
		cost := candidate.RunningTask.Cost
		requestAccountBalance, ok := state.Balances[requestAccountID]
		if !ok || requestAccountBalance.Less(*cost) {
			// Insufficient account balance to pay for preemption. Proceed
			// to next request candidate.
			requestIndex++
			continue
		}
		mut := MutPreemptJob{Priority: priority, RequestId: request.Request.Id, WorkerId: candidate.Id}
		output <- &mut
		mut.Mutate(state)
		request.Invalid = true
		requestIndex++
		candidateIndex++
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
