package priority

import (
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
	"reflect"
	"testing"
	"time"
)

func TestPrioritizeOneTaskWithQuota(t *testing.T) {
	t.Parallel()
	request := task.Request{AccountID: "a1", ID: "t1"}
	state := types.State{
		Balances:     map[account.ID]account.Balance{"a1": {1, 0, 0}},
		RequestQueue: map[task.ID]*task.Request{"t1": &request},
	}
	config := types.Config{}
	actual := PrioritizeRequests(&state, &config)
	expected := List([]Request{
		{Priority: 0, Request: &request},
	})
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected: %+v, actual: %+v", expected, actual)
	}
}

func TestPrioritizeOneTaskWithoutQuota(t *testing.T) {
	t.Parallel()
	request := task.Request{ID: "t1"}
	state := types.State{
		RequestQueue: map[task.ID]*task.Request{"t1": &request},
	}
	actual := PrioritizeRequests(&state, &types.Config{})
	expected := List([]Request{
		{Priority: account.FreeBucket, Request: &request},
	})
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected: %+v, actual: %+v", expected, actual)
	}
}

func TestPrioritizeWithEnqueueTimeTieBreaker(t *testing.T) {
	t.Parallel()
	earlier := time.Time{}
	later := earlier.Add(10 * time.Second)
	earlyRequest := task.Request{AccountID: "a1", ID: "t1", EnqueueTime: earlier}
	lateRequest := task.Request{AccountID: "a1", ID: "t2", EnqueueTime: later}
	state := types.State{
		Balances: map[account.ID]account.Balance{"a1": {1, 0, 0}},
		RequestQueue: map[task.ID]*task.Request{
			"t2": &lateRequest,
			"t1": &earlyRequest,
		},
	}
	actual := PrioritizeRequests(&state, &types.Config{})
	expected := List([]Request{
		{Priority: 0, Request: &earlyRequest},
		{Priority: 0, Request: &lateRequest},
	})
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected: %+v, actual: %+v", expected, actual)
	}
}

// TODO: this test doesn't belong in this module, it should move to along
// with the function it is testing.
func TestJobsAtPriority(t *testing.T) {
	t.Parallel()
	a1Req1 := &task.Request{AccountID: "a1", ID: "1"}
	a1Req2 := &task.Request{AccountID: "a1", ID: "2"}
	a1Req3 := &task.Request{AccountID: "a1", ID: "3"}

	a2Req1 := &task.Request{AccountID: "a2", ID: "4"}
	a2Req2 := &task.Request{AccountID: "a2", ID: "5"}

	a1R1 := &task.Running{Priority: 0, Request: a1Req1}
	a1R2 := &task.Running{Priority: 0, Request: a1Req2}
	a1R3 := &task.Running{Priority: 1, Request: a1Req3}

	a2R1 := &task.Running{Priority: 0, Request: a2Req1}
	a2R2 := &task.Running{Priority: account.FreeBucket, Request: a2Req2}

	balances := map[account.ID]account.Balance{
		"a1": {},
		"a2": {},
	}

	running := []*task.Running{a1R1, a1R2, a1R3, a2R1, a2R2}

	actual := countJobsPerPriorityPerAccount(running, &balances)
	expected := map[account.ID]*account.IntVector{
		"a1": {2, 1, 0},
		"a2": {1, 0, 0},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected: %+v, actual: %+v", expected, actual)
	}
}

func TestDemoteJobsBeyondFanout(t *testing.T) {
	t.Parallel()
	config := &types.Config{
		AccountConfigs: map[account.ID]account.Config{
			"a1": {MaxFanout: 3},
			"a2": {},
		},
	}
	running := []*task.Running{
		{Priority: 0, Request: &task.Request{AccountID: "a1", ID: "1"}},
		{Priority: 0, Request: &task.Request{AccountID: "a1", ID: "2"}},
		{Priority: 0, Request: &task.Request{AccountID: "a2", ID: "3"}},
		{Priority: account.FreeBucket, Request: &task.Request{AccountID: "a3", ID: "4"}},
	}
	r1 := task.Request{AccountID: "a1", ID: "5"}
	r2 := task.Request{AccountID: "a1", ID: "6"}
	r3 := task.Request{AccountID: "a2", ID: "7"}
	r4 := task.Request{AccountID: "a3", ID: "8"}
	requestQueue := map[task.ID]*task.Request{
		"5": &r1,
		"6": &r2,
		"7": &r3,
		"8": &r4,
	}
	state := &types.State{
		Balances: map[account.ID]account.Balance{
			"a1": {},
			"a2": {},
		},
		RequestQueue: requestQueue,
		Running:      running,
	}

	prioritizedRequests := []Request{
		{Priority: 0, Request: &r1},
		{Priority: 0, Request: &r2},
		{Priority: 0, Request: &r3},
		{Priority: account.FreeBucket, Request: &r4},
	}

	expected := []Request{
		{Priority: 0, Request: &r1},
		{Priority: account.FreeBucket, Request: &r2},
		{Priority: 0, Request: &r3},
		{Priority: account.FreeBucket, Request: &r4},
	}

	demoteJobsBeyondFanout(prioritizedRequests, state, config)

	actual := prioritizedRequests
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected: %+v, actual: %+v", expected, actual)
	}
}

