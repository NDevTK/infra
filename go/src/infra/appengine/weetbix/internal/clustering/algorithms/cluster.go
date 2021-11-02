// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package algorithms

import (
	"encoding/hex"
	"errors"
	"time"

	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/clustering/algorithms/failurereason"
	"infra/appengine/weetbix/internal/clustering/algorithms/testname"
)

// Algorithm represents the interface that each clustering algorithm must
// implement.
type Algorithm interface {
	// Name returns the identifier of the clustering algorithm.
	Name() string
	// Cluster clusters the given test failure and returns its cluster ID (if
	// it can be clustered) or nil otherwise. THe returned cluster ID must be
	// at most 16 bytes.
	Cluster(failure *clustering.Failure) []byte
	// FailureAssociationRule returns a failure association rule that
	// captures the definition of cluster containing the given example.
	FailureAssociationRule(example *clustering.Failure) string
	// ClusterDisplayName returns a human-readable display name for the
	// cluster containing the given example.
	ClusterDisplayName(example *clustering.Failure) string
	// BugDescription returns a description of the cluster containing the
	// given example, to appear in newly-filed bugs.
	BugDescription(example *clustering.Failure) string
}

// AlgorithmsVersion is the version of the set of algorithms used.
// Changing the set of algorithms below (including add, update or
// deletion of an algorithm) should result in this version being
// incremented.
const AlgorithmsVersion = 1 + failurereason.AlgorithmVersion + testname.AlgorithmVersion

// algorithms is the set of clustering algorithms known to Weetbix.
// When this an algorithm is added or removed from the set,
// or when an algorithm is updated, ensure the AlgorithmsVersion
// above increments.
var algorithms = []Algorithm{
	&failurereason.Algorithm{},
	&testname.Algorithm{},
}

// ClusterResults represents the results of clustering test failures.
type ClusterResults struct {
	// RuleVersion is the version of failure association rules used
	// to cluster test results. This is the Spanner commit timestamp
	// of the most recent failure association rule used in clustering.
	RuleVersion time.Time
	// Clusters each test result is in, one slice of ClusterIDs
	// for each test result.
	Clusters [][]*clustering.ClusterID
}

// Cluster clusters the given test failures using all registered
// clustering algorithms.
func Cluster(failures []*clustering.Failure) *ClusterResults {
	var result [][]*clustering.ClusterID
	for _, f := range failures {
		var ids []*clustering.ClusterID
		for _, a := range algorithms {
			id := a.Cluster(f)
			if id == nil {
				continue
			}
			ids = append(ids, &clustering.ClusterID{
				Algorithm: a.Name(),
				ID:        hex.EncodeToString(id),
			})
		}
		result = append(result, ids)
	}
	return &ClusterResults{
		// TODO(crbug.com/1243174): Set when failure association rules
		// are implemented.
		RuleVersion: time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
		Clusters:    result,
	}
}

// ErrAlgorithmNotExist is returned if a clusterID is passed whose
// clustering algorithm is not supported. This may indicate the algorithm
// is newer or older than the current version.
var ErrAlgorithmNotExist = errors.New("algorithm does not exist")

// ByName returns the algorithm with the given name. If the algorithm
// does not exist, ErrAlgorithmNotExist is returned.
func ByName(algorithm string) (Algorithm, error) {
	for _, a := range algorithms {
		if a.Name() == algorithm {
			return a, nil
		}
	}
	// We may be running old code, or the caller may be asking
	// for an old (version of an) algorithm.
	return nil, ErrAlgorithmNotExist
}
