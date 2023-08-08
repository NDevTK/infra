// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package util contains common utility functions.
package util

// PrependString returns an array with an element at the beginning.
func PrependString(newElem string, arr []string) []string {
	return append([]string{newElem}, arr...)
}
