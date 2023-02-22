// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import "google.golang.org/protobuf/types/known/durationpb"

// CrosBrowserDUTDeployConfig provides config for deploy tasks for ChromeOS DUTs for Browser testing.
func CrosBrowserDUTDeployConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{PlanCrOS},
		Plans: map[string]*Plan{
			PlanCrOS: {
				AllowFail: false,
				CriticalActions: []string{
					"Set state: needs_deploy",
					"Check stable versions exist",
					"Device is SSHable",
					"DUT has expected dev firmware (for browser DUTs)",
					"Update inventory info",
					"Set state: ready",
				},
				Actions: CrOSBrowserDUTDeployActions(),
			},
		},
	}
}

func CrOSBrowserDUTDeployActions() map[string]*Action {
	browserActions := CrOSBrowserDUTRepairActions()
	browserActions["Update inventory info"] = &Action{
		Docs: []string{
			"Updating device info in inventory.",
		},
		ExecName: "sample_pass",
		Dependencies: []string{
			"cros_ssh",
			"cros_update_hwid_to_inventory",
			"cros_update_serial_number_inventory",
		},
	}
	browserActions["DUT has expected dev firmware (for browser DUTs)"] = &Action{
		Docs: []string{
			"Verify that FW on the DUT has dev keys.",
		},
		Conditions: []string{
			"dut_is_not_browser_legacy_duts",
		},
		Dependencies: []string{
			"Device is SSHable",
		},
		ExecName:    "cros_has_dev_signed_firmware",
		ExecTimeout: &durationpb.Duration{Seconds: 600},
		RecoveryActions: []string{
			"Update DUT firmware with factory mode and restart by host",
		},
	}
	for name, action := range deployActions() {
		if _, ok := browserActions[name]; ok {
			continue
		}
		browserActions[name] = action
	}
	return browserActions
}
