// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package schedulingunit

import (
	"fmt"

	"infra/libs/skylab/inventory/swarming"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsUtil "infra/unifiedfleet/app/util"
)

func schedulingUnitDutState(states []string) string {
	dutStateMap := map[string]int{
		"ready":               1,
		"needs_repair":        2,
		"repair_failed":       3,
		"needs_manual_repair": 4,
		"needs_replacement":   4,
		"needs_deploy":        4,
	}
	record := 0
	for _, s := range states {
		if dutStateMap[s] > record {
			record = dutStateMap[s]
		}
	}
	suStateMap := map[int]string{
		0: "unknown",
		1: "ready",
		2: "needs_repair",
		3: "repair_failed",
		4: "needs_manual_attention",
	}
	return suStateMap[record]
}

func joinSingleValueLabel(labels []string) []string {
	res := make([]string, 0)
	occurrences := make(map[string]int)
	for _, l := range labels {
		occurrences[l] += 1
		suffix := ""
		if occurrences[l] > 1 {
			suffix = fmt.Sprintf("%d", occurrences[l])
		}
		res = append(res, l+suffix)
	}
	return res
}

func dutLabelValues(label string, dims []swarming.Dimensions) []string {
	res := make([]string, 0)
	for _, dim := range dims {
		if v, ok := dim[label]; ok {
			if len(v) > 0 {
				res = append(res, v[0])
			}
		}
	}
	return res
}

func SchedulingUnitDimensions(su *ufspb.SchedulingUnit, dutsDims []swarming.Dimensions) map[string][]string {
	suDims := map[string][]string{
		"dut_name":        {ufsUtil.RemovePrefix(su.GetName())},
		"dut_id":          {su.GetName()},
		"label-pool":      su.GetPools(),
		"label-dut_count": {fmt.Sprintf("%d", len(dutsDims))},
		"label-multiduts": {"True"},
		"dut_state":       {schedulingUnitDutState(dutLabelValues("dut_state", dutsDims))},
	}
	singleValueLabels := []string{"label-board", "label-model"}
	for _, l := range singleValueLabels {
		joinedLabels := joinSingleValueLabel(dutLabelValues(l, dutsDims))
		if len(joinedLabels) > 0 {
			suDims[l] = joinedLabels
		}
	}
	return suDims
}
