// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package collection

import (
	"google.golang.org/api/iterator"
)

func Subtract[T, U any](sliceA []T, sliceB []U, compare func(a T, b U) bool) []T {
	var acc []T

	for _, a := range sliceA {
		find := false
		for _, b := range sliceB {
			if compare(a, b) {
				find = true
				break
			}
		}
		if !find {
			acc = append(acc, a)
		}
	}

	return acc
}

// Collect the iterator, and then use parser function to compose the result type.
func Collect[T, U any](nextFunc func() (T, error), parser func(T) (U, error)) ([]U, error) {
	var res []U

	for {
		item, err := nextFunc()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		data, err := parser(item)
		if err != nil {
			return nil, err
		}
		res = append(res, data)
	}

	return res, nil
}

// Contains check the element is in the given slice.
func Contains[T comparable](slice []T, elem T) bool {
	for _, a := range slice {
		if a == elem {
			return true
		}
	}

	return false
}
