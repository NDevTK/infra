// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"google.golang.org/protobuf/types/known/durationpb"
)

// LabstationDeployConfig provides config for deploy labstation task.
func LabstationDeployConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{PlanCrOS},
		Plans: map[string]*Plan{
			PlanCrOS: {
				AllowFail: false,
				CriticalActions: []string{
					"dut_state_needs_deploy",
					"check_host_info",
					"Device is SSHable",
					"Update inventory info",
					"Installed OS is stable",
					"Remove reboot requests from host",
					"Update provisioned info",
					"Validate RPM info",
					"dut_state_ready",
				},
				Actions: map[string]*Action{
					"check_host_info": {
						Docs: []string{
							"Check basic info for deployment.",
						},
						ExecName: "sample_pass",
						Dependencies: []string{
							"dut_has_name",
							"dut_has_board_name",
							"dut_has_model_name",
						},
					},
					"Update inventory info": {
						Docs: []string{
							"Updating device info in inventory.",
						},
						ExecName: "sample_pass",
						Dependencies: []string{
							"cros_ssh",
							"cros_update_hwid_to_inventory",
							"cros_update_serial_number_inventory",
						},
					},
					"Installed OS is stable": {
						Docs: []string{
							"Verify that OS on the device is stable.",
							"Labstation will be rebooted to make it ready for use.",
						},
						Conditions: []string{
							"has_stable_version_cros_image",
						},
						ExecName: "cros_is_on_stable_version",
						RecoveryActions: []string{
							"Install stable OS",
							"Power cycle by RPM",
						},
					},
					"Install stable OS": {
						Docs: []string{
							"Install stable OS on the device.",
							"Labstation will be rebooted to make it ready for use.",
						},
						Conditions: []string{
							"has_stable_version_cros_image",
							"cros_not_on_stable_version",
						},
						ExecName:    "cros_provision",
						ExecTimeout: &durationpb.Duration{Seconds: 3600},
					},
					"Remove reboot requests from host": {
						Docs: []string{
							"Remove reboot request flag files.",
						},
						ExecName:               "cros_remove_all_reboot_request",
						AllowFailAfterRecovery: true,
					},
					"Update provisioned info": {
						Docs: []string{
							"Update OS version for provision info.",
						},
						ExecName: "cros_update_provision_info",
					},
					"Validate RPM info": {
						Docs: []string{
							"Validate and update rpm_state.",
							"The execs is not ready yet.",
						},
						Conditions: []string{
							"has_rpm_info",
						},
						ExecName:    "rpm_audit_without_battery",
						ExecTimeout: &durationpb.Duration{Seconds: 900},
						RecoveryActions: []string{
							"Power cycle by RPM",
						},
					},
					"Device is SSHable": {
						Docs: []string{
							"This verifier checks whether the host is accessible over ssh.",
						},
						ExecName:    "cros_ssh",
						ExecTimeout: &durationpb.Duration{Seconds: 30},
						RunControl:  RunControl_ALWAYS_RUN,
						RecoveryActions: []string{
							"Power cycle by RPM",
						},
					},
					"Power cycle by RPM": {
						Docs: []string{
							"Action is always runnable.",
						},
						Conditions: []string{
							"has_rpm_info",
						},
						ExecName:   "rpm_power_cycle",
						RunControl: RunControl_ALWAYS_RUN,
					},
				},
			},
		},
	}
}
