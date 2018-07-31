package priority

import (
	"github.com/golang/protobuf/ptypes"
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
	. "infra/qscheduler/qslib/types/vector"
	"reflect"
	"testing"
	"time"
)

func TestPrioritizeOneTaskWithQuota(t *testing.T) {
	t.Parallel()
	request := task.Request{AccountId: "a1", Id: "t1"}
	state := types.State{
		Balances:     map[string]*Vector{"a1": Ref(V{1, 0, 0})},
		RequestQueue: map[string]*task.Request{"t1": &request},
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
	request := task.Request{Id: "t1"}
	state := types.State{
		RequestQueue: map[string]*task.Request{"t1": &request},
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
	earlier := time.Now()
	later := earlier.Add(10 * time.Second)
	earlierTs, err1 := ptypes.TimestampProto(earlier)
	laterTs, err2 := ptypes.TimestampProto(later)
	if err1 != nil || err2 != nil {
		t.Errorf("errors when computing timestamp %+v, %+v", err1, err2)
	}
	earlyRequest := task.Request{AccountId: "a1", Id: "t1", EnqueueTime: earlierTs}
	lateRequest := task.Request{AccountId: "a1", Id: "t2", EnqueueTime: laterTs}
	state := types.State{
		Balances: map[string]*Vector{"a1": Ref(V{1, 0, 0})},
		RequestQueue: map[string]*task.Request{
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

func TestDemoteJobsBeyondFanout(t *testing.T) {
	t.Parallel()
	config := &types.Config{
		AccountConfigs: map[string]*account.Config{
			"a1": {MaxFanout: 3},
			"a2": {},
		},
	}
	running := []*task.Run{
		{Priority: 0, Request: &task.Request{AccountId: "a1", Id: "1"}},
		{Priority: 0, Request: &task.Request{AccountId: "a1", Id: "2"}},
		{Priority: 0, Request: &task.Request{AccountId: "a2", Id: "3"}},
		{Priority: account.FreeBucket, Request: &task.Request{AccountId: "a3", Id: "4"}},
	}
	r1 := task.Request{AccountId: "a1", Id: "5"}
	r2 := task.Request{AccountId: "a1", Id: "6"}
	r3 := task.Request{AccountId: "a2", Id: "7"}
	r4 := task.Request{AccountId: "a3", Id: "8"}
	requestQueue := map[string]*task.Request{
		"5": &r1,
		"6": &r2,
		"7": &r3,
		"8": &r4,
	}
	state := &types.State{
		Balances: map[string]*Vector{
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
	a1 := string("a1")
	a2 := string("a2")
	a3 := string("a3")
	a4 := string("a4")
	// A1: Account with P0 quota, fanout limit 3.``
	a1Balance := Val(V{1, 0, 0})
	a1Config := account.Config{MaxFanout: 3}
	// A2: Account with P1 quota, no fanout limit.
	a2Balance := Val(V{0, 1, 0})
	a2Config := account.Config{}
	// A3: Account with not quota.
	a3Balance := Vector{}
	a3Config := account.Config{}
	// A4: Invalid / nonexistant account.
	balances := map[string]*Vector{
		a1: &a1Balance,
		a2: &a2Balance,
		a3: &a3Balance,
	}

	// 6 Jobs are already running. 2 for A1, 2 for A2, 1 for each of A3, A4
	running1 := task.Run{Priority: 0, Request: &task.Request{AccountId: a1}}
	running2 := task.Run{Priority: 0, Request: &task.Request{AccountId: a1}}
	running3 := task.Run{Priority: 1, Request: &task.Request{AccountId: a2}}
	running4 := task.Run{Priority: 1, Request: &task.Request{AccountId: a2}}
	running5 := task.Run{Priority: 3, Request: &task.Request{AccountId: a3}}
	running6 := task.Run{Priority: 3, Request: &task.Request{AccountId: a4}}
	running := []*task.Run{
		&running1,
		&running2,
		&running3,
		&running4,
		&running5,
		&running6,
	}

	t0 := time.Now()
	t1 := t0.Add(1 * time.Second)
	t2 := t0.Add(2 * time.Second)
	t3 := t0.Add(3 * time.Second)
	t4 := t0.Add(4 * time.Second)
	t5 := t0.Add(5 * time.Second)
	ts0, _ := ptypes.TimestampProto(t0)
	ts1, _ := ptypes.TimestampProto(t1)
	ts2, _ := ptypes.TimestampProto(t2)
	ts3, _ := ptypes.TimestampProto(t3)
	ts4, _ := ptypes.TimestampProto(t4)
	ts5, _ := ptypes.TimestampProto(t5)
	// 6 Jobs are requested. 3 for A1, 1 for each of the remaining

	// A3's requests are the earliest, and 1 second apart.
	request1 := task.Request{AccountId: a1, EnqueueTime: ts0, Id: "1"}
	request2 := task.Request{AccountId: a1, EnqueueTime: ts1, Id: "2"}
	request3 := task.Request{AccountId: a1, EnqueueTime: ts2, Id: "3"}
	// The remaining requests are later by 1 second each.
	request4 := task.Request{AccountId: a2, EnqueueTime: ts3, Id: "4"}
	request5 := task.Request{AccountId: a3, EnqueueTime: ts4, Id: "5"}
	request6 := task.Request{AccountId: a4, EnqueueTime: ts5, Id: "6"}

	requests := map[string]*task.Request{
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
		AccountConfigs: map[string]*account.Config{
			a1: &a1Config,
			a2: &a2Config,
			a3: &a3Config,
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
		actual := pRequests.ForPriority(int32(priority))
		expected := expecteds[priority]
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Incorrect p%d slice, expected: %+v actual: %+v",
				priority, expected, actual)
		}
	}
}
