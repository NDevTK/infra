// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package try

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/cmd"
	"infra/cros/internal/gerrit"
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
	// The output for GetRelatedChanges for this path.
	actualRelatedChanges map[string]map[int][]gerrit.Change
	// Expected patches after including all ancestors.
	expectedPatches []string
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
	if tc.expectedPatches == nil || len(tc.expectedPatches) == 0 {
		tc.expectedPatches = tc.patches
	}
	for _, patch := range tc.expectedPatches {
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
			cmdRunner: f,
			gerritClient: &gerrit.MockClient{
				ExpectedRelatedChanges: tc.actualRelatedChanges,
			},
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
		actualRelatedChanges: map[string]map[int][]gerrit.Change{
			"https://chromium-review.googlesource.com": {1234567: {}},
		},
	})
}

func TestRun_patches_withAncestors(t *testing.T) {
	t.Parallel()
	doTestRun(t, &runTestConfig{
		branch:       "release-R106.15054.B",
		skipPaygen:   false,
		expectedOrch: "staging-release-R106.15054.B-orchestrator",
		patches:      []string{"crrev.com/c/1234567"},
		actualRelatedChanges: map[string]map[int][]gerrit.Change{
			"https://chromium-review.googlesource.com": {
				1234567: {{ChangeNumber: 1234565}, {ChangeNumber: 1234567}, {ChangeNumber: 1234568}, {ChangeNumber: 1234560}},
			},
		},
		expectedPatches: []string{"crrev.com/c/1234560", "crrev.com/c/1234568", "crrev.com/c/1234567"},
	})
}

