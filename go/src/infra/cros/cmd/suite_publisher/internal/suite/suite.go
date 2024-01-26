// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package suite

import (
	"go.chromium.org/chromiumos/config/go/test/api"
)

// Suite implements the CentralizedSuite interface for a Suite.
type Suite struct {
	suite *api.Suite
}

// Metadata returns the metadata for a Suite.
func (s *Suite) Metadata() *Metadata {
	return convertMetadata(s.suite.GetMetadata())
}

// ID returns the id of the Suite, if the struct does not
// hold a Suite it returns an empty string.
func (s *Suite) ID() string {
	return s.suite.GetId().GetValue()
}

// Type returns whether the CentralizedSuite holds a Suite.
func (*Suite) Type() CentralizedSuiteType {
	return SuiteType
}

// Tests returns the tests for the Suite.
func (s *Suite) Tests() []string {
	tests := make([]string, 0, len(s.suite.GetTests()))
	for _, test := range s.suite.GetTests() {
		tests = append(tests, test.GetValue())
	}
	return tests
}

// Suites returns an empty list since this is not a SuiteSet.
func (s *Suite) Suites() []string {
	return nil
}

// SuiteSets returns an empty list since this is not a SuiteSet.
func (s *Suite) SuiteSets() []string {
	return nil
}

// NewSuite returns constructs a suite object from an api.Suite Protobuf.
func NewSuite(suite *api.Suite) *Suite {
	return &Suite{suite: suite}
}
