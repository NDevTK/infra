// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package stableversion

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseImage(t *testing.T) {
	t.Parallel()

	data := []struct {
		name string
		in   string
		out  *FirmwareImageResult
		isOk bool
	}{
		{
			"empty",
			"",
			nil,
			false,
		},
		{
			"no prefix",
			"R1-2.3.4",
			nil,
			false,
		},
		{
			"basic good",
			"a-firmware/R1-2.3.4",
			&FirmwareImageResult{
				Platform:     "a",
				Release:      1,
				Tip:          2,
				Branch:       3,
				BranchBranch: 4,
			},
			true,
		},
		{
			"realistic good",
			"octopus-firmware/R72-11297.75.0",
			&FirmwareImageResult{
				Platform:     "octopus",
				Release:      72,
				Tip:          11297,
				Branch:       75,
				BranchBranch: 0,
			},
			true,
		},
		{
			"firmware version -- wrong entry type",
			"Google_Rammus.11275.41.0",
			nil,
			false,
		},
	}

	for _, subtest := range data {
		t.Run(subtest.name, func(t *testing.T) {
			wanted := subtest.out
			got, e := ParseFirmwareImage(subtest.in)
			if diff := cmp.Diff(wanted, got); diff != "" {
				t.Errorf("wanted: (%#v) got: (%#v)\n(%s)", wanted, got, diff)
			}
			if subtest.isOk {
				if e != nil {
					t.Errorf("unexpected error %s", errToString(e))
				}
			} else {
				if e == nil {
					t.Errorf("error should not have been nil")
				}
			}
		})
	}
}

func errToString(e error) string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("[%s]", e.Error())
}
