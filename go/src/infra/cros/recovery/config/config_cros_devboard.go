// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

// CrosDevBoardConfig uses for DevBoards devices.
func CrosDevBoardConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanClosing: {
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
						RunControl:    RunControl_RUN_ONCE,
						MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
					},
				},
			},
		}}
}
