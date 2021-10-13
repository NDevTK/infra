package algorithms

import (
	"infra/appengine/weetbix/internal/clustering"
	re "infra/appengine/weetbix/internal/clustering/algorithms/regexp"
	"infra/appengine/weetbix/internal/clustering/algorithms/testname"
	cpb "infra/appengine/weetbix/internal/clustering/proto"
	"time"
)

// Algorithm represents the interface that each clustering algorithm must
// implement.
type Algorithm interface {
	Name() string
	Cluster(failure *cpb.Failure) []byte
}

// AlgorithmsVersion is the version of the set of algorithms used.
// Changing the set of algorithms below should result in this
// version being incremented. Changes to a clustering algorithm
// (that result in new cluster IDs being output for existing test
// results) should result in this version being incremented and the
// algorithm name being changed.
const AlgorithmsVersion = 1

// algs is the set of clustering algorithms known to Weetbix.
// When this set is changed or any individual algorithm is changed,
// bump the AlgorithmsVersion above.
var algs = []Algorithm{
	&re.Algorithm{},
	&testname.Algorithm{},
}

// ClusterResults represents the results of clustering test failures.
type ClusterResults struct {
	// RuleVersion is the version of failure association rules used
	// to cluster test results.
	RuleVersion time.Time
	// Clusters each test result is in, one slice of ClusterRefs
	// for each test result.
	Clusters [][]*clustering.ClusterRef
}

// Cluster clusters the given test failures using all registered
// clustering algorithms.
func Cluster(failures []*cpb.Failure) *ClusterResults {
	var result [][]*clustering.ClusterRef
	for _, f := range failures {
		var refs []*clustering.ClusterRef
		for _, a := range algs {
			id := a.Cluster(f)
			if id == nil {
				continue
			}
			refs = append(refs, &clustering.ClusterRef{
				Algorithm: a.Name(),
				ID:        id,
			})
		}
		result = append(result, refs)
	}
	return &ClusterResults{
		// TODO(crbug.com/1243174): Set when failure association rules
		// are implemented.
		RuleVersion: time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
		Clusters:    result,
	}
}
