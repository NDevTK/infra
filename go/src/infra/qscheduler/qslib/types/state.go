package types

import (
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
)

// State represents the overall state of a quota scheduler worker pool,
// account set, and task queue. This is represented separately from
// configuration information. The state is expected to be updated frequently,
// on each scheduler tick.
type State struct {
	RequestQueue map[task.ID]*task.Request      // Requests that are waiting to be assigned to a worker.
	Running      []*task.Running                // Requests that are running on a worker.
	Balances     map[account.ID]account.Balance // Balance of all quota accounts for this pool.
	Workers      map[WorkerID]*Worker           // Workers that may run tasks, and their states.
}

// Mutater is an interface that represents mutations to State that the
// scheduler may emit.
type Mutater interface {
	Mutate(state *State)
}
