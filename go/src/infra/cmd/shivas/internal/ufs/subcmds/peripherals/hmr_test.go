// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package peripherals

import (
	"testing"

	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
)

func TestHmrCleanAndValidateFlags(t *testing.T) {
	// Invalid cases
	errTests := []struct {
		cmd  *manageHmrCmd
		want []string
	}{
		{
			cmd:  &manageHmrCmd{},
			want: []string{errDUTMissing, errEmptyHmrModel},
		},
		{
			cmd:  &manageHmrCmd{dutName: "d"},
			want: []string{errEmptyHmrPiHostname, errEmptyTouchHostPiHostname, errEmptyHmrModel},
		},
		{
			cmd:  &manageHmrCmd{dutName: "d", touchHostPi: "touch-host-pi"},
			want: []string{errEmptyHmrPiHostname, errEmptyHmrModel},
		},
		{
			cmd:  &manageHmrCmd{dutName: "d", hmrPi: "hmr-pi"},
			want: []string{errEmptyTouchHostPiHostname, errEmptyHmrModel},
		},
		{
			cmd:  &manageHmrCmd{dutName: "d", touchHostPi: "touch-host-pi", hmrPi: "hmr-pi"},
			want: []string{errEmptyHmrModel},
		},
		{
			cmd:  &manageHmrCmd{mode: actionDelete},
			want: []string{errDUTMissing},
		},
	}

	for _, tt := range errTests {
		err := tt.cmd.cleanAndValidateFlags()
		if err == nil {
			t.Errorf("cleanAndValidateFlags = nil; want errors: %v", tt.want)
			continue
		}
	}

}

func TestAddHmr(t *testing.T) {
	// Valid case
	c := &manageHmrCmd{
		dutName:     "d",
		touchHostPi: "touch-host-pi",
		hmrPi:       "hmr-pi",
		hmrModel:    "hmr-model",
		mode:        actionAdd,
	}
	if err := c.cleanAndValidateFlags(); err != nil {
		t.Errorf("cleanAndValidateFlags = %v; want nil", err)
	}

	// Valid case: create HMR
	if _, err := c.createHmr(); err != nil {
		t.Errorf("unable to create HMR: %v; want nil", err)
	}
}

func TestDeleteHmr(t *testing.T) {
	// Valid case
	c := &manageHmrCmd{
		dutName: "d",
		mode:    actionDelete,
	}
	if err := c.cleanAndValidateFlags(); err != nil {
		t.Errorf("cleanAndValidateFlags = %v; want nil", err)
	}
	// Valid case: delete HMR
	current := &lab.HumanMotionRobot{
		Hostname:        "hmr-pi",
		HmrModel:        "hmr-model",
		GatewayHostname: "touch-host-pi",
	}
	if _, err := c.runHmrAction(current); err != nil {
		t.Errorf("unable to delete HMR: %v; want nil", err)
	}
}
