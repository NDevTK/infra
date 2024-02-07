// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dolos

import (
	"context"
	"strings"
	"testing"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/tlw"
)

var setDolosStateExecTestCases = []struct {
	testName           string
	actionArgs         []string
	expectedDolosState tlw.Dolos_State
	Dolos              *tlw.Dolos
	expectedErr        error
}{
	{
		"success: set dolos state to WORKING",
		[]string{
			"state:WORKING",
		},
		tlw.Dolos_WORKING,
		&tlw.Dolos{
			Hostname: "fake-dolos-host",
			State:    tlw.Dolos_STATE_UNSPECIFIED,
		},
		nil,
	},
	{
		"fail: missing state key",
		[]string{
			"test:WORKING",
		},
		tlw.Dolos_STATE_UNSPECIFIED,
		&tlw.Dolos{
			Hostname: "fake-dolos-host",
			State:    tlw.Dolos_STATE_UNSPECIFIED,
		},
		errors.Reason("set dolos state: state is not provided").Err(),
	},
	{
		"fail: state info is empty",
		[]string{
			"state:",
		},
		tlw.Dolos_STATE_UNSPECIFIED,
		&tlw.Dolos{
			Hostname: "fake-dolos-host",
			State:    tlw.Dolos_STATE_UNSPECIFIED,
		},
		errors.Reason("set dolos state: state is not provided").Err(),
	},
	{
		"fail: passed in a miss matched state",
		[]string{
			"state:wrong_state",
		},
		tlw.Dolos_STATE_UNSPECIFIED,
		&tlw.Dolos{
			Hostname: "fake-dolos-host",
			State:    tlw.Dolos_STATE_UNSPECIFIED,
		},
		errors.Reason("set dolos state: state is \"WRONG_STATE\" not found").Err(),
	},
	{
		"fail: do not update if dolos is not supported in structure",
		[]string{
			"state:WORKING",
		},
		tlw.Dolos_WORKING,
		nil,
		errors.Reason("set dolos state: dolos is not supported").Err(),
	},
	{
		"success: set dolos state to BROKEN",
		[]string{
			"state:BROKEN",
		},
		tlw.Dolos_BROKEN,
		&tlw.Dolos{
			Hostname: "fake-dolos-host",
			State:    tlw.Dolos_BROKEN,
		},
		nil,
	},
}

func TestSetServoStateExec(t *testing.T) {
	t.Parallel()
	for _, tt := range setDolosStateExecTestCases {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			args := &execs.RunArgs{
				DUT: &tlw.Dut{
					Chromeos: &tlw.ChromeOS{
						Dolos: tt.Dolos,
					},
				},
			}
			info := execs.NewExecInfo(args, "name", tt.actionArgs, 0, nil)
			actualErr := setDolosStateExec(ctx, info)
			if actualErr != nil && tt.expectedErr != nil {
				if !strings.Contains(actualErr.Error(), tt.expectedErr.Error()) {
					t.Errorf("Expected error %q, but got %q", tt.expectedErr, actualErr)
				}
			}
			if (actualErr == nil && tt.expectedErr != nil) || (actualErr != nil && tt.expectedErr == nil) {
				t.Errorf("Expected error %q, but got %q", tt.expectedErr, actualErr)
			}
			if tt.Dolos != nil {
				actualDolosState := tt.Dolos.GetState()
				if actualDolosState != tt.expectedDolosState {
					t.Errorf("Expected dolos state %q, but got %q", tt.expectedDolosState, actualDolosState)
				}
			}
		})
	}
}
