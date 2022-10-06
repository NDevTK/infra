// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"reflect"
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

func TestGetReleaseBuilderNames(t *testing.T) {
	for i, testCase := range []struct {
		staging      bool
		branch       string
		buildTargets []string
		expected     []string
	}{
		{false, "main", []string{"eve", "kevin-kernelnext"}, []string{"eve-release-main", "kevin-kernelnext-release-main"}},
		{true, "main", []string{"eve", "kevin-kernelnext"}, []string{"staging-eve-release-main", "staging-kevin-kernelnext-release-main"}},
		{false, "release-R106.15054.B", []string{"eve", "kevin-kernelnext"}, []string{"eve-release-R106.15054.B", "kevin-kernelnext-release-R106.15054.B"}},
		{true, "release-R106.15054.B", []string{"eve", "kevin-kernelnext"}, []string{"staging-eve-release-R106.15054.B", "staging-kevin-kernelnext-release-R106.15054.B"}},
	} {
		r := releaseRun{
			tryRunBase: tryRunBase{
				branch:       testCase.branch,
				staging:      testCase.staging,
				buildTargets: testCase.buildTargets,
			},
		}
		if actual := r.getReleaseBuilderNames(); !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("#%d: Incorrect release builder names: got %s; want %s", i, actual, testCase.expected)
		}
	}
}
