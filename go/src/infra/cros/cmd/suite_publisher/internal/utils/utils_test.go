// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnionSets(t *testing.T) {
	testCases := []struct {
		name      string
		sets      []map[string]struct{}
		wantUnion map[string]struct{}
	}{
		{
			name: "one set",
			sets: []map[string]struct{}{
				{
					"one": {},
				},
			},
			wantUnion: map[string]struct{}{
				"one": {},
			},
		},
		{
			name: "two disjoint sets",
			sets: []map[string]struct{}{
				{
					"one": {},
				},
				{
					"two": {},
				},
			},
			wantUnion: map[string]struct{}{
				"one": {},
				"two": {},
			},
		},
		{
			name: "three intersecting sets",
			sets: []map[string]struct{}{
				{
					"one": {},
					"two": {},
				},
				{
					"two":   {},
					"three": {},
				},
				{
					"three": {},
					"four":  {},
				},
			},
			wantUnion: map[string]struct{}{
				"one":   {},
				"two":   {},
				"three": {},
				"four":  {},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gotUnion := UnionSets(testCase.sets...)
			if !cmp.Equal(testCase.wantUnion, gotUnion) {
				t.Fatalf("computed union does not match expected, want: %v, got %v", testCase.wantUnion, gotUnion)
			}
		})
	}
}
