// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package peripherals

import (
	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	"strings"
	"testing"
)

func TestChameleonCleanAndValidateFlags(t *testing.T) {
	// Test invalid flags
	errTests := []struct {
		cmd  *manageChamCmd
		want []string
	}{
		{
			cmd:  &manageChamCmd{},
			want: []string{errDUTMissing, errNoHostname},
		},
		{
			cmd:  &manageChamCmd{hostname: " "},
			want: []string{errDUTMissing, errNoHostname},
		},
		{
			cmd:  &manageChamCmd{hostname: "h1"},
			want: []string{errDUTMissing},
		},
		{
			cmd: &manageChamCmd{
				hostname:  "h1",
				dutName:   "d",
				rpmOutlet: "a",
			},
			want: []string{"Need both rpm and its outlet"},
		},
		{
			cmd: &manageChamCmd{
				hostname:    "h1",
				dutName:     "d",
				rpmHostname: "rmp",
			},
			want: []string{"Need both rpm and its outlet"},
		},
		{
			cmd: &manageChamCmd{
				hostname: "h1",
				dutName:  "d",
				types:    []string{"  "},
			},
			want: []string{errEmptyType},
		},
		{
			cmd: &manageChamCmd{
				hostname: "h1",
				dutName:  "d",
				types:    []string{"v2", "v2"},
			},
			want: []string{errDuplicateType},
		},
	}

	for _, tt := range errTests {
		err := tt.cmd.cleanAndValidateFlags()
		if err == nil {
			t.Errorf("cleanAndValidateFlags = nil; want errors: %v", tt.want)
			continue
		}
		for _, errStr := range tt.want {
			if !strings.Contains(err.Error(), errStr) {
				t.Errorf("cleanAndValidateFlags = %q; want err %q included", err, errStr)
			}
		}
	}

	// Test valid flags with hostname cleanup
	c := &manageChamCmd{
		dutName:     "d",
		hostname:    "h1",
		types:       []string{"v2", "v3"},
		rpmHostname: "rmp",
		rpmOutlet:   "123",
		mode:        actionAdd,
	}
	if err := c.cleanAndValidateFlags(); err != nil {
		t.Errorf("cleanAndValidateFlags = %v; want nil", err)
	}
}

func TestAddCham(t *testing.T) {
	cmd := &manageChamCmd{dutName: "d", hostname: "h", mode: actionAdd}
	cmd.cleanAndValidateFlags()

	// Test adding a chameleon when there already exist another chameleon
	current := &lab.Chameleon{Hostname: "h2"}
	if _, err := cmd.newCham(current); err == nil {
		t.Errorf("newCham(%v) succeded, expect duplication failure", current)
	}
}

func TestDeleteCham(t *testing.T) {
	cmd := &manageChamCmd{dutName: "d", hostname: "h", mode: actionDelete}
	cmd.cleanAndValidateFlags()

	// Test deleting non existent
	current := &lab.Chameleon{}
	if _, err := cmd.newCham(current); err == nil {
		t.Errorf("deleteCham(%v) succeeded, expected non-existent delete failure", current)
	}

	current = &lab.Chameleon{Hostname: "h2"}
	if _, err := cmd.newCham(current); err == nil {
		t.Errorf("deleteCham(%v) succeeded, expected non-existent delete failure", current)
	}

	current = &lab.Chameleon{Hostname: "h"}
	_, err := cmd.newCham(current)
	if err != nil {
		t.Errorf("newCham(%v) = %v, expect success", current, err)
	}
}
