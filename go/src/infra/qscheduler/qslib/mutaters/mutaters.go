package mutaters

import (
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/task"
)

// TODO: Turn mutaters code into an independent package called mutaters.

// Mutater is an interface that describes operations that mutate a types.State
type Mutater interface {
	Mutate(state *types.State)
}

// Mutate implements Mutaters.Mutate().
func (m *AssignIdleWorker) Mutate(state *types.State) {
	rt := &task.Run{
		Priority: m.Priority,
		Request:  state.RequestQueue[m.RequestId],
	}
	delete(state.RequestQueue, m.RequestId)
	state.Running = append(state.Running, rt)
	state.Workers[m.WorkerId].RunningTask = rt
}

// Mutate implements Mutaters.Mutate().
func (m *ChangePriority) Mutate(state *types.State) {
	state.Workers[m.WorkerId].RunningTask.Priority = m.NewPriority
}

// Mutate implements Mutaters.Mutate().
func (m *PreemptTask) Mutate(state *types.State) {
	w := state.Workers[m.WorkerId]
	cost := w.RunningTask.Cost
	oT := w.RunningTask
	nT := state.RequestQueue[m.RequestId]
	oAcc := w.RunningTask.Request.AccountId
	nAcc := nT.AccountId

	oBal := state.Balances[oAcc].Plus(*cost)
	state.Balances[oAcc] = &oBal

	nBal := state.Balances[nAcc].Minus(*cost)
	state.Balances[nAcc] = &nBal

	delete(state.RequestQueue, m.RequestId)
	state.RequestQueue[oT.Request.Id] = oT.Request
	w.RunningTask = &task.Run{Cost: cost, Priority: m.Priority, Request: nT}
}
