// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"testing"
)

// TestGetReleaseOrchestratorName tests getReleaseOrchestratorName.
func TestGetReleaseOrchestratorName(t *testing.T) {
	for i, testCase := range []struct {
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
			tryRunBase: tryRunBase{
				branch:  testCase.branch,
				staging: testCase.staging,
			},
		}
		if actual := r.getReleaseOrchestratorName(); actual != testCase.expected {
			t.Errorf("#%d: Incorrect release orch name: got %s; want %s", i, actual, testCase.expected)
		}
	}
}
