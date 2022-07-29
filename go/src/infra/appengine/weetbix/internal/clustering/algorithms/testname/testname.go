// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package testname contains the test name-based clustering algorithm for Weetbix.
package testname

import (
	"crypto/sha256"
	"fmt"
	"strconv"

	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/config/compiledcfg"
)

// AlgorithmVersion is the version of the clustering algorithm. The algorithm
// version should be incremented whenever existing test results may be
// clustered differently (i.e. Cluster(f) returns a different value for some
// f that may have been already ingested).
const AlgorithmVersion = 3

// AlgorithmName is the identifier for the clustering algorithm.
// Weetbix requires all clustering algorithms to have a unique identifier.
// Must match the pattern ^[a-z0-9-.]{1,32}$.
//
// The AlgorithmName must encode the algorithm version, so that each version
// of an algorithm has a different name.
var AlgorithmName = fmt.Sprintf("%sv%v", clustering.TestNameAlgorithmPrefix, AlgorithmVersion)

// Algorithm represents an instance of the test name-based clustering
// algorithm.
type Algorithm struct{}

// Name returns the identifier of the clustering algorithm.
func (a *Algorithm) Name() string {
	return AlgorithmName
}

// clusterLike returns the test name LIKE expression that defines
// the cluster the given test result belongs to.
//
// By default this LIKE expression encodes just the test
// name itself. However, by using rules, projects can configure
// it to mask out parts of the test name (e.g. corresponding
// to test variants).
// "ninja://chrome/test:interactive_ui_tests/ColorSpaceTest.testNullTransform/%"
func clusterLike(config *compiledcfg.ProjectConfig, failure *clustering.Failure) (like string, ok bool) {
	testID := failure.TestID
	for _, r := range config.TestNameRules {
		like, ok := r(testID)
		if ok {
			return like, true
		}
	}
	// No rule matches. Match the test name literally.
	return "", false
}

// clusterKey returns the unhashed key for the cluster. Absent an extremely
// unlikely hash collision, this value is the same for all test results
// in the cluster.
func clusterKey(config *compiledcfg.ProjectConfig, failure *clustering.Failure) string {
	// Get the like expression that defines the cluster.
	key, ok := clusterLike(config, failure)
	if !ok {
		// Fall back to clustering on the exact test name.
		key = failure.TestID
	}
	return key
}

// Cluster clusters the given test failure and returns its cluster ID (if it
// can be clustered) or nil otherwise.
func (a *Algorithm) Cluster(config *compiledcfg.ProjectConfig, failure *clustering.Failure) []byte {
	key := clusterKey(config, failure)

	// Hash the expressionto generate a unique fingerprint.
	h := sha256.Sum256([]byte(key))
	// Take first 16 bytes as the ID. (Risk of collision is
	// so low as to not warrant full 32 bytes.)
	return h[0:16]
}

const bugDescriptionTemplateLike = `This bug is for all test failures with a test name like: %s`
const bugDescriptionTemplateExact = `This bug is for all test failures with the test name: %s`

// ClusterDescription returns a description of the cluster, for use when
// filing bugs, with the help of the given example failure.
func (a *Algorithm) ClusterDescription(config *compiledcfg.ProjectConfig, summary *clustering.ClusterSummary) (*clustering.ClusterDescription, error) {
	// Get the like expression that defines the cluster.
	like, ok := clusterLike(config, &summary.Example)
	if ok {
		return &clustering.ClusterDescription{
			Title:       clustering.EscapeToGraphical(like),
			Description: fmt.Sprintf(bugDescriptionTemplateLike, clustering.EscapeToGraphical(like)),
		}, nil
	} else {
		// No matching clustering rule. Fall back to the exact test name.
		return &clustering.ClusterDescription{
			Title:       clustering.EscapeToGraphical(summary.Example.TestID),
			Description: fmt.Sprintf(bugDescriptionTemplateExact, clustering.EscapeToGraphical(summary.Example.TestID)),
		}, nil
	}
}

// ClusterKey returns the unhashed clustering key which is common
// across all test results in a cluster. For display on the cluster
// page or cluster listing.
func (a *Algorithm) ClusterKey(config *compiledcfg.ProjectConfig, example *clustering.Failure) string {
	key := clusterKey(config, example)
	return clustering.EscapeToGraphical(key)
}

// FailureAssociationRule returns a failure association rule that
// captures the definition of cluster containing the given example.
func (a *Algorithm) FailureAssociationRule(config *compiledcfg.ProjectConfig, example *clustering.Failure) string {
	like, ok := clusterLike(config, example)
	if ok {
		return fmt.Sprintf("test LIKE %s", strconv.QuoteToGraphic(like))
	} else {
		return fmt.Sprintf("test = %s", strconv.QuoteToGraphic(example.TestID))
	}
}
