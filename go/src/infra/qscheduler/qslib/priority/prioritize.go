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

	"github.com/golang/protobuf/ptypes"

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
type List []Request

// PrioritizeRequests computes the priority of requests from state.Requests
//
// The computed prority is based on quota account balances, max fanout, and FIFO
// ordering. It returns the prioritized requests sorted in descending order
// (i.e. most important = first).
func PrioritizeRequests(state *types.State, config *types.Config) List {
	// TODO Use container/heap rather than slices to make this faster.

	// Initial pass: compute priority for each task based purely on account
	// balance.
	list := List(make([]Request, 0, len(state.Requests)))
	for _, req := range state.Requests {
		p := account.FreeBucket
		if accountBalance, ok := state.Balances[req.AccountId]; ok {
			p = account.BestPriorityFor(*accountBalance)
		}
		list = append(list, Request{
			Priority: p,
			Request:  req,
		})
	}

	less := func(i, j int) bool { return requestsListLess(list, i, j) }

	// Sort requests by priority, then demote in priority those that are beyond
	// an account's MaxFanout (because we want to demote the lowest priority jobs
	// that we can), and sort again.
	// TODO: Use a heap instead of a slice here, then use heap fix when demoting
	// a job to avoid needing to re sort the full list.
	sort.SliceStable(list, less)
	demoteTasksBeyondFanout(list, state, config)
	sort.SliceStable(list, less)
	return list
}

// requestListLess is a helper function used when sorting prioritized requests
//
// It compares items at index i and j, using first their priority, then
// their enqueue time as tiebreaker.
func requestsListLess(l List, i, j int) bool {
	a := l[i]
	b := l[j]
	if a.Priority == b.Priority {
		// Tiebreaker: enqueue time.
		// Ignore unlikely timestamp parsing error.
		timeA, _ := ptypes.Timestamp(a.Request.EnqueueTime)
		timeB, _ := ptypes.Timestamp(b.Request.EnqueueTime)
		return timeA.Before(timeB)
	}
	return a.Priority < b.Priority
}

// ForPriority takes an already sorted prioritizedRequests slice, and
// returns the sub slice of it for the given priority
// TODO: Consider turning this into a generator, so that it can iterate only
// once through the list rather than once per priority level.
func (s List) ForPriority(priority int32) List {
	// TODO: Use binary search instead of linear search to find start-end
	// boundaries.
	start := 0
	for ; start < len(s); start++ {
		if s[start].Priority >= priority {
			break
		}
	}
	end := start
	for ; end < len(s); end++ {
		if s[end].Priority > priority {
			break
		}
	}

	return s[start:end]
}

// demoteTasksBeyondFanout enforces that no account will have more than that
// account's MaxFanout tasks running concurrently (aside from in the
// FreeBucket).
//
// TODO: Possible optimizations:
//   - Pass in a heap instead of a sorted list, and when demoting a job, demote
//     it and heap fix it. Rather than modify the list and re-sort. This should
//     be a bit faster, because this function is likely to only demote a
//     fraction of jobs in the list.
func demoteTasksBeyondFanout(prioritizedRequests List, state *types.State, config *types.Config) {
	tasksPerAccount := make(map[string]int32)
	for _, w := range state.Workers {
		if !w.IsIdle() {
			tasksPerAccount[w.RunningTask.Request.AccountId]++
		}
	}

	for i, r := range prioritizedRequests {
		id := r.Request.AccountId
		c := config.AccountConfigs[id]
		// Jobs without a valid accound id / config are already assigned
		// to the free bucket, so ignore them here.
		if c != nil {
			maxFanout := c.MaxFanout

			if maxFanout > 0 && tasksPerAccount[id] >= maxFanout {
				prioritizedRequests[i].Priority = account.FreeBucket
			}
			tasksPerAccount[id]++
		}
	}
}
