// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dns

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestMakeNewContent tests updating lines in a DNS file
func TestMakeNewContent(t *testing.T) {
	t.Parallel()

	expected := strings.Join([]string{
		tabify("addr1-UPDATE host1"),
		tabify("addr2 host2"),
		tabify("addr3-NEW host3") + "\n", // temporarily "strange" newline behavior resolved in next CL
		tabify("addr4-NEW host4"),
	}, "\n") + string("\n")

	newRecords := map[string]string{
		"host1": "addr1-UPDATE",
		"host4": "addr4-NEW",
		"host3": "addr3-NEW",
	}

	input := strings.Join([]string{
		tabify("addr1 host1"),
		tabify("addr2 host2"),
	}, "\n")
	actual, err := makeNewContent(input, newRecords)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		fmt.Printf("new: %s\n", actual)
		fmt.Printf("exp: %s\n", expected)
		t.Errorf("unexpected diff: %s", diff)
	}
}

// Tabify replaces arbitrary whitespace with tabs.
func tabify(s string) string {
	return strings.Join(strings.Fields(s), "\t")
}
