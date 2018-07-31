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

/*
Package task describes either a queued or running task as part of the quota
scheduler algorithm.
*/
package task

import (
	"sort"
)

// LabelSet represents a set of provisionable labels.
//
// In practice, these will almost always be of size 1, so implement set-like
// semantics with a slice instead of the overkill of using a map.
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

	sort.Strings(a)
	sort.Strings(b)
	for i, aVal := range a {
		if aVal != b[i] {
			return false
		}
	}

	return true
}
