package scheduler

import (
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/task"
)

// MutAssignRequestIdleWorker represents a state mutation in which a request
// is assigned to an already idle worker.
type MutAssignRequestIdleWorker struct {
	WorkerID  string
	RequestID string
	Priority  int32
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
	WorkerID    string
	NewPriority int32
}

// Mutate applies a mutation to a state, and is part of the types.Mutater
// interface.
func (m *MutChangePriority) Mutate(state *types.State) {
	state.Workers[m.WorkerID].RunningTask.Priority = m.NewPriority
}

// MutPreemptJob represents a state mutation in which a running request
// is interrupted, and replaced by a new request.
type MutPreemptJob struct {
	WorkerID  string
	RequestID string
	Priority  int32
}

// Mutate applies a mutation to a state, and is part of the types.Mutater
// interface.
func (m *MutPreemptJob) Mutate(state *types.State) {
	worker := state.Workers[m.WorkerID]
	cost := worker.RunningTask.Cost
	oldTask := worker.RunningTask
	newTask := state.RequestQueue[m.RequestID]
	creditAccountID := worker.RunningTask.Request.AccountId
	debitAccountID := newTask.AccountId

	newCreditBalance := state.Balances[creditAccountID].Plus(*cost)
	state.Balances[creditAccountID] = &newCreditBalance

	newDebitBalance := state.Balances[debitAccountID].Minus(*cost)
	state.Balances[debitAccountID] = &newDebitBalance
	delete(state.RequestQueue, m.RequestID)
	state.RequestQueue[oldTask.Request.Id] = oldTask.Request
	worker.RunningTask = &task.Running{Cost: cost, Priority: m.Priority, Request: newTask}
}
