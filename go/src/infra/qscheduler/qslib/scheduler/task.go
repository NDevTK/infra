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

package scheduler

import (
	"sort"
	"time"

	"infra/qscheduler/qslib/tutils"

	"go.chromium.org/luci/common/data/stringset"
)

// LabelSet represents a set of provisionable labels.
//
// A "provisionable label" refers to a dimension of a worker that can be
// changed by running a task on it (main example: the version of operating
// system running on that worker, which is also the version-under-test).
//
// In swarming, labels (dimensions) are also used to describe hardware capabilities
// or a worker (e.g. the machine type). These labels are opaque and unknown to
// quotascheduler; quotascheduler expects to run on a pool of interchangeable
// devices.
//
// In practice, this set will almost always be of size 1. And it is convenient
// for the .proto representation of a task to use a slice, because that corresponds
// directly to a repeated string proto field. So, implement set-like
// semantics with a slice instead of the using a map, which is the conventional
// means.
type LabelSet []string

// Equal returns true if and only if a and b are set-wise equal.
func (a LabelSet) Equal(b LabelSet) bool {
	if len(a) != len(b) {
		return false
	}
	// Most LabelSets are of size 1, so make those calculations efficient
	// and simple.
	if len(a) == 1 {
		return a[0] == b[0]
	}

	acopy := make([]string, len(a))
	bcopy := make([]string, len(b))
	copy(acopy, a)
	copy(bcopy, b)

	sort.Strings(acopy)
	sort.Strings(bcopy)
	for i, aVal := range acopy {
		if aVal != bcopy[i] {
			return false
		}
	}

	return true
}

// Contains returns true iff the left set contains all elements of the right set.
func (a LabelSet) Contains(b LabelSet) bool {
	// Most LabelSets are of size 1, so make those calculations efficient and simple.
	if len(a) == len(b) {
		return a.Equal(b)
	}

	s := stringset.NewFromSlice(a...)
	return s.HasAll(b...)
}

// NewRequest creates a new TaskRequest.
func NewRequest(accountID string, labels []string, enqueueTime time.Time) *TaskRequest {
	return &TaskRequest{
		AccountId:   accountID,
		EnqueueTime: tutils.TimestampProto(enqueueTime),
		Labels:      labels,
	}
}

// confirm updates a request's confirmed time (to acknowledge that its state
// is consistent with authoritative source as of this time). The update is
// only applied if it is a forward-in-time update or if the existing time
// was undefined.
func (r *TaskRequest) confirm(t time.Time) {
	if r.ConfirmedTime == nil || tutils.Timestamp(r.ConfirmedTime).Before(t) {
		r.ConfirmedTime = tutils.TimestampProto(t)
	}
}
