// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package labels

import (
	"infra/libs/skylab/inventory"
)

func init() {
	converters = append(converters, hwidCompConverter)
	reverters = append(reverters, hwidCompReverter)
}

const HWID_COMPONENT_NAME = "hwid_component"

func hwidCompConverter(ls *inventory.SchedulableLabels) []string {
	var labels []string
	for _, v := range ls.GetHwidComponent() {
		if v != "" {
			lv := HWID_COMPONENT_NAME + v.String()
			labels = append(labels, lv)
		}
	}
	return labels
}

func hwidCompReverter(ls *inventory.SchedulableLabels, labels []string) []string {
	var hwid_component = &inventory.SchedulableLabels.HwidComponent
	*hwid_component = nil
	for i := len(labels) - 1; i >= 0; i-- {
		if k, v := splitLabel(labels[i]); k == HWID_COMPONENT_NAME {
			*hwid_component = append(*hwid_component, v)
			labels = removeLabel(labels, i)
		}
	}
	return labels
}
