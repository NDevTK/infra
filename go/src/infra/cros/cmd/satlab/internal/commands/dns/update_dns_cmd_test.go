// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dns

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// fakeDBHGetter emulates behavior of fetching satlab ID with a constant return value
func fakeDHBGetter() (string, error) {
	return "123", nil
}

// TestRunCommandValidates tests various inputs to our runCmdInjected function
func TestRunCommandValidates(t *testing.T) {
	t.Parallel()

	type input struct {
		command         *updateDNSRun
		args            []string
		satlabIDFetcher DockerHostBoxIdentifierGetter
	}
	type output struct {
		errored bool
		command *updateDNSRun
	}
	type test struct {
		name   string
		input  input
		output output
	}

	tests := []test{{
		"happy path",
		input{&updateDNSRun{host: "satlab-123-eli", address: "127.0.0.1"}, make([]string, 0), fakeDHBGetter},
		output{false, &updateDNSRun{host: "satlab-123-eli", address: "127.0.0.1"}},
	}, {
		"no host",
		input{&updateDNSRun{address: "127.0.0.1"}, make([]string, 0), fakeDHBGetter},
		output{true, &updateDNSRun{host: "", address: "127.0.0.1"}},
	}, {
		"no address",
		input{&updateDNSRun{host: "satlab-123-eli"}, make([]string, 0), fakeDHBGetter},
		output{true, &updateDNSRun{host: "satlab-123-eli", address: ""}},
	}, {
		"positional args",
		input{&updateDNSRun{host: "satlab-123-eli", address: "127.0.0.1"}, []string{"hi"}, fakeDHBGetter},
		output{true, &updateDNSRun{host: "satlab-123-eli", address: "127.0.0.1"}},
	}, {
		"prepend host",
		input{&updateDNSRun{host: "eli", address: "127.0.0.1"}, make([]string, 0), fakeDHBGetter},
		output{false, &updateDNSRun{host: "satlab-123-eli", address: "127.0.0.1"}},
	}}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i, o := tc.input, tc.output
			err := i.command.runCmdInjected(i.args, i.satlabIDFetcher)

			if o.errored != (err != nil) {
				t.Errorf("Testing(%+v) failed. Got error: %t, expected error: %t", tc, err, o.errored)
			}
			if diff := cmp.Diff(i.command.host, o.command.host); diff != "" {
				t.Errorf("Testing(%+v) failed with diff in host of command: %s", tc, diff)
			}
		})
	}
}
