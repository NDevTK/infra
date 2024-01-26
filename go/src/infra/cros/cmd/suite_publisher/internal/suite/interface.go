// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package suite defines interfaces for abstracting the fields of a Suite or SuiteSet.
package suite

import (
	"go.chromium.org/chromiumos/config/go/test/api"
)

// CentralizedSuiteType is an enum for the type of a CentralizedSuite.
type CentralizedSuiteType string

const (
	SuiteType    CentralizedSuiteType = "Suite"
	SuiteSetType CentralizedSuiteType = "SuiteSet"
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

	// Type returns whether the CentralizedSuite holds a Suite or SuiteSet.
	Type() CentralizedSuiteType

	// Tests returns the tests for a Suite or empty list if the Type is SuiteSetType.
	Tests() []string

	// Suites returns the child suites for a SuiteSet or empty list if the Type
	// is SuiteType.
	Suites() []string

	// SuiteSets returns the child suitesets for a SuiteSet or empty list if the Type is SuiteType.
	SuiteSets() []string
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
