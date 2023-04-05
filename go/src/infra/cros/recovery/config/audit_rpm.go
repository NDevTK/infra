// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

// CrosAuditRPMConfig audits the RPM information for ChromeOS DUTs only.
//
// This is a port of VerifyRPMConfig.
//
// https://chromium.googlesource.com/chromiumos/third_party/labpack/+/refs/heads/main/site_utils/admin_audit/verifiers.py#245 .
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
					// See the link below for the python labpack equivalent.
					//   _check_rpm_power_delivery_without_battery
					//   https://chromium.googlesource.com/chromiumos/third_party/labpack/+/refs/heads/main/site_utils/admin_audit/rpm_validator.py#110
					"Audit RPM config (without battery)",
					// See the link below for the python labpack equivalent.
					//   _check_rpm_power_delivery_with_battery
					//   https://chromium.googlesource.com/chromiumos/third_party/labpack/+/refs/heads/main/site_utils/admin_audit/rpm_validator.py#69
					"Verify RPM config with battery",
				},
				Actions:   crosRepairActions(),
				AllowFail: false,
			},
			PlanClosing: setAllowFail(crosClosePlan(), true),
		},
	}
}
