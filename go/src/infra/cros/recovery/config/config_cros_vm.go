// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

func CrosVMSuccessConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanCrOS,
		},
		Plans: map[string]*Plan{
			PlanCrOS: {
				CriticalActions: []string{
					"Set state: ready",
				},
				Actions: map[string]*Action{
					"Set state: ready": {
						Docs: []string{
							"The action set devices with state ready for the testing.",
						},
						ExecName: "dut_set_state",
						ExecExtraArgs: []string{
							"state:ready",
						},
						RunControl: RunControl_RUN_ONCE,
					},
				},
			},
		}}
}
