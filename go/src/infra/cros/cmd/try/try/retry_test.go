// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package try

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/cmd"
	bb "infra/cros/lib/buildbucket"

	pb "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	bbpb "go.chromium.org/luci/buildbucket/proto"
)

func TestValidate_retryRim(t *testing.T) {
	t.Parallel()
	r := retryRun{}
	assert.ErrorContains(t, r.validate(), "--bbid")

	r = retryRun{
		originalBBID: "123",
	}
	assert.NilError(t, r.validate())
}

const (
	retryTestGoodJSON = `{
	"id": "8794230068334833057",
	"builder": {
		"project": "chromeos",
		"bucket": "staging",
		"builder": "staging-release-main-orchestrator"
	},
	"status": "SUCCESS",
	"input": {
		"properties": {
			"recipe": "orchestrator",
			"input_prop": 102
		}
	},
	"output": {
		"properties": {
			"retry_summary": {
				"CREATE_BUILDSPEC": "SUCCESS",
				"RUN_CHILDREN": "SUCCESS"
			},
			"child_builds": [
				"8794230068334833058",
				"8794230068334833059"
			]
		}
	}
}`

	successfulChildJSON = `{
		"id": "8794230068334833058",
		"builder": {
			"project": "chromeos",
			"bucket": "staging",
			"builder": "staging-eve-release-main"
		},
		"status": "SUCCESS",
		"input": {
			"properties": {
				"recipe": "build_release",
				"input_prop": 102
			}
		},
		"output": {
			"properties": {
				"retry_summary": {
					"COLLECT_SIGNING": "SUCCESS",
					"DEBUG_SYMBOLS": "SUCCESS",
					"EBUILD_TESTS": "SUCCESS",
					"PAYGEN": "SUCCESS",
					"PUSH_IMAGES": "SUCCESS",
					"STAGE_ARTIFACTS": "SUCCESS"
				}
			}
		}
	}`

	failedChildJSON = `{
		"id": "8794230068334833059",
		"builder": {
			"project": "chromeos",
			"bucket": "staging",
			"builder": "staging-zork-release-main"
		},
		"status": "FAILURE",
		"input": {
			"properties": {
				"recipe": "build_release",
				"input_prop": 102
			}
		},
		"output": {
			"properties": {
				"retry_summary": {
					"COLLECT_SIGNING": "FAILURE",
					"DEBUG_SYMBOLS": "FAILURE",
					"PUSH_IMAGES": "SUCCESS",
					"STAGE_ARTIFACTS": "SUCCESS"
				}
			}
		}
	}`

	emptyRetrySummaryJSON = `{
		"id": "8794230068334833059",
		"builder": {
			"project": "chromeos",
			"bucket": "staging",
			"builder": "staging-zork-release-main"
		},
		"status": "FAILURE",
		"input": {
			"properties": {
				"recipe": "build_release",
				"input_prop": 102
			}
		},
		"output": {
			"properties": {
				"retry_summary": {
				}
			}
		}
	}`
)

func stripNewlines(s string) string {
	return strings.ReplaceAll(s, "\n", "")
}

type retryTestConfig struct {
	dryrun bool
}

