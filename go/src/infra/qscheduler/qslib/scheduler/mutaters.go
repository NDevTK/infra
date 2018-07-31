package scheduler

import (
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/task"
)

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
	worker := state.Workers[m.WorkerID]
	cost := worker.RunningTask.Cost
	oldTask := worker.RunningTask
	newTask := state.RequestQueue[m.RequestID]
	creditAccountID := worker.RunningTask.Request.AccountID
	debitAccountID := newTask.AccountID

	state.Balances[creditAccountID] = state.Balances[creditAccountID].Plus(cost)
	state.Balances[debitAccountID] = state.Balances[debitAccountID].Minus(cost)
	delete(state.RequestQueue, m.RequestID)
	state.RequestQueue[oldTask.Request.ID] = oldTask.Request
	worker.RunningTask = &task.Running{Cost: cost, Priority: m.Priority, Request: newTask}
}
