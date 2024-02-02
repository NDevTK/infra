// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package suite

import (
	"go.chromium.org/chromiumos/config/go/test/api"
)

// SuiteSet implements the CentralizedSuite interface for a SuiteSet.
type SuiteSet struct {
	suiteSet *api.SuiteSet
}

// Metadata returns the metadata for a Suite.
func (s *SuiteSet) Metadata() *Metadata {
	return convertMetadata(s.suiteSet.GetMetadata())
}

// ID returns the id of the Suite, if the struct does not
// hold a Suite it returns an empty string.
func (s *SuiteSet) ID() string {
	return s.suiteSet.GetId().GetValue()
}

// Tests returns an empty list since this is not a Suite.
func (s *SuiteSet) Tests() []string {
	return []string{}
}

// Suites returns the IDs of the child Suites of the SuiteSet.
func (s *SuiteSet) Suites() []string {
	suites := make([]string, 0, len(s.suiteSet.GetSuites()))
	for _, suite := range s.suiteSet.GetSuites() {
		suites = append(suites, suite.GetValue())
	}
	return suites
}

// SuiteSets returns the IDs of the child SuiteSet of the SuiteSet.
func (s *SuiteSet) SuiteSets() []string {
	suiteSets := make([]string, 0, len(s.suiteSet.GetSuiteSets()))
	for _, suiteSet := range s.suiteSet.GetSuiteSets() {
		suiteSets = append(suiteSets, suiteSet.GetValue())
	}
	return suiteSets
}

// NewSuiteSet constructs a new SuiteSet from a Protobuf.
func NewSuiteSet(suiteSet *api.SuiteSet) *SuiteSet {
	return &SuiteSet{suiteSet: suiteSet}
}