func doOrchestratorRetryTestRun(t *testing.T, tc *retryTestConfig) {
	t.Helper()
	propsFile, err := os.CreateTemp("", "input_props")
	defer os.Remove(propsFile.Name())
	assert.NilError(t, err)

	bbid := "8794230068334833057"
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
			{
				ExpectedCmd: []string{"bb", "get", bbid, "-p", "-json"},
				Stdout:      stripNewlines(retryTestGoodJSON),
			},
			{
				ExpectedCmd: []string{"bb", "get", "8794230068334833058", "-p", "-json"},
				Stdout:      stripNewlines(successfulChildJSON),
			},
			{
				ExpectedCmd: []string{"bb", "get", "8794230068334833059", "-p", "-json"},
				Stdout:      stripNewlines(failedChildJSON),
			},
		},
	}
	expectedBucket := "chromeos/staging"
	expectedBuilder := "staging-release-main-orchestrator"
	expectedAddCmd := []string{"bb", "add", fmt.Sprintf("%s/%s", expectedBucket, expectedBuilder)}
	expectedAddCmd = append(expectedAddCmd, "-t", "tryjob-launcher:sundar@google.com")
	expectedAddCmd = append(expectedAddCmd, "-p", fmt.Sprintf("@%s", propsFile.Name()))
	if !tc.dryrun {
		f.CommandRunners = append(f.CommandRunners,
			cmd.FakeCommandRunner{
				ExpectedCmd: expectedAddCmd,
				Stdout:      bbAddOutput(bbid),
			},
		)
	}

	r := retryRun{
		propsFile:    propsFile,
		originalBBID: bbid,
		tryRunBase: tryRunBase{
			cmdRunner: f,
			dryrun:    tc.dryrun,
		},
	}
	ret := r.Run(nil, nil, nil)
	assert.IntsEqual(t, ret, Success)

	properties, err := bb.ReadStructFromFile(propsFile.Name())
	assert.NilError(t, err)

	checkpointProps := properties.GetFields()["$chromeos/checkpoint"].GetStructValue()

	assert.Assert(t, checkpointProps.GetFields()["retry"].GetBoolValue())

	signingProps := properties.GetFields()["$chromeos/signing"].GetStructValue()
	assert.Assert(t, signingProps.GetFields()["ignore_already_exists_errors"].GetBoolValue())

	originalBuildBBID := checkpointProps.GetFields()["original_build_bbid"].GetStringValue()
	assert.StringsEqual(t, originalBuildBBID, bbid)

	execSteps := checkpointProps.GetFields()["exec_steps"].GetStructValue().GetFields()["steps"].GetListValue().AsSlice()
	assert.IntsEqual(t, len(execSteps), 1)
	assert.IntsEqual(t, int(execSteps[0].(float64)), int(pb.RetryStep_RUN_FAILED_CHILDREN.Number()))

	builderExecSteps := checkpointProps.GetFields()["builder_exec_steps"].GetStructValue()
	_, exists := builderExecSteps.GetFields()["staging-eve-release-main"]
	assert.Assert(t, !exists)

	zorkExecSteps := builderExecSteps.GetFields()["staging-zork-release-main"].GetStructValue().GetFields()["steps"].GetListValue().AsSlice()
	assert.IntsEqual(t, len(zorkExecSteps), 1)
	assert.IntsEqual(t, int(zorkExecSteps[0].(float64)), int(pb.RetryStep_DEBUG_SYMBOLS.Number()))

	supportedBuild, exists := properties.GetFields()["$chromeos/cros_try"].GetStructValue().GetFields()["supported_build"]
	assert.Assert(t, exists && supportedBuild.GetBoolValue())
}

func TestRetry_dryRun(t *testing.T) {
	t.Parallel()
	doOrchestratorRetryTestRun(t, &retryTestConfig{
		dryrun: true,
	})
}
func TestRetry_fullRun(t *testing.T) {
	t.Parallel()
	doOrchestratorRetryTestRun(t, &retryTestConfig{
		dryrun: false,
	})
}

type childRetryTestConfig struct {
	dryrun           bool
	testNoRun        bool
	bbid             string
	builderName      string
	builderJSON      string
	expectedExecStep pb.RetryStep
	expectError      bool
	paygenRetry      bool
}

