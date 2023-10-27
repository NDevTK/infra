// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

// CrosAuditUSBConfig audits the USB drive for a servo associated with a ChromeOS DUT.
func CrosAuditUSBConfig() *Configuration {

	return &Configuration{
		PlanNames: []string{
			// First we prepare servo to work on setup.
			// Make sure that the servo is in a good state and servod is up.
			PlanServo,
			// Core plan to validate access and check USB-drive.
			PlanCrOSAudit,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanServo: setAllowFail(servoRepairPlan(), false),
			PlanCrOSAudit: {
				CriticalActions: []string{
					// We defensively set the state to needs repair before every task so that we force
					// a repair once the audit task is complete.
					"Set state: needs_repair",
					// Check that we can SSH to the DUT in question.
					"Device is SSHable",
					// Attempt to audit the USB from the DUT side.
					"Audit USB-drive from DUT",
				},
				// We use CrOS repair actions as it has all action it requires.
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
