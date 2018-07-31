package qslib

import (
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
	"reflect"
	"testing"
)

func assertMutations(t *testing.T, expects []types.Mutater, actual chan types.Mutater) {
	for mut := range actual {
		expect := expects[0]
		expects = expects[1:]
		if !reflect.DeepEqual(mut, expect) {
			t.Errorf("expected: %+v actual: %+v", expect, mut)
		}
	}
	if len(expects) != 0 {
		t.Errorf("fewer than expected muts")
	}
}

func TestMatchWithIdleWorkers(t *testing.T) {
	t.Parallel()
	state := types.State{
		Workers: map[types.WorkerID]*types.Worker{
			"w0": &types.Worker{ID: "w0"},
			"w1": &types.Worker{ID: "w1", Labels: []string{"label1"}},
		},
		RequestQueue: map[task.ID]*task.Request{
			"t1": &task.Request{ID: "t1", AccountID: "a1", Labels: []string{"label1"}},
			"t2": &task.Request{ID: "t2", AccountID: "a1", Labels: []string{"label2"}},
		},
		Balances: map[account.ID]account.Vector{
			"a1": account.Vector{2, 0, 0},
		},
	}

	config := types.Config{
		AccountConfigs: map[account.ID]account.Config{
			"a1": account.Config{},
		},
	}

	expects := []types.Mutater{
		&MutAssignRequestIdleWorker{Priority: 0, RequestID: "t1", WorkerID: "w1"},
		&MutAssignRequestIdleWorker{Priority: 0, RequestID: "t2", WorkerID: "w0"},
	}

	muts := make(chan types.Mutater)
	go QuotaSchedule(&state, &config, muts)
	assertMutations(t, expects, muts)
}

func TestReprioritize(t *testing.T) {
	t.Parallel()
	// Prepare a situation in which one P0 job (out of 2 running) will be
	// demoted, and a separate P2 job will be promoted to P1.
	config := types.Config{
		AccountConfigs: map[account.ID]account.Config{
			"a1": account.Config{ChargeRate: account.Vector{1.5, 1.5}},
		},
	}
	state := types.State{
		Balances: map[account.ID]account.Vector{
			"a1": account.Vector{2 * account.DemoteThreshold, 2 * account.PromoteThreshold, 0},
		},
		Workers: map[types.WorkerID]*types.Worker{
			"w1": &types.Worker{ID: "w1", RunningTask: &task.Running{
				Cost:     account.Vector{1},
				Priority: 0,
				Request:  &task.Request{ID: "t1", AccountID: "a1"},
			},
			},
			"w2": &types.Worker{ID: "w2", RunningTask: &task.Running{
				Priority: 0,
				Request:  &task.Request{ID: "t2", AccountID: "a1"},
			},
			},
			"w3": &types.Worker{
				ID: "w3",
				RunningTask: &task.Running{
					Cost:     account.Vector{1},
					Priority: 2,
					Request:  &task.Request{ID: "t3", AccountID: "a1"},
				},
			},
			"w4": &types.Worker{
				ID: "w4",
				RunningTask: &task.Running{
					Priority: 2,
					Request:  &task.Request{ID: "t4", AccountID: "a1"},
				},
			},
		},
	}

	expects := []types.Mutater{
		&MutChangePriority{NewPriority: 1, WorkerID: "w2"},
		&MutChangePriority{NewPriority: 1, WorkerID: "w3"},
	}

	muts := make(chan types.Mutater)
	go QuotaSchedule(&state, &config, muts)
	assertMutations(t, expects, muts)
}

func TestPreempt(t *testing.T) {
	t.Parallel()
	config := types.Config{
		AccountConfigs: map[account.ID]account.Config{
			"a1": account.Config{},
			"a2": account.Config{},
		},
	}
	state := types.State{
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
		},
	}

	expects := []types.Mutater{
		&MutPreemptJob{Priority: 0, WorkerID: "w1", RequestID: "t1"},
	}

	muts := make(chan types.Mutater)
	go QuotaSchedule(&state, &config, muts)
	assertMutations(t, expects, muts)
}

// TODO: Add tests for mutators.
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
