// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"log"

	"google.golang.org/protobuf/types/known/durationpb"
)

func crosDeployPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Set state: needs_deploy",
			"Check stable versions exist",
			"Clean up",
			"Download stable version OS image to servo usbkey if necessary",
			"Device is pingable before deploy",
			"DUT is on test channel OS",
			"Set dev_boot_usb is enabled",
			"DUT has expected dev firmware",
			"Switch to secure-mode and reboot",
			"Deployment checks",
			"Collect DUT labels",
			"DUT verify",
		},
		Actions: crosDeployAndRepairActions(),
	}
}

func deployActions() map[string]*Action {
	// Prepare critical actions as part of DUT verify.
	var repairCriticalActions []string
	for _, a := range crosRepairPlan().GetCriticalActions() {
		// Exclude repair state to keep need_deploy state as default.
		if a == "Set state: repair_failed" {
			continue
		}
		repairCriticalActions = append(repairCriticalActions, a)
	}

	return map[string]*Action{
		"DUT is in dev-mode and allowed to boot from USB-key": {
			Docs:        []string{"Verify that device is set to boot in DEV mode and enabled to boot from USB-drive."},
			ExecTimeout: &durationpb.Duration{Seconds: 2000},
			ExecName:    "cros_read_gbb_by_servo",
			ExecExtraArgs: []string{
				"validate_in_dev_mode:true",
				"validate_usb_boot_enabled:true",
				"remove_file:false",
			},
			RecoveryActions: []string{
				"Switch DUT to dev mode by servo",
			},
		},
		"Device is pingable before deploy": {
			Docs: []string{
				"Verify that device is present in setup.",
				"All devices is pingable by default even they have prod images on them.",
				"If device is not pingable then device is off on not connected",
			},
			ExecName:    "cros_ping",
			ExecTimeout: &durationpb.Duration{Seconds: 15},
			RecoveryActions: []string{
				"Cold reset DUT by servo and wait to boot",
				"Power cycle DUT by RPM and wait",
			},
		},
		"DUT is on test channel OS": {
			Docs: []string{
				"Verify that device has OS version from test channel, if not then install it.",
			},
			Dependencies: []string{
				"Device is pingable before deploy",
				"Recovery version has OS image path",
				"Device NOT booted from USB-drive",
			},
			ExecName: "cros_is_os_test_channel",
			RecoveryActions: []string{
				"Quick provision OS",
				"Install OS in DEV mode",
				"Install OS in DEV mode, with force to DEV-mode",
				"Install OS in DEV mode with fresh image",
				"Install OS in DEV mode, with force to DEV-mode with test firmware",
			},
		},
		"DUT has expected dev firmware": {
			Docs: []string{
				"Verify that FW on the DUT has dev keys.",
			},
			Conditions: []string{
				//TODO(b:231627918): Flex does not have own firmware for EC/AP
				"Is not Flex device",
			},
			Dependencies: []string{
				"Device is SSHable",
			},
			ExecName:    "cros_has_dev_signed_firmware",
			ExecTimeout: &durationpb.Duration{Seconds: 600},
			RecoveryActions: []string{
				"Update DUT firmware with factory mode and restart by servo",
				"Update DUT firmware with factory mode and restart by host",
			},
		},
		"Update DUT firmware with factory mode and restart by servo": {
			Docs: []string{
				"Force update FW on the DUT by factory mode.",
				"Reboot device by servo",
			},
			Conditions: []string{
				"dut_servo_host_present",
				"servod_echo",
			},
			Dependencies: []string{
				"Device is SSHable",
				"Disable software-controlled write-protect for 'host'",
				"Disable software-controlled write-protect for 'ec'",
				"Try to update FW from firmware image with factory mode",
				"Try to update FW from OS image with factory mode",
				"Cold reset DUT by servo",
				"Wait to be pingable (normal boot)",
				"Wait to be SSHable (normal boot)",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Update DUT firmware with factory mode and restart by host": {
			Docs: []string{
				"Force update FW on the DUT by factory mode.",
				"Reboot device by host",
			},
			Dependencies: []string{
				"Device is SSHable",
				"Disable software-controlled write-protect for 'host'",
				"Disable software-controlled write-protect for 'ec'",
				"Try to update FW from firmware image with factory mode",
				"Try to update FW from OS image with factory mode",
				"Simple reboot",
				"Wait to be pingable (normal boot)",
				"Wait to be SSHable (normal boot)",
			},
			ExecName:   "sample_pass",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Try to update FW from firmware image with factory mode": {
			Docs: []string{
				"Update firmware from faft stable image with chromeos-firmwareupdate tool",
				"--mode=facotry will be specified when run chromeos-firmwareupdate",
				"Set timeout to 120 minutes = 10 minutes for download + 100 minutes for find and extract AP/EC images + 10 minutes for run updater.",
			},
			Conditions: []string{
				"has_stable_version_fw_image",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 7200},
			ExecExtraArgs: []string{
				"mode:factory",
				"force:true",
				"updater_timeout:600",
				"update_ec_attempt_count:1",
				"update_ap_attempt_count:1",
				"no_strict_update:true",
			},
			ExecName:   "cros_update_firmware_from_firmware_image",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Try to update FW from OS image with factory mode": {
			Docs: []string{
				"Run chromeos-firmwareupdate with factory mode.",
				"The reboot is not triggered as part of the action.",
				"The action is not strict to not block repair actions.",
				"Only runs when the DUT doesn't have a model specific faft stable_version, e.g. it's an early stage device use satlab or flex device.",
			},
			Conditions: []string{
				"Missing stable fw image",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 900},
			ExecName:    "cros_run_firmware_update",
			ExecExtraArgs: []string{
				"mode:factory",
				"force:true",
				"updater_timeout:600",
				"no_strict_update:true",
			},
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Need to run deployment checks": {
			Docs: []string{
				"Check if deployment check not need to be run.",
				"If HWID or serial-number already collected from DUT then we already test it before.",
			},
			Conditions: []string{
				"Is HWID known",
				"Is serial-number known",
			},
			ExecName: "sample_fail",
		},
		"Deployment checks": {
			Docs: []string{
				"Run some special checks as part of deployment.",
			},
			Conditions: []string{
				"Not Satlab device",
				"Need to run deployment checks",
			},
			Dependencies: []string{
				"Verify battery charging level",
				"Verify boot in recovery mode",
				"Verify RPM config (without battery)",
				"Verify RPM config with battery",
				"Wait to be SSHable (normal boot)",
			},
			ExecName: "sample_pass",
		},
		"Verify battery charging level": {
			Docs: []string{
				"Battery will be checked that it can be charged to the 80% as if device cannot then probably device is not fully prepared for deployment.",
				"If battery is not charged, then we will re-check every 15 minutes for 8 time to allows to charge the battery.",
				"Dues overheat battery in audio boxes mostly it deployed ",
			},
			Conditions: []string{
				"Is not in audio box",
				"Battery is expected on device",
				"Battery is present on device",
			},
			Dependencies: []string{
				"Wait to be SSHable (normal boot)",
			},
			ExecName: "cros_battery_changable_to_expected_level",
			ExecExtraArgs: []string{
				"charge_retry_count:8",
				"charge_retry_interval:900",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 9000},
		},
		"Verify boot in recovery mode": {
			Docs: []string{
				"Devices deployed with servo in the pools required secure mode need to be able to be boot in recovery mode.",
			},
			Conditions: []string{
				"Pools required to be in Secure mode",
			},
			Dependencies: []string{
				"dut_servo_host_present",
				"servod_echo",
				"Wait to be SSHable (normal boot)",
			},
			ExecName: "cros_verify_boot_in_recovery_mode",
			ExecExtraArgs: []string{
				"boot_timeout:480",
				"boot_interval:10",
				"halt_timeout:120",
				"ignore_reboot_failure:false",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 1200},
			RecoveryActions: []string{
				"Cold reset DUT by servo and wait to boot",
				// The other reason why it fail on good DUT is that USB-key has not good image.
				"Download stable image to USB-key",
			},
		},
		"DUT verify": {
			Docs: []string{
				"Run all repair critcal actions.",
			},
			Dependencies: repairCriticalActions,
			ExecName:     "sample_pass",
		},
		"Install OS in DEV mode": {
			Docs: []string{
				"Install OS on the device from USB-key when device is in DEV-mode.",
			},
			Conditions: []string{
				"dut_servo_host_present",
				"servod_echo",
			},
			Dependencies: []string{
				"Install OS in DEV mode by USB-drive",
			},
			ExecName: "sample_pass",
		},
		"Install OS in DEV mode with fresh image": {
			Docs: []string{
				"Download fresh usb image and Install OS from it in DEV-mode.",
			},
			Conditions: []string{
				"dut_servo_host_present",
				"servod_echo",
			},
			Dependencies: []string{
				"Download stable image to USB-key",
				"Install OS in DEV mode by USB-drive",
			},
			ExecName: "sample_pass",
		},
		"Install OS in DEV mode, with force to DEV-mode": {
			Docs: []string{
				"Install OS on the device from USB-key when device is in DEV-mode.",
			},
			Conditions: []string{
				"dut_servo_host_present",
				"servod_echo",
			},
			Dependencies: []string{
				"Set GBB flags to 0x18 by servo",
				"Install OS in DEV mode by USB-drive",
			},
			ExecName: "sample_pass",
		},
		"Install OS in DEV mode, with force to DEV-mode with test firmware": {
			Docs: []string{
				"Second attempt to install image in DEV mode",
			},
			Conditions: []string{
				"dut_servo_host_present",
				"servod_echo",
			},
			Dependencies: []string{
				"Update FW from fw-image by servo and set GBB to 0x18",
				"Install OS in DEV mode by USB-drive",
			},
			ExecName: "sample_pass",
		},
		"Clean up": {
			Docs: []string{
				"Verify that device is set to boot in DEV mode and enabled to boot from USB-drive.",
			},
			Conditions: []string{
				"dut_servo_host_present",
			},
			Dependencies: []string{
				"cros_remove_default_ap_file_servo_host",
			},
			ExecName: "sample_pass",
		},
		"Collect DUT labels": {
			Docs: []string{
				"Updating device info in inventory.",
			},
			Dependencies: []string{
				"Read HWID from DUT",
				"Read HWID from DUT (Satlab)",
				"Read DUT serial-number from DUT",
				"Read DUT serial-number from DUT (Satlab)",
				"Read device SKU",
				"servo_type_label",
				"Read RO_VPD from DUT",
			},
			ExecName: "sample_pass",
		},
		"servo_type_label": {
			Docs: []string{
				"Update the servo type label for the DUT info.",
			},
			ExecName:               "servo_update_servo_type_label",
			AllowFailAfterRecovery: true,
		},
		"Check stable versions exist": {
			Docs: []string{
				"Check the DUT has model specific cros, firmware and faft stable_version configured.",
			},
			Dependencies: []string{
				"has_stable_version_cros_image",
				"Check stable firmware version exists",
				// Disabled faft version check until b/241150358 got resolved.
				//"Check stable faft version exists",
			},
			ExecName: "sample_pass",
		},
		"Check stable firmware version exists": {
			Docs: []string{
				"Check the DUT has model specific firmware stable_version configured.",
				"Flex device are exampted from this check as they don't run cros firmware",
			},
			Conditions: []string{
				"Is not Flex device",
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
			},
			ExecName: "has_stable_version_fw_image",
		},
		"Missing stable fw image": {
			Docs: []string{
				"Verify that the DUT doesn't have model specific stable_version record in faft section",
			},
			Conditions: []string{
				"has_stable_version_fw_image",
			},
			ExecName: "sample_fail",
		},
	}
}

func crosDeployAndRepairActions() map[string]*Action {
	combo := deployActions()
	for name, action := range crosRepairActions() {
		if _, ok := combo[name]; ok {
			log.Fatalf("duplicate name in crosDeploy and crosRepair plan actions: %s", name)
		}
		combo[name] = action
	}
	return combo
}
