// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package utils

// UnionSets returns a new set that is a union of the given sets
func UnionSets[X comparable](sets ...map[X]struct{}) map[X]struct{} {
	unionedSet := make(map[X]struct{})
	for _, set := range sets {
		for key := range set {
			unionedSet[key] = struct{}{}
		}
	}
	return unionedSet
}

// Keys returns a slice of all the keys in the map.
func Keys[X comparable, Y any](m map[X]Y) []X {
	keys := make([]X, 0)
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
