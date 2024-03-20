// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

var dutStateWeights = map[string]int{
	"ready":               1,
	"needs_repair":        2,
	"repair_failed":       3,
	"needs_manual_repair": 4,
	"needs_deploy":        5,
	"needs_replacement":   6,
	"reserved":            7,
}

var suStateMap = map[int]string{
	0: "unknown",
	1: "ready",
	2: "needs_repair",
	3: "repair_failed",
	4: "needs_manual_repair",
	5: "needs_deploy",
	6: "needs_replacement",
	7: "reserved",
}

// SchedulingUnitDutState calculates a weighted state based on all DUTs
// to represent the scheduling unit
// Copied from infra/go/src/infra/cmd/shivas/utils/schedulingunit/su_utility.go
func SchedulingUnitDutState(states []string) string {
	record := 0
	for _, s := range states {
		if dutStateWeights[s] > record {
			record = dutStateWeights[s]
		}
	}
	return suStateMap[record]
}