func TestIncludeAllAncestors(t *testing.T) {
	/*
		Using ["crrev.com/c/4279213"] to includeAncestors should return ["crrev.com/c/4279210", "crrev.com/c/4279211", "crrev.com/c/4279212", "crrev.com/c/4279213"]
		Using ["4279210", "4279210"] is the same as above because they are in the same relation chain
		Using ["crrev.com/c/4279212", "crrev.com/i/5279212"] will provide a list of 6 elements
		  - ["crrev.com/c/4279210", "crrev.com/c/4279211", "crrev.com/c/4279212", "crrev.com/i/5279210", "crrev.com/i/5279211", "crrev.com/i/5279212"] is valid
		  - ["crrev.com/i/5279210", "crrev.com/c/4279210", "crrev.com/c/4279211", "crrev.com/c/4279212", "crrev.com/i/5279211", "crrev.com/i/5279212"] is also valid
		  - ["crrev.com/i/5279211", "crrev.com/i/5279210", "crrev.com/c/4279210", "crrev.com/c/4279211", "crrev.com/c/4279212", "crrev.com/i/5279212"] is not valid
		    crrev.com/i/5279211 is newer than crrev.com/c/4279210 in the chain and this ordering must be maintained in the output.
	*/
	emptyChain := []gerrit.Change{}
	externalChain := []gerrit.Change{
		{ChangeNumber: 4279218},
		{ChangeNumber: 4279217},
		{ChangeNumber: 4279216},
		{ChangeNumber: 4279215},
		{ChangeNumber: 4279214},
		{ChangeNumber: 4279213},
		{ChangeNumber: 4279212},
		{ChangeNumber: 4279211},
		{ChangeNumber: 4279210},
	}
	internalChain := []gerrit.Change{
		{ChangeNumber: 5279218},
		{ChangeNumber: 5279217},
		{ChangeNumber: 5279216},
		{ChangeNumber: 5279215},
		{ChangeNumber: 5279214},
		{ChangeNumber: 5279213},
		{ChangeNumber: 5279212},
		{ChangeNumber: 5279211},
		{ChangeNumber: 5279210},
	}
	externalChangeMap := map[int][]gerrit.Change{
		4279218: externalChain,
		4279212: externalChain,
		4279217: externalChain,
		4279210: externalChain,
		4273260: emptyChain,
	}
	internalChangeMap := map[int][]gerrit.Change{
		5279218: internalChain,
		5279212: internalChain,
		5279217: internalChain,
		5279210: internalChain,
		5273260: emptyChain,
	}
	patchChains := map[string]map[int][]gerrit.Change{
		"https://chromium-review.googlesource.com":        externalChangeMap,
		"https://chrome-internal-review.googlesource.com": internalChangeMap,
	}
	mockClient := &gerrit.MockClient{
		T:                      t,
		ExpectedRelatedChanges: patchChains,
	}
	ctx := context.Background()
	t.Run("GetRelated error", func(t *testing.T) {
		t.Parallel()
		// An error getting related changes for any patch returns an empty list and error.
		patchesWithAncestors, err := includeAllAncestors(ctx, mockClient, []string{"crrev.com/i/5273269"})
		assert.NonNilError(t, err)
		assert.IntsEqual(t, 0, len(patchesWithAncestors))
		patchesWithAncestors, err = includeAllAncestors(ctx, mockClient, []string{"crrev.com/i/5279210", "crrev.com/i/5273269"})
		assert.NonNilError(t, err)
		assert.IntsEqual(t, 0, len(patchesWithAncestors))
	})
	t.Run("GetRelated empty list", func(t *testing.T) {
		t.Parallel()
		// Patches with no related changes return the specified patch itself.
		patchesWithAncestors, err := includeAllAncestors(ctx, mockClient, []string{"crrev.com/c/4273260"})
		assert.NilError(t, err)
		assert.IntsEqual(t, 1, len(patchesWithAncestors))
		assert.StringArrsEqual(t, []string{"crrev.com/c/4273260"}, patchesWithAncestors)
		patchesWithAncestors, _ = includeAllAncestors(ctx, mockClient, []string{"crrev.com/c/4273260", "crrev.com/i/5273260"})
		assert.IntsEqual(t, 2, len(patchesWithAncestors))
		assert.StringArrsEqual(t, []string{"crrev.com/c/4273260", "crrev.com/i/5273260"}, patchesWithAncestors)
	})
	t.Run("CrosTry duplicate input", func(t *testing.T) {
		t.Parallel()
		// Duplicated inputs in the PatchList count as one.
		patchesWithAncestors, _ := includeAllAncestors(ctx, mockClient, []string{"crrev.com/c/4273260", "crrev.com/i/5273260", "crrev.com/c/4273260"})
		assert.IntsEqual(t, 2, len(patchesWithAncestors))
		assert.StringArrsEqual(t, []string{"crrev.com/c/4273260", "crrev.com/i/5273260"}, patchesWithAncestors)
		patchesWithAncestors, _ = includeAllAncestors(ctx, mockClient, []string{"crrev.com/i/5279212", "crrev.com/i/5279210", "crrev.com/i/5279212"})
		assert.IntsEqual(t, 3, len(patchesWithAncestors))
		for index := 8; index > 5; index-- {
			assert.StringsEqual(t, fmt.Sprintf("crrev.com/i/%d", internalChain[index].ChangeNumber), patchesWithAncestors[8-index])
		}
	})
	t.Run("CrosTry single patch", func(t *testing.T) {
		t.Parallel()
		// Only required patches from a chain.
		patchesWithAncestors, _ := includeAllAncestors(ctx, mockClient, []string{"crrev.com/i/5279210"})
		assert.StringArrsEqual(t, []string{"crrev.com/i/5279210"}, patchesWithAncestors)
		patchesWithAncestors, _ = includeAllAncestors(ctx, mockClient, []string{"crrev.com/i/5279218"})
		assert.IntsEqual(t, len(internalChain), len(patchesWithAncestors))
		for index := 8; index >= 0; index-- {
			assert.StringsEqual(t, fmt.Sprintf("crrev.com/i/%d", internalChain[index].ChangeNumber), patchesWithAncestors[8-index])
		}
		patchesWithAncestors, _ = includeAllAncestors(ctx, mockClient, []string{"crrev.com/i/5279212"})
		assert.IntsEqual(t, 3, len(patchesWithAncestors))
		for index := 8; index > 5; index-- {
			assert.StringsEqual(t, fmt.Sprintf("crrev.com/i/%d", internalChain[index].ChangeNumber), patchesWithAncestors[8-index])
		}
		patchesWithAncestors, _ = includeAllAncestors(ctx, mockClient, []string{"crrev.com/i/5279212", "crrev.com/i/5279210"})
		assert.IntsEqual(t, 3, len(patchesWithAncestors))
		for index := 8; index > 5; index-- {
			assert.StringsEqual(t, fmt.Sprintf("crrev.com/i/%d", internalChain[index].ChangeNumber), patchesWithAncestors[8-index])
		}
		patchesWithAncestors, _ = includeAllAncestors(ctx, mockClient, []string{"crrev.com/i/5279212", "crrev.com/i/5279217"})
		assert.IntsEqual(t, 8, len(patchesWithAncestors))
		for index := 8; index > 0; index-- {
			assert.StringsEqual(t, fmt.Sprintf("crrev.com/i/%d", internalChain[index].ChangeNumber), patchesWithAncestors[8-index])
		}
	})
	t.Run("CrosTry patch ordering", func(t *testing.T) {
		t.Parallel()
		// Maintaining patch ordering in relation chain.
		patchesWithAncestors, _ := includeAllAncestors(ctx, mockClient, []string{"crrev.com/i/5279212", "crrev.com/c/4279217", "crrev.com/c/4273260", "crrev.com/i/5273260", "crrev.com/i/5279217"})
		// Patches "crrev.com/c/4279211" and "crrev.com/i/5279211" will each have 8 related changes in the output.
		// Patches "crrev.com/c/4273260" and "crrev.com/i/5273260" have no related changes so they are each included once.
		// Patch "crrev.com/i/5279216" counts towards patches needed for "crrev.com/i/5279211".
		assert.IntsEqual(t, 8+8+2, len(patchesWithAncestors))
		// The maps are used to make sure patches are ordered according to the relation chain.
		expectedFromInternal := make(map[string]int)
		expectedFromExternal := make(map[string]int)
		for index := 8; index >= 0; index-- {
			expectedFromInternal[fmt.Sprintf("crrev.com/i/%d", internalChain[index].ChangeNumber)] = 8 - index
			expectedFromExternal[fmt.Sprintf("crrev.com/c/%d", externalChain[index].ChangeNumber)] = 8 - index
		}
		lastInternalVisited := 0
		lastExternalVisited := 0
		// These flags mark "crrev.com/i/5273260" and "crrev.com/c/4273260" in the output.
		var expectSingleInternal bool
		var expectSingleExternal bool
		for _, patch := range patchesWithAncestors {
			if patch == "crrev.com/i/5273260" {
				expectSingleInternal = true
				continue
			}
			if patch == "crrev.com/c/4273260" {
				expectSingleExternal = true
				continue
			}
			if strings.Contains(patch, "crrev.com/i/") {
				if expectedFromInternal[patch] < lastInternalVisited {
					t.Errorf("Unexpected order for patch %s", patch)
				} else {
					// The index of this patch in the chain has been visited.
					lastInternalVisited = expectedFromInternal[patch]
				}
			} else {
				if expectedFromExternal[patch] < lastExternalVisited {
					t.Errorf("Unexpected order for patch %s", patch)
				} else {
					lastExternalVisited = expectedFromExternal[patch]
				}
			}
		}
		assert.Assert(t, expectSingleInternal)
		assert.Assert(t, expectSingleExternal)
	})
}
