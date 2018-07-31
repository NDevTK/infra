package scheduler

import (
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
	"reflect"
	"testing"
)

func stateForMutTest() *types.State {
	return &types.State{
		Balances: map[account.ID]account.Vector{
			"a1": account.Vector{},
			"a2": account.Vector{1},
		},
		RequestQueue: map[task.ID]*task.Request{
			"t1": &task.Request{AccountID: "a2", ID: "t1"},
		},
		Workers: map[types.WorkerID]*types.Worker{
			"w1": &types.Worker{
				ID: "w1",
				RunningTask: &task.Running{
					Cost:     account.Vector{.5, .5, .5},
					Priority: 1,
					Request:  &task.Request{ID: "t2", AccountID: "a1"},
				},
			},
			"w2": &types.Worker{ID: "w2"},
		},
	}
}

func TestMutMatch(t *testing.T) {
	t.Parallel()
	state := stateForMutTest()
	mut := MutAssignRequestIdleWorker{Priority: 1, RequestID: "t1", WorkerID: "w2"}
	mut.Mutate(state)
	w2 := state.Workers["w2"]
	if w2.RunningTask.Priority != 1 {
		t.Errorf("incorrect priority")
	}
	if w2.RunningTask.Request.ID != "t1" {
		t.Errorf("incorect task")
	}
	_, ok := state.RequestQueue["t1"]
	if ok {
		t.Errorf("task remains in queue")
	}
}

func TestMutReprioritize(t *testing.T) {
	t.Parallel()
	state := stateForMutTest()
	mut := MutChangePriority{NewPriority: 2, WorkerID: "w1"}
	mut.Mutate(state)
	if state.Workers["w1"].RunningTask.Priority != 2 {
		t.Errorf("incorrect priority")
	}
}

func TestMutPreempt(t *testing.T) {
	t.Parallel()
	state := stateForMutTest()
	mut := MutPreemptJob{Priority: 0, RequestID: "t1", WorkerID: "w1"}
	mut.Mutate(state)
	if state.Workers["w1"].RunningTask.Request.ID != "t1" {
		t.Errorf("incorrect task on worker")
	}
	if state.Workers["w1"].RunningTask.Priority != 0 {
		t.Errorf("wrong priority")
	}
	if !reflect.DeepEqual(state.Workers["w1"].RunningTask.Cost,
		account.Vector{.5, .5, .5}) {
		t.Errorf("task has wrong cost")
	}
	if !reflect.DeepEqual(state.Balances["a2"], account.Vector{.5, -.5, -.5}) {
		t.Errorf("paying account balance incorrect %+v", state.Balances["a2"])
	}
	if !reflect.DeepEqual(state.Balances["a1"], account.Vector{.5, .5, .5}) {
		t.Errorf("receiving account balance incorrect %+v", state.Balances["a1"])
	}
}
