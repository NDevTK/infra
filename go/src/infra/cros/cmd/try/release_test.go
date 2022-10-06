// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/cmd"
)

// TestGetReleaseOrchestratorName tests getReleaseOrchestratorName.
func TestGetReleaseOrchestratorName(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func TestValidate_releaseRun(t *testing.T) {
	t.Parallel()
	r := releaseRun{
		tryRunBase: tryRunBase{
			branch:  "release-R106.15054.B",
			staging: false,
		},
	}
	assert.ErrorContains(t, r.validate(), "currently unsupported")

	r = releaseRun{
		tryRunBase: tryRunBase{
			branch:  "release-R106.15054.B",
			staging: false,
		},
		skipPaygen: true,
	}
	assert.ErrorContains(t, r.validate(), "not supported for production")
}

type runTestConfig struct {
	// e.g. "eve"
	buildTargets []string
	// e.g. "staging-eve-release-R106.15054.B"
	expectedChildren []string
	skipPaygen       bool
	dryrun           bool
}

func doTestRun(t *testing.T, tc *runTestConfig) {
	t.Helper()
	propsFile, err := os.CreateTemp("", "input_props")
	defer os.Remove(propsFile.Name())
	assert.NilError(t, err)
	f := &cmd.FakeCommandRunnerMulti{
		CommandRunners: []cmd.FakeCommandRunner{
			fakeAuthInfoRunner("bb", 0),
			fakeAuthInfoRunner("led", 0),
			{
				ExpectedCmd: []string{
					"led",
					"get-builder",
					"chromeos/staging:staging-release-R106.15054.B-orchestrator",
				},
				Stdout: validJSON,
			},
			{
				ExpectedCmd: []string{"bb", "add",
					"chromeos/staging/staging-release-R106.15054.B-orchestrator",
					"-cl", "crrev.com/c/1234567", "-cl", "crrev.com/i/7654321",
					"-p", fmt.Sprintf("@%s", propsFile.Name())},
			},
		},
	}
	r := releaseRun{
		propsFile: propsFile,
		tryRunBase: tryRunBase{
			cmdRunner:    f,
			branch:       "release-R106.15054.B",
			staging:      true,
			patches:      []string{"crrev.com/c/1234567", "crrev.com/i/7654321"},
			buildTargets: tc.buildTargets,
			buildspec:    "gs://chromiumos-manifest-versions/staging/108/15159.0.0.xml",
		},
		useProdTests: true,
	}
	ret := r.Run(nil, nil, nil)
	assert.IntsEqual(t, ret, Success)

	properties, err := readStructFromFile(propsFile.Name())
	assert.NilError(t, err)

	if len(tc.buildTargets) > 0 {
		if len(tc.buildTargets) != len(tc.expectedChildren) {
			t.Fatalf("len(buildTargets) != len(expectedChildren), invalid test")
		}
		child_builds := properties.GetFields()["$chromeos/orch_menu"].GetStructValue().GetFields()["child_builds"].GetListValue().AsSlice()
		assert.StringArrsEqual(t, interfaceSliceToStr(child_builds), tc.expectedChildren)
	} else {
		_, exists := properties.GetFields()["$chromeos/orch_menu"].GetStructValue().GetFields()["child_builds"]
		assert.Assert(t, !exists)
	}

	skipPaygen, exists := properties.GetFields()["$chromeos/orch_menu"].GetStructValue().GetFields()["skip_paygen"]
	if !tc.skipPaygen {
		assert.Assert(t, !exists)
	} else {
		assert.Assert(t, exists && skipPaygen.GetBoolValue())
	}

	manifestInfo := properties.GetFields()["$chromeos/cros_source"].GetStructValue().GetFields()["syncToManifest"].GetStructValue()
	syncToManifest := manifestInfo.GetFields()["manifestGsPath"].GetStringValue()
	assert.StringsEqual(t, r.buildspec, syncToManifest)

	disable_build_plan_pruning := properties.GetFields()["$chromeos/build_plan"].GetStructValue().GetFields()["disable_build_plan_pruning"].GetBoolValue()
	assert.Assert(t, disable_build_plan_pruning)

	use_prod_tests := properties.GetFields()["$chromeos/cros_test_plan"].GetStructValue().GetFields()["use_prod_config"].GetBoolValue()
	assert.Assert(t, use_prod_tests)
}

func TestRun_dryrun(t *testing.T) {
	doTestRun(t, &runTestConfig{
		dryrun: true,
	})
}

func TestRun_noBuildTargets(t *testing.T) {
	doTestRun(t, &runTestConfig{
		skipPaygen: false,
	})
}

func TestRun_buildTargets(t *testing.T) {
	doTestRun(t, &runTestConfig{
		buildTargets:     []string{"eve", "kevin-kernelnext"},
		expectedChildren: []string{"staging-eve-release-R106.15054.B", "staging-kevin-kernelnext-release-R106.15054.B"},
	})
}
