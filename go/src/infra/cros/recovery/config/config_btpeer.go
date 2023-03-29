// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"google.golang.org/protobuf/types/known/durationpb"
)

func btpeerRepairPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Set state: BROKEN",
			"Device is pingable",
			"Device is SSHable",
			// TODO(b:261631000) Chameleond is responsive",
			"Set state: WORKING",
		},
		Actions: map[string]*Action{
			"Set state: BROKEN": {
				Docs: []string{
					"The device state BROKEN.",
				},
				ExecName:      "btpeer_state_broken",
				RunControl:    RunControl_RUN_ONCE,
				MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
			},
			"Set state: WORKING": {
				Docs: []string{
					"The device state WORKING.",
				},
				ExecName:      "btpeer_state_working",
				RunControl:    RunControl_RUN_ONCE,
				MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
			},
			"Device is pingable": {
				Docs: []string{
					"Wait device to be pingable.",
					"Waiting time 15 seconds.",
				},
				ExecName:    "cros_ping",
				ExecTimeout: &durationpb.Duration{Seconds: 15},
				RunControl:  RunControl_ALWAYS_RUN,
			},
			"Device is SSHable": {
				Docs: []string{
					"Try to wait device to be sshable.",
					"Waiting time 150 seconds.",
				},
				ExecName:   "cros_ssh",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Chameleond is responsive": {
				Docs: []string{
					"Verify chameleond is responsive.",
					"Expected to receive not empty list of detected statuses.",
				},
				ExecName: "btpeer_get_detected_statuses",
				RecoveryActions: []string{
					"Restart chameleond and wait",
				},
			},
			"Restart chameleond and wait": {
				Docs: []string{
					"Restart chameleond and wait for service to be ready.",
				},
				Dependencies: []string{
					"Restart chameleond command",
					"Sleep for chameleond restart",
				},
				ExecName: "sample_pass",
			},
			"Restart chameleond command": {
				Docs: []string{
					"Restart chameleond service.",
				},
				ExecName: "cros_run_shell_command",
				ExecExtraArgs: []string{
					"sudo service chameleond restart",
				},
			},
			"Sleep for chameleond restart": {
				Docs: []string{
					"When we restart chameleond we need wait about 10 seconds to be recovered.",
				},
				ExecName: "sample_sleep",
				ExecExtraArgs: []string{
					"sleep:10",
				},
				RunControl:             RunControl_ALWAYS_RUN,
				AllowFailAfterRecovery: true,
				MetricsConfig:          &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
			},
		},
	}
}
