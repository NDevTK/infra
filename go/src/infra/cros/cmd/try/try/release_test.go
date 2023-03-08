// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package try

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/cmd"
	bb "infra/cros/lib/buildbucket"
)

const (
	validJSON = `{
		"buildbucket": {
			"bbagent_args": {
				"build": {
					"input": {
						"properties": {
							"$chromeos/my_module": {
								"my_prop": 100
							},
							"my_other_prop": 101
						}
					},
					"infra": {
						"buildbucket": {
							"experiment_reasons": {
								"chromeos.cros_artifacts.use_gcloud_storage": 1
							}
						}
					}
				}
			}
		}
	}`
)

func bbAddOutput(bbid string) string {
	return fmt.Sprintf("http://ci.chromium.org/b/%s SCHEDULED ...\n", bbid)
}

// TestGetReleaseOrchestratorName tests getReleaseOrchestratorName.
func TestGetReleaseOrchestratorName(t *testing.T) {
	t.Parallel()
	for i, testCase := range []struct {
		production bool
		branch     string
		expected   string
	}{
		{false, "main", "chromeos/staging-try/staging-release-main-orchestrator"},
		{true, "main", "chromeos/release/release-main-orchestrator"},
		{false, "release-R106.15054.B", "chromeos/staging-try/staging-release-R106.15054.B-orchestrator"},
		{true, "release-R106.15054.B", "chromeos/release/release-R106.15054.B-orchestrator"},
	} {
		r := releaseRun{
			tryRunBase: tryRunBase{
				branch:     testCase.branch,
				production: testCase.production,
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
				production:   !testCase.staging,
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
			branch:     "release-R106.15054.B",
			production: true,
		},
		skipPaygen: true,
	}
	assert.ErrorContains(t, r.validate(), "not supported for production")
}

func TestValidate_stabilizeRun(t *testing.T) {
	t.Parallel()
	r := releaseRun{
		tryRunBase: tryRunBase{
			branch:     "stabilize-15185.B",
			production: false,
		},
	}
	assert.NilError(t, r.validate())

	r = releaseRun{
		tryRunBase: tryRunBase{
			branch:     "stabilize-15185.B",
			production: true,
		},
	}
	assert.NilError(t, r.validate())
}

type runTestConfig struct {
	// Whether to fail the child build check. Only used if buildTargets is set.
	failChildCheck bool
	// e.g. ["crrev.com/c/1234567"]
	patches []string
	// e.g. "eve"
	buildTargets []string
	// e.g. staging-release-R106.15054.B-orchestrator
	expectedOrch string
	// e.g. "staging-eve-release-R106.15054.B"
	expectedChildren []string
	skipPaygen       bool
	production       bool
	dryrun           bool
	branch           string
}

func doTestRun(t *testing.T, tc *runTestConfig) {
	t.Helper()
	propsFile, err := os.CreateTemp("", "input_props")
	defer os.Remove(propsFile.Name())
	assert.NilError(t, err)

	var expectedBucket string
	expectedBuilder := tc.expectedOrch
	if tc.production {
		expectedBucket = "chromeos/release"
	} else {
		expectedBucket = "chromeos/staging-try"
	}

	f := &cmd.FakeCommandRunnerMulti{
		CommandRunners: []cmd.FakeCommandRunner{
			bb.FakeAuthInfoRunner("bb", 0),
			bb.FakeAuthInfoRunner("led", 0),
			{
				ExpectedCmd: []string{
					"led", "auth-info",
				},
				Stdout: "Logged in as sundar@google.com.\n\nOAuth token details:\n...",
			},
		},
	}
	for _, childBuilder := range tc.expectedChildren {
		expectedCmd := []string{
			"led",
			"get-builder",
			fmt.Sprintf("%s:%s", expectedBucket, childBuilder),
		}
		if tc.failChildCheck {
			f.CommandRunners = append(f.CommandRunners,
				cmd.FakeCommandRunner{
					ExpectedCmd: expectedCmd,
					FailCommand: true,
					FailError:   errors.New("return code 1"),
					Stderr:      ("... not found ..."),
				})
		} else {
			f.CommandRunners = append(f.CommandRunners,
				cmd.FakeCommandRunner{
					ExpectedCmd: expectedCmd,
				})
		}
	}
	f.CommandRunners = append(f.CommandRunners,
		cmd.FakeCommandRunner{
			ExpectedCmd: []string{
				"led",
				"get-builder",
				fmt.Sprintf("%s:%s", expectedBucket, expectedBuilder),
			},
			Stdout: validJSON,
		})
	expectedAddCmd := []string{"bb", "add", fmt.Sprintf("%s/%s", expectedBucket, expectedBuilder)}
	expectedAddCmd = append(expectedAddCmd, "-t", "tryjob-launcher:sundar@google.com")
	for _, patch := range tc.patches {
		expectedAddCmd = append(expectedAddCmd, "-cl", patch)
	}
	expectedAddCmd = append(expectedAddCmd, "-p", fmt.Sprintf("@%s", propsFile.Name()))
	if !tc.dryrun {
		f.CommandRunners = append(f.CommandRunners,
			cmd.FakeCommandRunner{
				ExpectedCmd: expectedAddCmd,
				Stdout:      bbAddOutput("12345"),
			},
		)
	}

	r := releaseRun{
		propsFile: propsFile,
		tryRunBase: tryRunBase{
			cmdRunner:            f,
			dryrun:               tc.dryrun,
			branch:               tc.branch,
			production:           tc.production,
			patches:              tc.patches,
			buildTargets:         tc.buildTargets,
			skipProductionPrompt: true,
		},
		useProdTests: true,
	}
	ret := r.Run(nil, nil, nil)
	if tc.failChildCheck {
		assert.IntsEqual(t, ret, CmdError)
		return
	} else {
		assert.IntsEqual(t, ret, Success)
	}

	properties, err := bb.ReadStructFromFile(propsFile.Name())
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

	noPublicBuild, exists := properties.GetFields()["$chromeos/orch_menu"].GetStructValue().GetFields()["schedule_public_build"]
	assert.Assert(t, exists && !noPublicBuild.GetBoolValue())

	disable_build_plan_pruning, exists := properties.GetFields()["$chromeos/build_plan"].GetStructValue().GetFields()["disable_build_plan_pruning"]
	if len(tc.patches) > 0 {
		assert.Assert(t, disable_build_plan_pruning.GetBoolValue())
	} else {
		assert.Assert(t, !exists)
	}

	use_prod_tests := properties.GetFields()["$chromeos/cros_test_plan"].GetStructValue().GetFields()["use_prod_config"].GetBoolValue()
	assert.Assert(t, use_prod_tests)

	supportedBuild, exists := properties.GetFields()["$chromeos/cros_try"].GetStructValue().GetFields()["supported_build"]
	assert.Assert(t, exists && supportedBuild.GetBoolValue())
}

func TestRun_dryrun(t *testing.T) {
	t.Parallel()
	doTestRun(t, &runTestConfig{
		branch:       "release-R106.15054.B",
		dryrun:       true,
		expectedOrch: "staging-release-R106.15054.B-orchestrator",
	})
}

func TestRun_staging_noBuildTargets(t *testing.T) {
	t.Parallel()
	doTestRun(t, &runTestConfig{
		branch:       "release-R106.15054.B",
		skipPaygen:   false,
		expectedOrch: "staging-release-R106.15054.B-orchestrator",
	})
}

func TestRun_staging_buildTargets(t *testing.T) {
	doTestRun(t, &runTestConfig{
		branch:           "release-R106.15054.B",
		buildTargets:     []string{"eve", "kevin-kernelnext"},
		expectedOrch:     "staging-release-R106.15054.B-orchestrator",
		expectedChildren: []string{"staging-eve-release-R106.15054.B", "staging-kevin-kernelnext-release-R106.15054.B"},
	})
}

func TestRun_staging_buildTargets_fail(t *testing.T) {
	doTestRun(t, &runTestConfig{
		failChildCheck:   true,
		branch:           "release-R106.15054.B",
		buildTargets:     []string{"eve", "kevin-kernelnext"},
		expectedOrch:     "staging-release-R106.15054.B-orchestrator",
		expectedChildren: []string{"staging-eve-release-R106.15054.B", "staging-kevin-kernelnext-release-R106.15054.B"},
	})
}

func TestRun_production(t *testing.T) {
	t.Parallel()
	doTestRun(t, &runTestConfig{
		branch:           "release-R106.15054.B",
		production:       true,
		buildTargets:     []string{"eve", "kevin-kernelnext"},
		expectedOrch:     "release-R106.15054.B-orchestrator",
		expectedChildren: []string{"eve-release-R106.15054.B", "kevin-kernelnext-release-R106.15054.B"},
	})
}

func TestRun_stabilize(t *testing.T) {
	t.Parallel()
	doTestRun(t, &runTestConfig{
		branch:           "stabilize-15185.B",
		production:       true,
		buildTargets:     []string{"eve", "kevin-kernelnext"},
		expectedOrch:     "release-stabilize-15185.B-orchestrator",
		expectedChildren: []string{"eve-release-stabilize-15185.B", "kevin-kernelnext-release-stabilize-15185.B"},
	})
}

func TestRun_patches(t *testing.T) {
	t.Parallel()
	doTestRun(t, &runTestConfig{
		branch:       "release-R106.15054.B",
		skipPaygen:   false,
		expectedOrch: "staging-release-R106.15054.B-orchestrator",
		patches:      []string{"crrev.com/c/1234567"},
	})
}
