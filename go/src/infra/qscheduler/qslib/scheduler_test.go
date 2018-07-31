package qslib

import (
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
	"reflect"
	"testing"
)

func TestMatchWithIdleWorkers(t *testing.T) {
	state := types.State{
		Workers: map[types.WorkerID]*types.Worker{
			"w0": &types.Worker{ID: "w0"},
			"w1": &types.Worker{ID: "w1", Labels: []string{"label1"}},
		},
		RequestQueue: map[task.ID]*task.Request{
			"t1": &task.Request{ID: "t1", AccountID: "a1", Labels: []string{"label1"}},
			"t2": &task.Request{ID: "t2", AccountID: "a1", Labels: []string{"label2"}},
		},
		Balances: map[account.ID]account.Balance{
			"a1": account.Balance{2, 0, 0},
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
	i := 0
	for mut := range muts {
		if !reflect.DeepEqual(mut, expects[i]) {
			t.Errorf("expected: %+v actual: %+v", expects[i], mut)
		}
		i++
	}
}
