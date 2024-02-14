// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package suite

import "fmt"

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

func (s *Suite) Closures(suites map[string]CentralizedSuite) ([]*SuiteClosure, error) {
	return closuresWithParent(suites, s.ID(), s, 0)
}

// Closures takes in map of all known Suites/SuiteSets and generates closure
// relationships to be uploaded to database for efficient queries, only generates
// closures for the CentralizedSuite not for all Suites/SuiteSets in suites arg.
func (s *SuiteSet) Closures(suites map[string]CentralizedSuite) ([]*SuiteClosure, error) {
	return closuresWithParent(suites, s.ID(), s, 0)
}

func closuresWithParent(suites map[string]CentralizedSuite, parent string, s CentralizedSuite, depth int) ([]*SuiteClosure, error) {
	closures := []*SuiteClosure{{
		ID:    parent,
		Child: s.ID(),
		Depth: depth,
	}}
	ids := s.Suites()
	ids = append(ids, s.SuiteSets()...)
	for _, id := range ids {
		subsuite := suites[id]
		if subsuite == nil {
			return nil, fmt.Errorf("unknown suite or suite set: %q", id)
		}
		newClosures, err := closuresWithParent(suites, parent, subsuite, depth+1)
		if err != nil {
			return nil, err
		}
		closures = append(closures, newClosures...)
	}
	return closures, nil
}
