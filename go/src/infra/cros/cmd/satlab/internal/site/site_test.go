// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package site

import (
	"fmt"
	"testing"
)

// TestGetFullyQualifiedHostname tests that we produce correct FQ hostnames when passed different satlab ids and hosts
func TestGetFullyQualifiedHostname(t *testing.T) {
	t.Parallel()

	type input struct {
		specifiedSatlabID string
		fetchedSatlabID   string
		prefix            string
		content           string
	}

	type test struct {
		name   string
		input  input
		output string
	}

	tests := []test{
		{"prepend fields", input{"", "abc", "satlab", "host"}, "satlab-abc-host"},
		{"prepend fields manual override", input{"def", "abc", "satlab", "host"}, "satlab-def-host"},
		{"dont prepend", input{"abc", "abc", "satlab", "satlab-abc-host"}, "satlab-abc-host"},
	}

	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("TestFullyQualifiedHostname%d", i), func(t *testing.T) {
			t.Parallel()
			got := GetFullyQualifiedHostname(tc.input.specifiedSatlabID, tc.input.fetchedSatlabID, tc.input.prefix, tc.input.content)
			if got != tc.output {
				t.Errorf("got: %s, expected: %s for input %+v", got, tc.output, tc.input)
			}
		})
	}
}
