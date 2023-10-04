// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package collection

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/iterator"
)

func TestSubtractShouldWork(t *testing.T) {
	inA := []int{1, 2, 3, 4, 5}
	inB := []int{1, 2, 3}
	expected := []int{4, 5}

	out := Subtract(inA, inB, func(a, b int) bool { return a == b })

	if diff := cmp.Diff(expected, out); diff != "" {
		t.Errorf("diff: %v\n", diff)
	}
}

func TestCollectShouldWork(t *testing.T) {
	in := []int{1, 2, 3, 4}
	count := 0
	max := len(in)
	expected := []int{2, 4, 6, 8}
	out, err := Collect(func() (int, error) {
		if count < max {
			out := in[count]
			count = count + 1
			return out, nil
		}
		return 0, iterator.Done
	}, func(a int) (int, error) {
		return a * 2, nil
	})

	if err != nil {
		t.Errorf("Should success, but got an error:%v\n", err)
		return
	}

	if diff := cmp.Diff(expected, out); diff != "" {
		t.Errorf("diff: %v\n", diff)
		return
	}
}

func TestCollectWhenParseFailedShouldFail(t *testing.T) {
	in := []int{1, 2, 3, 4}
	count := 0
	max := len(in)
	out, err := Collect(func() (int, error) {
		if count < max {
			out := in[count]
			count = count + 1
			return out, nil
		}
		return 0, iterator.Done
	}, func(a int) (int, error) {
		return 0, errors.New("parse error")
	})

	if err == nil {
		t.Errorf("should parse error")
	}

	if out != nil {
		t.Errorf("result should be nil")
	}
}

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
