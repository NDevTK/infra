package scheduler

import (
	"math"
)
import (
	"infra/qscheduler/qslib/mutaters"
	"infra/qscheduler/qslib/priority"
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
	"infra/qscheduler/qslib/types/vector"
)

// QuotaSchedule performs a single round of the quota scheduler algorithm
// on a given state and config, and emits state mutations to its output.
func QuotaSchedule(state *types.State, config *types.Config) <-chan mutaters.Mutater {
	output := make(chan mutaters.Mutater)
	go func() {
		defer close(output)
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

	}()
	return output
}

// matchIdleBotsWithLabels matches requests with idle workers that already
// share all of that request's provisionable labels.
func matchIdleBotsWithLabels(state *types.State, requestsAtP priority.List, output chan<- mutaters.Mutater) {
	for i, request := range requestsAtP {
		if request.Scheduled {
			continue
		}
		for _, worker := range state.Workers {
			if worker.IsIdle() && task.LabelSet(worker.Labels).Equal(request.Request.Labels) {
				m := mutaters.AssignIdleWorker{
					WorkerId:  worker.Id,
					RequestId: request.Request.Id,
					Priority:  request.Priority,
				}
				output <- &m
				m.Mutate(state)
				requestsAtP[i] = priority.Request{Scheduled: true}
				break
			}
		}
	}
}

// matchIdleBots matches requests with any idle workers.
// Returns true if any job was matched.
func matchIdleBots(state *types.State, requestsAtP []priority.Request, output chan<- mutaters.Mutater) (anyMatched bool) {
	anyMatched = false

	i := 0
	idleWorkers := make([]*types.Worker, 0, len(state.Workers))
	for _, worker := range state.Workers {
		if worker.IsIdle() {
			idleWorkers = append(idleWorkers, worker)
		}
	}

	for rI, wI := 0, 0; rI < len(requestsAtP) && wI < len(idleWorkers); {
		request := requestsAtP[rI]
		worker := idleWorkers[wI]
		if request.Scheduled {
			rI++
			continue
		}
		m := mutaters.AssignIdleWorker{
			WorkerId:  worker.Id,
			RequestId: request.Request.Id,
			Priority:  request.Priority,
		}
		output <- &m
		m.Mutate(state)
		requestsAtP[i] = priority.Request{Scheduled: true}
		anyMatched = true
		rI++
		wI++
	}

	return
}

// reprioritizeRunningTasks changes the priority of running tasks by either
// demoting jobs out of the given priority, or promoting them to it. Running
// tasks are demoted if their quota account has too negative a balance, and are
// promoted if their quota account could afford them running at a higher
// priority.
func reprioritizeRunningTasks(state *types.State, config *types.Config, priority int32, output chan<- mutaters.Mutater) {
	for accountID, fullBalance := range state.Balances {
		accountConfig, ok := config.AccountConfigs[accountID]
		if !ok {
			// TODO: consider panic here instead of ignoring missing account.
			continue
		}
		balance := fullBalance.At(priority)
		demote := balance < account.DemoteThreshold
		promote := balance > account.PromoteThreshold
		if !demote && !promote {
			continue
		}

		runningAtP := workersAt(state.Workers, priority, accountID)

		chargeRate := accountConfig.ChargeRate.At(priority) - float64(len(runningAtP))

		switch {
		case demote && chargeRate < 0:
			doDemote(state, runningAtP, chargeRate, priority, output)
		case promote && chargeRate > 0:
			runningBelowP := workersBelow(state.Workers, priority, accountID)
			doPromote(state, runningBelowP, chargeRate, priority, output)
		}
	}
}

// TODO: Consider unifying doDemote and doPromote somewhat
// to reuse more code.

func doDemote(state *types.State, candidates []*types.Worker, chargeRate float64, priority int32, output chan<- mutaters.Mutater) {
	types.SortAscendingCost(candidates)

	numberToDemote := minInt(len(candidates), int(math.Ceil(-chargeRate)))
	for _, toDemote := range candidates[:numberToDemote] {
		mut := &mutaters.ChangePriority{
			NewPriority: priority + 1,
			WorkerId:    toDemote.Id,
		}
		output <- mut
		mut.Mutate(state)
	}
}

func doPromote(state *types.State, candidates []*types.Worker, chargeRate float64, priority int32, output chan<- mutaters.Mutater) {
	// We sort here in decreasing cost order.
	types.SortDescendingCost(candidates)

	numberToPromote := minInt(len(candidates), int(math.Ceil(chargeRate)))

	for _, toPromote := range candidates[:numberToPromote] {
		mut := &mutaters.ChangePriority{
			NewPriority: priority,
			WorkerId:    toPromote.Id,
		}
		output <- mut
		mut.Mutate(state)
	}
}

func workersAt(ws map[string]*types.Worker, priority int32, accountID string) []*types.Worker {
	ans := make([]*types.Worker, 0, len(ws))
	for _, worker := range ws {
		if !worker.IsIdle() &&
			worker.RunningTask.Request.AccountId == accountID &&
			worker.RunningTask.Priority == priority {
			ans = append(ans, worker)
		}
	}
	return ans
}

func workersBelow(ws map[string]*types.Worker, priority int32, accountID string) []*types.Worker {
	ans := make([]*types.Worker, 0, len(ws))
	for _, worker := range ws {
		if !worker.IsIdle() &&
			worker.RunningTask.Request.AccountId == accountID &&
			worker.RunningTask.Priority > priority {
			ans = append(ans, worker)
		}
	}
	return ans
}

// preemptRunningTasks interrupts lower priority already-running tasks, and
// replaces them with higher priority tasks. When doing so, it also reimburses
// the account that had been charged for the task.
func preemptRunningTasks(state *types.State, jobsAtP []priority.Request, priority int32, output chan<- mutaters.Mutater) {
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

	types.SortAscendingCost(candidates)

	for rI, cI := 0, 0; rI < len(jobsAtP) && cI < len(candidates); {
		request := jobsAtP[rI]
		candidate := candidates[cI]
		if request.Scheduled {
			rI++
			continue
		}
		requestAccountID := request.Request.AccountId
		if _, ok := bannedAccounts[requestAccountID]; ok {
			rI++
			continue
		}
		cost := candidate.RunningTask.Cost
		requestAccountBalance, ok := state.Balances[requestAccountID]
		if !ok || requestAccountBalance.Less(*cost) {
			rI++
			continue
		}
		mut := mutaters.PreemptTask{Priority: priority, RequestId: request.Request.Id, WorkerId: candidate.Id}
		output <- &mut
		mut.Mutate(state)
		request.Scheduled = true
		rI++
		cI++
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
