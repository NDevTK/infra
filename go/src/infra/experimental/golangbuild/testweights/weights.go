// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate go run computeweights.go

// Package testweights contains test weights for the test sharding strategy
// for golangbuild.
package testweights

// GoDistTest returns a weight for the test by name.
func GoDistTest(name string, longtest, race bool) int {
	var weight int
	var ok bool
	switch {
	case !longtest && !race:
		weight, ok = goDistTestWeights[name]
	case longtest && !race:
		weight, ok = goDistTestLongtestWeights[name]
	case !longtest && race:
		weight, ok = goDistTestRaceWeights[name]
	case longtest && race:
		weight, ok = goDistTestLongtestRaceWeights[name]
	}
	if !ok {
		return 1
	}
	return weight
}