func doChildRetryTestRun(t *testing.T, tc *childRetryTestConfig) {
	t.Helper()
	propsFile, err := os.CreateTemp("", "input_props")
	defer os.Remove(propsFile.Name())
	assert.NilError(t, err)

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
			{
				ExpectedCmd: []string{"bb", "get", tc.bbid, "-p", "-json"},
				Stdout:      tc.builderJSON,
			},
		},
	}
	expectedBucket := "chromeos/staging"
	expectedBuilder := tc.builderName
	expectedAddCmd := []string{"bb", "add", fmt.Sprintf("%s/%s", expectedBucket, expectedBuilder)}
	expectedAddCmd = append(expectedAddCmd, "-t", "tryjob-launcher:sundar@google.com")
	expectedAddCmd = append(expectedAddCmd, "-p", fmt.Sprintf("@%s", propsFile.Name()))
	if !tc.dryrun && !tc.testNoRun {
		f.CommandRunners = append(f.CommandRunners,
			cmd.FakeCommandRunner{
				ExpectedCmd: expectedAddCmd,
				Stdout:      bbAddOutput(tc.bbid),
			},
		)
	}

	r := retryRun{
		propsFile:    propsFile,
		originalBBID: tc.bbid,
		paygenRetry:  tc.paygenRetry,
		tryRunBase: tryRunBase{
			cmdRunner: f,
			dryrun:    tc.dryrun,
		},
	}
	ret := r.Run(nil, nil, nil)
	if !tc.expectError {
		assert.IntsEqual(t, ret, Success)
	} else {
		assert.IntsNotEqual(t, ret, Success)
		return
	}

	if tc.testNoRun {
		return
	}

	properties, err := bb.ReadStructFromFile(propsFile.Name())
	assert.NilError(t, err)

	checkpointProps := properties.GetFields()["$chromeos/checkpoint"].GetStructValue()

	assert.Assert(t, checkpointProps.GetFields()["retry"].GetBoolValue())

	originalBuildBBID := checkpointProps.GetFields()["original_build_bbid"].GetStringValue()
	assert.StringsEqual(t, originalBuildBBID, tc.bbid)

	execSteps := checkpointProps.GetFields()["exec_steps"].GetStructValue().GetFields()["steps"].GetListValue().AsSlice()
	assert.IntsEqual(t, len(execSteps), 1)
	assert.IntsEqual(t, int(execSteps[0].(float64)), int(tc.expectedExecStep.Number()))

	supportedBuild, exists := properties.GetFields()["$chromeos/cros_try"].GetStructValue().GetFields()["supported_build"]
	assert.Assert(t, exists && supportedBuild.GetBoolValue())
}

func TestRetry_childBuilder_fullRun(t *testing.T) {
	doChildRetryTestRun(t, &childRetryTestConfig{
		dryrun:           false,
		bbid:             "8794230068334833050",
		builderName:      "staging-zork-release-main",
		builderJSON:      stripNewlines(failedChildJSON),
		expectedExecStep: pb.RetryStep_DEBUG_SYMBOLS,
	})
}
func TestRetry_childBuilder_dryRun(t *testing.T) {
	doChildRetryTestRun(t, &childRetryTestConfig{
		dryrun:           true,
		bbid:             "8794230068334833050",
		builderName:      "staging-zork-release-main",
		builderJSON:      stripNewlines(failedChildJSON),
		expectedExecStep: pb.RetryStep_DEBUG_SYMBOLS,
	})
}

func TestRetry_childBuilder_successfulNoRetry(t *testing.T) {
	doChildRetryTestRun(t, &childRetryTestConfig{
		bbid:        "8794230068334833058",
		builderName: "staging-eve-release-main",
		builderJSON: stripNewlines(successfulChildJSON),
		testNoRun:   true,
	})
}

func TestRetry_childBuilder_failedNoRetrySummary(t *testing.T) {
	doChildRetryTestRun(t, &childRetryTestConfig{
		bbid:             "8794230068334833059",
		builderName:      "staging-zork-release-main",
		builderJSON:      stripNewlines(emptyRetrySummaryJSON),
		expectedExecStep: pb.RetryStep_STAGE_ARTIFACTS,
	})
}

const (
	noRetrySummaryJSON = `{
	"id": "879423006833483308",
	"builder": {
		"project": "chromeos",
		"bucket": "staging",
		"builder": "staging-zork-release-main"
	},
	"status": "FAILURE",
	"input": {
		"properties": {
			"recipe": "build_release",
			"input_prop": 102
		}
	},
	"output": {
		"properties": {
		}
	}
}`
)

func TestRetry_childBuilder_paygen_fail_noSummary(t *testing.T) {
	// This build has no retry_summary and failed so we can't retry it.
	doChildRetryTestRun(t, &childRetryTestConfig{
		dryrun:      false,
		bbid:        "8794230068334833058",
		builderName: "staging-zork-release-main",
		builderJSON: stripNewlines(noRetrySummaryJSON),
		expectError: true,
		paygenRetry: true,
	})
}

func TestRetry_childBuilder_paygen_fail_hasSummary(t *testing.T) {
	// This build has failed steps before paygen and thus should not run.
	doChildRetryTestRun(t, &childRetryTestConfig{
		dryrun:      false,
		bbid:        "8794230068334833058",
		builderName: "staging-zork-release-main",
		builderJSON: stripNewlines(failedChildJSON),
		expectError: true,
		paygenRetry: true,
	})
}

