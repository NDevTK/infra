// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

// CrosAuditStorageConfig audits the internal storage for a ChromeOS DUT.
func CrosAuditStorageConfig() *Configuration {
	crosPlan := crosRepairPlan()
	crosPlan.CriticalActions = []string{
		"Set state: needs_repair",
		"Device is SSHable",
		"Audit storage (SMART only)",
		"Audit device storage using badblocks",
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
