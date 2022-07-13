// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dns

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestValidate tests various flows
func TestValidate(t *testing.T) {
	t.Parallel()

	type input struct {
		command  *updateDNSRun
		args     []string
		satlabID string
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
		input{&updateDNSRun{host: "satlab-123-eli", address: "127.0.0.1"}, make([]string, 0), "123"},
		output{false, &updateDNSRun{host: "satlab-123-eli", address: "127.0.0.1"}},
	}, {
		"no host",
		input{&updateDNSRun{address: "127.0.0.1"}, make([]string, 0), "123"},
		output{true, &updateDNSRun{host: "", address: "127.0.0.1"}},
	}, {
		"no address",
		input{&updateDNSRun{host: "satlab-123-eli"}, make([]string, 0), "123"},
		output{true, &updateDNSRun{host: "satlab-123-eli", address: ""}},
	}, {
		"positional args",
		input{&updateDNSRun{host: "satlab-123-eli", address: "127.0.0.1"}, []string{"hi"}, "123"},
		output{true, &updateDNSRun{host: "satlab-123-eli", address: "127.0.0.1"}},
	}, {
		"prepend host",
		input{&updateDNSRun{host: "eli", address: "127.0.0.1"}, make([]string, 0), "123"},
		output{false, &updateDNSRun{host: "satlab-123-eli", address: "127.0.0.1"}},
	}}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i, o := tc.input, tc.output
			err := i.command.validate(i.args, i.satlabID)

			if o.errored != (err != nil) {
				t.Errorf("Testing(%+v) failed. Got error: %t, expected error: %t", tc, err, o.errored)
			}
			if diff := cmp.Diff(i.command.host, o.command.host); diff != "" {
				t.Errorf("Testing(%+v) failed with diff in host of command: %s", tc, diff)
			}
		})
	}
}
