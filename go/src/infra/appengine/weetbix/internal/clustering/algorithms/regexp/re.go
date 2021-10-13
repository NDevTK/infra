// Package re contains the regular expression clustering algorithm for Weetbix.
//
// This algorithm removes ips, temp file names, numbers and other such tokens to cluster
// errors together.
package re

import (
	"crypto/sha256"
	"regexp"

	cpb "infra/appengine/weetbix/internal/clustering/proto"
)

// AlgorithmName is the identifier for the clustering algorithm.
// Weetbix requires all clustering algorithms to have a unique identifier.
// Must match the pattern ^[a-z0-9-.]{1,32}$.
const AlgorithmName = "regexp-v0.1"

// Replaces any 1 or more digit numbers, or hex values (often appear in temp file names or prints of pointers)
var clusterExp = regexp.MustCompile(`[0-9]+|[0-9a-fx]{8,}`)

// Algorithm represents an instance of the regular-expression clustering algorithm.
type Algorithm struct{}

// Name returns the identifier of the clustering algorithm.
func (a *Algorithm) Name() string {
	return AlgorithmName
}

// Cluster clusters the given test failure and returns its cluster ID (if it
// can be clustered) or nil otherwise.
func (a *Algorithm) Cluster(failure *cpb.Failure) []byte {
	if failure.FailureReason == nil || failure.FailureReason.PrimaryErrorMessage == "" {
		return nil
	}
	// Replace numbers and hex values.
	id := clusterExp.ReplaceAllString(failure.FailureReason.PrimaryErrorMessage, "0")
	// sha256 hash the resulting string.
	h := sha256.Sum256([]byte(id))
	// Take first 16 bytes as the ID. (Risk of collision is
	// so low as to not warrant full 32 bytes.)
	return h[0:16]
}