const (
	failedEbuildTestJSON = `{
		"id": "8794230068334833051",
		"builder": {
			"project": "chromeos",
			"bucket": "staging",
			"builder": "staging-eve-release-main"
		},
		"status": "FAILURE",
		"input": {
			"properties": {
				"recipe": "build_release",
				"input_prop": 102
			}
		},
		"output": {
			"properties": {
				"retry_summary": {
					"COLLECT_SIGNING": "FAILURE",
					"DEBUG_SYMBOLS": "FAILURE",
					"EBUILD_TESTS": "FAILED",
					"PAYGEN": "FAILED",
					"PUSH_IMAGES": "SUCCESS",
					"STAGE_ARTIFACTS": "SUCCESS"
				}
			}
		}
	}`
)

func TestRetry_childBuilder_ebuildTestsFailNoRetry(t *testing.T) {
	doChildRetryTestRun(t, &childRetryTestConfig{
		dryrun:      false,
		bbid:        "8794230068334833051",
		builderName: "staging-eve-release-main",
		builderJSON: stripNewlines(failedEbuildTestJSON),
		expectError: true,
		testNoRun:   true,
	})
}

func TestGetExecStep(t *testing.T) {
	t.Parallel()

	for i, tc := range []struct {
		recipe           string
		buildStatus      bbpb.Status
		retrySummary     map[pb.RetryStep]string
		signingSummary   map[string]string
		expectedExecStep pb.RetryStep
		expectError      bool
	}{
		{
			recipe: "orchestrator",
			retrySummary: map[pb.RetryStep]string{
				pb.RetryStep_CREATE_BUILDSPEC: "FAILED",
			},
			expectedExecStep: pb.RetryStep_CREATE_BUILDSPEC,
		},
		{
			recipe:           "orchestrator",
			retrySummary:     map[pb.RetryStep]string{},
			expectedExecStep: pb.RetryStep_CREATE_BUILDSPEC,
		},
		{
			// Signing retry.
			recipe: "build_release",
			retrySummary: map[pb.RetryStep]string{
				pb.RetryStep_STAGE_ARTIFACTS: "SUCCESS",
				pb.RetryStep_PUSH_IMAGES:     "SUCCESS",
				pb.RetryStep_DEBUG_SYMBOLS:   "SUCCESS",
				pb.RetryStep_COLLECT_SIGNING: "SUCCESS",
				pb.RetryStep_PAYGEN:          "FAILED",
			},
			signingSummary: map[string]string{
				"gs://chromeos-releases/canary-channel/...instructions": "PASSED",
				"gs://chromeos-releases/dev-channel/...instructions":    "FAILED",
			},
			expectedExecStep: pb.RetryStep_PUSH_IMAGES,
		},
		{
			// Signing retry.
			recipe: "build_release",
			retrySummary: map[pb.RetryStep]string{
				pb.RetryStep_STAGE_ARTIFACTS: "SUCCESS",
				pb.RetryStep_PUSH_IMAGES:     "SUCCESS",
				pb.RetryStep_DEBUG_SYMBOLS:   "SUCCESS",
				pb.RetryStep_COLLECT_SIGNING: "FAILED",
			},
			signingSummary: map[string]string{
				"gs://chromeos-releases/canary-channel/...instructions": "PASSED",
				"gs://chromeos-releases/dev-channel/...instructions":    "TIMED_OUT",
			},
			expectedExecStep: pb.RetryStep_PUSH_IMAGES,
		},
		{
			recipe: "build_release",
			retrySummary: map[pb.RetryStep]string{
				pb.RetryStep_STAGE_ARTIFACTS: "SUCCESS",
				pb.RetryStep_PUSH_IMAGES:     "SUCCESS",
				pb.RetryStep_DEBUG_SYMBOLS:   "SUCCESS",
				pb.RetryStep_COLLECT_SIGNING: "SUCCESS",
				pb.RetryStep_PAYGEN:          "SUCCESS",
			},
			expectedExecStep: pb.RetryStep_UNDEFINED,
		},
		{
			recipe: "build_release",
			retrySummary: map[pb.RetryStep]string{
				pb.RetryStep_STAGE_ARTIFACTS: "SUCCESS",
				pb.RetryStep_PUSH_IMAGES:     "SUCCESS",
				pb.RetryStep_DEBUG_SYMBOLS:   "FAILED",
			},
			expectedExecStep: pb.RetryStep_DEBUG_SYMBOLS,
		},
		{
			recipe:           "build_release",
			retrySummary:     map[pb.RetryStep]string{},
			expectedExecStep: pb.RetryStep_STAGE_ARTIFACTS,
		},
		{
			recipe:      "paygen-orchestrator",
			expectError: true,
		},
		{
			// Violates suffix constraint.
			recipe: "build_release",
			retrySummary: map[pb.RetryStep]string{
				pb.RetryStep_STAGE_ARTIFACTS: "SUCCESS",
				pb.RetryStep_PUSH_IMAGES:     "SUCCESS",
				pb.RetryStep_DEBUG_SYMBOLS:   "FAILURE",
				pb.RetryStep_COLLECT_SIGNING: "SUCCESS",
				pb.RetryStep_PAYGEN:          "SUCCESS",
			},
			expectError: true,
		},
		{
			// Violates suffix constraint.
			recipe: "build_release",
			retrySummary: map[pb.RetryStep]string{
				pb.RetryStep_STAGE_ARTIFACTS: "SUCCESS",
				pb.RetryStep_PUSH_IMAGES:     "SUCCESS",
				// Missing DEBUG_SYMBOLS.
				pb.RetryStep_COLLECT_SIGNING: "SUCCESS",
				pb.RetryStep_PAYGEN:          "SUCCESS",
			},
			expectError: true,
		},
		{
			recipe:      "build_release",
			buildStatus: bbpb.Status_FAILURE,
			retrySummary: map[pb.RetryStep]string{
				pb.RetryStep_STAGE_ARTIFACTS: "SUCCESS",
				pb.RetryStep_PUSH_IMAGES:     "SUCCESS",
				pb.RetryStep_DEBUG_SYMBOLS:   "SUCCESS",
				pb.RetryStep_COLLECT_SIGNING: "SUCCESS",
				pb.RetryStep_PAYGEN:          "SUCCESS",
			},
			expectError: true,
		},
	} {
		if tc.buildStatus == bbpb.Status_STATUS_UNSPECIFIED {
			tc.buildStatus = bbpb.Status_SUCCESS
		}
		execStep, err := getExecStep(tc.recipe, buildInfo{
			status:         tc.buildStatus,
			retrySummary:   tc.retrySummary,
			signingSummary: tc.signingSummary,
		})
		if tc.expectError && err == nil {
			t.Errorf("#%d: expected error from GetExecStep, got none", i)
		}
		if !tc.expectError && err != nil {
			t.Errorf("#%d: unexpected error from GetExecStep: %v", i, err)
		}
		if execStep != tc.expectedExecStep {
			t.Errorf("#%d: unexpected return from GetExecStep: expected %+v, got %+v", i, tc.expectedExecStep, execStep)
		}
	}
}

