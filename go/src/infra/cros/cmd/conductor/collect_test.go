// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"infra/cros/internal/assert"
	"infra/cros/internal/cmd"
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

func TestValidate(t *testing.T) {
	t.Parallel()
	r := &collectRun{}
	assert.ErrorContains(t, r.validate(), "--input_json")
	r = &collectRun{
		inputJSON: "foo",
	}
	assert.ErrorContains(t, r.validate(), "BBID")
	r = &collectRun{
		inputJSON: "foo",
		bbids:     []string{"123"},
	}
	assert.NilError(t, r.validate())
}

type collectResult struct {
	status bbpb.Status
}

type collectTestConfig struct {
	configJSON     string
	bbids          []int64
	collectResults []map[int64]collectResult
}

func doTestRun(t *testing.T, tc *collectTestConfig) {
	t.Helper()

	inputFile, err := os.CreateTemp("", "input_json")
	defer os.Remove(inputFile.Name())
	assert.NilError(t, err)

	_, err = inputFile.WriteString(tc.configJSON)
	assert.NilError(t, err)
	assert.NilError(t, inputFile.Close())

	var initialBBIDs []string
	commandRunners := []cmd.FakeCommandRunner{}
	for _, collectResults := range tc.collectResults {
		var stdout string
		bbids := []string{}
		for bbid, collectResult := range collectResults {
			build := bbpb.Build{
				Id:     bbid,
				Status: collectResult.status,
			}
			buildJSON, err := json.Marshal(build)
			assert.NilError(t, err)
			bbids = append(bbids, fmt.Sprintf("%d", bbid))
			stdout += string(buildJSON) + "\n"
		}
		sort.Strings(bbids)
		if len(initialBBIDs) == 0 {
			initialBBIDs = bbids
		}
		args := []string{"bb", "get"}
		args = append(args, bbids...)
		args = append(args, "-p", "-json")
		commandRunners = append(commandRunners, cmd.FakeCommandRunner{
			ExpectedCmd: args,
			Stdout:      stdout,
		})
	}
	c := collectRun{
		cmdRunner: &cmd.FakeCommandRunnerMulti{
			CommandRunners: commandRunners,
		},
		inputJSON:              inputFile.Name(),
		pollingIntervalSeconds: 0,
		bbids:                  initialBBIDs,
	}
	ret := c.Run(nil, nil, nil)
	assert.IntsEqual(t, ret, 0)
}

func TestCollect_NoRetries(t *testing.T) {
	t.Parallel()
	doTestRun(t, &collectTestConfig{
		configJSON: "{}",
		collectResults: []map[int64]collectResult{
			{
				12345: {
					bbpb.Status_SCHEDULED,
				},
				12346: {
					bbpb.Status_STARTED,
				},
				12347: {
					bbpb.Status_STARTED,
				},
			},
			{
				12345: {
					bbpb.Status_SUCCESS,
				},
				12346: {
					bbpb.Status_STARTED,
				},
				12347: {
					bbpb.Status_STARTED,
				},
			},
			{
				12346: {
					bbpb.Status_FAILURE,
				},
				12347: {
					bbpb.Status_INFRA_FAILURE,
				},
			},
		},
	})
}
