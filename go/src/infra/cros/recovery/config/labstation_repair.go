// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"google.golang.org/protobuf/types/known/durationpb"
)

// LabstationRepairConfig provides config for repair labstation task.
func LabstationRepairConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanCrOS,
		},
		Plans: map[string]*Plan{
			PlanCrOS: {
				AllowFail: false,
				CriticalActions: []string{
					"dut_state_repair_failed",
					"check_host_info",
					"Device is SSHable",
					"System services is up",
					"Clean up logs if necessary",
					"Filesystem is writable",
					"Check servod dependencies",
					"cros_is_on_stable_version",
					"Update provisioned info",
					"booted_from_right_kernel",
					"reboot_by_request",
					"Reboot labstation if uptime longer than 7 days",
					// TODO(b/245824583): remove this action once the bug fixed.
					"Cleanup bluetooth",
					"Update inventory info",
					"dut_state_ready",
				},
				Actions: map[string]*Action{
					"dut_state_repair_failed": {
						RunControl: RunControl_RUN_ONCE,
					},
					"check_host_info": {
						Docs:     []string{"Check basic info for deployment."},
						ExecName: "sample_pass",
						Dependencies: []string{
							"dut_has_name",
							"dut_has_board_name",
							"dut_has_model_name",
						},
					},
					"cros_is_on_stable_version": {
						Conditions: []string{
							"has_stable_version_cros_image",
							"cros_kernel_priority_has_not_changed",
							"Labstation not in auto-update exempted pool",
						},
						RecoveryActions: []string{
							"Install stable labstation image without reboot",
						},
						AllowFailAfterRecovery: true,
					},
					"Install stable labstation image without reboot": {
						Docs: []string{
							"Install stable labstation image but do not reboot.",
						},
						Conditions: []string{
							"has_stable_version_cros_image",
							"cros_kernel_priority_has_not_changed",
						},
						ExecName: "cros_provision",
						ExecExtraArgs: []string{
							"no_reboot",
						},
						ExecTimeout: &durationpb.Duration{Seconds: 3600},
					},
					"Labstation not in auto-update exempted pool": {
						Docs: []string{
							"There are some labstations we don't want they receive auto-update, e.g. labstations that used for image qualification purpose",
						},
						ExecName: "dut_not_in_pool",
						ExecExtraArgs: []string{
							"servo_verification",
							"labstation_tryjob",
							"labstation_canary",
							"labstation_phone_station",
							"labstation_block_autoupdate",
						},
					},
					"Update provisioned info": {
						Docs: []string{
							"Update OS version for provision info.",
						},
						ExecName:               "cros_update_provision_info",
						AllowFailAfterRecovery: true,
					},
					"labstation_langid_check": {
						Docs: []string{
							"This part is not ready.",
							"The action and will validate present of lang_id issue",
						},
						ExecName:               "sample_pass",
						AllowFailAfterRecovery: true,
					},
					"cros_stop_powerd": {
						ExecName: "cros_run_shell_command",
						ExecExtraArgs: []string{
							"stop",
							"powerd",
						},
						AllowFailAfterRecovery: true,
						RunControl:             RunControl_ALWAYS_RUN,
					},
					"cros_clean_tmp_owner_request": {
						Docs: []string{
							"In some cases, the update flow puts the TPM into a state such that it fails verification.",
							"We don't know why. However, this call papers over the problem by clearing the TPM during the reboot.",
							"We ignore failures from 'crossystem'.",
							"Although failure here is unexpected, and could signal a bug, the point of the exercise is to paper over problems.",
						},
						AllowFailAfterRecovery: true,
						RunControl:             RunControl_ALWAYS_RUN,
					},
					"labstation_uptime_6_hours": {
						ExecName: "cros_validate_uptime",
						ExecExtraArgs: []string{
							"min_duration:6",
						},
					},
					"Remove reboot requests": {
						Docs: []string{
							"Remove all requests for reboot on the host.",
							"The action has to be called after reboot of the device.",
						},
						ExecName:               "cros_remove_all_reboot_request",
						AllowFailAfterRecovery: true,
					},
					"reboot_by_request": {
						Docs: []string{
							"Some DUTs can request reboot labstation if they has issue with servo-nic or other issues with servo-host.",
							"We allowed to remove requests for reboot if we rebooted per request.",
						},
						Conditions: []string{
							"cros_has_reboot_request",
							"cros_has_no_servo_in_use",
							"labstation_uptime_6_hours",
						},
						// If condition passed then action will fail and request recovery actions.
						ExecName: "sample_fail",
						RecoveryActions: []string{
							"Labstation reboot",
							"Power cycle by RPM",
						},
					},
					"Reboot labstation if uptime longer than 7 days": {
						Docs: []string{
							"Check labstation uptime and trigger a reboot if it's longer than 7 days (168 hours).",
						},
						Conditions: []string{
							// No need to run this action if there is servo in use as we don't want reboot interrupt active servos.
							"cros_has_no_servo_in_use",
						},
						ExecName: "cros_validate_uptime",
						ExecExtraArgs: []string{
							"max_duration:168",
						},
						RecoveryActions: []string{
							"Labstation reboot",
							"Power cycle by RPM",
						},
					},
					"booted_from_right_kernel": {
						Docs: []string{
							"Verified if kernel has update and waiting for update.",
							"Kernel can wait for reboot as provisioning is not doing reboot by default for labstations.",
						},
						Conditions: []string{
							"cros_has_no_servo_in_use",
						},
						ExecName: "cros_kernel_priority_has_not_changed",
						RecoveryActions: []string{
							"Labstation reboot",
							"Power cycle by RPM",
						},
					},
					"Device is SSHable": {
						Docs: []string{
							"This verifier checks whether the host is accessible over ssh.",
						},
						RecoveryActions: []string{
							"Power cycle by RPM",
						},
						ExecName:    "cros_ssh",
						ExecTimeout: &durationpb.Duration{Seconds: 30},
						RunControl:  RunControl_ALWAYS_RUN,
					},
					"Filesystem is writable": {
						Docs: []string{
							"This verifier checks whether the host filesystem is writable.",
						},
						ExecName:               "cros_is_file_system_writable",
						AllowFailAfterRecovery: true,
					},
					"Labstation reboot": {
						Docs: []string{
							"Perform reboot of the host and perform additional actions as necessary.",
							"If reboot succeed then we can remove all request for reboot as we just did it.",
						},
						Dependencies: []string{
							"cros_stop_powerd",
							"cros_clean_tmp_owner_request",
							"cros_allowed_reboot",
							"Simple reboot",
							"Sysrq reboot",
							"Sleep 10s",
							// Waiting to tell if success.
							"Wait to be SSHable",
							"Start system services",
							"Remove reboot requests",
						},
						ExecName:   "sample_pass",
						RunControl: RunControl_ALWAYS_RUN,
					},
					"Power cycle by RPM": {
						Docs: []string{
							"Action is always runnable.",
						},
						Conditions: []string{
							"has_rpm_info",
						},
						Dependencies: []string{
							"rpm_power_cycle",
							// Waiting to tell if success.
							"Wait to be SSHable",
							"Start system services",
							"Remove reboot requests",
						},
						ExecName:   "sample_pass",
						RunControl: RunControl_ALWAYS_RUN,
					},
					"Simple reboot": {
						Docs: []string{
							"Simple un-blocker reboot.",
							"The action will not run if the labstation's filesystem I/O is blocked because /sbin/reboot may not work if the filesystem is hosed.",
						},
						Conditions: []string{
							"cros_filesystem_io_not_blocked",
						},
						ExecName: "cros_run_command",
						ExecExtraArgs: []string{
							"command:reboot",
							"background:true",
						},
						RunControl: RunControl_ALWAYS_RUN,
					},
					"Wait to be SSHable": {
						Docs: []string{
							"Try to wait device to be sshable during after the device being rebooted.",
						},
						// Labstation may take some time to fully up(e.g. network service ready) after an update.
						// So giving it 10 minutes in here to allow more buffer.
						ExecTimeout: &durationpb.Duration{Seconds: 600},
						ExecName:    "cros_ssh",
						RunControl:  RunControl_ALWAYS_RUN,
					},
					"Update inventory info": {
						Docs: []string{
							"Updating device info in inventory.",
						},
						ExecName: "sample_pass",
						Dependencies: []string{
							"cros_update_hwid_to_inventory",
							"Read serial number from labstation",
						},
					},
					"Sysrq reboot": {
						Docs: []string{
							"Immediately reboot the system, without unmounting or syncing filesystems",
							"The action only runs when the filesystem is hosed where regular reboot executable will not work.",
						},
						Conditions: []string{
							"Filesystem IO blocked",
						},
						ExecName: "cros_run_shell_command",
						ExecExtraArgs: []string{
							"echo b > /proc/sysrq-trigger",
						},
						RunControl:             RunControl_ALWAYS_RUN,
						AllowFailAfterRecovery: true,
					},
					"Filesystem IO blocked": {
						Docs: []string{
							"Filesystem I/O is blocked on the labstation.",
							"The action is expected to fail when filesystem I/O is not blocked on the labstation.",
						},
						Conditions: []string{
							"cros_filesystem_io_not_blocked",
						},
						ExecName:   "sample_fail",
						RunControl: RunControl_ALWAYS_RUN,
					},
					"Read serial number from labstation": {
						ExecName:               "cros_update_serial_number_inventory",
						AllowFailAfterRecovery: true,
					},
					"Clean up logs if necessary": {
						Docs: []string{
							"Check size of messages logs on labstation and cleanup if necessary.",
						},
						ExecName:               "cros_log_clean_up",
						AllowFailAfterRecovery: true,
					},
					"Attempt to remove bluetooth device": {
						Docs: []string{
							"Attempt to remove bluetooth device from the labstation.",
						},
						ExecName:               "cros_remove_bt_devices",
						AllowFailAfterRecovery: true,
					},
					"Attempt to power off bluetooth adapter": {
						Docs: []string{
							"Attempt to power off bluetooth adapter on the labstation.",
						},
						ExecName: "cros_run_shell_command",
						ExecExtraArgs: []string{
							"bluetoothctl power off",
						},
						AllowFailAfterRecovery: true,
					},
					"Cleanup bluetooth": {
						Docs: []string{
							"Attempt to remove bluetooth device and then power off BT adapter.",
							"This action should be removed once b/245824583 got fixed.",
						},
						Dependencies: []string{
							"Attempt to remove bluetooth device",
							"Attempt to power off bluetooth adapter",
						},
						ExecName:               "sample_pass",
						AllowFailAfterRecovery: true,
					},
					"Labstation image contains target GenesysLogic firmware": {
						Docs: []string{
							"Check if the current labstation OS image contains required GenesysLogic firmware",
						},
						ExecName: "cros_genesys_logic_firmware_image_exists",
					},
					"Update GenesysLogic Firmware for servos": {
						Docs: []string{
							"Attempt to update GenesysLogic firmware for all servos on the labstation if needed.",
							"The update run will be a no-op if a servo is already updated to the target firmware.",
						},
						Conditions: []string{
							"Labstation image contains target GenesysLogic firmware",
						},
						ExecName:               "cros_update_genesys_logic_firmware",
						AllowFailAfterRecovery: true,
					},
					"System services is up": {
						Docs: []string{
							"Check whether system-services is up and running",
						},
						Dependencies: []string{
							"Device is SSHable",
						},
						ExecName: "cros_wait_for_system",
						RecoveryActions: []string{
							"Start system services",
						},
					},
					"Start system services": {
						Docs: []string{
							"Start system-services on the labstation",
						},
						Dependencies: []string{
							"Device is SSHable",
						},
						ExecName: "cros_run_command",
						ExecExtraArgs: []string{
							"host:dut",
							"command:start system-services",
						},
						AllowFailAfterRecovery: true,
						RunControl:             RunControl_ALWAYS_RUN,
					},
					"Write factory-install-reset to file system": {
						ExecName: "cros_run_shell_command",
						ExecExtraArgs: []string{
							"echo \"fast safe\" > /mnt/stateful_partition/factory_install_reset",
						},
						MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
					},
					"Install stable labstation image with reboot": {
						Docs: []string{
							"Install stable labstation image with reboot during provision process.",
						},
						Conditions: []string{
							"has_stable_version_cros_image",
						},
						ExecName:      "cros_provision",
						ExecTimeout:   &durationpb.Duration{Seconds: 3600},
						MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
					},
					"Powerwash repair labstation": {
						Docs: []string{
							"Powerwash and then install stable_version image on the labstation.",
						},
						Conditions: []string{
							"Device is SSHable",
						},
						Dependencies: []string{
							"Write factory-install-reset to file system",
							"Labstation reboot",
							"Install stable labstation image without reboot",
							"Labstation reboot",
						},
						ExecName: "sample_pass",
					},
					"Check servod dependencies": {
						Docs: []string{
							"Ensure critical dependencies for servod is there, this check may fail if labstation had a incomplete provision.",
						},
						Dependencies: []string{
							"Device is SSHable",
						},
						ExecName: "cros_run_command",
						ExecExtraArgs: []string{
							"host:dut",
							"command:servod --sversion",
						},
						AllowFailAfterRecovery: true,
					},
					"Sleep 10s": {
						ExecName: "sample_sleep",
						ExecExtraArgs: []string{
							"sleep:10",
						},
						RunControl:             RunControl_ALWAYS_RUN,
						AllowFailAfterRecovery: true,
						MetricsConfig:          &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
					},
				},
			},
		},
	}
}
