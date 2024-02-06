// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package closurse exposes a logic that helps generate closure database tables
// for Centralized Suites.
package closures

import (
	"infra/cros/cmd/suite_publisher/internal/utils"

	csuite "go.chromium.org/chromiumos/platform/dev-util/src/chromiumos/test/suite/centralizedsuite"
)

// SuiteClosure holds the information for a row in the closure table
// this is used to publish SuiteSet relationships and allow quick
// easy lookup.
type SuiteClosure struct {
	ID    string
	Child string
	Depth int
}

// Closures takes in map of all known Suites/SuiteSets and generates closure
// relationships to be uploaded to database for efficient queries, only generates
// closures for the CentralizedSuite not for all Suites/SuiteSets in suites arg.
func Closures(s csuite.CentralizedSuite, mappings csuite.Mappings) []*SuiteClosure {
	return closuresWithParent(s, mappings, s.ID(), 0)
}

func closuresWithParent(s csuite.CentralizedSuite, mappings csuite.Mappings, parent string, depth int) []*SuiteClosure {
	closures := []*SuiteClosure{{
		ID:    parent,
		Child: s.ID(),
		Depth: depth,
	}}
	descendants := utils.UnionSets(s.Suites(), s.SuiteSets())
	for id := range descendants {
		closures = append(closures, closuresWithParent(mappings[id], mappings, parent, depth+1)...)
	}
	return closures
}
