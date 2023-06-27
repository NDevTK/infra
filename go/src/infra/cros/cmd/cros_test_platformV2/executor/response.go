// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executor

import "go.chromium.org/chromiumos/config/go/test/api"

type HWRequirement struct {
	Board string
	Model string
	Deps  []string
}

type ProvisionInfo struct {
}

type SuiteMetadata struct {
	HWRequirements []*HWRequirement
	ProvisionInfo  []*ProvisionInfo
	Builds         []string
	Pool           string
}

type TestCase struct {
	name           string
	metadata       *api.TestCaseMetadata
	HWRequirements []*HWRequirement
}

// Each "HW Requirement" will have its own "Testsuite" for its response.
// These will be tracked in list.
type TestPlanResponse struct {
	SuiteMetadata *SuiteMetadata
	TestCases     []*TestCase
}
