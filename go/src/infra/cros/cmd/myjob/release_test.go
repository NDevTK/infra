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
		expected string
	}{
		{true, "chromeos/staging/staging-release-main-orchestrator"},
		{false, "chromeos/release/release-main-orchestrator"},
	} {
		r := releaseRun{myjobRunBase: myjobRunBase{staging: testCase.staging}}
		if actual := r.getReleaseOrchestratorName(); actual != testCase.expected {
			t.Errorf("Incorrect release orch name with staging=%v: got %s; want %s", testCase.staging, actual, testCase.expected)
		}
	}
}
