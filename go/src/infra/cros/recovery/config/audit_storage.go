// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

// CrosAuditStorageConfig audits the internal storage for a ChromeOS DUT.
func CrosAuditStorageConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanCrOSAudit,
		},
		Plans: map[string]*Plan{
			PlanCrOSAudit: {
				CriticalActions: []string{
					"Set state: needs_repair",
					"Device is SSHable",
					"Audit storage (SMART only)",
					"Audit device storage using badblocks",
				},
				Actions: crosRepairActions(),
			},
		},
	}
}
