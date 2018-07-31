package task

import (
	"infra/qscheduler/qslib/types/account"
	"sort"
	"time"
)

// ID is an opaque globally unique identifier for a task request.
type ID string

// LabelSet represents a set of provisionable labels. In practice, these will
// almost always be of size 1, so implement set-like semantics with a go slice
// instead of the overkill of using a map.
type LabelSet struct {
	Strings []string
}

// Equals returns true if and only if label sets a and b are set-equal.
// Most LabelSets are of size 1, so make those calculations efficient
// and simple.
func (a LabelSet) Equals(b LabelSet) bool {
	if len(a.Strings) != len(b.Strings) {
		return false
	}
	if len(a.Strings) == 1 {
		return a.Strings[0] == b.Strings[0]
	}

	sort.Strings(a.Strings)
	sort.Strings(b.Strings)
	for i, aVal := range a.Strings {
		if aVal != b.Strings[i] {
			return false
		}
	}

	return true
}

// Request represents a requested task in the queue, and refers to the
// quota account to run it against. Note: this representation intentionally
// excludes most of the details of a true Swarming task request.
type Request struct {
	ID          ID
	AccountID   account.ID
	EnqueueTime time.Time
	Labels      LabelSet // The set of Provisionable Labels for this task.
}

// Running represents a task that has been assigned to a worker and is
// now running.
type Running struct {
	Cost     account.Vector // The total cost that has been spent on this RunningTask.
	Request  *Request       // The request that this running task corresponds to.
	Priority int            // The current priority level of the running task.
}
