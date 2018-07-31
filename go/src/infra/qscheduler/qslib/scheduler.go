package qslib

import (
	"infra/qscheduler/qslib/priority"
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
	"math"
	"sort"
)

// QuotaSchedule performs a single round of the quota scheduler algorithm
// on a given state and config, and emits state mutations to its output.
func QuotaSchedule(state *types.State, config *types.Config, output chan types.Mutater) {
	list := priority.PrioritizeRequests(state, config)
	for priority := 0; priority < account.NumPriorities; priority++ {
		jobsAtP := list.ForPriority(priority)

		matchIdleBotsWithLabels(state, jobsAtP, output)
		matchIdleBots(state, jobsAtP, output)
		reprioritizeRunningTasks(state, config, priority, output)
		preemptRunningTasks(state, jobsAtP, output)
	}
	close(output)
}

// MutAssignRequestIdleWorker represents a state mutation in which a request
// is assigned to an already idle worker.
type MutAssignRequestIdleWorker struct {
	WorkerID  types.WorkerID
	RequestID task.ID
	Priority  int
}

// Mutate applies a mutation to a state, and is part of the types.Mutater
// interface.
func (m *MutAssignRequestIdleWorker) Mutate(state *types.State) {
	rt := &task.Running{
		Priority: m.Priority,
		Request:  state.RequestQueue[m.RequestID],
	}
	delete(state.RequestQueue, m.RequestID)
	state.Running = append(state.Running, rt)
	state.Workers[m.WorkerID].RunningTask = rt
}

// MutChangePriority represents a state mutation in which a running request
// has its priority changed.
type MutChangePriority struct {
	WorkerID    types.WorkerID
	NewPriority int
}

// Mutate applies a mutation to a state, and is part of the types.Mutater
// interface.
func (m *MutChangePriority) Mutate(state *types.State) {
	state.Workers[m.WorkerID].RunningTask.Priority = m.NewPriority
}

// MutPreemptJob represents a state mutation in which a running request
// is interrupted, and replaced by a new request.
type MutPreemptJob struct {
	WorkerID  types.WorkerID
	RequestID task.ID
	Priority  int
}

// Mutate applies a mutation to a state, and is part of the types.Mutater
// interface.
func (m *MutPreemptJob) Mutate(state *types.State) {
	// TODO: Implement me.
}

// matchIdleBotsWithLabels matches requests with idle workers that already
// share all of that request's provisionable labels.
func matchIdleBotsWithLabels(state *types.State, jobsAtP priority.List, output chan types.Mutater) {
	for i, request := range jobsAtP {
		if request.Invalid {
			continue
		}
		for _, worker := range state.Workers {
			if worker.IsIdle() && worker.Labels.Equals(request.Request.Labels) {
				m := MutAssignRequestIdleWorker{
					WorkerID:  worker.ID,
					RequestID: request.Request.ID,
					Priority:  request.Priority,
				}
				output <- &m
				m.Mutate(state)
				jobsAtP[i] = priority.Request{Invalid: true}
			}
		}
	}
}

// matchIdleBotsWithLabels matches requests with any idle workers.
func matchIdleBots(state *types.State, jobsAtP []priority.Request, output chan types.Mutater) {
	for i, request := range jobsAtP {
		if request.Invalid {
			continue
		}

		// TODO: Replace this O(Jobs) * O(Workers) loop with a single
		// O(Jobs + Workers) loop.
		for _, worker := range state.Workers {
			if worker.IsIdle() {
				m := MutAssignRequestIdleWorker{
					WorkerID:  worker.ID,
					RequestID: request.Request.ID,
					Priority:  request.Priority,
				}
				output <- &m
				m.Mutate(state)
				jobsAtP[i] = priority.Request{Invalid: true}
			}

		}
	}
}

// reprioritizeRunningTasks changes the priority of running tasks by either
// demoting jobs out of the given priority, or promoting them to it. Running
// tasks are demoted if their quota account has too negative a balance, and are
// promoted if their quota account could afford them running at a higher
// priority.
func reprioritizeRunningTasks(state *types.State, config *types.Config, priority int, output chan types.Mutater) {
	for accountID, fullBalance := range state.Balances {
		accountConfig, ok := config.AccountConfigs[accountID]
		if !ok {
			// This should not be possible, but guard against it anyway.
			continue
		}
		balance := fullBalance[priority]
		demote := balance < account.DemoteThreshold
		promote := balance > account.PromoteThreshold
		if !demote && !promote {
			continue
		}

		// TODO: Consider unifying the promote and demote pathways somewhat
		// to reuse more code.
		if demote {
			candidates := make([]*types.Worker, 0, len(state.Workers))
			for _, worker := range state.Workers {
				if !worker.IsIdle() && worker.RunningTask.Request.AccountID == accountID {
					candidates = append(candidates, worker)
				}
			}
			less := func(i, j int) bool {
				return candidates[i].RunningTask.Cost.Less(candidates[j].RunningTask.Cost)
			}
			sort.SliceStable(candidates, less)

			drainRate := float64(len(candidates)) - accountConfig.ChargeRate[priority]
			numberToDemote := minInt(len(candidates), int(math.Ceil(drainRate)))

			for _, toDemote := range candidates[:numberToDemote] {
				output <- &MutChangePriority{
					NewPriority: priority + 1,
					WorkerID:    toDemote.ID,
				}
			}
		} else if promote {
			// TODO: Implement me.
			// The idea is similar to above, except that we promote the most costly
			// currently-lower-priority jobs to this priority bucket.
		}
	}
}

// preemptRunningTasks interrupts lower priority already-running tasks, and
// replaces hem with higher priority tasks. When doing so, it also reimburses
// the account that had been charged for the task.
func preemptRunningTasks(state *types.State, jobsAtP []priority.Request, output chan types.Mutater) {
	// TODO: Implementme.
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
