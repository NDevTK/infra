// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"google.golang.org/protobuf/types/known/durationpb"
)

// dolosRepairPlan describe the plan to repair Dolos device.
func dolosRepairPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Set state:BROKEN",
			"Device is pingable",
			"Device is sshable",
			"Set state:WORKING",
		},
		Actions: map[string]*Action{
			"Device is pingable": {
				ExecName:    "cros_ping",
				ExecTimeout: &durationpb.Duration{Seconds: 15},
			},
			"Device is sshable": {
				ExecName:    "cros_ssh",
				ExecTimeout: &durationpb.Duration{Seconds: 30},
			},
			"Set state:BROKEN": {
				ExecName: "set_dolos_state",
				ExecExtraArgs: []string{
					"state:BROKEN",
				},
				MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
			},
			"Set state:WORKING": {
				ExecName: "set_dolos_state",
				ExecExtraArgs: []string{
					"state:WORKING",
				},
				MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
			},
		},
	}
}