func Test_DirectEntry(t *testing.T) {
	t.Parallel()
	propsFile, err := os.CreateTemp("", "input_props")
	defer os.Remove(propsFile.Name())
	assert.NilError(t, err)

	bbid := "8794230068334833058"

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
			{
				ExpectedCmd: []string{"bb", "get", bbid, "-p", "-json"},
				Stdout:      stripNewlines(failedChildJSON),
			},
		},
	}
	expectedBucket := "chromeos/staging"
	expectedBuilder := "staging-zork-release-main"
	expectedAddCmd := []string{"bb", "add", fmt.Sprintf("%s/%s", expectedBucket, expectedBuilder)}
	expectedAddCmd = append(expectedAddCmd, "-t", "tryjob-launcher:sundar@google.com")
	expectedAddCmd = append(expectedAddCmd, "-p", fmt.Sprintf("@%s", propsFile.Name()))
	f.CommandRunners = append(f.CommandRunners,
		cmd.FakeCommandRunner{
			ExpectedCmd: expectedAddCmd,
			Stdout:      bbAddOutput("12345679"),
		},
	)

	retryOpts := &RetryRunOpts{
		CmdRunner: f,
		PropsFile: propsFile,
		BBID:      bbid,
		Dryrun:    false,
	}
	retryClient := &Client{}
	newBBID, err := retryClient.DoRetry(retryOpts)
	assert.NilError(t, err)
	assert.StringsEqual(t, newBBID, "12345679")

	properties, err := bb.ReadStructFromFile(propsFile.Name())
	assert.NilError(t, err)

	checkpointProps := properties.GetFields()["$chromeos/checkpoint"].GetStructValue()

	assert.Assert(t, checkpointProps.GetFields()["retry"].GetBoolValue())

	originalBuildBBID := checkpointProps.GetFields()["original_build_bbid"].GetStringValue()
	assert.StringsEqual(t, originalBuildBBID, bbid)

	execSteps := checkpointProps.GetFields()["exec_steps"].GetStructValue().GetFields()["steps"].GetListValue().AsSlice()
	assert.IntsEqual(t, len(execSteps), 1)
	assert.IntsEqual(t, int(execSteps[0].(float64)), int(pb.RetryStep_DEBUG_SYMBOLS.Number()))
}

