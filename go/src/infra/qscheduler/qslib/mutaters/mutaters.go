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

package mutaters

import (
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/task"
	"infra/qscheduler/qslib/types/vector"
)

// Mutater is an interface that describes operations that mutate a types.State
// TODO: Consider moving this interface definition to scheduler package.
type Mutater interface {
	Mutate(state *types.State)
}

// Mutate implements Mutaters.Mutate().
func (m *AssignIdleWorker) Mutate(state *types.State) {
	rt := &task.Run{
		Priority: m.Priority,
		Request:  state.Requests[m.RequestId],
		Cost:     vector.New(),
	}
	delete(state.Requests, m.RequestId)
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
	nT := state.Requests[m.RequestId]
	oAcc := w.RunningTask.Request.AccountId
	nAcc := nT.AccountId

	oBal := state.Balances[oAcc].Plus(*cost)
	state.Balances[oAcc] = &oBal

	nBal := state.Balances[nAcc].Minus(*cost)
	state.Balances[nAcc] = &nBal

	delete(state.Requests, m.RequestId)
	state.Requests[oT.Request.Id] = oT.Request
	w.RunningTask = &task.Run{Cost: cost, Priority: m.Priority, Request: nT}
}
