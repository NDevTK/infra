package scheduler

import (
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/task"
)

// Mutater is an interface that represents mutations to State that the
// scheduler may emit.
type Mutater interface {
	Mutate(state *types.State)
}

// Mutate applies a mutation to a state, and is part of the types.Mutater
// interface.
func (m *MutAssignRequestIdleWorker) Mutate(state *types.State) {
	rt := &task.Run{
		Priority: m.Priority,
		Request:  state.RequestQueue[m.RequestId],
	}
	delete(state.RequestQueue, m.RequestId)
	state.Running = append(state.Running, rt)
	state.Workers[m.WorkerId].RunningTask = rt
}

// Mutate applies a mutation to a state, and is part of the types.Mutater
// interface.
func (m *MutChangePriority) Mutate(state *types.State) {
	state.Workers[m.WorkerId].RunningTask.Priority = m.NewPriority
}

// Mutate applies a mutation to a state, and is part of the types.Mutater
// interface.
func (m *MutPreemptJob) Mutate(state *types.State) {
	worker := state.Workers[m.WorkerId]
	cost := worker.RunningTask.Cost
	oldTask := worker.RunningTask
	newTask := state.RequestQueue[m.RequestId]
	creditAccountID := worker.RunningTask.Request.AccountId
	debitAccountID := newTask.AccountId

	newCreditBalance := state.Balances[creditAccountID].Plus(*cost)
	state.Balances[creditAccountID] = &newCreditBalance

	newDebitBalance := state.Balances[debitAccountID].Minus(*cost)
	state.Balances[debitAccountID] = &newDebitBalance
	delete(state.RequestQueue, m.RequestId)
	state.RequestQueue[oldTask.Request.Id] = oldTask.Request
	worker.RunningTask = &task.Run{Cost: cost, Priority: m.Priority, Request: newTask}
}
