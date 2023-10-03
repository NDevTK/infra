// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSliceContainsShouldWork(t *testing.T) {
	t.Parallel()

	slice := []string{"1", "2", "3"}

	cases := []struct {
		input  string
		output bool
	}{
		{
			"1",
			true,
		},
		{
			"4",
			false,
		},
	}

	for _, tt := range cases {
		input := tt.input
		expected := tt.output
		actual := Contains(slice, input)
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("diff: %v\n", diff)
		}
	}

}
