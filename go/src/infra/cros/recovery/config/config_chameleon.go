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
				Dependencies: []string{
					"Has chameleon rpm info",
					"Power cycle chameleon by RPM",
					"Wait for pingable",
					"Wait for SSHable",
				},
				ExecName: "sample_pass",
			},
			"Has chameleon rpm info": {
				ExecName: "device_has_rpm_info",
				ExecExtraArgs: []string{
					"device_type:chameleon",
				},
			},
			"Power cycle chameleon by RPM": {
				ExecName: "device_rpm_power_cycle",
				ExecExtraArgs: []string{
					"device_type:chameleon",
				},
			},
			"Wait for SSHable": {
				// No recovery actions as that is help action.
				Docs: []string{
					"Try to wait device to be sshable after the device being rebooted.",
					"Waiting time 150 seconds.",
				},
				ExecName:    "cros_ssh",
				ExecTimeout: &durationpb.Duration{Seconds: 150},
				RunControl:  RunControl_ALWAYS_RUN,
			},
			"Wait for pingable": {
				// No recovery actions as that is help action.
				Docs: []string{
					"Wait DUT to be pingable after some action on it.",
					"Waiting time 150 seconds.",
				},
				ExecName:    "cros_ping",
				ExecTimeout: &durationpb.Duration{Seconds: 150},
				RunControl:  RunControl_ALWAYS_RUN,
			},
		},
	}
}
