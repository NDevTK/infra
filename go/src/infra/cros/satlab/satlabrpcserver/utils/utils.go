// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package utils

import (
	"log"

	"google.golang.org/api/iterator"
)

type Pair[T, U any] struct {
	First  T
	Second U
}

type BoardAndModelPair struct {
	Board string
	Model string
}

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
//
// TODO Maybe we can think a better way to handle parse error.
// If the parser function get an error, it will ignore the record.
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
			log.Printf("Parser data failed %v", item)
			return nil, err
		}
		res = append(res, data)
	}

	return res, nil
}
