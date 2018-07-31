package priority

import (
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
	"sort"
	"github.com/golang/protobuf/ptypes"
)

// Request represents a request along with the computed priority
// for it (based on quota account balance, max fanout for that account, and
// FIFO ordering).
type Request struct {
	Priority int32
	Request  *task.Request
	Invalid  bool // Flag used within scheduler to indicate that a request is already handled.
}

// List represents a priority-sorted list of requests.
type List []Request

func demoteJobsBeyondFanout(prioritizedRequests List, state *types.State, config *types.Config) {
	jobsPerAccount := make(map[string]int32)
	for accountID := range state.Balances {
		jobsPerAccount[accountID] = 0
	}
	for _, running := range state.Running {
		count, ok := jobsPerAccount[running.Request.AccountId]
		if ok {
			jobsPerAccount[running.Request.AccountId] = count + 1
		}
	}

	for i, prioritizedRequest := range prioritizedRequests {
		accountID := prioritizedRequest.Request.AccountId
		count, ok := jobsPerAccount[accountID]
		if ok {
			accountConfig, ok := config.AccountConfigs[accountID]
			if ok {
				maxFanout := accountConfig.MaxFanout
				if maxFanout > 0 && count >= maxFanout {
					prioritizedRequests[i].Priority = account.FreeBucket
				}
			}
			jobsPerAccount[accountID] = count + 1
		}
	}
}

// PrioritizeRequests computes the priority of requests within
// state.RequestQueue, based on quota account balances, max fanout, and FIFO
// ordering. It returns the prioritized requests sorted in descending order
// (i.e. most important = first).
func PrioritizeRequests(state *types.State, config *types.Config) List {
	// TODO Use container/heap rather than slices to make this
	// function somewhat faster.
	prioritizedRequests := List(make([]Request, len(state.RequestQueue)))
	i := 0
	for _, request := range state.RequestQueue {
		priority := account.FreeBucket
		accountBalance, ok := state.Balances[request.AccountId]
		if ok {
			priority = account.BestPriorityFor(*accountBalance)
		}
		prioritizedRequests[i] = Request{
			Priority: priority,
			Request:  request,
		}
		i++
	}

	// Comparison operator for PrioritizedRequest, which looks
	// at Priority, followed by EnqueueTime in case of tie.
	less := func(i, j int) bool {
		a := prioritizedRequests[i]
		b := prioritizedRequests[j]
		if a.Priority < b.Priority {
			return true
		}
		if a.Priority > b.Priority {
			return false
		}
		// Tiebreaker: enqueue time.
		// Ignore unlikely timestamp parsing error.
		timeA, _ := ptypes.Timestamp(a.Request.EnqueueTime)
		timeB, _ := ptypes.Timestamp(b.Request.EnqueueTime)
		return timeA.Before(timeB)
	}

	sort.SliceStable(prioritizedRequests, less)
	demoteJobsBeyondFanout(prioritizedRequests, state, config)
	sort.SliceStable(prioritizedRequests, less)
	return prioritizedRequests
}

// ForPriority takes an already sorted prioritizedRequests slice, and
// returns the sub slice of it for the given priority
// TODO: Consider turning this into a generator, so that it can iterate only
// once through the list rather than once per priority level.
func (pR List) ForPriority(priority int32) List {
	start := len(pR)
	end := len(pR)
	for i := 0; i < len(pR); i++ {
		p := pR[i].Priority
		if p == priority {
			start = i
			break
		}
		if p > priority {
			return []Request{}
		}
	}

	for i := start + 1; i < len(pR); i++ {
		p := pR[i].Priority
		if p > priority {
			end = i
			break
		}
	}

	return pR[start:end]
}
