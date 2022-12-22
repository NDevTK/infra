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
			"input_prop": 102
		}
	},
	"output": {
		"properties": {
			"$chromeos/my_module": {
				"my_prop": 100
			},
			"my_other_prop": 101
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

	bbid := "12345"
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

	otherProp := int(properties.GetFields()["input_prop"].GetNumberValue())
	assert.IntsEqual(t, otherProp, 102)
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
