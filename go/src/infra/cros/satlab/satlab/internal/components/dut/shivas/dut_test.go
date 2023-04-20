// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shivas

import (
	"os/exec"
	"testing"

	"infra/cros/satlab/satlab/internal/site"
)

// cmdRunTracker holds the last command run.
type cmdRunTracker struct {
	lastCall string
}

// fakeRunCmd returns benign data but tracks the command that was executed.
func (c *cmdRunTracker) fakeRunCmd(e *exec.Cmd) error {
	c.lastCall = e.String()
	return nil
}

// TestDUT_add ensures for given inputs, we run a specific command.
// This works largely because we sort the flags map beforehand so we have
// deterministic outputs.
func TestDUT_add(t *testing.T) {
	tests := []struct {
		name     string
		inputEnv map[string]string
		wantCall string
	}{
		{
			name:     "tmp",
			inputEnv: map[string]string{},
			wantCall: "/usr/local/bin/shivas add dut -deploy-bucket labpack_runner -deploy-project chromeos -name name -pools swimming -servo servo",
		},
		{
			name: "tmp",
			inputEnv: map[string]string{
				site.LUCIProjectEnv:         "nyc-rocks",
				site.DeployBuilderBucketEnv: "special-deploy",
			},
			wantCall: "/usr/local/bin/shivas add dut -deploy-bucket special-deploy -deploy-project nyc-rocks -name name -pools swimming -servo servo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel() setting env vars

			// Set package level `commandRunner` to replace command execution
			// with our own function
			c := cmdRunTracker{}
			commandRunner = c.fakeRunCmd

			for key, val := range tt.inputEnv {
				t.Setenv(key, val)
			}

			d := &DUT{
				Namespace:  "os",
				Zone:       "zone",
				Name:       "name",
				Servo:      "servo",
				Rack:       "rack",
				ShivasArgs: map[string][]string{"pools": {"swimming"}},
			}
			err := d.add()
			if err != nil {
				t.Errorf("unexpected err: %s", err)
			}

			if c.lastCall != tt.wantCall {
				t.Errorf("got: %s, expected: %s", c.lastCall, tt.wantCall)
			}
		})
	}
}
