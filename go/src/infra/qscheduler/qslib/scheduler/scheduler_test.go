package scheduler

import (
	"reflect"
	"testing"
)
import (
	"infra/qscheduler/qslib/mutaters"
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
	. "infra/qscheduler/qslib/types/vector"
)

func assertMutations(t *testing.T, expects []mutaters.Mutater, actual <-chan mutaters.Mutater) {
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
		Workers: map[string]*types.Worker{
			"w0": &types.Worker{Id: "w0"},
			"w1": &types.Worker{Id: "w1", Labels: []string{"label1"}},
		},
		RequestQueue: map[string]*task.Request{
			"t1": &task.Request{Id: "t1", AccountId: "a1", Labels: []string{"label1"}},
			"t2": &task.Request{Id: "t2", AccountId: "a1", Labels: []string{"label2"}},
		},
		Balances: map[string]*Vector{
			"a1": Ref(V{2, 0, 0}),
		},
	}

	config := types.Config{
		AccountConfigs: map[string]*account.Config{
			"a1": account.NewConfig(),
		},
	}

	expects := []mutaters.Mutater{
		&mutaters.AssignIdleWorker{Priority: 0, RequestId: "t1", WorkerId: "w1"},
		&mutaters.AssignIdleWorker{Priority: 0, RequestId: "t2", WorkerId: "w0"},
	}

	muts := QuotaSchedule(&state, &config)
	assertMutations(t, expects, muts)
}

func TestReprioritize(t *testing.T) {
	t.Parallel()
	// Prepare a situation in which one P0 job (out of 2 running) will be
	// demoted, and a separate P2 job will be promoted to P1.
	config := types.Config{
		AccountConfigs: map[string]*account.Config{
			"a1": &account.Config{ChargeRate: Ref(V{1.5, 1.5})},
		},
	}
	state := types.State{
		Balances: map[string]*Vector{
			"a1": Ref(V{2 * account.DemoteThreshold, 2 * account.PromoteThreshold, 0}),
		},
		Workers: map[string]*types.Worker{
			"w1": &types.Worker{Id: "w1", RunningTask: &task.Run{
				Cost:     Ref(V{1}),
				Priority: 0,
				Request:  &task.Request{Id: "t1", AccountId: "a1"},
			},
			},
			"w2": &types.Worker{Id: "w2", RunningTask: &task.Run{
				Priority: 0,
				Request:  &task.Request{Id: "t2", AccountId: "a1"},
				Cost:     Ref(V{}),
			},
			},
			"w3": &types.Worker{
				Id: "w3",
				RunningTask: &task.Run{
					Cost:     Ref(V{1}),
					Priority: 2,
					Request:  &task.Request{Id: "t3", AccountId: "a1"},
				},
			},
			"w4": &types.Worker{
				Id: "w4",
				RunningTask: &task.Run{
					Priority: 2,
					Request:  &task.Request{Id: "t4", AccountId: "a1"},
					Cost:     Ref(V{}),
				},
			},
		},
	}

	expects := []mutaters.Mutater{
		&mutaters.ChangePriority{NewPriority: 1, WorkerId: "w2"},
		&mutaters.ChangePriority{NewPriority: 1, WorkerId: "w3"},
	}

	muts := QuotaSchedule(&state, &config)
	assertMutations(t, expects, muts)
}

func TestPreempt(t *testing.T) {
	t.Parallel()
	config := types.Config{
		AccountConfigs: map[string]*account.Config{
			"a1": account.NewConfig(),
			"a2": account.NewConfig(),
		},
	}
	state := types.State{
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
				RunningTask: &task.Run{
					Cost:     Ref(V{.5, .5, .5}),
					Priority: 1,
					Request:  &task.Request{Id: "t2", AccountId: "a1"},
				},
			},
		},
	}

	expects := []mutaters.Mutater{
		&mutaters.PreemptTask{Priority: 0, WorkerId: "w1", RequestId: "t1"},
	}

	muts := QuotaSchedule(&state, &config)
	assertMutations(t, expects, muts)
}
