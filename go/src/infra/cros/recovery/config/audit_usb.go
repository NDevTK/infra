// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

// CrosAuditUSBConfig audits the USB storage for a servo associated with a ChromeOS DUT.
func CrosAuditUSBConfig() *Configuration {
	// In order to audit a servo USB, we will start by repairing the servo, and then
	// add detecting the USB drive and auditing the USB drive as additional required steps.

	// TODO(gregorynisbet): Make the servo plan critical.
	servoPlan := servoRepairPlan()
	servoPlan.CriticalActions = []string{
		// Check for an SSHable device. In the future we will initiate the badblocks check
		// from the device itself and not from the labstation.
		"Device is SSHable",
		// TODO(gregorynisbet): Make the USB drive detectable step critical.
		"Verify that USB drive is detectable",
		// TODO(gregorynisbet): Run only from DUT side.
		//
		// TODO(gregorynisbet): The Audit-of-USB-drive action would work here, but it would
		//                      make more sense for it to be part of the CrOS plan. Fix this
		//                      in a subsequent CL.
		// "Audit of USB drive",
	}

	crosPlan := &Plan{
		CriticalActions: []string{
			// We defensively set the state to needs repair before every task so that we force
			// a repair once the audit task is complete.
			"Set state: needs_repair",
			"Device is SSHable",
		},
		Actions:   crosRepairActions(),
		AllowFail: false,
	}

	return &Configuration{
		PlanNames: []string{
			PlanServo,
			PlanCrOS,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanCrOS:    setAllowFail(crosPlan, true),
			PlanServo:   setAllowFail(servoPlan, true),
			PlanClosing: setAllowFail(crosClosePlan(), true),
		},
	}
}
