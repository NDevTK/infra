// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package starfish

import (
	"strings"
	"testing"
)

func TestParseActiveSIMSlots(t *testing.T) {
	tests := []struct {
		name           string
		text           []string
		wantErr        bool
		expectedResult []int
	}{
		{
			name: "pass_4_slots",
			text: []string{
				"[0000000027318797] <inf> console: SIM 0 = Found",
				"[0000000027318797] <inf> console: SIM 1 = Found",
				"[0000000027318797] <inf> console: SIM 2 = Found",
				"[0000000027318797] <inf> console: SIM 3 = Found",
				"[0000000027318797] <inf> console: SIM 4 = None",
				"[0000000027318798] <inf> console: SIM 5 = None",
				"[0000000027318798] <inf> console: SIM 6 = None",
				"[0000000027318798] <inf> console: SIM 7 = None",
			},
			expectedResult: []int{0, 1, 2, 3},
		},
		{
			name:           "pass_empty",
			text:           []string{},
			expectedResult: []int{},
		},
		{
			name: "fail_duplicate_slots",
			text: []string{
				"[0000000027318797] <inf> console: SIM 0 = Found",
				"[0000000027318798] <inf> console: SIM 6 = None",
				"[0000000027318798] <inf> console: SIM 7 = None",
				"[0000000027318797] <inf> console: SIM 0 = Found",
			},
			expectedResult: []int{},
			wantErr:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			slots, err := parseActiveSIMSlots(strings.Join(test.text, "\n"))
			if len(slots) != len(test.expectedResult) {
				t.Errorf("TestParseActiveSIMSlots: invalid result, got %d results expected %d", len(slots), len(test.expectedResult))
				return
			}
			for i := range test.expectedResult {
				if slots[i] != test.expectedResult[i] {
					t.Errorf("TestParseActiveSIMSlots: invalid result, got %v expected %v", slots[i], test.expectedResult[i])
					return
				}
			}
			if (err != nil) != test.wantErr {
				t.Errorf("TestParseActiveSIMSlots: error = %v, wantErr %v", err, test.wantErr)
				return
			}
		})
	}
}
