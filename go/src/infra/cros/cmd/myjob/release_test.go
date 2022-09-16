// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"testing"
)

// TestGetReleaseOrchestratorName tests getReleaseOrchestratorName.
func TestGetReleaseOrchestratorName(t *testing.T) {
	for _, testCase := range []struct {
		staging  bool
		branch   string
		expected string
	}{
		{true, "main", "chromeos/staging/staging-release-main-orchestrator"},
		{false, "main", "chromeos/release/release-main-orchestrator"},
		{true, "release-R106.15054.B", "chromeos/staging/staging-release-R106.15054.B-orchestrator"},
		{false, "release-R106.15054.B", "chromeos/release/release-R106.15054.B-orchestrator"},
	} {
		r := releaseRun{
			myjobRunBase: myjobRunBase{
				branch:  testCase.branch,
				staging: testCase.staging,
			},
		}
		if actual := r.getReleaseOrchestratorName(); actual != testCase.expected {
			t.Errorf("Incorrect release orch name with staging=%v, branch=%s: got %s; want %s", testCase.staging, testCase.branch, actual, testCase.expected)
		}
	}
}
