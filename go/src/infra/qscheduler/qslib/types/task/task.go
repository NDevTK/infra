/*
Package task describes either a queued or running task as part of the quota
scheduler algorithm.
*/
package task

import (
	"sort"
)

// LabelSet represents a set of provisionable labels. In practice, these will
// almost always be of size 1, so implement set-like semantics with a go slice
// instead of the overkill of using a map.
type LabelSet []string

// Equals returns true if and only if label sets a and b are set-equal.
func (a LabelSet) Equals(b LabelSet) bool {
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
