package scheduler

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
