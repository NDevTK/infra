// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

// CrosAuditRPMConfig audits the RPM information for ChromeOS DUTs only.
func CrosAuditRPMConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanServo,
			PlanCrOSAudit,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanServo: setAllowFail(servoRepairPlan(), true),
			PlanCrOSAudit: {
				CriticalActions: []string{
					// Set a repair failed state to call auto-repair after that task is completed.
					"Set state: needs_repair",
					"Device is SSHable",
					"Verify RPM config",
				},
				Actions:   crosRepairActions(),
				AllowFail: false,
			},
			PlanClosing: {
				CriticalActions: []string{
					"Close Servo-host",
				},
				Actions:   crosRepairClosingActions(),
				AllowFail: true,
			},
		},
	}
}
