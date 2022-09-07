// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cbi

import (
	"reflect"
	"testing"
)

// TestBuildCBILocation tests that the BuildCBILocation function can correctly
// parse the output of the `ectool locatechip` command to get the port and
// address of CBI on a DUT and return it as a CBILocation. Does not actually
// invoke the `ectool locatechip` command.
func TestBuildCBILocation(t *testing.T) {
	testCases := []struct {
		locateCBIOutput     string
		expectedCBILocation *CBILocation
	}{
		{
			"Bus: I2C; Port: 0; Address: 0x50 (7-bit format)",
			&CBILocation{
				port:    "0",
				address: "0x50",
			},
		},
		{
			"Bus: I2C; Port: 9999; Address: 0xF00BAD (7-bit format)",
			&CBILocation{
				port:    "9999",
				address: "0xF00BAD",
			},
		},
		{
			"Port: 9999; Address: 0xF00BAD",
			&CBILocation{
				port:    "9999",
				address: "0xF00BAD",
			},
		},
		{
			"Port: 9999; 0xF00BAD",
			nil,
		},
		{
			"Port: 9999; 0xF00BAD",
			nil,
		},
		{
			"Port: 9999",
			nil,
		},
		{
			"Address: 0xF00BAD",
			nil,
		},
		{
			"",
			nil,
		},
	}
	t.Parallel()
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.locateCBIOutput, func(t *testing.T) {
			cbiLocation, _ := buildCBILocation(tt.locateCBIOutput)
			if !reflect.DeepEqual(cbiLocation, tt.expectedCBILocation) {
				t.Errorf(
					"Expected CBI Location %+v\n but got %+v\n",
					tt.expectedCBILocation,
					cbiLocation)
			}
		})
	}
}
