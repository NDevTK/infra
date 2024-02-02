// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package suite defines interfaces for abstracting the fields of a Suite or SuiteSet.
package suite

import (
	"go.chromium.org/chromiumos/config/go/test/api"
)

type Metadata struct {
	BugComponent string
	Owners       []string
	Criteria     string
}

// CentralizedSuite is an interface that allows generic
// access to the fields of a Suite or SuiteSet.
type CentralizedSuite interface {
	// Metadata returns the metadata for a Suite or SuiteSet.
	Metadata() *Metadata

	// ID returns the id of the Suite or SuiteSet, if the struct does not
	// hold a Suite or SuiteSet it returns an empty string.
	ID() string

	// Tests returns the tests for a Suite or empty list if the interface contains
	// a SuiteSet.
	Tests() []string

	// Suites returns the child suites for a SuiteSet or empty list if the interface
	// contains a Suite.
	Suites() []string

	// SuiteSets returns the child suitesets for a SuiteSet or empty listif the interface
	// contains a Suite.
	SuiteSets() []string

	// Closures takes in map of all known Suites/SuiteSets and generates closure
	// relationships to be uploaded to database for efficient queries, only generates
	// closures for the CentralizedSuite not for all Suites/SuiteSets in suites arg.
	Closures(suites map[string]CentralizedSuite) []*SuiteClosure
}

func convertMetadata(metadata *api.Metadata) *Metadata {
	owners := []string{}
	for _, owner := range metadata.GetOwners() {
		owners = append(owners, owner.GetEmail())
	}
	return &Metadata{
		BugComponent: metadata.GetBugComponent().GetValue(),
		Owners:       owners,
		Criteria:     metadata.GetCriteria().GetValue(),
	}
}
