// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package util contains common utility functions.
package util

import (
	"reflect"
	"testing"
)

// TestPrependString tests PrependString.
func TestPrependString(t *testing.T) {
	t.Parallel()
	for i, tc := range []struct {
		newElem        string
		arr            []string
		expectedResult []string
	}{
		{"foo", []string{}, []string{"foo"}},
		{"foo", []string{"bar", "baz"}, []string{"foo", "bar", "baz"}},
	} {
		actualResult := PrependString(tc.newElem, tc.arr)
		if !reflect.DeepEqual(actualResult, tc.expectedResult) {
			t.Errorf("#%d: PrependString(%s, %s) returned %s; want %s", i, tc.newElem, tc.arr, actualResult, tc.expectedResult)
		}
	}
}
