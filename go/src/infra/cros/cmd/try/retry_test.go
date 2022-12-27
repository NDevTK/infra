// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"fmt"
	"os"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/cmd"

	pb "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
)

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
)

type retryTestConfig struct {
	dryrun bool
}

func doRetryTestRun(t *testing.T, tc *retryTestConfig) {
	t.Helper()
	propsFile, err := os.CreateTemp("", "input_props")
	defer os.Remove(propsFile.Name())
	assert.NilError(t, err)

	bbid := "8794230068334833057"
	f := &cmd.FakeCommandRunnerMulti{
		CommandRunners: []cmd.FakeCommandRunner{
			fakeAuthInfoRunner("bb", 0),
			fakeAuthInfoRunner("led", 0),
			{
				ExpectedCmd: []string{
					"led", "auth-info",
				},
				Stdout: "Logged in as sundar@google.com.\n\nOAuth token details:\n...",
			},
			{
				ExpectedCmd: []string{"bb", "get", bbid, "-p", "-json"},
				Stdout:      retryTestGoodJSON,
			},
			{
				ExpectedCmd: []string{"bb", "get", "8794230068334833058", "-p", "-json"},
				Stdout:      successfulChildJSON,
			},
			{
				ExpectedCmd: []string{"bb", "get", "8794230068334833059", "-p", "-json"},
				Stdout:      failedChildJSON,
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

	properties, err := readStructFromFile(propsFile.Name())
	assert.NilError(t, err)

	checkpointProps := properties.GetFields()["$chromeos/checkpoint"].GetStructValue()

	assert.Assert(t, checkpointProps.GetFields()["retry"].GetBoolValue())

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
}

func TestRetry_dryRun(t *testing.T) {
	t.Parallel()
	doRetryTestRun(t, &retryTestConfig{
		dryrun: true,
	})
}
func TestRetry_fullRun(t *testing.T) {
	t.Parallel()
	doRetryTestRun(t, &retryTestConfig{
		dryrun: false,
	})
}

func TestGetExecStep(t *testing.T) {
	t.Parallel()

	for i, tc := range []struct {
		recipe           string
		retrySummary     map[pb.RetryStep]string
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
	} {
		execStep, err := getExecStep(tc.recipe, tc.retrySummary)
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