// Run a thorough test of the full set of prioritization behaviors.
func TestPrioritize(t *testing.T) {
	t.Parallel()
	// Setup common variables.
	a1 := account.ID("a1")
	a2 := account.ID("a2")
	a3 := account.ID("a3")
	a4 := account.ID("a4")
	// A1: Account with P0 quota, fanout limit 3.
	a1Balance := account.Balance{1, 0, 0}
	a1Config := account.Config{MaxFanout: 3}
	// A2: Account with P1 quota, no fanout limit.
	a2Balance := account.Balance{0, 1, 0}
	a2Config := account.Config{}
	// A3: Account with not quota.
	a3Balance := account.Balance{}
	a3Config := account.Config{}
	// A4: Invalid / nonexistant account.
	balances := map[account.ID]account.Balance{
		a1: a1Balance,
		a2: a2Balance,
		a3: a3Balance,
	}

	// 6 Jobs are already running. 2 for A1, 2 for A2, 1 for each of A3, A4
	running1 := task.Running{Priority: 0, Request: &task.Request{AccountID: a1}}
	running2 := task.Running{Priority: 0, Request: &task.Request{AccountID: a1}}
	running3 := task.Running{Priority: 1, Request: &task.Request{AccountID: a2}}
	running4 := task.Running{Priority: 1, Request: &task.Request{AccountID: a2}}
	running5 := task.Running{Priority: 3, Request: &task.Request{AccountID: a3}}
	running6 := task.Running{Priority: 3, Request: &task.Request{AccountID: a4}}
	running := []*task.Running{
		&running1,
		&running2,
		&running3,
		&running4,
		&running5,
		&running6,
	}

	earliestTime := time.Now()
	// 6 Jobs are requested. 3 for A1, 1 for each of the remaining

	// A3's requests are the earliest, and 1 second apart.
	request1 := task.Request{AccountID: a1, EnqueueTime: earliestTime, ID: "1"}
	request2 := task.Request{AccountID: a1, EnqueueTime: earliestTime.Add(time.Second), ID: "2"}
	request3 := task.Request{AccountID: a1, EnqueueTime: earliestTime.Add(2 * time.Second), ID: "3"}
	// The remaining requests are later by 1 second each.
	request4 := task.Request{AccountID: a2, EnqueueTime: earliestTime.Add(3 * time.Second), ID: "4"}
	request5 := task.Request{AccountID: a3, EnqueueTime: earliestTime.Add(4 * time.Second), ID: "5"}
	request6 := task.Request{AccountID: a4, EnqueueTime: earliestTime.Add(5 * time.Second), ID: "6"}

	requests := map[task.ID]*task.Request{
		"1": &request1,
		"2": &request2,
		"3": &request3,
		"4": &request4,
		"5": &request5,
		"6": &request6,
	}

	state := &types.State{
		Balances:     balances,
		Running:      running,
		RequestQueue: requests,
	}

	config := &types.Config{
		AccountConfigs: map[account.ID]account.Config{
			a1: a1Config,
			a2: a2Config,
			a3: a3Config,
		},
	}

	// Expectation:
	expected := List([]Request{
		// A1 gets one additional job at P0, prior to overflowing fanout.
		{Priority: 0, Request: &request1},
		// A2 gets a P1 job.
		{Priority: 1, Request: &request4},
		// Remaining jobs are all in the free bucket, ordered by enqueue time.
		{Priority: account.FreeBucket, Request: &request2},
		{Priority: account.FreeBucket, Request: &request3},
		{Priority: account.FreeBucket, Request: &request5},
		{Priority: account.FreeBucket, Request: &request6},
	})

	actual := PrioritizeRequests(state, config)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected: %+v, actual: %+v", expected, actual)
	}
}

func TestForPriority(t *testing.T) {
	pRequests := List([]Request{
		Request{Priority: 0},
		Request{Priority: 0},
		Request{Priority: 1},
		Request{Priority: 3},
		Request{Priority: 3},
		Request{Priority: 4},
	})

	expecteds := []List{
		pRequests[0:2],
		pRequests[2:3],
		[]Request{},
		pRequests[3:5],
		pRequests[5:6],
		[]Request{},
	}

	for priority := 0; priority < 6; priority++ {
		actual := pRequests.ForPriority(priority)
		expected := expecteds[priority]
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Incorrect p%d slice, expected: %+v actual: %+v",
				priority, expected, actual)
		}
	}
}
