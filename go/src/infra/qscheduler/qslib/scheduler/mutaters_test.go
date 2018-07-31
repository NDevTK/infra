package scheduler

import (
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/task"
	. "infra/qscheduler/qslib/types/vector"
	"testing"
)

func stateForMutTest() *types.State {
	return &types.State{
		Balances: map[string]*Vector{
			"a1": Ref(V{}),
			"a2": Ref(V{1}),
		},
		RequestQueue: map[string]*task.Request{
			"t1": &task.Request{AccountId: "a2", Id: "t1"},
		},
		Workers: map[string]*types.Worker{
			"w1": &types.Worker{
				Id: "w1",
				RunningTask: &task.Running{
					Cost:     Ref(V{.5, .5, .5}),
					Priority: 1,
					Request:  &task.Request{Id: "t2", AccountId: "a1"},
				},
			},
			"w2": &types.Worker{Id: "w2"},
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
	if w2.RunningTask.Request.Id != "t1" {
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
	if state.Workers["w1"].RunningTask.Request.Id != "t1" {
		t.Errorf("incorrect task on worker")
	}
	if state.Workers["w1"].RunningTask.Priority != 0 {
		t.Errorf("wrong priority")
	}
	if !state.Workers["w1"].RunningTask.Cost.Equals(Val(V{.5, .5, .5})) {
		t.Errorf("task has wrong cost")
	}
	if !state.Balances["a2"].Equals(Val(V{.5, -.5, -.5})) {
		t.Errorf("paying account balance incorrect %+v", state.Balances["a2"])
	}
	if !state.Balances["a1"].Equals(Val(V{.5, .5, .5})) {
		t.Errorf("receiving account balance incorrect %+v", state.Balances["a1"])
	}
}
