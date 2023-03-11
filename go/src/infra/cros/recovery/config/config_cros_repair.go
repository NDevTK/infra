// Copyright 2021 The Chromium OS Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"google.golang.org/protobuf/types/known/durationpb"
)

func crosRepairPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Collect logs and crashinfo",
			"Set state: repair_failed",
			"Device is pingable",
			"Device is SSHable",
			"Collect logs and crashinfo",
			"Verify internal storage",
			"Set dev_boot_usb is enabled",
			"Check if last provision was good",
			"Check if OS on required version for camerabox tablet",
			"Python is present",
			"Verify that device is not enrolled",
			"Check power sources",
			"Check TPM statuses",
			"Read TPM ownership",
			"Verify tmp_fwver is updated correctly",
			"Verify tpm_kernver is updated correctly",
			"Verify present of gsctool",
			"Audit battery",
			"Audit storage (SMART only)",
			"Audit wifi",
			"Audit bluetooth",
			"Audit cellular",
			"Stop if DUT needs replacement",
			"Firmware validations",
			"Login UI is up",
			"Can list RW VPD Keys",
			"Verify keys of RW_VPD",
			"Verify RO_VPD sku_number",
			"Verify RO_VPD data on DUT",
			"Verify system info",
			"Update Servo NIC mac address",
			"Match provision labels",
			"Backup CBI",
			"Check CBI",
			"Set state: ready",
			"Update special device labels",
			"Collect dmesg logs from DUT",
			"Record type C status",
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
			RunControl:    RunControl_RUN_ONCE,
			MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
		},
		"Set state: needs_repair": {
			Docs: []string{
				"The action set devices with state means that repair tsk did not success to recover the devices.",
			},
			ExecName: "dut_set_state",
			ExecExtraArgs: []string{
				"state:needs_repair",
			},
			RunControl:    RunControl_RUN_ONCE,
			MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
		},
		"Set state: repair_failed": {
			Docs: []string{
				"The action set devices with state means that repair tsk did not success to recover the devices.",
			},
			ExecName: "dut_set_state",
			ExecExtraArgs: []string{
				"state:repair_failed",
			},
			RunControl:    RunControl_RUN_ONCE,
			MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
		},
		"Set state: needs_deploy": {
			Docs: []string{
				"The action set devices with request to be redeployed.",
			},
			ExecName: "dut_set_state",
			ExecExtraArgs: []string{
				"state:needs_deploy",
			},
			MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
		},
		"DUT has board info": {
			ExecName:   "dut_has_board_name",
			RunControl: RunControl_RUN_ONCE,
		},
		"DUT has model info": {
			ExecName:   "dut_has_model_name",
			RunControl: RunControl_RUN_ONCE,
		},
		"Device is pingable": {
			Docs: []string{
				"Verify that device is reachable by ping.",
				"Limited to 15 seconds.",
			},
			Dependencies: []string{
				"DUT has board info",
				"DUT has model info",
			},
			ExecName: "cros_ping",
			ExecTimeout: &durationpb.Duration{
				Seconds: 15,
			},
			RecoveryActions: []string{
				"Cold reset by servo and wait for SSH",
				"Power cycle DUT by RPM and wait",
				"Trigger kernel panic to reset the whole board and try ssh to DUT",
				"Restore AC detection by EC console and wait for ping",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Update FW from fw-image by servo and wait for boot",
				"Install OS in recovery mode by booting from servo USB-drive (special pools)",
				"Install OS in DEV mode by USB-drive",
				"Reset power using servo if booted from USB",
				"Check if request labstation reboot",
			},
			RunControl: RunControl_ALWAYS_RUN,
			MetricsConfig: &MetricsConfig{
				// Always upload so we can track recovery.
				UploadPolicy: MetricsConfig_DEFAULT_UPLOAD_POLICY,
			},
		},
		"Device is SSHable": {
			Docs: []string{
				"Verify that device is reachable by SSH.",
				"Limited to 15 seconds.",
			},
			ExecName:    "cros_ssh",
			ExecTimeout: &durationpb.Duration{Seconds: 15},
			RecoveryActions: []string{
				// The DUT is pingable, so no need extra reboot actions.
				"Cold reset by servo and wait for SSH",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Update FW from fw-image by servo and wait for boot",
				"Install OS in recovery mode by booting from servo USB-drive (special pools)",
				"Install OS in DEV mode by USB-drive",
				"Reset power using servo if booted from USB",
			},
			RunControl: RunControl_ALWAYS_RUN,
			MetricsConfig: &MetricsConfig{
				UploadPolicy: MetricsConfig_DEFAULT_UPLOAD_POLICY,
			},
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
			},
			ExecName: "sample_pass",
		},
		"Stop if DUT needs replacement": {
			Docs: []string{
				"Plan stopper if the DUT has state 'needs_replacement'.",
			},
			ExecName: "dut_state_match",
			ExecExtraArgs: []string{
				"state:needs_replacement",
				"invert:true",
			},
		},
		"Audit storage (SMART only)": {
			Docs: []string{
				"Quick audit internal storage by reading SMART data.",
				"The check updates storage state.",
				"The check is not critical as update storage state.",
			},
			ExecName:               "cros_audit_storage_smart",
			AllowFailAfterRecovery: true,
		},
		"Audit device storage using badblocks": {
			Docs: []string{
				"Use the badblocks command to audit the storage on the DUT",
			},
			ExecName: "cros_audit_storage_bad_blocks",
			ExecExtraArgs: []string{
				"badblocks_mode:auto",
				"rw_badblocks_timeout:5400",
				"ro_badblocks_timeout:3600",
			},
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
			},
			ExecName: "sample_pass",
		},
		"Restore HWID from inventory": {
			Docs: []string{
				"Restoring HWID on the host from the inventory data.",
				"Using recovery from the host as flashing firmware by servo is very slow.",
			},
			Dependencies: []string{
				"Is not Flex device",
				"Is HWID known",
				"Device is SSHable",
				"Disable software-controlled write-protect for 'host'",
				"Disable software-controlled write-protect for 'ec'",
				"cros_update_hwid_from_inventory_to_host",
				"Simple reboot",
				"Sleep 1s",
				"Wait to be SSHable (normal boot)",
			},
			ExecName: "cros_match_hwid_to_inventory",
		},
		"Read OS version": {
			Docs: []string{
				"Read and log current OS version.",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			ExecName: "cros_read_os_version",
			RecoveryActions: []string{
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive",
			},
		},
		"Python is present": {
			Docs: []string{
				"Verify that device has python on it.",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			ExecName: "cros_has_python_interpreter_working",
			RecoveryActions: []string{
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive",
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
				"Install OS in DEV mode by USB-drive",
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
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive",
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
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive",
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
				"Install OS in DEV mode by USB-drive",
			},
		},
		"Read TPM ownership": {
			Docs: []string{
				"That is initiate action to detect issues for b/246476353",
				"No recovery action just verify and report of issue.",
				"Verify that we can read ownership on device.",
			},
			Conditions: []string{
				"Device is SSHable",
				"Is hwsec-ownership-id expected",
			},
			ExecName: "cros_run_command",
			ExecExtraArgs: []string{
				"host:dut",
				"command:hwsec-ownership-id id",
			},
			// Action is only for analysis at this stage.
			AllowFailAfterRecovery: true,
		},
		"Print block devices of the DUT": {
			Docs: []string{
				"Lsblk is used to display details about block devices and these blocks.",
			},
			ExecName: "cros_run_command",
			ExecExtraArgs: []string{
				"host:dut",
				"command:lsblk",
			},
			// Action is only for analysis at this stage.
			AllowFailAfterRecovery: true,
			MetricsConfig: &MetricsConfig{
				UploadPolicy: MetricsConfig_SKIP_ALL,
			},
		},
		"Is hwsec-ownership-id expected": {
			Docs: []string{
				"The hwsec-ownership-id is expected from R101 version of ChromeOS on the DUT.",
			},
			Conditions: []string{
				"Device is SSHable",
			},
			ExecName: "cros_is_on_expected_version",
			ExecExtraArgs: []string{
				"min_version:101",
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
				"RO Firmware version matches the recovery-version",
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
				"Fix FW on the DUT to match stable-version and wait to boot",
				"Update FW from fw-image by servo and wait for boot",
			},
		},
		"Check CBI": {
			Docs: []string{
				"Checks the CBI contents for corruption. go/cbi-auto-recovery-dd",
			},
			Conditions: []string{
				"CBI is present",
				"UFS contains CBI contents",
			},
			Dependencies: []string{
				"CBI contents contain valid magic",
			},
			ExecName:               "cros_cbi_contents_match",
			AllowFailAfterRecovery: true,
		},
		"Recover CBI With Contents From Inventory": {
			Docs: []string{
				"Manages the CBI repair workflow.",
			},
			Conditions: []string{
				"Hardware write protection is disabled",
				"CBI is present",
				"UFS contains CBI contents",
				"CBI contents do not match",
			},
			Dependencies: []string{
				"Restore CBI contents from UFS",
				"Simple reboot",
				"Wait to be SSHable (normal boot)",
				"Invalidate CBI cache",
				"CBI contents match",
			},
			ExecName: "sample_pass",
		},
		"Restore CBI contents from UFS": {
			Docs: []string{
				"Restore backup CBI contents from UFS.",
			},
			Dependencies: []string{
				"Hardware write protection is disabled",
				"CBI is present",
				"UFS contains CBI contents",
				"CBI contents do not match",
			},
			ExecName: "cros_restore_cbi_contents_from_ufs",
		},
		"CBI contents contain valid magic": {
			Docs: []string{
				"Check if CBI contents on the DUT contain valid CBI magic",
			},
			Dependencies: []string{
				"CBI is present",
			},
			ExecName: "cros_cbi_contents_contain_valid_magic",
		},
		"Invalidate CBI cache": {
			Docs: []string{
				"Invalidate the current CBI cache to ensure that any existing contents are up to date.",
			},
			Dependencies: []string{
				"CBI is present",
			},
			ExecName: "cros_invalidate_cbi_cache",
		},
		"Backup CBI": {
			Docs: []string{
				"Store CBI contents in UFS",
			},
			Conditions: []string{
				"CBI is present",
				"UFS does not contain CBI contents",
			},
			Dependencies: []string{
				"CBI contents contain valid magic",
			},
			ExecName:               "cros_backup_cbi",
			AllowFailAfterRecovery: true,
		},
		"CBI contents match": {
			Docs: []string{
				"Check if the contents on the DUT match the contents stored in UFS.",
			},
			Dependencies: []string{
				"CBI is present",
				"UFS contains CBI contents",
			},
			ExecName: "cros_cbi_contents_match",
		},
		"CBI contents do not match": {
			Docs: []string{
				"Check if the contents on the DUT do not match the contents stored in UFS.",
			},
			Dependencies: []string{
				"CBI is present",
				"UFS contains CBI contents",
			},
			ExecName: "cros_cbi_contents_do_not_match",
		},
		"UFS contains CBI contents": {
			Docs: []string{
				"Check if UFS contents are stored in UFS.",
			},
			ExecName: "cros_ufs_contains_cbi_contents",
		},
		"UFS does not contain CBI contents": {
			Docs: []string{
				"Check if UFS contents are not stored in UFS.",
			},
			ExecName: "cros_ufs_does_not_contain_cbi_contents",
		},
		"CBI is present": {
			Docs: []string{
				"Check if CBI is present on the DUT (most devices manufactured after 2020 should have CBI) go/cros-board-info",
			},
			ExecName: "cros_cbi_is_present",
		},
		"Hardware write protection is disabled": {
			Docs: []string{
				"Checks if crossystem wpsw_cur is set to 0 (hardware write protection is disabled). Required before writing to CBI EEPROM or other on board EC chips.",
			},
			ExecName: "cros_is_hardware_write_protection_disabled",
		},
		"Login UI is up": {
			Docs: []string{
				"Check the command 'stop ui' won't crash the DUT.",
			},
			ExecName:    "cros_stop_start_ui",
			ExecTimeout: &durationpb.Duration{Seconds: 45},
			RecoveryActions: []string{
				"Cold reset by servo and wait for SSH",
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive",
			},
		},
		"Start UI": {
			Docs: []string{
				"If the stop and start of UI timesout, it is an indication that ",
				"the UI has crashed. This action attempts to start the UI as an ",
				"attempt to recover the crashed UI.",
			},
			ExecName: "cros_start_ui",
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
				// TODO(b/248630303): Need run tmp reset.
				"Restore RW VPD Keys",
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
			},
			AllowFailAfterRecovery: true,
		},
		"Can list RW VPD Keys": {
			Docs: []string{
				"Check whether the RW VPD keys can be listed without any errors.",
			},
			ExecName: "cros_can_list_rw_vpd_keys",
			RecoveryActions: []string{
				"Recover from RW VPD keys listing errors",
			},
		},
		"Recover from RW VPD keys listing errors": {
			Docs: []string{
				"Check whether the RW VPD keys can be listed without any errors.",
			},
			Dependencies: []string{
				"Provision OS if needed",
				"Erase RW VPD Keys",
				"Restore RW VPD Keys",
				"Check RW VPD Keys",
			},
			ExecName: "sample_pass",
		},
		"Check RW VPD Keys": {
			Docs: []string{
				"Verify that keys: 'should_send_rlz_ping', 'gbind_attribute', 'ubind_attribute' are present in vpd RW_VPD partition.",
				"This action is not critical, and only logs any missing RW VPD keys.",
			},
			Conditions: []string{
				"Is not Flex device",
			},
			ExecName:               "cros_are_required_rw_vpd_keys_present",
			AllowFailAfterRecovery: true,
		},
		"Device has incorrect cros image version": {
			Docs: []string{
				"Check whether the cros image version on the device is not as expected.",
			},
			Dependencies: []string{
				"Recovery version has OS image path",
				"cros_is_on_stable_version",
			},
			ExecName: "sample_fail",
		},
		"Erase RW VPD Keys": {
			Docs: []string{
				"Reset the RW VPD keys by resetting the flash memory of RW_VPD on the device.",
			},
			ExecName: "cros_run_shell_command",
			ExecExtraArgs: []string{
				"flashrom -p host -i RW_VPD -E",
			},
		},
		"Restore RW VPD Keys": {
			Docs: []string{
				"Restore any possible RW VPD keys from the known default values.",
			},
			ExecName: "cros_restore_rw_vpd_keys",
		},
		"Verify servo keyboard firmware": {
			Conditions: []string{
				"Is servod running",
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
				"Flash keyboard map",
			},
			AllowFailAfterRecovery: true,
		},
		"Flash keyboard map": {
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"set_at_hwb_on",
				"set_atmega_rst_on",
				"Sleep for atmega reset",
				"set_atmega_rst_off",
				"Sleep for atmega reset",
				"set_at_hwb_off",
				"Sleep for usb present delay",
				"Check if expected Atmel chip",
				"Transfer keyboard hex",
				"Erase keyboard map",
				"Write keyboard map",
			},
			ExecName: "cros_run_shell_command",
			ExecExtraArgs: []string{
				"lsusb -vv -d 03eb:2042 |grep \"Remote Wakeup\"",
			},
			AllowFailAfterRecovery: true,
		},
		"set_at_hwb_on": {
			Docs: []string{
				"set servo's 'at_hwb' command to 'on' value.",
			},
			Dependencies: []string{
				"Setup has servo info",
			},
			ExecName: "servo_set",
			ExecExtraArgs: []string{
				"command:at_hwb",
				"string_value:on",
			},
		},
		"set_atmega_rst_on": {
			Docs: []string{
				"set servo's 'atmega_rst' command to 'on' value.",
			},
			Dependencies: []string{
				"Setup has servo info",
			},
			ExecName: "servo_set",
			ExecExtraArgs: []string{
				"command:atmega_rst",
				"string_value:on",
			},
		},
		"Sleep for atmega reset": {
			Docs: []string{
				"In legacy repair, the atmega reset delay in 0.2 seconds. ",
				"However, here we are being more conservative, and sleep ",
				"for 1 second.",
			},
			ExecName: "sample_sleep",
			ExecExtraArgs: []string{
				"sleep:1",
			},
			RunControl: RunControl_ALWAYS_RUN,
		},
		"set_atmega_rst_off": {
			Docs: []string{
				"set servo's 'atmega_rst' command to 'off' value.",
			},
			Dependencies: []string{
				"Setup has servo info",
			},
			ExecName: "servo_set",
			ExecExtraArgs: []string{
				"command:atmega_rst",
				"string_value:off",
			},
		},
		"set_at_hwb_off": {
			Docs: []string{
				"set servo's 'at_hwb' command to 'off' value.",
			},
			Dependencies: []string{
				"Setup has servo info",
			},
			ExecName: "servo_set",
			ExecExtraArgs: []string{
				"command:at_hwb",
				"string_value:off",
			},
		},
		"Sleep for usb present delay": {
			Docs: []string{
				"We need to wait for USB Present Delay, which is 1 second.",
			},
			ExecName: "sample_sleep",
			ExecExtraArgs: []string{
				"sleep:1",
			},
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Check if expected Atmel chip": {
			Docs: []string{
				"We check whether the chip is of the expected type.",
			},
			ExecName: "cros_run_shell_command",
			ExecExtraArgs: []string{
				"lsusb -d 03eb: | grep \"Atmel Corp. atmega32u4 DFU bootloader\"",
			},
			ExecTimeout: &durationpb.Duration{
				Seconds: 30,
			},
		},
		"Transfer keyboard hex": {
			Docs: []string{
				"We will transfer the keyboard hex embedded in repair package to the DUT.",
			},
			ExecName: "transfer_keyboard_hex",
		},
		"Erase keyboard map": {
			Docs: []string{
				"Erase pre-existing keyboard map.",
			},
			ExecName: "cros_run_shell_command",
			ExecExtraArgs: []string{
				"dfu-programmer atmega32u4 erase --force",
			},
			ExecTimeout: &durationpb.Duration{
				Seconds: 120,
			},
		},
		"Write keyboard map": {
			Docs: []string{
				"Write new keyboard map using the copied keyboard hex.",
			},
			ExecName: "cros_run_shell_command",
			ExecExtraArgs: []string{
				"dfu-programmer atmega32u4 flash /tmp/keyboard.hex",
			},
			ExecTimeout: &durationpb.Duration{
				Seconds: 120,
			},
		},
		"Update Servo NIC mac address": {
			Docs: []string{
				"Update mac address of servo-NIC in servod.",
			},
			Conditions: []string{
				"Setup has servo info",
				"servod_control_exist_for_mac_address",
			},
			ExecName:               "servo_audit_nic_mac_address",
			AllowFailAfterRecovery: true,
		},
		"servod_control_exist_for_mac_address": {
			Conditions: []string{
				"Setup has servo info",
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
				"Setup has servo info",
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
				"Setup has servo info",
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
				"Setup has servo info",
			},
			ExecName: "servo_match_state",
			ExecExtraArgs: []string{
				"state:WORKING",
			},
		},
		"RO Firmware version matches the recovery-version": {
			Docs: []string{
				"Check if the version of RO firmware on DUT matches the stable firmware version.",
			},
			Conditions: []string{
				"Check stable firmware version exists",
				"Recovery version has firmware image path",
				"Pools required to manage FW on the device",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			ExecName: "cros_is_on_ro_firmware_stable_version",
			RecoveryActions: []string{
				"Fix FW on the DUT to match stable-version and wait to boot",
				"Update FW from fw-image by servo and wait for boot",
			},
			AllowFailAfterRecovery: true,
		},
		"Fix FW on the DUT to match stable-version and wait to boot": {
			Docs: []string{
				"Update firmware from the host and reboot, then wait for host be available for SSH.",
			},
			Dependencies: []string{
				"Fix FW on the DUT to match stable-version",
				"Simple reboot",
				"Wait to be SSHable (normal boot)",
			},
			ExecName: "sample_pass",
		},
		"Fix FW on the DUT to match stable-version": {
			Docs: []string{
				"Download firmware image based on stable_version and install via firmware updater from DUT",
				"Update FW required the DUT to be run on stable-version OS.",
				"The reboot is not triggered as part of the action.",
				"Set timeout to 120 minutes = 10 minutes for download + 100 minutes for find and extract AP/EC images + 10 minutes for run updater.",
			},
			Conditions: []string{
				"Recovery version has OS image path",
				"Recovery version has firmware image path",
			},
			Dependencies: []string{
				"Provision OS if needed",
				"Disable software-controlled write-protect for 'host'",
				"Disable software-controlled write-protect for 'ec'",
			},
			ExecName:    "cros_update_firmware_from_firmware_image",
			ExecTimeout: &durationpb.Duration{Seconds: 7200},
			ExecExtraArgs: []string{
				"mode:recovery",
				"force:true",
				"update_ec_attempt_count:1",
				"update_ap_attempt_count:1",
				"updater_timeout:600",
				"use_cache_extractor:true",
			},
			// Allowed to fail as part of b/236417969 to check affect of it.
			AllowFailAfterRecovery: true,
		},
		"Provision OS if needed": {
			Docs: []string{
				"Perform provision OS if device is not running on it.",
			},
			Conditions: []string{
				"Recovery version has OS image path",
				"DUT not on stable version",
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
			Dependencies: []string{
				"Read OS version",
			},
			ExecName: "cros_run_shell_command",
			ExecExtraArgs: []string{
				"gsctool -a -f",
			},
			RecoveryActions: []string{
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive",
			},
			AllowFailAfterRecovery: true,
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
				//TODO(b:234761994, Flex device does not have charge_full file)
				"Is not Flex device",
			},
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
		"Audit USB-drive from DUT": {
			Docs: []string{
				"Audit the USB drive.",
				"Run badblocks to test USB-drive from DUT side.",
				"Timeout is 2 hours.",
			},
			Dependencies: []string{
				"servo_state_is_working",
				"Is servod running",
				"Device NOT booted from USB-drive",
				"Print block devices of the DUT",
			},
			ExecName: "audit_usb_from_dut_side",
			ExecExtraArgs: []string{
				// Set 118 minutes.
				// That is 2 minutes to prepare and clean up before timeout.
				"audit_timeout_min:118",
			},
			ExecTimeout: &durationpb.Duration{
				Seconds: 7200,
			},
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
		"Audit cellular": {
			Docs: []string{
				"Check cellular modem on the DUT is normal and update cellular modem state accordingly.",
			},
			Conditions: []string{
				"cros_has_mmcli",
				"has_cellular_info",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_audit_cellular",
			ExecExtraArgs: []string{
				"wait_manager_when_not_expected:120",
				"wait_manager_when_expected:15",
			},
			ExecTimeout: &durationpb.Duration{
				Seconds: 180,
			},
			AllowFailAfterRecovery: true,
		},
		"Verify tmp_fwver is updated correctly": {
			Docs: []string{
				"For dev-signed firmware, tpm_fwver reported from crossystem should always be 0x10001.",
				"Firmware update on DUTs with incorrect tmp_fwver may fail due to firmware rollback protection.",
			},
			Conditions: []string{
				"Is not Flex device",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			ExecName: "cros_match_dev_tpm_firmware_version",
			RecoveryActions: []string{
				"ChromeOS TMP recovery (not critical)",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive",
			},
		},
		"Verify tpm_kernver is updated correctly": {
			Docs: []string{
				"For dev-signed firmware, tpm_kernver reported from crossystem should always be 0x10001.",
				"Firmware update on DUTs with incorrect tpm_kernver may fail due to firmware rollback protection.",
			},
			Conditions: []string{
				"Is not Flex device",
			},
			Dependencies: []string{
				"Internal storage is responsive",
			},
			ExecName: "cros_match_dev_tpm_kernel_version",
			RecoveryActions: []string{
				"ChromeOS TMP recovery (not critical)",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive",
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
				"Restore HWID from inventory",
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
		"Is not booted in secure mode (condition)": {
			Docs: []string{
				"Check if the device is not booted in secure mode.",
			},
			Conditions: []string{
				"Booted in secure mode (condition)",
			},
			ExecName:   "sample_fail",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Booted in secure mode (condition)": {
			Docs: []string{
				"Check if the device booted in secure mode.",
			},
			ExecName:   "cros_is_booted_in_secure_mode",
			RunControl: RunControl_ALWAYS_RUN,
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
				"Allowed to fail if HWID is not matched",
			},
			Conditions: []string{
				"Is not Flex device",
				"Read OS version",
				"Is HWID known",
			},
			ExecName: "cros_match_hwid_to_inventory",
			RecoveryActions: []string{
				"Restore HWID from inventory",
			},
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
			ExecName: "cros_match_serial_number_inventory",
			RecoveryActions: []string{
				"Restore serial-number",
			},
			AllowFailAfterRecovery: true,
		},
		"Restore serial-number": {
			Docs: []string{
				"Restore serial number from inventory to the host",
			},
			ExecName: "cros_restore_serial_number",
		},
		"Is HWID known": {
			Docs: []string{
				"Check whether the DUT information includes its HWID.",
			},
			ExecName:   "dut_has_hwid",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Is HWID empty": {
			Docs: []string{
				"Check whether the DUT information includes its HWID.",
			},
			Conditions: []string{
				"Is HWID known",
			},
			ExecName:   "sample_fail",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Is serial-number known": {
			Docs: []string{
				"Check whether the DUT information includes its ",
				"serial number.",
			},
			ExecName:   "dut_has_serial_number",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Is serial-number empty": {
			Docs: []string{
				"Check whether the DUT information includes its ",
				"serial number.",
			},
			Conditions: []string{
				"Is serial-number known",
			},
			ExecName:   "sample_fail",
			RunControl: RunControl_ALWAYS_RUN,
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
			ExecName:   "cros_update_serial_number_inventory",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Read DUT serial-number from DUT (Satlab)": {
			Conditions: []string{
				"Is Satlab device",
			},
			ExecName:               "cros_update_serial_number_inventory",
			RunControl:             RunControl_ALWAYS_RUN,
			AllowFailAfterRecovery: true,
		},
		"Read HWID from DUT": {
			Conditions: []string{
				"Not Satlab device",
			},
			ExecName:   "cros_update_hwid_to_inventory",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Read HWID from DUT (Satlab)": {
			Conditions: []string{
				"Is Satlab device",
			},
			ExecName:               "cros_update_hwid_to_inventory",
			RunControl:             RunControl_ALWAYS_RUN,
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
			ExecExtraArgs: []string{
				"paths:/mnt/stateful_partition,/var/tmp,/mnt/stateful_partition/encrypted",
			},
			RecoveryActions: []string{
				"Cold reset by servo and wait for SSH",
				"Quick provision OS",
				"Repair by powerwash",
				"Install OS in recovery mode by booting from servo USB-drive",
				"Install OS in DEV mode by USB-drive",
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
				"Install OS in DEV mode by USB-drive",
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
				"Install OS in DEV mode by USB-drive",
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
				"Install OS in DEV mode by USB-drive",
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
				"Install OS in DEV mode by USB-drive",
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
				"Is gsctool present on the host",
			},
			ExecName:               "cros_update_cr50_label",
			AllowFailAfterRecovery: true,
		},
		"Read Cr50 key ID": {
			Docs: []string{
				"Update the cr50 key ID from device.",
			},
			Conditions: []string{
				"Is gsctool present on the host",
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
		"Is gsctool present on the host": {
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
		"Stable version image is missing from servo usbkey": {
			Docs: []string{
				"This is a reverse action which fails when required image is already cached in servo usbkey.",
				"The purpose is to serve as a condition of Download stable image to USB-key action, so that we don't do duplicate download.",
				"If this action fails, it means the servo usbkey already have required stable_version OS image cached.",
			},
			Conditions: []string{
				"Servo usbkey has stable image",
			},
			ExecName: "sample_fail",
		},
		"Servo usbkey has stable image": {
			Docs: []string{
				"Check if the usbkey has the stable_version OS image.",
				"TODO: Collect data on the usual number of retries and tweak the default",
			},
			ExecName: "servo_usbkey_has_stable_image",
			ExecExtraArgs: []string{
				"retry_count:3",
				"retry_interval:1",
			},
		},
		"Download stable version OS image to servo usbkey if necessary": {
			Docs: []string{
				"This action will download model specific stable version OS image to servo usbkey.",
				"The action will be skipped if the required image is already loaded.",
			},
			Conditions: []string{
				"Setup has servo info",
				"Has a stable-version service",
				"Stable version image is missing from servo usbkey",
			},
			Dependencies: []string{
				"servo_servod_echo_host",
				"Is servo USB key detected",
			},
			ExecName:    "servo_download_image_to_usb",
			ExecTimeout: &durationpb.Duration{Seconds: 3000},
		},
		"Is servo USB key detected": {
			Docs: []string{
				"The action used as codiion.",
			},
			Dependencies: []string{
				"Is servod running",
			},
			ExecName: "servo_usbkey_is_detected",
		},
		"Download stable version OS image to servo usbkey if necessary (allow fail)": {
			Docs: []string{
				"This action will download model specific stable version OS image to servo usbkey.",
				"The action will be skipped if the required image is already loaded.",
			},
			Conditions: []string{
				"Setup has servo info",
				"Has a stable-version service",
				"Stable version image is missing from servo usbkey",
			},
			Dependencies: []string{
				"servo_servod_echo_host",
				"Is servo USB key detected",
			},
			ExecName:               "servo_download_image_to_usb",
			ExecTimeout:            &durationpb.Duration{Seconds: 3000},
			AllowFailAfterRecovery: true,
			RunControl:             RunControl_RUN_ONCE,
		},
		"Download stable image to USB-key": {
			Docs: []string{
				"Download lab stable image on servo USB-key",
				"Download the image can take longer if labstation download parallel a few images.",
				"This step is allowed to complete successfully even if some",
				" errors happen during download because the image can already",
				" be present on the USB-drive.",
			},
			Dependencies: []string{
				"Setup has servo info",
				// If servo is responsive then probably we can download image to USB drive.
				// Present of DUT connection is not critical.
				"servo_servod_echo_host",
				"Is servo USB key detected",
			},
			ExecName:               "servo_download_image_to_usb",
			ExecTimeout:            &durationpb.Duration{Seconds: 3000},
			RunControl:             RunControl_ALWAYS_RUN,
			AllowFailAfterRecovery: true,
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
				"Is not booted in secure mode (condition)",
			},
			Dependencies: []string{
				"Reset GBB flags by host",
				"cros_switch_to_secure_mode",
				"Simple reboot",
				"Wait to be pingable (normal boot)",
				"Wait to be SSHable (normal boot)",
			},
			ExecName: "sample_pass",
		},
		"Reset GBB flags by host": {
			Docs: []string{
				"This action sets the GBB flags to 0x0.",
			},
			Dependencies: []string{
				"Disable software-controlled write-protect for 'host'",
				"Disable software-controlled write-protect for 'ec'",
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
				"Device is SSHable",
				"Internal storage is responsive",
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
				"Is servod running",
			},
			Dependencies: []string{
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
				"Is servod running",
			},
			Dependencies: []string{
				"Setup has servo info",
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
			},
			Conditions: []string{
				"Not Satlab device",
				"Is servod running",
			},
			Dependencies: []string{
				"servod_has_control_cr50_reboot",
				"Trigger power_state:cr50_reset",
				"Re-initialize DUT part of servo",
				"Wait to be SSHable (normal boot)",
			},
			ExecName: "sample_pass",
		},
		"Skip condition": {
			Docs: []string{
				"Condition which always fail to make possible to skip any action when needed.",
			},
			ExecName: "sample_fail",
		},
		"Re-initialize DUT part of servo": {
			Docs: []string{
				"cr50 reset will clear some some init like `ccd testlab open` so we want to re-initialize servo after cr50 reset if the main device uses cr50/gsc console commands.",
			},
			Conditions: []string{
				"Servo main device is GSC chip",
			},
			Dependencies: []string{
				"Setup has servo info",
				"Sleep 1s",
			},
			ExecName: "init_dut_for_servo",
		},
		"Servo main device is GSC chip": {
			Docs: []string{
				"Verify that main device is c2d2/cr50/GSC",
			},
			Dependencies: []string{
				"Setup has servo info",
				"servo_host_is_labstation",
			},
			ExecName: "servo_main_device_is_gsc",
		},
		"servod_has_control_cr50_reboot": {
			Docs: []string{
				"Checks whether the servod has the command control: cr50_reboot.",
			},
			Conditions: []string{
				"Setup has servo info",
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
				"Setup has servo info",
			},
			ExecExtraArgs: []string{
				"command:power_state",
				"string_value:cr50_reset",
			},
			ExecName:               "servo_set",
			AllowFailAfterRecovery: true,
		},
		"Sleep 1s": {
			ExecName: "sample_sleep",
			ExecExtraArgs: []string{
				"sleep:1",
			},
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Read BIOS from DUT by servo": {
			Docs: []string{
				"Read GBB flags from the DUT by servo.",
				"Set 40 minutes as some FW BIOS is too big and take time to flash it.",
			},
			Dependencies: []string{
				"Setup has servo info",
				"Is servod running",
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
				"Is servod running",
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
				"Perform RPM cycle and wait to device to boot back.",
			},
			Conditions: []string{
				"has_rpm_info",
			},
			Dependencies: []string{
				"rpm_power_cycle",
				"Set servo PD to src",
				"Wait to be pingable (normal boot)",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Set servo PD to src": {
			Docs: []string{
				"Set servo PD to src to power the DUT.",
				"If servo is type-c it can switch PD to snk.",
			},
			ExecName: "servo_set",
			ExecExtraArgs: []string{
				"command:servo_pd_role",
				"expected_string_value:src",
			},
			RunControl:             RunControl_ALWAYS_RUN,
			AllowFailAfterRecovery: true,
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
		"Collect dmesg logs from DUT": {
			Docs: []string{
				"Collect the entire output of dmesg",
			},
			ExecName:               "cros_dmesg",
			AllowFailAfterRecovery: true,
		},
		"Restore AC detection by EC console and wait for ping": {
			Docs: []string{
				"Try to recover AC detection through servod's ec control.",
				"This action wraps the recovery action and waits for the device to come back online.",
			},
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"Servo recover AC power",
				"Wait to be pingable (normal boot)",
			},
			ExecName: "sample_pass",
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
		"Servo recover AC power": {
			Docs: []string{
				"Try to recover AC detection through servod's ec control.",
				"The DUT may take time to become pingable again,",
				"so we use a wrapper action to wait.",
			},
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"DUT has CrOS EC",
				"cros_is_battery_expected",
			},
			ExecName: "servo_recover_ac_power",
			ExecExtraArgs: []string{
				"wait_timeout:120",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 600},
		},
		"DUT has CrOS EC": {
			Docs: []string{
				"Verify if DUT has ChromeOS firmware for EC",
			},
			Dependencies: []string{
				"Is servod running",
			},
			ExecExtraArgs: []string{
				"command:supports_cros_ec_communication",
				"expected_string_value:yes",
			},
			ExecName: "servo_check_servod_control",
		},
		"Install OS in recovery mode by booting from servo USB-drive": {
			Docs: []string{
				"This action installs the test image on DUT utilizing the features of servo.",
				"DUT will be booted in recovery mode.",
			},
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"Servo USB-Key needs to be reflashed",
				"Download stable version OS image to servo usbkey if necessary (allow fail)",
				"Boot DUT in recovery and install from USB-drive",
				"Wait to be SSHable (normal boot)",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Install OS in recovery mode by booting from servo USB-drive (special pools)": {
			Docs: []string{
				"This action installs the test image on DUT utilizing the features of servo.",
				"DUT will be booted in recovery mode. This action is targeted at devices ",
				"in special pools only.",
			},
			Conditions: []string{
				"Pools allowed to stay in DEV mode",
				"Is servod running",
			},
			Dependencies: []string{
				"Boot DUT in recovery and install from USB-drive",
				"Wait to be SSHable (normal boot)",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Download and install OS in DEV mode using USB-drive": {
			Docs: []string{
				"This action installs the test image on DUT after booking the DUT in dev mode.",
				"The action is only for deployment as not limited by pools.",
			},
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"Download stable version OS image to servo usbkey if necessary (allow fail)",
				"Install OS in DEV mode by USB-drive",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Install OS in DEV mode by USB-drive": {
			Docs: []string{
				"This action installs the test image on DUT after booking the DUT in dev mode.",
			},
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"Boot DUT from USB in DEV mode",
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
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
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
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"Cold reset DUT by servo",
				"Wait to be SSHable (normal boot)",
			},
			ExecName: "sample_pass",
		},
		"Cold reset by servo and wait for SSH": {
			Docs: []string{
				"This repair action will use servod command to reset power_state on the DUT.",
				"TODO: (blocked by: b/221083688) Collect logs from a successfully repaired DUT.",
			},
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"Cold reset DUT by servo",
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
				"Is servod running",
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
		"Check stable firmware version exists": {
			Docs: []string{
				"Check the DUT has model specific firmware stable_version configured.",
				"Flex device are exampted from this check as they don't run cros firmware",
			},
			Conditions: []string{
				"Is not Flex device",
				"Has a stable-version service",
			},
			ExecName: "has_stable_version_fw_version",
		},
		"Check stable faft version exists": {
			Docs: []string{
				"Check the DUT has model specific faft stable_version configured.",
				"Flex device are exampted from this check as they don't run cros firmware",
				"Satlab DUTs are exampted from this check given some early stage device don't have firmware branch GS bucket setup yet.",
			},
			Conditions: []string{
				"Is not Flex device",
				"Not Satlab device",
				"Has a stable-version service",
			},
			ExecName: "has_stable_version_fw_image",
		},
		// TODO: Resolve duplication with another action when satlab condition resolved.
		"Recovery version has firmware image path": {
			Docs: []string{
				"Verify that recovery version has firmware image path.",
			},
			Dependencies: []string{
				"Has a stable-version service",
			},
			ExecName: "has_stable_version_fw_image",
		},
		"Recovery version has OS image path": {
			Docs: []string{
				"Verify that recovery version has OS image path.",
			},
			Dependencies: []string{
				"Has a stable-version service",
			},
			ExecName: "has_stable_version_cros_image",
		},
		"Simple reboot": {
			Docs: []string{
				"Simple un-blocker reboot.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_run_command",
			ExecExtraArgs: []string{
				"host:dut",
				"command:reboot",
				"background:true",
			},
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Set default boot as disk": {
			Docs: []string{
				"Set default boot from disk by crossystem.",
			},
			ExecName: "cros_update_crossystem",
			ExecExtraArgs: []string{
				"command:dev_default_boot",
				"value:disk",
				"check_after_update:true",
			},
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
			Conditions: []string{
				"Device is SSHable",
				"Internal storage is responsive",
			},
			Dependencies: []string{
				"Write factory-install-reset to file system",
				"Simple reboot",
				"Wait to be SSHable (normal boot)",
			},
			ExecName:    "cros_provision",
			ExecTimeout: &durationpb.Duration{Seconds: 3600},
		},
		"Flash AP (FW) with GBB 0x18 by servo": {
			Docs: []string{
				"Download fw-image specified in stable version and flash AP to the DUT by servo",
				"Set timeout for 90 minutes for now as = 10m(download)+2*20m(find/extract file)+40m(ap-update with retry).",
				"We will retry up to 3 times since there may be flakiness on flash AP via servo.",
			},
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"Recovery version has firmware image path",
			},
			ExecName: "cros_update_fw_with_fw_image_by_servo",
			ExecExtraArgs: []string{
				"update_ap_attempt_count:3",
				"download_timeout:600",
				"gbb_flags:0x18",
				"use_cache_extractor:true",
			},
			ExecTimeout: &durationpb.Duration{
				Seconds: 5400,
			},
			AllowFailAfterRecovery: true,
		},
		"Flash EC (FW) by servo": {
			Docs: []string{
				"Download fw-image specified in stable version and flash EC to the DUT by servo",
				"Set timeout for 110 minutes for now as = 10m(download)+4*20m(find/extract file)+20m(ec-update with retry).",
				"We will retry up to 5 times since there is flakiness on flash EC.",
			},
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"Recovery version has firmware image path",
			},
			ExecName: "cros_update_fw_with_fw_image_by_servo",
			ExecExtraArgs: []string{
				"update_ec_attempt_count:5",
				"download_timeout:600",
				"use_cache_extractor:true",
			},
			ExecTimeout: &durationpb.Duration{
				Seconds: 6600,
			},
			AllowFailAfterRecovery: true,
		},
		"Flash AP (FW) and set GBB to 0x18 from fw-image by servo (without reboot)": {
			Docs: []string{
				"Download fw-image specified in stable version and flash AP only to the DUT by servo",
				"Set timeout for 90 minutes for now as = 10m(download)+2*20m(find/extract file)+40m(ap-update with retry).",
				"The time will be updated later based on collected metrics",
				"Each operation with extraction files can take up to a few minutes.",
				"The AP update on the DUT can take up to 30 minutes",
			},
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"Recovery version has firmware image path",
			},
			ExecName: "cros_update_fw_with_fw_image_by_servo",
			ExecExtraArgs: []string{
				"update_ec_attempt_count:0",
				"update_ap_attempt_count:3",
				"download_timeout:600",
				"gbb_flags:0x18",
				"use_cache_extractor:true",
			},
			ExecTimeout: &durationpb.Duration{
				Seconds: 5400,
			},
		},
		"Update FW from fw-image by servo and wait for boot": {
			Docs: []string{
				"This action will repair the firmware on the DUT, and ",
				"then reboot and wait for the DUT to again become ",
				"available. This action exists to wrap these component ",
				"actions into a single repair action.",
			},
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"Flash EC (FW) by servo",
				"Sleep 60 seconds",
				"Flash AP (FW) with GBB 0x18 by servo",
				// TODO(b:267549690) close the bug if we determine cold reset is not needed after flash AP.
				"Wait to be SSHable (normal boot)",
			},
			ExecName: "sample_pass",
		},
		"Update FW from fw-image by servo and set GBB to 0x18": {
			Docs: []string{
				"Download fw-image specified in stable version and flash EC/AP to the DUT by servo",
				"Set timeout for 180 minutes for now as = 10m(download)+ 6*20m(extraction-file)+10m(ec-update)+40m(ap-update).",
				"The time will be updated later based on collected metrics",
				"Each operation with extraction files can take up to a few minutes.",
				"Ap update on the DUT can take up to 30 minutes",
				"The GBB will set to 0x18 which equal to switch to DEV mode and enable boot from USB drive in DEV mode.",
			},
			Conditions: []string{
				"Is servod running",
			},
			Dependencies: []string{
				"Recovery version has firmware image path",
			},
			ExecName: "cros_update_fw_with_fw_image_by_servo",
			ExecExtraArgs: []string{
				"update_ec_attempt_count:3",
				"update_ap_attempt_count:3",
				"download_timeout:600",
				"gbb_flags:0x18",
				"use_cache_extractor:true",
			},
			ExecTimeout: &durationpb.Duration{
				Seconds: 10800,
			},
		},
		"Boot DUT in recovery and install from USB-drive": {
			Docs: []string{
				"This action installs the test image on DUT utilizing ",
				"the features of servo. DUT will be booted in recovery ",
				"mode. In some cases RO FW is not allowed to boot in ",
				"recovery mode with active PD, so we will change it to ",
				"sink-mode if required.",
			},
			ExecName: "cros_install_in_recovery_mode",
			ExecExtraArgs: []string{
				"run_tpm_reset:true",
				"run_os_install:true",
				"boot_timeout:480",
				"boot_interval:10",
				"boot_retry:2",
				"halt_timeout:120",
				"install_timeout:1200",
				"tpm_reset_timeout:60",
				"post_install_boot_time:15",
				"badblocks_mode:auto",
				"rw_badblocks_timeout:5400",
				"ro_badblocks_timeout:3600",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 8000},
			RunControl:  RunControl_ALWAYS_RUN,
		},
		"Restore FW from USB drive": {
			Docs: []string{
				"The goal to force update DUT fw when devices booted in the recovery mode from USB-stick",
			},
			ExecName: "cros_install_in_recovery_mode",
			ExecExtraArgs: []string{
				"run_custom_commands:true",
				"boot_timeout:480",
				"boot_interval:10",
				"boot_retry:1",
				"halt_timeout:120",
				"custom_command_allowed_to_fail:true",
				"custom_command_timeout:60",
				"custom_commands:chromeos-firemwareupdate --mode=recovery",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 1000},
			RunControl:  RunControl_ALWAYS_RUN,
		},
		"Restore from TPM 0x54 error": {
			Docs: []string{
				"Process based on b/272310645#comment39.",
			},
			ExecName: "cros_install_in_recovery_mode",
			ExecExtraArgs: []string{
				"run_custom_commands:true",
				"boot_timeout:480",
				"boot_interval:10",
				"boot_retry:1",
				"halt_timeout:120",
				"custom_command_allowed_to_fail:false",
				"custom_command_timeout:60",
				"custom_commands:tpm_manager_client status##tpm_manager_client take_ownership##cryptohome --action=remove_firmware_management_parameters",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 1000},
			RunControl:  RunControl_ALWAYS_RUN,
		},
		"Boot DUT from USB in DEV mode": {
			Docs: []string{
				"Restart and try to boot from USB-drive",
				"First boot in dev mode can take time so set boot time to 10 minutes.",
			},
			Dependencies: []string{
				"Setup has servo info",
			},
			ExecName: "cros_dev_mode_boot_from_servo_usb_drive",
			ExecExtraArgs: []string{
				"boot_retry:2",
				"boot_timeout:600",
				"retry_interval:1",
				"verify_usbkey_boot:true",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 650},
			RunControl:  RunControl_ALWAYS_RUN,
		},
		"Run install after boot from USB-drive": {
			Docs: []string{
				"Perform install process when device booted from USB-drive.",
			},
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
		"Perform RPM config verification": {
			Docs: []string{
				"Verify if RPM verification is required fo setup",
				"Setup with PD control temprarely excluded from testing.",
			},
			Dependencies: []string{
				"Setup has servo info",
				"has_rpm_info",
				"Verify servod is responsive",
				// The servo setup has PD control if and only if
				// it supports GSC (Google Security Chip) firmware (e.g. cr50, ti50).
				"Setup does't have Servo PD control",
			},
			ExecName: "sample_pass",
		},
		"Perform RPM config verification for audit tasks": {
			Docs: []string{
				"Verify when the RPM verification is required for setup",
			},
			Dependencies: []string{
				// For audit tasks, we should consider making RPM info
				// a necessary condition for running this check.
				// For right now, let's let the audit task fail if the RPM info is absent.
				"has_rpm_info",

				// For audit, we don't need a servo host.
				// - "dut_servo_host_present"
				// - "servod_echo"
				//
				// For audit tests, we do not require the servo to have PD control.
				// For that reason, the dependency below is excluded.
				// - "Setup doesn't have Servo PD control"
			},
			Conditions: []string{
				// Since we're performing an audit task, this action is always applicable.
			},
			ExecName: "sample_pass",
		},
		"Setup has Servo PD control": {
			Docs: []string{
				"Verify that servo has build in PD control.",
			},
			Conditions: []string{
				"Setup has servo info",
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
		"Audit RPM config (without battery)": {
			Docs: []string{
				"Verify RPM configs and set RPM state",
				// For audit tasks, run the RPM config check regardless of whether
				// the servo uses a cr50 or not.
				//
				// This matches the behavior of '_check_rpm_power_delivery_without_battery' in
				// python labpack.
				//
				// https://chromium.googlesource.com/chromiumos/third_party/labpack/+/refs/heads/main/site_utils/admin_audit/rpm_validator.py#69
				//
				// "Not applicable for cr50 servos based on b/205728276",
			},
			Dependencies: []string{
				"Perform RPM config verification for audit tasks",
			},
			Conditions: []string{
				"No Battery is present on device",
			},
			ExecName:    "rpm_audit_without_battery",
			ExecTimeout: &durationpb.Duration{Seconds: 600},
		},
		"Verify RPM config (without battery)": {
			Docs: []string{
				"Verify RPM configs and set RPM state",
				"Not applicable for cr50 servos based on b/205728276",
				"Action is not critical as it updates own state.",
			},
			Conditions: []string{
				"Perform RPM config verification",
				"No Battery is present on device",
			},
			Dependencies: []string{
				"Wait to be SSHable (normal boot)",
			},
			ExecName:    "rpm_audit_without_battery",
			ExecTimeout: &durationpb.Duration{Seconds: 600},
		},
		"Audit RPM config with battery": {
			Docs: []string{
				"Verify RPM when battery is present",
				// For audit tasks, run the RPM config check regardless of whether
				// the servo uses a cr50 or not.
				//
				// This matches the behavior of '_check_rpm_power_delivery_with_battery' in
				// python labpack.
				//
				// https://chromium.googlesource.com/chromiumos/third_party/labpack/+/refs/heads/main/site_utils/admin_audit/rpm_validator.py#110
				//
				// "Not applicable for cr50 servos based on b/205728276",
				"Action is not critical as it updates own state.",
			},
			Conditions: []string{
				"Perform RPM config verification",
				"Battery is present on device",
			},
			ExecName:    "rpm_audit_with_battery",
			ExecTimeout: &durationpb.Duration{Seconds: 600},
			ExecExtraArgs: []string{
				"timeout:120",
				"wait_interval:5",
			},
		},
		"Verify RPM config with battery": {
			Docs: []string{
				"Verify RPM when battery is present",
				"Not applicable for cr50 servos based on b/205728276",
				"Action is not critical as it updates own state.",
			},
			Conditions: []string{
				"Perform RPM config verification",
				"Battery is present on device",
			},
			Dependencies: []string{
				"Wait to be SSHable (normal boot)",
			},
			ExecName:    "rpm_audit_with_battery",
			ExecTimeout: &durationpb.Duration{Seconds: 600},
			ExecExtraArgs: []string{
				"timeout:120",
				"wait_interval:5",
			},
		},
		"Is servod running": {
			Docs: []string{
				"Verify that servo host specified in setup and servod is running.",
			},
			Dependencies: []string{
				"Setup has servo info",
				"Verify servod is responsive",
			},
			ExecName: "sample_pass",
		},
		"Setup has servo info": {
			ExecName: "dut_servo_host_present",
			MetricsConfig: &MetricsConfig{
				UploadPolicy: MetricsConfig_SKIP_ALL,
			},
		},
		"Verify servod is responsive": {
			Conditions: []string{
				"Setup has servo info",
			},
			ExecName:    "servod_echo",
			ExecTimeout: &durationpb.Duration{Seconds: 10},
			ExecExtraArgs: []string{
				"ssh_check:false",
			},
			RunControl: RunControl_ALWAYS_RUN,
			MetricsConfig: &MetricsConfig{
				UploadPolicy: MetricsConfig_UPLOAD_ON_ERROR,
			},
		},
		"Record type C status": {
			Docs: []string{
				"Record the type C status reported by the DUT",
			},
			Conditions: []string{
				"Device is SSHable",
			},
			ExecName:               "cros_log_typec_status",
			AllowFailAfterRecovery: true,
		},
		"Verify RO_VPD sku_number": {
			Docs: []string{
				"Verify that sku_number is present in RO_VPD, if required.",
			},
			Conditions: []string{
				"Is not Flex device",
				"RO_VPD sku_number is required",
			},
			ExecName: "cros_verify_ro_vpd_sku_number",
			RecoveryActions: []string{
				"Set fake RO_VPD sku_number",
			},
			AllowFailAfterRecovery: true,
		},
		"RO_VPD sku_number is required": {
			Docs: []string{
				"Verifies if RO_VPD sku_number is required on this device.",
			},
			ExecName: "cros_is_ro_vpd_sku_number_required",
		},
		"Set fake RO_VPD sku_number": {
			Docs: []string{
				"Set a fake sku_number in RO_VPD",
			},
			ExecName: "cros_set_fake_ro_vpd_sku_number",
			ExecExtraArgs: []string{
				"sku_number:FAKE-SKU",
			},
			AllowFailAfterRecovery: true,
		},
		"Read RO_VPD from DUT": {
			Docs: []string{
				"Record data that is present in RO_VPD from the following keys: wifi_sar.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_update_ro_vpd_inventory",
		},
		"Verify RO_VPD data on DUT": {
			Docs: []string{
				"Verify if data from RO_VPD key: wifi_sar that was present on deploy is still present.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_match_ro_vpd_inventory",
			RecoveryActions: []string{
				"Restore RO_VPD on DUT",
			},
		},
		"Restore RO_VPD on DUT": {
			Docs: []string{
				"Restore data from RO_VPD key: wifi_sar that was present on deploy.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_set_ro_vpd",
		},
		"Check if OS on required version for camerabox tablet": {
			Docs: []string{
				"Check if the camerabox tablet os is the same as required version",
			},
			Dependencies: []string{
				"Internal storage is responsive",
				"Read OS version",
			},
			Conditions: []string{
				"is_camerabox_tablet_pool",
			},
			ExecName: "is_camerabox_tablet_on_os_version",
			RecoveryActions: []string{
				"provision camerabox tablet",
			},
		},
		"provision camerabox tablet": {
			Docs: []string{
				"Provision camerabox tablet",
			},
			ExecName:    "provision_camerabox_tablet",
			ExecTimeout: &durationpb.Duration{Seconds: 3600},
		},
		"is_camerabox_tablet_pool": {
			Docs: []string{
				"Verify device is in camerabox_tablet pool.",
			},
			ExecName: "dut_is_in_pool",
			ExecExtraArgs: []string{
				"camerabox_tablet",
			},
		},
		"Check if request labstation reboot": {
			Docs: []string{
				"Check if there's a need to reboot the connected labstation.",
				"Request labstation reboot may bring DUT back but it happens",
				"in an asynchronous task so we make this action always fail",
				"to avoid an unnecessary retry of critical actions that triggered",
				"this action, e.g. 'Device is pingable' will get retried.",
			},
			Dependencies: []string{
				"Check if servo is not connected by hub",
				"Create request to reboot labstation",
			},
			ExecName: "sample_fail",
		},
		"Check if servo is not connected by hub": {
			Docs: []string{
				"Check if the servo is not connected by hub.",
			},
			ExecName: "servo_not_connected_by_hub",
		},
		"Create request to reboot labstation": {
			Docs: []string{
				"Try to create reboot flag file request in the connected labstation.",
			},
			ExecName:   "labstation_create_reboot_request",
			RunControl: RunControl_RUN_ONCE,
		},
		"Verify rootfs is on fs-verity": {
			Docs: []string{
				"Run rootdev to check rootfs",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName:               "cros_verify_rootfs_verity",
			RunControl:             RunControl_RUN_ONCE,
			AllowFailAfterRecovery: true,
		},
		"Set dev_boot_usb is enabled": {
			Docs: []string{
				"Set dev_boot_usb=1 to enable booting from USB drive in DEV mode.",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName: "cros_update_crossystem",
			ExecExtraArgs: []string{
				"command:dev_boot_usb",
				"value:1",
				"check_after_update:true",
			},
			RunControl:             RunControl_ALWAYS_RUN,
			AllowFailAfterRecovery: true,
		},
		"Collect logs and crashinfo": {
			Docs: []string{
				"We collect any pre-existing logs from before deletes such logs ",
				"on the DUT. Any logs collection are not critical, and we marks ",
				"that action attempt to perform to avoid repeating it.",
			},
			Conditions: []string{
				"Device is SSHable",
				"Confirm log collection info does not exist",
			},
			Dependencies: []string{
				"Create log collection info",
				"Collect logs from DUT on /var/log/*",
				"Collect dmesg",
				"Collect crash dumps",
			},
			ExecName:               "sample_pass",
			RunControl:             RunControl_RUN_ONCE,
			AllowFailAfterRecovery: true,
		},
		"Collect logs from DUT on /var/log/*": {
			Docs: []string{
				"We collect any pre-existing logs before executing repairs on ",
				"the DUT. Any failures with this initial log collection are ",
				"not critical, and we will proceed with the actual DUT repair ",
				"immediately after this.",
			},
			Conditions: []string{
				"Device is SSHable",
			},
			ExecName: "cros_copy_to_logs",
			ExecExtraArgs: []string{
				"src_host_type:dut",
				"src_path:/var/log",
				"src_type:dir",
				"use_host_dir:true",
				"dest_suffix:prior_logs",
			},
			RunControl:             RunControl_RUN_ONCE,
			AllowFailAfterRecovery: true,
		},
		"Collect dmesg": {
			Docs: []string{
				"We collect the dmesg output.",
			},
			Conditions: []string{
				"Device is SSHable",
			},
			ExecName: "cros_dmesg",
			ExecExtraArgs: []string{
				"human_readable:false",
				"create_crashinfo_dir:true",
			},
			RunControl:             RunControl_RUN_ONCE,
			AllowFailAfterRecovery: true,
		},
		"Collect crash dumps": {
			Docs: []string{
				"We collect the crash dumps on the DUT. Additionally, the files ",
				"on the source are deleted, irrespective of whether the ",
				"copy-attempt completes with success or not.",
			},
			Conditions: []string{
				"Device is SSHable",
			},
			ExecName: "cros_collect_crash_dumps",
			ExecExtraArgs: []string{
				"clean:true",
				"cleanup_timeout:10",
			},
			RunControl:             RunControl_RUN_ONCE,
			AllowFailAfterRecovery: true,
		},
		"Create log collection info": {
			Docs: []string{
				"When the log collection completes, we create an info file that ",
				"indicates the successful completion of the collection process.",
			},
			Conditions: []string{
				"Confirm log collection info does not exist",
			},
			ExecName: "cros_create_log_collection_info",
			ExecExtraArgs: []string{
				"info_file:log_collection_info",
			},
			RunControl: RunControl_RUN_ONCE,
		},
		"Confirm log collection info does not exist": {
			Docs: []string{
				"We need to check whether the log collection info file already ",
				"exists in the file system. A pre-existing file indicates that ",
				"the collection of any pre-existing logs has already been ",
				"tried to be collected.",
			},
			ExecName: "cros_confirm_file_not_exists",
			ExecExtraArgs: []string{
				"target_file:log_collection_info",
			},
		},
		"Erase DUT MRC cache via servo": {
			Docs: []string{
				"This action erases MRC cache of the DUT via applicable flash device(servo_micro, CCD) of its servo.",
			},
			Dependencies: []string{
				"Is servod running",
			},
			ExecName:    "cros_erase_mrc_cache_by_servo",
			RunControl:  RunControl_RUN_ONCE,
			ExecTimeout: &durationpb.Duration{Seconds: 1200},
		},
		"Sleep 60 seconds": {
			ExecName: "sample_sleep",
			ExecExtraArgs: []string{
				"sleep:60",
			},
			ExecTimeout: &durationpb.Duration{
				Seconds: 70,
			},
			MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
			RunControl:    RunControl_ALWAYS_RUN,
		},
		// This action is intended to be served as an temporary solution for b/255617349,
		// and it should not to be added as part of formal auto-repair task.
		// TODO(xianuowang@) remove this action once a permanent solution is in place.
		"Disable software write protection via servo": {
			Docs: []string{
				"Disable software write protection(for flash firmware) via servo.",
			},
			Dependencies: []string{
				"Is servod running",
			},
			ExecName:               "cros_disable_software_write_protection",
			ExecTimeout:            &durationpb.Duration{Seconds: 180},
			AllowFailAfterRecovery: true,
		},
		// This action is intended to be served as an temporary solution for b/255617349,
		// and it should not to be added as part of formal auto-repair task.
		// TODO(xianuowang@) remove this action once a permanent solution is in place.
		"Enable SPI mode via servo": {
			Docs: []string{
				"Enable SPI mode for flashing CPU firmware over servo.",
			},
			Dependencies: []string{
				"Is servod running",
			},
			ExecName:               "cros_enable_cpu_fw_spi",
			AllowFailAfterRecovery: true,
		},
		// This action is intended to be served as an temporary solution for b/255617349,
		// and it should not to be added as part of formal auto-repair task.
		// TODO(xianuowang@) remove this action once a permanent solution is in place.
		"Disable SPI mode via servo": {
			Docs: []string{
				"Disable SPI mode via servo",
			},
			Dependencies: []string{
				"Is servod running",
			},
			ExecName:               "cros_disable_cpu_fw_spi",
			AllowFailAfterRecovery: true,
		},
		"Has a stable-version service": {
			Docs: []string{
				"Verify if we have access to the service provided access to the stable version",
			},
			ExecName:   "has_stable_version_service_path",
			RunControl: RunControl_RUN_ONCE,
		},
	}
}
