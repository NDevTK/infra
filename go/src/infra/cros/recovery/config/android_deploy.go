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
			PlanAndroid: setAllowFail(androidRepairPlan(), false),
			PlanClosing: setAllowFail(androidClosePlan(), true),
		}}
}
