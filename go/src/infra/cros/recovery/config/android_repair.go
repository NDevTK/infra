// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"path/filepath"

	"google.golang.org/protobuf/types/known/durationpb"
)

// AndroidRepairConfig provides config for repair android task.
func AndroidRepairConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanAndroid,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanAndroid: setAllowFail(androidRepairPlan(), false),
			PlanClosing: setAllowFail(androidClosePlan(), true),
		}}
}

func androidRepairPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Set state: repair_failed",
			"Validate DUT info",
			"Validate associated host",
			"Lock associated host",
			"Associated host has vendor key",
			"Validate adb",
			"DUT is accessible over adb",
			"Reset DUT",
			"Configure DUT",
			"Set state: ready",
		},
		Actions: androidRepairDeployActions(),
	}
}

func androidRepairDeployActions() map[string]*Action {
	return map[string]*Action{
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
		"Set state: repair_failed": {
			Docs: []string{
				"Initial state of Android DUT before repair to indicate failure if recovery fails by any reason.",
			},
			ExecName:      "dut_set_state",
			ExecExtraArgs: []string{"state:repair_failed"},
			RunControl:    RunControl_RUN_ONCE,
			MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
		},
		"Set state: ready": {
			Docs: []string{
				"Final state of Android DUT indicating successful repair.",
			},
			ExecName:      "dut_set_state",
			ExecExtraArgs: []string{"state:ready"},
			RunControl:    RunControl_RUN_ONCE,
			MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
		},
		"Validate DUT info": {
			Docs: []string{"Check Android DUT info for repair."},
			Dependencies: []string{
				"dut_has_name",
				"android_dut_has_board_name",
				"android_dut_has_model_name",
				"android_dut_has_serial_number",
				"android_dut_has_associated_host",
			},
			ExecName: "sample_pass",
		},
		"Validate associated host": {
			Docs: []string{"Check availability of associated host of the DUT."},
			Dependencies: []string{
				"android_dut_has_serial_number",
				"android_dut_has_associated_host",
				"Associated host is pingable",
				"Associated host is accessible over SSH",
				"Associated host is labstation",
			},
			ExecName: "android_associated_host_fs_is_writable",
			RecoveryActions: []string{
				"Schedule associated host reboot and fail",
			},
		},
		"Associated host is pingable": {
			Docs: []string{
				"This verifier checks whether associated host of the DUT is reachable over ping. ",
				"This should happen as soon as the network driver gets loaded and the network becomes operational.",
			},
			ExecName:    "android_associated_host_ping",
			ExecTimeout: &durationpb.Duration{Seconds: 15},
		},
		"Associated host is accessible over SSH": {
			Docs: []string{
				"This verifier checks whether associated host of the DUT is accessible over ssh.",
			},
			ExecName:    "android_associated_host_ssh",
			ExecTimeout: &durationpb.Duration{Seconds: 30},
		},
		"Associated host is labstation": {
			Docs: []string{
				"This verifier checks whether associated host of the DUT is a labstation.",
			},
			ExecName: "android_associated_host_is_labstation",
		},
		"Lock associated host": {
			Docs: []string{
				"Creates a file to indicate that the associated host is in use.",
			},
			Dependencies: []string{
				"Validate associated host",
			},
			ExecName:   "android_associated_host_lock",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Validate adb": {
			Docs: []string{
				"This verifier checks whether adb and vendor key are properly provisioned on associated host of the DUT and adb server is running.",
			},
			Dependencies: []string{
				"Validate associated host",
				"Associated host has vendor key",
				"Associated host has adb",
				"Adb server is stopped",
			},
			ExecName:      "android_associated_host_start_adb",
			ExecExtraArgs: []string{"adb_vendor_key:" + filepath.Dir(adbPrivateVendorKeyFile)},
			RecoveryActions: []string{
				"Schedule associated host reboot and fail",
			},
		},
		"Associated host has vendor key": {
			Docs: []string{
				"This verifier checks whether a valid vendor key is provisioned to associated host of the DUT.",
			},
			ExecName:      "android_associated_host_has_vendor_key",
			ExecExtraArgs: []string{"adb_vendor_key:" + adbPrivateVendorKeyFile},
			RecoveryActions: []string{
				"Restore private vendor key on associated host",
			},
		},
		"Restore private vendor key on associated host": {
			Docs: []string{
				"This action restore the a pre-defined vendor key on the associated host.",
			},
			ExecName: "android_associated_host_restore_vendor_key",
			ExecExtraArgs: []string{
				"vendor_key_file:" + adbPrivateVendorKeyFile,
				"vendor_key_content:" + adbPrivateVendorKey,
			},
		},
		"Associated host has adb": {
			Docs: []string{
				"This verifier checks whether associated host of the DUT has adb installed.",
			},
			ExecName: "android_associated_host_has_adb",
		},
		"Adb server is stopped": {
			Docs: []string{
				"Stops Adb server if it is running on associated host of the DUT.",
			},
			Conditions: []string{
				"android_associated_host_has_no_other_locks",
			},
			ExecName: "android_associated_host_stop_adb",
		},
		"DUT is accessible over adb": {
			Docs: []string{
				"This verifier checks whether the DUT is accessible over adb.",
			},
			Dependencies: []string{
				"Validate associated host",
				"Validate adb",
			},
			ExecName:   "android_dut_is_accessible",
			RunControl: RunControl_ALWAYS_RUN,
			RecoveryActions: []string{
				"Reboot device if in fastboot mode",
				"Reconnect device if in offline state",
				"Schedule associated host reboot and fail",
			},
		},
		"Schedule associated host reboot and fail": {
			Docs: []string{
				"Schedules reboot of the DUT associated host and fails repair till the next run.",
			},
			Conditions: []string{
				"android_dut_has_serial_number",
				"android_dut_has_associated_host",
				"Associated host is pingable",
				"Associated host is accessible over SSH",
				"Associated host is labstation",
			},
			Dependencies: []string{
				"android_associated_host_schedule_reboot",
			},
			ExecName:   "sample_fail",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Unlock DUT screen": {
			Docs: []string{
				"Unlocks DUT screen.",
			},
			Conditions: []string{
				"DUT has userdebug build",
			},
			Dependencies: []string{
				"Validate associated host",
				"Validate adb",
				"Ensure adbd runs as root",
				"android_remove_screen_lock",
				"android_dut_reboot",
				"Wait for Online DUT",
			},
			ExecName:    "sample_pass",
			ExecTimeout: &durationpb.Duration{Seconds: 690},
		},
		"Reset DUT": {
			Docs: []string{"Resets DUT to factory settings."},
			Dependencies: []string{
				"Validate associated host",
				"Validate adb",
				"DUT is accessible over adb",
				"Reset public key",
			},
			ExecName: "android_enable_test_harness",
			RecoveryActions: []string{
				"Unlock DUT screen",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 690},
		},
		"Configure DUT": {
			Docs: []string{"Configures DUT after reset."},
			Dependencies: []string{
				"Wait for DUT to reboot",
				"Connect to WiFi network",
				"Unroot DUT",
			},
			ExecName:    "sample_pass",
			ExecTimeout: &durationpb.Duration{Seconds: 690},
		},
		"DUT has userdebug build": {
			Docs: []string{"This verifier checks whether the DUT has userdebug build."},
			Dependencies: []string{
				"Validate associated host",
				"Validate adb",
			},
			ExecName: "android_dut_has_userdebug_build",
		},
		"Wait for Offline DUT": {
			Docs:       []string{"Waits for DUT to become offline."},
			ExecName:   "android_wait_for_offline_dut",
			RunControl: RunControl_ALWAYS_RUN,
			ExecExtraArgs: []string{
				"timeout:90",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 90},
		},
		"Wait for Online DUT": {
			Docs:       []string{"Waits for DUT to become available."},
			ExecName:   "android_wait_for_online_dut",
			RunControl: RunControl_ALWAYS_RUN,
			ExecExtraArgs: []string{
				"timeout:600",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 600},
		},
		"Wait for DUT to reboot": {
			Docs: []string{"Waits for DUT till it reboots."},
			Dependencies: []string{
				"Wait for Offline DUT",
				"Wait for Online DUT",
			},
			ExecName:    "sample_pass",
			ExecTimeout: &durationpb.Duration{Seconds: 690},
		},
		"Connect to WiFi network": {
			Docs: []string{"Connects DUT to WiFi network."},
			Dependencies: []string{
				"Validate associated host",
				"Validate adb",
				"DUT is accessible over adb",
				"Enable WiFi",
			},
			ExecName: "android_connect_wifi_network",
			ExecExtraArgs: []string{
				"wifi_ssid:" + wifiSSID,
				"wifi_security:" + wifiSecurity,
				"wifi_password:" + wifiPassword,
				"retry_interval:5",
				"timeout:180",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 180},
		},
		"Enable WiFi": {
			Docs: []string{"Enables WiFi on DUT."},
			Dependencies: []string{
				"Validate associated host",
				"Validate adb",
				"DUT is accessible over adb",
				"Ensure adbd runs as root",
			},
			ExecName: "android_enable_wifi",
			ExecExtraArgs: []string{
				"retry_interval:5",
				"timeout:180",
			},
			ExecTimeout: &durationpb.Duration{Seconds: 180},
		},
		"Ensure adbd runs as root": {
			Docs: []string{"Restart adbd with root permission."},
			Conditions: []string{
				"DUT has userdebug build",
			},
			ExecName:   "android_restart_adbd_as_root",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Unroot DUT": {
			Docs: []string{
				"Ensures that adbd on DUT runs without root permissions after successful repair. If repair fails, this action is not required.",
			},
			Conditions: []string{
				"DUT has userdebug build",
			},
			ExecName:   "android_unroot_adbd",
			RunControl: RunControl_ALWAYS_RUN,
			// The action is not critical. It should not fail the repair process.
			AllowFailAfterRecovery: true,
		},
		"Reset public key": {
			Docs: []string{"Validates and restores ADB public vendor key."},
			Conditions: []string{
				"DUT has userdebug build",
			},
			Dependencies: []string{
				"Validate associated host",
				"Validate adb",
				"Ensure adbd runs as root",
			},
			ExecName: "android_reset_public_key",
			ExecExtraArgs: []string{
				"public_key:" + adbPublicVendorKey,
				"public_key_file:" + adbPublicVendorKeyFile,
			},
		},
		"Reboot device if in fastboot mode": {
			Docs: []string{
				"Reboot the device via fastboot if the device is in fastboot mode.",
			},
			Dependencies: []string{
				"android_device_in_fastboot_mode",
				"android_reboot_device_via_fastboot",
				"Wait for Online DUT",
			},
			ExecName:    "sample_pass",
			ExecTimeout: &durationpb.Duration{Seconds: 690},
		},
		"Reconnect device if in offline state": {
			Docs: []string{
				"Reconnect device if the device is in offline state.",
			},
			Conditions: []string{
				"android_dut_is_offline",
			},
			ExecName: "android_reconnect_offline_dut",
		},
	}
}

// androidClosePlan provides plan to close android repair tasks.
func androidClosePlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Unlock associated host",
		},
		Actions: map[string]*Action{
			"Unlock associated host": {
				Docs: []string{
					"Removes in-use file to release the associated host.",
				},
				Conditions: []string{
					"android_dut_has_serial_number",
					"android_dut_has_associated_host",
					"android_associated_host_ssh",
					"android_associated_host_is_labstation",
					"android_associated_host_fs_is_writable",
				},
				ExecName:   "android_associated_host_unlock",
				RunControl: RunControl_ALWAYS_RUN,
			},
		},
	}
}
