// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

// AndroidDeployConfig provides config for android deployment task.
func AndroidDeployConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanAndroid,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanAndroid: setAllowFail(androidDeployPlan(), false),
			PlanClosing: setAllowFail(androidClosePlan(), true),
		}}
}

func androidDeployPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Set state: needs_deploy",
			"Validate DUT info",
			"Validate associated host",
			"Lock associated host",
			"Validate adb",
			"DUT is accessible over adb",
			"Reset DUT",
			"Set state: ready",
		},
		Actions: androidRepairDeployActions(),
	}
}
