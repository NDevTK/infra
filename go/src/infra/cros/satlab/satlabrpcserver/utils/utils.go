// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package utils

import (
	"log"
	"math"
	"os"

	"golang.org/x/crypto/ssh"
	"google.golang.org/api/iterator"

	"infra/cros/satlab/satlabrpcserver/utils/constants"
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

// Contains check the element is in the given slice.
func Contains[T comparable](slice []T, elem T) bool {
	for _, a := range slice {
		if a == elem {
			return true
		}
	}

	return false
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

// ReadSSHKey read a ssh private key file and then parse it to `ssh.Signer`
func ReadSSHKey(path string) (ssh.Signer, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Can't read the ssh private key from %v", path)
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(b)
	if err != nil {
		log.Printf("Parse private key error, got %v", err)
		return nil, err
	}
	return signer, nil
}

// NearlyEqual check two float points are nearly equal.
func NearlyEqual(a, b float64) bool {
	return math.Abs(a-b) <= constants.F64Epsilon*(math.Abs(a)+math.Abs(b))
}