const (
	paygenOrchJSON = `{
	"id": "879423006833483308",
	"builder": {
		"project": "chromeos",
		"bucket": "staging",
		"builder": "staging-paygen-orchestrator"
	},
	"status": "FAILURE",
	"input": {
		"properties": {
			"recipe": "paygen_orchestrator",
			"input_prop": 102
		}
	},
	"output": {
		"properties": {
		}
	}
}`

	paygenJSON = `{
"id": "879423006833483308",
"builder": {
	"project": "chromeos",
	"bucket": "staging",
	"builder": "staging-paygen"
},
"status": "FAILURE",
"input": {
	"properties": {
		"recipe": "paygen",
		"input_prop": 102
	}
},
"output": {
	"properties": {
	}
}
}`
)

type paygenTestConfig struct {
	expectedBuilder string
	buildJSON       string
	dryrun          bool
}

func doPaygenTest(t *testing.T, tc *paygenTestConfig) {
	t.Helper()
	propsFile, err := os.CreateTemp("", "input_props")
	defer os.Remove(propsFile.Name())
	assert.NilError(t, err)

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
			{
				ExpectedCmd: []string{"bb", "get", "879423006833483308", "-p", "-json"},
				Stdout:      stripNewlines(tc.buildJSON),
			},
		},
	}
	expectedBuilder := fmt.Sprintf("chromeos/staging/%s", tc.expectedBuilder)
	expectedAddCmd := []string{"bb", "add", expectedBuilder}
	expectedAddCmd = append(expectedAddCmd, "-t", "tryjob-launcher:sundar@google.com")
	expectedAddCmd = append(expectedAddCmd, "-p", fmt.Sprintf("@%s", propsFile.Name()))
	if !tc.dryrun {
		f.CommandRunners = append(f.CommandRunners,
			cmd.FakeCommandRunner{
				ExpectedCmd: expectedAddCmd,
				Stdout:      bbAddOutput("12345679"),
			},
		)
	}

	r := retryRun{
		propsFile:    propsFile,
		originalBBID: "879423006833483308",
		tryRunBase: tryRunBase{
			cmdRunner: f,
			dryrun:    tc.dryrun,
		},
	}
	ret := r.Run(nil, nil, nil)
	assert.IntsEqual(t, ret, Success)

	properties, err := bb.ReadStructFromFile(propsFile.Name())
	assert.NilError(t, err)

	// Check that we haven't set any checkpoint properties.
	_, exists := properties.GetFields()["$chromeos/checkpoint"]
	assert.Assert(t, !exists)
}

func TestRetry_paygenOrch(t *testing.T) {
	doPaygenTest(t, &paygenTestConfig{
		expectedBuilder: "staging-paygen-orchestrator",
		buildJSON:       paygenOrchJSON,
	})
}

func TestRetry_paygen(t *testing.T) {
	doPaygenTest(t, &paygenTestConfig{
		expectedBuilder: "staging-paygen",
		buildJSON:       paygenJSON,
	})
}

func TestRetry_paygenOrch_dryrun(t *testing.T) {
	doPaygenTest(t, &paygenTestConfig{
		expectedBuilder: "staging-paygen-orchestrator",
		buildJSON:       paygenOrchJSON,
		dryrun:          true,
	})
}

func TestRetry_paygen_dryrun(t *testing.T) {
	doPaygenTest(t, &paygenTestConfig{
		expectedBuilder: "staging-paygen",
		buildJSON:       paygenJSON,
		dryrun:          true,
	})
}
