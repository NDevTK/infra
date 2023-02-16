// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"google.golang.org/protobuf/types/known/durationpb"
)

func chameleonPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Mark as bad",
			"Device is pingable",
			"cros_ssh",
			"Mark as good",
		},
		Actions: map[string]*Action{
			"Mark as bad":  {ExecName: "chameleon_state_broken"},
			"Mark as good": {ExecName: "chameleon_state_working"},
			"Device is pingable": {
				ExecTimeout: &durationpb.Duration{Seconds: 15},
				ExecName:    "cros_ping",
				RecoveryActions: []string{
					"Chameleon RPM power cycle",
				},
			},
			"Chameleon RPM power cycle": {
				Docs: []string{
					"Power cycle chameleon if rpm exists",
					"Ensure chameleon is SSHable on after power cycle",
				},
				Conditions: []string{
					"Has chameleon rpm info",
				},
				Dependencies: []string{
					"Power cycle chameleon by RPM",
					"Wait for SSHable (after rpm cycle)",
				},
				ExecName: "sample_pass",
			},
			"Has chameleon rpm info": {
				Docs: []string{
					"Check if chameleon rpm exists",
				},
				ExecName: "device_has_rpm_info",
				ExecExtraArgs: []string{
					"device_type:chameleon",
				},
			},
			"Power cycle chameleon by RPM": {
				Docs: []string{
					"Run rpm power cycle on chameleon",
				},
				ExecName: "device_rpm_power_cycle",
				ExecExtraArgs: []string{
					"device_type:chameleon",
				},
			},
			"Wait for SSHable (after rpm cycle)": {
				// No recovery actions as that is help action.
				Docs: []string{
					"Try to wait device to be sshable after the device being rebooted.",
					"Waiting time 150 seconds.",
				},
				ExecName:    "cros_ssh",
				ExecTimeout: &durationpb.Duration{Seconds: 150},
				RunControl:  RunControl_ALWAYS_RUN,
			},
		},
	}
}
