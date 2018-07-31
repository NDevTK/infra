// Copyright 2018 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package priority

import (
	"sort"
)
import (
	"github.com/golang/protobuf/ptypes"
)
import (
	"infra/qscheduler/qslib/types"
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
)

// Request represents a request along with the computed priority
// for it (based on quota account balance, max fanout for that account, and
// FIFO ordering).
type Request struct {
	Priority int32
	Request  *task.Request
	// Flag used within scheduler to indicate that a request is already handled.
	// TODO: This doesn't quite fit the abstraction of this package, consider
	// moving this tracking to scheduler package, where it is actually used.
	Scheduled bool
}

// List represents a priority-sorted list of requests.
// TODO: Consider making this a []*Request for slightly less overhead.
type List []Request

// PrioritizeRequests computes the priority of requests from state.RequestsQueue
//
// The computed prority is based on quota account balances, max fanout, and FIFO
// ordering. It returns the prioritized requests sorted in descending order
// (i.e. most important = first).
func PrioritizeRequests(state *types.State, config *types.Config) List {
	// TODO Use container/heap rather than slices to make this
	// function somewhat faster.
	prioritizedRequests := List(make([]Request, 0, len(state.RequestQueue)))
	for _, request := range state.RequestQueue {
		priority := account.FreeBucket
		if accountBalance, ok := state.Balances[request.AccountId]; ok {
			priority = account.BestPriorityFor(*accountBalance)
		}
		prioritizedRequests = append(prioritizedRequests, Request{
			Priority: priority,
			Request:  request,
		})
	}

	// Comparison operator for PrioritizedRequest, which looks
	// at Priority, followed by EnqueueTime in case of tie.
	less := func(i, j int) bool {
		a := prioritizedRequests[i]
		b := prioritizedRequests[j]
		switch {
		case a.Priority < b.Priority:
			return true
		case a.Priority > b.Priority:
			return false
		default:
			// Tiebreaker: enqueue time.
			// Ignore unlikely timestamp parsing error.
			timeA, _ := ptypes.Timestamp(a.Request.EnqueueTime)
			timeB, _ := ptypes.Timestamp(b.Request.EnqueueTime)
			return timeA.Before(timeB)
		}
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
func (s List) ForPriority(priority int32) List {
	start := len(s)
	end := len(s)
	for i := 0; i < len(s); i++ {
		p := s[i].Priority
		if p == priority {
			start = i
			break
		}
		if p > priority {
			return s[i:i]
		}
	}

	for i := start + 1; i < len(s); i++ {
		p := s[i].Priority
		if p > priority {
			end = i
			break
		}
	}

	return s[start:end]
}

// demoteJobsBeyondFanout enforces that no account will have more than that
// account's MaxFanout tasks running concurrently (aside from in the
// FreeBucket).
//
// TODO: Possible optimizations:
//   - Pass in a heap instead of a sorted list, and when demoting a job, demote
//     it and heap fix it. Rather than modify the list and re-sort. This should
//     be a bit faster, because this function is likely to only demote a
//     fraction of jobs in the list.
func demoteJobsBeyondFanout(prioritizedRequests List, state *types.State, config *types.Config) {
	jobsPerAccount := make(map[string]int32)
	for accountID := range state.Balances {
		jobsPerAccount[accountID] = 0
	}
	for _, running := range state.Running {
		if _, ok := jobsPerAccount[running.Request.AccountId]; ok {
			jobsPerAccount[running.Request.AccountId]++
		}
	}

	for i, prioritizedRequest := range prioritizedRequests {
		accountID := prioritizedRequest.Request.AccountId
		if count, ok := jobsPerAccount[accountID]; ok {
			// TODO: Consider panic or assert if this config doesn't exist
			// because that means we are tracking a balance that has no
			// corresponding config.
			if accountConfig, ok := config.AccountConfigs[accountID]; ok {
				maxFanout := accountConfig.MaxFanout
				if maxFanout > 0 && count >= maxFanout {
					prioritizedRequests[i].Priority = account.FreeBucket
				}
			}
			jobsPerAccount[accountID]++
		}
	}
}
