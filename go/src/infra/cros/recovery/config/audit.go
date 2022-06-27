// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

// CrosAuditRPMConfig is only for ChromeOS DUTs.
func CrosAuditRPMConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanServo,
			PlanCrOS,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			// First, ensure that the servo is in a good state.
			PlanServo: setAllowFail(servoRepairPlan(), true),
			// First thing: set the DUT state to needs_repair
			PlanCrOS: {
				CriticalActions: []string{
					// We defensively set the state to repair failed before every task so that we force
					// a repair once the audit task is complete.
					"Set state: repair_failed",
					"Device is SSHable",
					"Verify RPM config (without battery)",
					"Verify RPM config with battery",
				},
				Actions:   crosRepairActions(),
				AllowFail: false,
			},
			PlanClosing: setAllowFail(crosClosePlan(), true),
		},
	}
}

// CrosAuditStorageConfig audits the internal storage for a ChromeOS DUT.
func CrosAuditStorageConfig() *Configuration {
	crosPlan := crosRepairPlan()
	crosPlan.CriticalActions = []string{
		"Set state: needs_repair",
		"Device is SSHable",
		"Audit storage (SMART only)",
	}
	return &Configuration{
		PlanNames: []string{
			PlanCrOS,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanCrOS:    crosPlan,
			PlanClosing: setAllowFail(crosClosePlan(), true),
		},
	}
}
