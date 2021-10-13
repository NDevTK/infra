// Package re contains the regular expression clustering algorithm for Weetbix.
//
// This algorithm clusters tests based on their name.
package testname

import (
	"crypto/sha256"
	cpb "infra/appengine/weetbix/internal/clustering/proto"
)

// AlgorithmName is the identifier for the clustering algorithm.
// Weetbix requires all clustering algorithms to have a unique identifier.
// Must match the pattern ^[a-z0-9-.]{1,32}$.
const AlgorithmName = "testname-v0.1"

// Algorithm represents an instance of the regular-expression clustering algorithm.
type Algorithm struct{}

// Name returns the identifier of the clustering algorithm.
func (a *Algorithm) Name() string {
	return AlgorithmName
}

// Cluster clusters the given test failure and returns its cluster ID (if it
// can be clustered) or nil otherwise.
func (a *Algorithm) Cluster(failure *cpb.Failure) []byte {
	id := failure.TestId
	// Hash test ID to generate a unique fingerprint.
	h := sha256.Sum256([]byte(id))
	// Take first 16 bytes as the ID. (Risk of collision is
	// so low as to not warrant full 32 bytes.)
	return h[0:16]
}
