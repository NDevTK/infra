// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"google.golang.org/protobuf/types/known/durationpb"
)

func crosRepairPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Set state: repair_failed",
			"Device is SSHable",
			"Verify internal storage",
			"Check if last provision was good",
			"Verify system info",
			"Python is present",
			"Verify that device is not enrolled",
			"Check power sources",
			"Check TPM statuses",
			"Verify present of gsctool",
			"Audit wifi",
			"Audit bluetooth",
			"Firmware validations",
			"Login UI is up",
			"Verify keys of RW_VPD",
			"Update Servo NIC mac address",
			"Match provision labels",
			"Set state: ready",
			"Update special device labels",
			"Collect dmesg logs from DUT",
		},
		Actions: crosRepairActions(),
	}
}

func crosRepairActions() map[string]*Action {
	return map[string]*Action{
		"Set state: ready": {
			Docs: []string{
				"The action set devices with state ready for the testing.",
			},
			ExecName: "dut_set_state",
			ExecExtraArgs: []string{
				"state:ready",
			},
			RunControl: RunControl_RUN_ONCE,
		},
		"Set state: repair_failed": {
			Docs: []string{
				"The action set devices with state means that repair tsk did not success to recover the devices.",
			},
			ExecName: "dut_set_state",
			ExecExtraArgs: []string{
				"state:repair_failed",
			},
			RunControl: RunControl_RUN_ONCE,
		},
		"Set state: needs_deploy": {
			Docs: []string{
				"The action set devices with request to be redeployed.",
			},
			ExecName: "dut_set_state",
			ExecExtraArgs: []string{
				"state:needs_deploy",
			},
		},
		"Device is pingable": {
			Docs: []string{
				"Verify that device is reachable by ping.",
				"Limited to 15 seconds.",
			},
			Dependencies: []string{
				"dut_has_board_name",
				"dut_has_model_name",
			},
			ExecName: "cros_ping",
			ExecTimeout: &durationpb.Duration{
				Seconds: 15,
			},
			RecoveryActions: []string{
				"Power cycle DUT by RPM",
				"Cold reset by servo and wait for SSH",
				"Cr50 reset by servo wait for SSH",
				"Trigger kernel panic to reset the whole board and try ssh to DUT",
				"Update FW from fw-image by servo and reboot",
				"Restore AC detection by EC console",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
				"Reset power using servo if booted from USB",
				"Update firmware from USB-Drive and when booted in recovery mode",
			},
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Device is SSHable": {
			Docs: []string{
				"Verify that device is reachable by SSH.",
				"Limited to 15 seconds.",
			},
			Dependencies: []string{
				"Device is pingable",
			},
			ExecName:    "cros_ssh",
			ExecTimeout: &durationpb.Duration{Seconds: 15},
			RecoveryActions: []string{
				"Power cycle DUT by RPM",
				"Cold reset by servo and wait for SSH",
				"Cr50 reset by servo wait for SSH",
				"Trigger kernel panic to reset the whole board and try ssh to DUT",
				"Update FW from fw-image by servo and reboot",
				"Restore AC detection by EC console",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
				"Reset power using servo if booted from USB",
				"Update firmware from USB-Drive and when booted in recovery mode",
			},
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Verify internal storage": {
			Docs: []string{
				"Verify DUT internal storage",
			},
			Dependencies: []string{
				"Internal storage is responsive",
				"Kernel does not know issues",
				"Stateful partition has enough free index nodes",
				"Stateful partition has enough free space",
				"Stateful partition (encrypted) has enough free space",
				"Quick internal storage audit (SMART)",
				"Check DUT state and fail if needs replacement",
			},
			ExecName: "sample_pass",
		},
		"Check DUT state and fail if needs replacement": {
			Docs: []string{
				"Verify DUT internal storage",
			},
			ExecName: "dut_state_match",
			ExecExtraArgs: []string{
				"state:needs_replacement",
				"invert:true",
			},
		},
		"Quick internal storage audit (SMART)": {
			Docs: []string{
				"Quick audit internal storage by reading SMART data.",
				"The check updates storage state.",
				"The check is not critical as update storage state.",
			},
			ExecName:               "cros_audit_storage_smart",
			AllowFailAfterRecovery: true,
		},
		"Verify system info": {
			Conditions: []string{
				"Is not Flex device",
			},
			Dependencies: []string{
				"Default boot set as internal storage",
				"Verify that DUT is not in DEV mode",
				"Missing HWID",
				"Missing serial-number",
				"Match HWID",
				"Match serial-number",
				"Verify tmp_fwver is updated correctly",
				"Verify tpm_kernver is updated correctly",
			},
			ExecName: "sample_pass",
		},
		"Read OS version": {
			Docs: []string{
				"Read and log current OS version.",
			},
			ExecName: "cros_read_os_version",
			RecoveryActions: []string{
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Python is present": {
			Docs: []string{
				"Verify that device has python on it.",
			},
			Conditions: []string{
				// TODO(b/232854289): remove when closed the bug.
				"Is not Flex device",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			ExecName: "cros_has_python_interpreter_working",
			RecoveryActions: []string{
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Check if last provision was good": {
			Docs: []string{
				"Check if last provision fail on the DUT",
			},
			Dependencies: []string{
				"Internal storage is responsive",
				"Read OS version",
			},
			ExecName: "cros_is_last_provision_successful",
			RecoveryActions: []string{
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Verify that device is not enrolled": {
			Docs: []string{
				"Verify that the device's enrollment state is clean.",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			ExecName: "cros_is_enrollment_in_clean_state",
			RecoveryActions: []string{
				"Cleanup the enrollment state and wait for boot",
			},
			AllowFailAfterRecovery: true,
		},
		"Cleanup the enrollment state and wait for boot": {
			Docs: []string{
				"Cleanup the enrollment state.",
				"The recovery process can fail but still fix the issue.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_enrollment_cleanup",
			ExecExtraArgs: []string{
				"repair_timeout:120",
				"clear_tpm_owner_timeout:60",
				"file_deletion_timeout:120",
				"reboot_timeout:10",
				"tpm_timeout:150",
			},
			ExecTimeout:            &durationpb.Duration{Seconds: 600},
			AllowFailAfterRecovery: true,
		},
		"Check power sources": {
			Docs: []string{"Check for the AC power, and battery charging capability."},
			Conditions: []string{
				"Is not Flex device",
				"cros_is_not_virtual_machine",
			},
			Dependencies: []string{
				"Power is recognized by DUT",
				"Battery is changing or have accepted level",
				"Audit battery",
				"Check DUT state and fail if needs replacement",
			},
			ExecName: "sample_pass",
		},
		"Battery is changing or have accepted level": {
			Docs: []string{
				"Check the battery charging state.",
			},
			Conditions: []string{
				"cros_is_battery_expected",
				"cros_is_not_virtual_machine",
				"Battery is expected on device",
				"Battery is present on device",
			},
			ExecName: "cros_is_battery_chargable_or_good_level",
			RecoveryActions: []string{
				"Power cycle DUT by RPM and wait",
				"Cold reset by servo and wait for SSH",
				"Cr50 reset by servo wait for SSH",
				"Restore AC detection by EC console",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Power is recognized by DUT": {
			Docs: []string{
				"Verify that power is recognized by the DUT as marker as online.",
			},
			ExecName: "cros_is_ac_power_connected",
			RecoveryActions: []string{
				"Power cycle DUT by RPM and wait",
				"Cold reset by servo and wait for SSH",
				"Cr50 reset by servo wait for SSH",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Check TPM statuses": {
			Docs: []string{
				"Verify that TPM statuses is ok.",
			},
			Conditions: []string{
				"Is not Flex device",
				"cros_is_not_virtual_machine",
				"cros_is_tpm_present",
			},
			ExecName: "cros_is_tpm_in_good_status",
			RecoveryActions: []string{
				"ChromeOS TMP recovery (not critical)",
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Firmware validations": {
			Docs: []string{
				"Group action to combine all firmware checks in one place.",
			},
			Conditions: []string{
				"Is not Flex device",
			},
			Dependencies: []string{
				"Ensure firmware is in good state",
				"RW Firmware version matches the recovery-version",
				"Verify servo keyboard firmware",
			},
			ExecName: "sample_pass",
		},
		"Ensure firmware is in good state": {
			Docs: []string{
				"Ensure that firmware is in good state.",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			ExecName: "cros_is_firmware_in_good_state",
			RecoveryActions: []string{
				"Cold reset DUT by servo",
				"Fix FW on the DUT to match stable-version",
				"Update FW from fw-image by servo",
				"Update firmware from USB-Drive and when booted in recovery mode",
			},
		},
		"Login UI is up": {
			Docs: []string{
				"Check the command 'stop ui' won't crash the DUT.",
			},
			ExecName:    "cros_stop_start_ui",
			ExecTimeout: &durationpb.Duration{Seconds: 45},
			RecoveryActions: []string{
				// TODO: need action to start service only.
				"Cold reset by servo and wait for SSH",
				"Cr50 reset by servo wait for SSH",
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Verify keys of RW_VPD": {
			Docs: []string{
				"Verify that keys: 'should_send_rlz_ping', 'gbind_attribute', 'ubind_attribute' are present in vpd RW_VPD partition.",
			},
			Conditions: []string{
				"Is not Flex device",
			},
			ExecName: "cros_are_required_rw_vpd_keys_present",
			RecoveryActions: []string{
				// TODO: Need run tmp reset.
				// TODO: Missed original recovery action to set key values.
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
			},
			AllowFailAfterRecovery: true,
		},
		"Verify servo keyboard firmware": {
			Conditions: []string{
				"dut_servo_host_present",
				"servo_state_is_working",
				"is_servo_keyboard_image_tool_present",
			},
			Dependencies: []string{
				"servo_init_usb_keyboard",
				"lufa_keyboard_found",
			},
			ExecName: "cros_run_shell_command",
			ExecExtraArgs: []string{
				"lsusb -vv -d 03eb:2042 |grep \"Remote Wakeup\"",
			},
			RecoveryActions: []string{
				// TODO: update keyboard firmware.
			},
			AllowFailAfterRecovery: true,
		},
		"Update Servo NIC mac address": {
			Conditions: []string{
				"dut_servo_host_present",
				"Is not servo_v3",
				"servod_control_exist_for_mac_address",
			},
			ExecName:               "servo_audit_nic_mac_address",
			AllowFailAfterRecovery: true,
		},
		"Is not servo_v3": {
			Docs: []string{
				"Verify that servo_v3 isn ot used in setup.",
			},
			Conditions: []string{
				"is_servo_v3",
			},
			ExecName: "sample_fail",
		},
		"servod_control_exist_for_mac_address": {
			Conditions: []string{
				"dut_servo_host_present",
			},
			ExecName: "servo_check_servod_control",
			ExecExtraArgs: []string{
				"command:macaddr",
			},
		},
		"servo_init_usb_keyboard": {
			Docs: []string{
				"set servo's 'init_usb_keyboard' command to 'on' value.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
			},
			ExecName: "servo_set",
			ExecExtraArgs: []string{
				"command:init_usb_keyboard",
				"string_value:on",
			},
		},
		"is_servo_keyboard_image_tool_present": {
			Docs: []string{
				"check if the servo keyboard image specified by the name of dfu-programmer can be found in DUT cli.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
			},
			ExecName: "cros_is_tool_present",
			ExecExtraArgs: []string{
				"tools:dfu-programmer",
			},
		},
		"lufa_keyboard_found": {
			Docs: []string{
				"check if the lufa keyboard can be found by finding the match of the model information of it.",
			},
			ExecName: "cros_run_shell_command",
			ExecExtraArgs: []string{
				"lsusb -d 03eb:2042 |grep \"LUFA Keyboard Demo\"",
			},
		},
		"servo_state_is_working": {
			Docs: []string{
				"check the servo's state is WORKING.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
			},
			ExecName: "servo_match_state",
			ExecExtraArgs: []string{
				"state:WORKING",
			},
		},
		"servo_state_is_not_working": {
			Docs: []string{
				"check the servo's state is not WORKING.",
			},
			Conditions: []string{
				"servo_state_is_working",
			},
			ExecName: "sample_fail",
		},
		"RW Firmware version matches the recovery-version": {
			Docs: []string{
				"Check if the version of RW firmware on DUT matches the stable firmware version.",
			},
			Conditions: []string{
				"Recovery version has firmware version",
				"Pools required to manage FW on the device",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			ExecName: "cros_is_on_rw_firmware_stable_version",
			RecoveryActions: []string{
				"Fix FW on the DUT to match stable-version",
				"Update FW from fw-image by servo",
			},
		},
		"Fix FW on the DUT to match stable-version": {
			Docs: []string{
				"Run Fw update from the DUT to set expected FW version on the DUT",
				"Update FW required the DUT to be run on stable-version OS.",
			},
			Conditions: []string{
				"Recovery version has OS image path",
				"Recovery version has firmware version",
			},
			Dependencies: []string{
				"Provision OS if needed",
				"Check if stable version is availabe on installed OS",
				"Disable software-controlled write-protect for 'host'",
				"Disable software-controlled write-protect for 'ec'",
			},
			ExecName:    "cros_run_firmware_update",
			ExecTimeout: &durationpb.Duration{Seconds: 900},
			ExecExtraArgs: []string{
				"mode:recovery",
				"force:true",
				"reboot:by_host",
				"updater_timeout:600",
			},
		},
		"Provision OS if needed": {
			Docs: []string{
				"Perfrom provision OS if device is not running on it.",
			},
			Conditions: []string{
				"Recovery version has OS image path",
				"cros_not_on_stable_version",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName:    "cros_provision",
			ExecTimeout: &durationpb.Duration{Seconds: 3600},
		},
		"Verify present of gsctool": {
			Docs: []string{
				"Confirm that the GSC tool is function.",
				"Applicable only if device has Google security chip.",
			},
			Conditions: []string{
				//TODO(b:231609148: Flex device don't have security chip and gsctool.
				"Is not Flex device",
				"DUT has Cr50 phase label",
			},
			ExecName: "cros_is_gsc_tool_present",
			RecoveryActions: []string{
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Audit battery": {
			Docs: []string{
				"Check battery on the DUT is normal and update battery hardware state accordingly.",
			},
			Conditions: []string{
				"cros_is_battery_expected",
				"cros_is_not_virtual_machine",
				"Battery is expected on device",
				"Battery is present on device",
				"Internal storage is responsive",
			},
			// TODO: verify if the battery state is for replacement and escape from the finish the process.
			ExecName: "cros_audit_battery",
		},
		"Battery is expected on device": {
			Docs: []string{
				"Verifies that device is expected to have battery based on DUT info.",
			},
			ExecName: "dut_has_battery",
		},
		"Battery is present on device": {
			Docs: []string{
				"Verifies if battery present is reported as present in power supply info.",
			},
			ExecName:   "cros_is_battery_present",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"No Battery is present on device": {
			Conditions: []string{
				"Battery is present on device",
			},
			ExecName: "sample_fail",
		},
		"Audit wifi": {
			Docs: []string{
				"Check wifi on the DUT is normal and update wifi hardware state accordingly.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_audit_wifi",
			RecoveryActions: []string{
				"Cold reset by servo and wait for SSH",
			},
			AllowFailAfterRecovery: true,
		},
		"Audit bluetooth": {
			Docs: []string{
				"Check bluetooth on the DUT is normal and update bluetooth hardware state accordingly.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_audit_bluetooth",
			RecoveryActions: []string{
				"Cold reset by servo and wait for SSH",
			},
			AllowFailAfterRecovery: true,
		},
		"Verify tmp_fwver is updated correctly": {
			Docs: []string{
				"For dev-signed firmware, tpm_fwver reported from crossystem should always be 0x10001.",
				"Firmware update on DUTs with incorrect tmp_fwver may fail due to firmware rollback protection.",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			Conditions: []string{
				"Is not Flex device",
			},
			ExecName: "cros_match_dev_tpm_firmware_version",
			RecoveryActions: []string{
				"ChromeOS TMP recovery (not critical)",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Verify tpm_kernver is updated correctly": {
			Docs: []string{
				"For dev-signed firmware, tpm_kernver reported from crossystem should always be 0x10001.",
				"Firmware update on DUTs with incorrect tpm_kernver may fail due to firmware rollback protection.",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			Conditions: []string{
				"Is not Flex device",
			},
			ExecName: "cros_match_dev_tpm_kernel_version",
			RecoveryActions: []string{
				"ChromeOS TMP recovery (not critical)",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"ChromeOS TMP recovery (not critical)": {
			Docs: []string{
				"Run chromeos-tpm-recovery on DUT to reset TPM.",
				"That is experimental recovery action.",
			},
			ExecName: "cros_run_shell_command",
			ExecExtraArgs: []string{
				"chromeos-tpm-recovery",
			},
			AllowFailAfterRecovery: true,
		},
		"Default boot set as internal storage": {
			Docs: []string{
				"Check if the default boot drive is disk.",
			},
			Conditions: []string{
				"Is not Flex device",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			ExecName: "cros_is_default_boot_from_disk",
			RecoveryActions: []string{
				"Set default boot as disk",
				"Quick provision OS",
			},
		},
		"Verify that DUT is not in DEV mode": {
			Docs: []string{
				"Verify that devices is not in DEV mode.",
				"Mostly devices in the lab required to be in Secure mode, not DEV mode.",
			},
			Conditions: []string{
				"Is not Flex device",
				"Pools required to be in Secure mode",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			ExecName: "cros_is_not_in_dev_mode",
			RecoveryActions: []string{
				"Switch to secure-mode and reboot",
				"Quick provision OS",
			},
		},
		"Missing HWID": {
			Docs: []string{
				"Verify if device missing HWID because deployment was missed.",
			},
			Conditions: []string{
				"Is not Flex device",
				"Not Satlab device",
				"Read OS version",
				"Is HWID empty",
			},
			Dependencies: []string{
				"Set state: needs_deploy",
			},
			ExecName: "sample_fail",
		},
		"Match HWID": {
			Docs: []string{
				"Match HWID to value in inventory",
			},
			Conditions: []string{
				"Is not Flex device",
				"Read OS version",
				"Is HWID known",
			},
			ExecName:               "cros_match_hwid_to_inventory",
			AllowFailAfterRecovery: true,
		},
		"Missing serial-number": {
			Docs: []string{
				"Verify if device missing serial number because deployment was missed.",
			},
			Conditions: []string{
				"Is not Flex device",
				"Not Satlab device",
				"Read OS version",
				"Is serial-number empty",
			},
			Dependencies: []string{
				"Set state: needs_deploy",
			},
			ExecName: "sample_fail",
		},
		"Match serial-number": {
			Docs: []string{
				"Match serial number to value in inventory",
			},
			Conditions: []string{
				"Is not Flex device",
				"Read OS version",
				"Is serial-number known",
			},
			ExecName:               "cros_match_serial_number_inventory",
			AllowFailAfterRecovery: true,
		},
		"Is HWID known": {
			Docs: []string{
				"Check whether the DUT information includes its HWID.",
			},
			ExecName: "dut_has_hwid",
		},
		"Is HWID empty": {
			Docs: []string{
				"Check whether the DUT information includes its HWID.",
			},
			Conditions: []string{
				"Is HWID known",
			},
			ExecName: "sample_fail",
		},
		"Is serial-number known": {
			Docs: []string{
				"Check whether the DUT information includes its ",
				"serial number.",
			},
			ExecName: "dut_has_serial_number",
		},
		"Is serial-number empty": {
			Docs: []string{
				"Check whether the DUT information includes its ",
				"serial number.",
			},
			Conditions: []string{
				"Is serial-number known",
			},
			ExecName: "sample_fail",
		},
		"Not Satlab device": {
			Docs: []string{
				"Verify that DUT name is not belong Satlab.",
			},
			Conditions: []string{
				"Is Satlab device",
			},
			ExecName: "sample_fail",
		},
		"Is Satlab device": {
			Docs: []string{
				"Verify that DUT name is belong Satlab.",
			},
			ExecName: "dut_regex_name_match",
			ExecExtraArgs: []string{
				"regex:^satlab",
			},
		},
		"Read DUT serial-number from DUT": {
			Conditions: []string{
				"Not Satlab device",
			},
			ExecName: "cros_update_serial_number_inventory",
		},
		"Read DUT serial-number from DUT (Satlab)": {
			Conditions: []string{
				"Is Satlab device",
			},
			ExecName:               "cros_update_serial_number_inventory",
			AllowFailAfterRecovery: true,
		},
		"Read HWID from DUT": {
			Conditions: []string{
				"Not Satlab device",
			},
			ExecName: "cros_update_hwid_to_inventory",
		},
		"Read HWID from DUT (Satlab)": {
			Conditions: []string{
				"Is Satlab device",
			},
			ExecName:               "cros_update_hwid_to_inventory",
			AllowFailAfterRecovery: true,
		},
		"Internal storage is responsive": {
			Docs: []string{
				"Verify that internal storage is responsive",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_is_file_system_writable",
			RecoveryActions: []string{
				"Cold reset by servo and wait for SSH",
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Kernel does not know issues": {
			Docs: []string{
				"Verify instrenal storage is writable.",
				"If linux has some hardware error then file system can became read-only.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_has_critical_kernel_error",
			RecoveryActions: []string{
				"Cold reset by servo and wait for SSH",
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Stateful partition has enough free index nodes": {
			Docs: []string{
				"Check the stateful partition path has enough index nodes.",
			},
			ExecName: "cros_has_enough_index_nodes",
			ExecExtraArgs: []string{
				"/mnt/stateful_partition:100",
			},
			RecoveryActions: []string{
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Stateful partition has enough free space": {
			Docs: []string{
				"Check the stateful partition have enough disk space.",
				"Expected to have free 0.1GB storage unit.",
			},
			ExecName: "cros_has_enough_storage_space",
			ExecExtraArgs: []string{
				"/mnt/stateful_partition:0.7",
			},
			RecoveryActions: []string{
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Stateful partition (encrypted) has enough free space": {
			Docs: []string{
				"Check the encrypted stateful partition have enough disk space.",
				"Expected to have free 0.1GB storage unit.",
			},
			ExecName: "cros_has_enough_storage_space",
			ExecExtraArgs: []string{
				"/mnt/stateful_partition/encrypted:0.1",
			},
			RecoveryActions: []string{
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive (for special pools)",
			},
		},
		"Update special device labels": {
			Docs: []string{
				"Read special labels everytime as part of repair process.",
			},
			Dependencies: []string{
				"Read device SKU",
				"Read Cr50 PHASE",
				"Read Cr50 key ID",
				"Read if audio loopback present",
			},
			ExecName: "sample_pass",
		},
		"Read Cr50 PHASE": {
			Docs: []string{
				"Update the cr50 phase label from device.",
			},
			Conditions: []string{
				"Cr50 firmware present on the device",
			},
			ExecName:               "cros_update_cr50_label",
			AllowFailAfterRecovery: true,
		},
		"Read Cr50 key ID": {
			Docs: []string{
				"Update the cr50 key ID from device.",
			},
			Conditions: []string{
				"Cr50 firmware present on the device",
			},
			ExecName:               "cros_update_cr50_key_id_label",
			AllowFailAfterRecovery: true,
		},
		"Read if audio loopback present": {
			Docs: []string{
				"Update the audio_loop_back label on the cros Device.",
			},
			Conditions: []string{
				"Audio loopback state is not WORKING",
			},
			ExecName:               "cros_update_audio_loopback_state_label",
			AllowFailAfterRecovery: true,
		},
		"Audio loopback state is not WORKING": {
			Docs: []string{
				"Confirm that the DUT's audio loopback state is in not working state",
			},
			Conditions: []string{
				"cros_is_audio_loopback_state_working",
			},
			ExecName: "sample_fail",
		},
		"Cr50 firmware present on the device": {
			Docs: []string{
				"Checks if the cr 50 firmware exists on the DUT by running the gsctool version command.",
				"The action working as condition. Please do not exclude based on labels.",
			},
			Dependencies: []string{
				"Read OS version",
			},
			ExecName: "cros_run_shell_command",
			ExecExtraArgs: []string{
				"gsctool -a -f",
			},
		},
		"DUT has Cr50 phase label": {
			Docs: []string{
				"Check if the DUT has Cr50.",
			},
			ExecName: "dut_has_cr50",
		},
		"Read device SKU": {
			Docs: []string{
				"Update the device_sku label from the device if not present in inventory data.",
			},
			Conditions: []string{
				"dut_does_not_have_device_sku",
			},
			ExecName:               "cros_update_device_sku",
			AllowFailAfterRecovery: true,
		},
		"dut_does_not_have_device_sku": {
			Docs: []string{
				"Confirm that the DUT itself does not have device_sku label.",
			},
			Conditions: []string{
				"dut_has_device_sku",
			},
			ExecName: "sample_fail",
		},
		"Servo USB-Key needs to be reflashed": {
			Docs: []string{
				"Check if it is time to download image to servo usbkey.",
				"If so, then download the stable image to usbkey.",
			},
			Conditions: []string{
				"cros_is_time_to_force_download_image_to_usbkey",
			},
			Dependencies: []string{
				"Download stable image to USB-key",
			},
			ExecName: "sample_pass",
		},
		"Servo has USB-key with require image": {
			Docs: []string{
				"USB-drive contains stable image on it.",
			},
			Conditions: []string{
				"dut_servo_host_present",
				"servo_state_is_working",
			},
			ExecName:    "servo_usbkey_has_stable_image",
			ExecTimeout: &durationpb.Duration{Seconds: 120},
			RecoveryActions: []string{
				"Download stable image to USB-key",
			},
		},
		"Download stable image to USB-key": {
			Docs: []string{
				"Download lab stable image on servo USB-key",
				"Download the image can take longer if labstation download parallel a few images.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
				"servo_state_is_working",
			},
			ExecName:    "servo_download_image_to_usb",
			ExecTimeout: &durationpb.Duration{Seconds: 3000},
			RunControl:  RunControl_ALWAYS_RUN,
		},
		"cros_is_time_to_force_download_image_to_usbkey": {
			Docs: []string{
				"Check if it is time to force download image to usbkey",
				"from the number of failed recoveries since last successful PARIS repair task.",
			},
			ExecExtraArgs: []string{
				"task_name:recovery",
				"repair_failed_count:1",
				"repair_failed_interval:10",
			},
		},
		"Match provision labels": {
			Docs: []string{
				"Verify that provision labels is correct.",
			},
			Dependencies: []string{
				"Match CrOS version with provision label",
				"Match job_repo_url with provision label",
			},
			ExecName: "sample_pass",
		},
		"Match CrOS version with provision label": {
			Docs: []string{
				"Verify that cros-version match version on the host.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_match_cros_version_to_inventory",
			RecoveryActions: []string{
				"Update provisioned info",
			},
		},
		"Match job_repo_url with provision label": {
			Docs: []string{
				"Verify that job_repo_url matches the version on the host.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_match_job_repo_url_version_to_inventory",
			RecoveryActions: []string{
				"Update provisioned info",
			},
		},
		"Update provisioned info": {
			Docs: []string{
				"Read and update cros-provision labels.",
			},
			Dependencies: []string{
				"cros_update_provision_os_version",
				"cros_update_job_repo_url",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Switch to secure-mode and reboot": {
			Docs: []string{
				"This repair action utilizes the dependent actions to set the",
				" GBB flags and disable booting into dev-mode. Then it reboots",
				" the DUT.",
			},
			Conditions: []string{
				//TODO(b:231640496): flex board unpingable after switching to secure-mode.
				"Is not Flex device",
				"Pools required to be in Secure mode",
			},
			Dependencies: []string{
				"Reset GBB flags by host",
				"cros_switch_to_secure_mode",
				"Simple reboot",
				"Wait to be pingable (normal boot)",
			},
			ExecName: "sample_pass",
		},
		"Reset GBB flags by host": {
			Docs: []string{
				"This action sets the GBB flags to 0x0.",
			},
			ExecName:               "cros_set_gbb_flags",
			ExecTimeout:            &durationpb.Duration{Seconds: 3600},
			AllowFailAfterRecovery: true,
		},
		"cros_switch_to_secure_mode": {
			Docs: []string{
				"This action disables booting into dev-mode.",
			},
			ExecTimeout:            &durationpb.Duration{Seconds: 3600},
			AllowFailAfterRecovery: true,
		},
		"Is not Flex device": {
			Docs: []string{
				"Verify that device is belong Reven models",
			},
			ExecExtraArgs: []string{
				"string_values:reven",
				"invert_result:true",
			},
			ExecName: "dut_check_board",
		},
		"Quick provision OS": {
			Docs: []string{
				"Install stable OS on the device.",
			},
			Conditions: []string{
				"Recovery version has OS image path",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName:    "cros_provision",
			ExecTimeout: &durationpb.Duration{Seconds: 3600},
		},
		"Wait to be SSHable (normal boot)": {
			// No recovery actions as that is help action.
			Docs: []string{
				"Try to wait device to be sshable after the device being rebooted.",
				"Waiting time 150 seconds.",
			},
			ExecName:    "cros_ssh",
			ExecTimeout: &durationpb.Duration{Seconds: 150},
			RunControl:  RunControl_ALWAYS_RUN,
		},
		"Wait to be pingable (normal boot)": {
			// No recovery actions as that is help action.
			Docs: []string{
				"Wait DUT to be pingable after some action on it.",
				"Waiting time 150 seconds.",
			},
			ExecName:    "cros_ping",
			ExecTimeout: &durationpb.Duration{Seconds: 150},
			RunControl:  RunControl_ALWAYS_RUN,
		},
		"Trigger kernel panic to reset the whole board and try ssh to DUT": {
			Docs: []string{
				"This repair action repairs a Chrome device by sending a system request to the kernel.",
				"TODO: (blocked by: b/221083688) Collect logs from a successfully repaired DUT.",
			},
			Conditions: []string{
				"servod_echo",
			},
			Dependencies: []string{
				"dut_servo_host_present",
				"Trigger kernel panic by servod",
				"Wait to be SSHable (normal boot)",
			},
			ExecName: "sample_pass",
		},
		"Trigger kernel panic by servod": {
			Docs: []string{
				"This repair action repairs a Chrome device by sending a system request to the kernel.",
			},
			Conditions: []string{
				"servod_echo",
			},
			Dependencies: []string{
				"dut_servo_host_present",
			},
			ExecExtraArgs: []string{
				"count:3",
				"retry_interval:2",
			},
			ExecName: "servo_trigger_kernel_panic",
		},
		"Cr50 reset by servo wait for SSH": {
			Docs: []string{
				"Repair a Chrome Device by resetting cr50 by servo.",
				"Then, using servo to initialize dut again.",
				"TODO: (blocked by: b/221083688) Collect logs from a successfully repaired DUT.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
				"servod_echo",
				"servo_host_is_labstation",
				"servod_has_control_cr50_reboot",
				"Trigger power_state:cr50_reset",
				"Re-initialize DUT part of servo",
				"Wait to be SSHable (normal boot)",
			},
			ExecName: "sample_pass",
		},
		"Re-initialize DUT part of servo": {
			Docs: []string{
				"cr50 reset will clear some some init like `ccd testlab open` so we want to re-initialize servo after cr50 reset if the main device uses cr50/gsc console commands.",
			},
			Conditions: []string{
				"Is not servo_v3",
				"Servo main device is GSC chip",
			},
			Dependencies: []string{
				"dut_servo_host_present",
				"sleep_1_second",
			},
			ExecName: "init_dut_for_servo",
		},
		"Servo main device is GSC chip": {
			Docs: []string{
				"Verify that main device is c2d2/cr50/GSC",
			},
			Dependencies: []string{
				"dut_servo_host_present",
				"servo_host_is_labstation",
			},
			ExecName: "servo_main_device_is_gcs",
		},
		"servod_has_control_cr50_reboot": {
			Docs: []string{
				"Checks whether the servod has the command control: cr50_reboot.",
			},
			Conditions: []string{
				"dut_servo_host_present",
			},
			ExecExtraArgs: []string{
				"command:cr50_reboot",
			},
			ExecName: "servo_check_servod_control",
		},
		"Trigger power_state:cr50_reset": {
			Docs: []string{
				"Repair a ChromeOS Device by resetting cr50 by servo.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
			},
			ExecExtraArgs: []string{
				"command:power_state",
				"string_value:cr50_reset",
			},
			ExecName:               "servo_set",
			AllowFailAfterRecovery: true,
		},
		"sleep_1_second": {
			ExecExtraArgs: []string{
				"sleep:1",
			},
			ExecName: "sample_sleep",
		},
		"Read BIOS from DUT by servo": {
			Docs: []string{
				"Read GBB flags from the DUT by servo.",
				"Set 40 minutes as some FW BIOS is too big and take time to flash it.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
				"servo_state_is_working",
			},
			ExecName: "cros_read_gbb_by_servo",
			ExecExtraArgs: []string{
				"remove_file:false",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 2400},
		},
		"Switch DUT to dev mode by servo": {
			Docs: []string{
				"Force to set GBB flags to 0x18 to boot in DEV mode and enable to boot from USB-drive.",
				"Reboot and wait for device to be back.",
			},
			Dependencies: []string{
				"Set GBB flags to 0x18 by servo",
				"Wait to be pingable (normal boot)",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Set GBB flags to 0x18 by servo": {
			Docs: []string{
				"Force to set GBB flags to 0x18 to boot in DEV mode and enable to boot from USB-drive.",
				"Set 40 minutes as some FW BIOS is too big and take time to flash it.",
				"Allowed to fail as flags can applied but fail by some reason",
			},
			Dependencies: []string{
				"Read BIOS from DUT by servo",
			},
			ExecName: "cros_set_gbb_by_servo",
			ExecExtraArgs: []string{
				"gbb_flags:0x18",
			},
			ExecTimeout:            &durationpb.Duration{Seconds: 2400},
			AllowFailAfterRecovery: true,
		},
		"Power cycle DUT by RPM and wait": {
			Docs: []string{
				"Perfrom RPM cycle and wait to device to boot back.",
			},
			Conditions: []string{
				"has_rpm_info",
			},
			Dependencies: []string{
				"Power cycle DUT by RPM",
				"Wait to be pingable (normal boot)",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Is not in audio box": {
			Docs: []string{
				"Verify that setup is not audio box",
			},
			Conditions: []string{
				"dut_is_in_audio_box",
			},
			ExecName: "sample_fail",
		},
		"Power cycle DUT by RPM": {
			Docs: []string{
				"Power cycle the DUT by RPM outlet.",
			},
			Conditions: []string{
				"has_rpm_info",
			},
			ExecName:   "rpm_power_cycle",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Collect dmesg logs from DUT": {
			Docs: []string{
				"Collect the entire output of dmesg",
			},
			ExecName:               "cros_dmesg",
			AllowFailAfterRecovery: true,
		},
		"Restore AC detection by EC console": {
			Docs: []string{
				"Try to recover AC detection through servod's ec control",
			},
			Conditions: []string{
				"dut_servo_host_present",
				"cros_is_battery_expected",
			},
			ExecExtraArgs: []string{
				"wait_timeout:120",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 600},
			ExecName:    "servo_recover_ac_power",
		},
		"Disable software-controlled write-protect for 'host'": {
			Docs: []string{
				"Disable write-protect fprom 'host'.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_disable_fprom_write_protect",
			ExecExtraArgs: []string{
				"fprom:host",
			},
			ExecTimeout:            &durationpb.Duration{Seconds: 300},
			AllowFailAfterRecovery: true,
			RunControl:             RunControl_ALWAYS_RUN,
		},
		"Disable software-controlled write-protect for 'ec'": {
			Docs: []string{
				"Disable write-protect fprom 'ec'.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_disable_fprom_write_protect",
			ExecExtraArgs: []string{
				"fprom:ec",
			},
			ExecTimeout:            &durationpb.Duration{Seconds: 300},
			AllowFailAfterRecovery: true,
			RunControl:             RunControl_ALWAYS_RUN,
		},
		"Install OS in recovery mode by booting from servo USB-drive": {
			Docs: []string{
				"This action installs the test image on DUT utilizing ",
				"the features of servo. DUT will be booted in recovery ",
				"mode. In some cases RO FW is not allowed to boot in ",
				"recovery mode with active PD, so we will change it to ",
				"sink-mode if required.",
			},
			Conditions: []string{
				"dut_servo_host_present",
				"Pools required to be in Secure mode",
			},
			Dependencies: []string{
				"Servo USB-Key needs to be reflashed",
				"Servo has USB-key with require image",
			},
			ExecName: "cros_install_in_recovery_mode",
			ExecExtraArgs: []string{
				"run_tpm_reset:true",
				"run_os_install:true",
				"boot_timeout:480",
				"boot_interval:10",
				"halt_timeout:120",
				"install_timeout:1200",
				"tpm_reset_timeout:60",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 3600},
		},
		"Install OS in DEV mode by USB-drive (for special pools)": {
			Docs: []string{
				"This action installs the test image on DUT after booking the DUT in dev mode.",
				"The action is only for deployment as not limited by pools.",
			},
			Conditions: []string{
				"Pools allowed to stay in DEV mode",
			},
			Dependencies: []string{
				"Install OS in DEV mode by USB-drive",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Install OS in DEV mode by USB-drive": {
			Docs: []string{
				"This action installs the test image on DUT after booking the DUT in dev mode.",
			},
			Dependencies: []string{
				"Boot DUT from USB in DEV mode",
				"Device booted from USB-drive",
				"Run install after boot from USB-drive",
				"Cold reset DUT by servo and wait to boot",
				"Wait to be SSHable (normal boot)",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Cold reset DUT by servo and wait to boot": {
			Docs: []string{
				"Cold reset device by servo and wait for DUT to become ping-able.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
				"servo_state_is_working",
				"Cold reset DUT by servo",
				"Wait to be pingable (normal boot)",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Reset power using servo if booted from USB": {
			Docs: []string{
				"This action will reboot the DUT using servo if device ",
				"is not booted after off/on performed as part of ",
				"re-imaging the device from USB device.",
			},
			Dependencies: []string{
				"Cold reset by servo and wait for SSH",
			},
			ExecName: "sample_pass",
		},
		"Cold reset by servo and wait for SSH": {
			Docs: []string{
				"This repair action will use servod command to reset power_state on the DUT.",
				"TODO: (blocked by: b/221083688) Collect logs from a successfully repaired DUT.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
				"servod_echo",
				"servo_power_state_reset",
				"Wait to be SSHable (normal boot)",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Cold reset DUT by servo": {
			Docs: []string{
				"Cold reset device by servo and do not wait.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
				"servo_state_is_working",
			},
			ExecName: "servo_set",
			ExecExtraArgs: []string{
				"command:power_state",
				"string_value:reset",
			},
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Pools required to manage FW on the device": {
			Docs: []string{
				"Verify that device we check in the pool which not required fw management.",
			},
			ExecName: "dut_not_in_pool",
			ExecExtraArgs: []string{
				"faft-test",
				"faft-test-tot",
				"faft-test-experiment",
				"faft_test_debug",
				"faft-cr50",
				"faft-cr50-debug",
				"faft-cr50-experimental",
				"faft-cr50-tot",
				"faft-experimental",
				"satlab_faft",
			},
		},
		"Pools allowed to stay in DEV mode": {
			Docs: []string{
				"Verify that pools are allowed to stay in DEV mode.",
			},
			ExecName: "dut_is_in_pool",
			ExecExtraArgs: []string{
				"crouton",
				"faft-test",
				"faft-test-au",
				"faft-test-tot",
				"nyc-meet-lab",
				"satlab_faft",
			},
		},
		"Pools required to be in Secure mode": {
			Docs: []string{
				"Verify that DUT need to be in Secure mode.",
			},
			Conditions: []string{
				"Pools allowed to stay in DEV mode",
			},
			ExecName: "sample_fail",
		},
		"Set default boot as disk and reboot": {
			Docs: []string{
				"Set default boot from disk and reboot.",
			},
			Dependencies: []string{
				"Set default boot as disk",
				"Simple reboot",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Recovery version has firmware version": {
			Docs: []string{
				"Verify that recovery version has firmware version.",
			},
			ExecName: "has_stable_version_fw_version",
		},
		"Recovery version has OS image path": {
			Docs: []string{
				"Verify that recovery version has OS image path.",
			},
			ExecName: "has_stable_version_cros_image",
		},
		"Simple reboot": {
			Docs: []string{
				"Simple un-blocker reboot.",
			},
			ExecName: "cros_run_shell_command",
			ExecExtraArgs: []string{
				"reboot && exit",
			},
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Set default boot as disk": {
			Docs: []string{
				"Set default boot from disk by crossystem.",
			},
			ExecExtraArgs: []string{
				"command:dev_default_boot",
				"value:disk",
				"check_after_update:true",
			},
			ExecName: "cros_update_crossystem",
		},
		"Device NOT booted from USB-drive": {
			Docs: []string{
				"Verify that device was not booted from USB-drive.",
			},
			Conditions: []string{
				//TODO(b:231627956): Flex board cannot run crossystem set_default_boot
				"Is not Flex device",
				"Device booted from USB-drive",
			},
			RecoveryActions: []string{
				"Set default boot as disk and reboot",
			},
			ExecName:   "sample_fail",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Device booted from USB-drive": {
			Docs: []string{
				"Verify that device was booted from USB-drive.",
			},
			ExecName:   "cros_booted_from_external_storage",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Write factory-install-reset to file system": {
			ExecName: "cros_run_shell_command",
			ExecExtraArgs: []string{
				"echo \"fast safe\" > /mnt/stateful_partition/factory_install_reset",
			},
			AllowFailAfterRecovery: true,
		},
		"Repair by powerwash": {
			Docs: []string{
				"Install the stable test image designated for the DUT.",
			},
			Dependencies: []string{
				"Write factory-install-reset to file system",
				"Quick provision OS",
			},
			ExecName: "sample_pass",
		},
		"Update FW from fw-image by servo": {
			Docs: []string{
				"Download fw-image specified in stable version and flash EC/AP to the DUT by servo",
				"Set timeout for 100 minutes for now as = 10m(download)+ 7*3m(extraction-file)+10m(ec-update)+30m(ap-update).",
				"The time will be updated later based on collected metrics",
				"Each operation with extraction files can take up to a few minutes.",
				"Ap update on the DUT can take up to 30 minutes",
			},
			Conditions: []string{
				"dut_servo_host_present",
				"servo_state_is_working",
			},
			Dependencies: []string{
				"Recovery version has firmware version",
			},
			ExecName: "cros_update_fw_with_fw_image_by_servo_from",
			ExecExtraArgs: []string{
				"update_ec:true",
				"update_ap:true",
				"download_timeout:600",
			},
			ExecTimeout: &durationpb.Duration{
				Seconds: 6000,
			},
		},
		"Update FW from fw-image by servo and reboot": {
			Docs: []string{
				"This action will repair the firmware on the DUT, and ",
				"then reboot and wait for the DUT to again become ",
				"available. This action exists to wrap these component ",
				"actions into a single repair action.",
			},
			Conditions: []string{
				"dut_servo_host_present",
			},
			Dependencies: []string{
				"Update FW from fw-image by servo",
				"Cold reset by servo and wait for SSH",
			},
			ExecName: "sample_pass",
		},
		"Update FW from fw-image by servo and set GBB to 0x18": {
			Docs: []string{
				"Download fw-image specified in stable version and flash EC/AP to the DUT by servo",
				"Set timeout for 100 minutes for now as = 10m(download)+ 7*3m(extraction-file)+10m(ec-update)+30m(ap-update).",
				"The time will be updated later based on collected metrics",
				"Each operation with extraction files can take up to a few minutes.",
				"Ap update on the DUT can take up to 30 minutes",
				"The GBB will set to 0x18 which equal to switch to DEV mode and enable boot from USB drive in DEV mode.",
			},
			Conditions: []string{
				"dut_servo_host_present",
				"servo_state_is_working",
			},
			Dependencies: []string{
				"Recovery version has firmware version",
			},
			ExecName: "cros_update_fw_with_fw_image_by_servo_from",
			ExecExtraArgs: []string{
				"update_ec:true",
				"update_ap:true",
				"download_timeout:600",
				"gbb_flags:0x18",
			},
			ExecTimeout: &durationpb.Duration{
				Seconds: 6000,
			},
		},
		"Boot DUT from USB in DEV mode": {
			Docs: []string{
				"Restart and try to boot from USB-drive",
				"First boot in dev mode can take time so set boot time to 10 minutes.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
			},
			ExecName: "cros_dev_mode_boot_from_servo_usb_drive",
			ExecExtraArgs: []string{
				"boot_timeout:600",
				"retry_interval:2",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 900},
			RunControl:  RunControl_ALWAYS_RUN,
		},
		"Run install after boot from USB-drive": {
			Docs:        []string{"Perform install process"},
			ExecName:    "cros_run_chromeos_install_command_after_boot_usbdrive",
			ExecTimeout: &durationpb.Duration{Seconds: 1200},
			RunControl:  RunControl_ALWAYS_RUN,
		},
		"DUT not on stable version": {
			Docs: []string{
				"Confirm that DUT does not have stable version.",
			},
			ExecName: "cros_not_on_stable_version",
		},
		"Check if stable version is availabe on installed OS": {
			Docs: []string{
				"Check that firmware manifest in OS tells that expected version nis present.",
			},
			Dependencies: []string{
				"Recovery version has firmware version",
			},
			ExecName: "cros_is_rw_firmware_stable_version_available",
		},
		"Update firmware from USB-Drive and when booted in recovery mode": {
			Docs: []string{
				"Update FW on the DUT using the USB-Drive.",
				"The process will required to boot device in recovery mode.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
				"Recovery version has OS image path",
				"Servo has USB-key with require image",
			},
			ExecName: "cros_install_in_recovery_mode",
			ExecExtraArgs: []string{
				"run_fw_update:true",
				"boot_timeout:480",
				"boot_interval:10",
				"halt_timeout:120",
				"fw_update_mode:recovery",
				"fw_update_use_force:false",
				"fw_update_timeout:600",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 2000},
		},
		"Perfrom RPM config verification": {
			Docs: []string{
				"Verify if RPM verification is required fo setup",
				"Setup with PD control temprarely excluded from testing.",
			},
			Conditions: []string{
				"dut_servo_host_present",
				"has_rpm_info",
				"servo_state_is_working",
				"Setup does't have Servo PD control",
			},
			ExecName: "sample_pass",
		},
		"Setup has Servo PD control": {
			Docs: []string{
				"Verify that servo has build in PD control.",
			},
			Conditions: []string{
				"dut_servo_host_present",
			},
			ExecName: "servo_build_in_pd_present",
		},
		"Setup does't have Servo PD control": {
			Docs: []string{
				"Verify that servo does not have build in PD control.",
			},
			Conditions: []string{
				"Setup has Servo PD control",
			},
			ExecName: "sample_fail",
		},
		"Verify RPM config (without battery)": {
			Docs: []string{
				"Verify RPM configs and set RPM state",
				"Not applicable for cr50 servos based on b/205728276",
				"Action is not critical as it updates own state.",
			},
			Conditions: []string{
				"Perfrom RPM config verification",
				"No Battery is present on device",
			},
			ExecName:               "rpm_audit_without_battery",
			ExecTimeout:            &durationpb.Duration{Seconds: 600},
			AllowFailAfterRecovery: true,
		},
		"Verify RPM config with battery": {
			Docs: []string{
				"Verify RPM when battery is present",
				"Not applicable for cr50 servos based on b/205728276",
				"Action is not critical as it updates own state.",
			},
			Conditions: []string{
				"Perfrom RPM config verification",
				"Battery is present on device",
			},
			ExecName:    "rpm_audit_with_battery",
			ExecTimeout: &durationpb.Duration{Seconds: 600},
			ExecExtraArgs: []string{
				"timeout:120",
				"wait_interval:5",
			},
			AllowFailAfterRecovery: true,
		},
		"servo_power_state_reset": {
			Docs: []string{
				"Reset the DUT using servod power_state control.",
			},
			Dependencies: []string{
				"dut_servo_host_present",
			},
			ExecName: "sample_fail",
		},
	}
}
