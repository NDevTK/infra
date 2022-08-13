// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"google.golang.org/protobuf/types/known/durationpb"
)

func servoRepairPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Set state:MISSING_CONFIG",
			"Servo is know in the setup",
			"Set state:WRONG_CONFIG",
			"Servod port specified",
			"Servo serial is specified",
			"Set state:SERVO_HOST_ISSUE",
			"Initialize docker container",
			"Set state:NO_SSH",
			"Device is pingable",
			"Device is SSHable",
			"Servo_v3 uptime is not long",
			"Power-cycle by smart-hub",
			"Power-cycle servo-v4p1 network",
			"Set state:SERVO_HOST_ISSUE",
			"Mark labstation as servod is in-use",
			"Set state:BROKEN",
			"Has enough free disk space",
			"Cache latest servod start time",
			"Set state:NOT_CONNECTED",
			"Servo_v4(p1) main present",
			"Set state:NEED_REPLACEMENT",
			"Servo_v3 root present",
			"Set state:SERVO_UPDATER_ISSUE",
			"All servo's fw updated",
			"Set state:SERVO_HOST_ISSUE",
			"Start servod daemon",
			"Set state:SERVOD_ISSUE",
			"Servod is responsive to dut-control",
			"Set state:SERVO_HOST_ISSUE",
			"Read servo serial by servod harness",
			"Set state:DUT_NOT_CONNECTED",
			"Verify servo connected to the DUT",
			"Set state:COLD_RESET_PIN_ISSUE",
			"Cold reset pin is detected",
			"Set state:WARM_RESET_PIN_ISSUE",
			"Warm reset pin is detected",
			"Set state:SERVOD_ISSUE",
			"Check if PD is src state",
			"Verify Cr50 detected",
			"Set state:DUT_NOT_CONNECTED",
			"Servod detect all children components",
			"Set state:TOPOLOGY_ISSUE",
			"Servo topology",
			"Verify that USB drive is detectable",
			"Update USB drive info",
			"Set state:SERVOD_PROXY_ISSUE",
			"Initialize DUT part for servo",
			"Set state:CR50_CONSOLE_MISSING",
			"Verify cr50 console",
			"Set state:CCD_TESTLAB_ISSUE",
			"Cr50 testlab is enabled",
			"Verify EC",
			"Set state:BROKEN",
			"Record good servo type",
			"Set state:WORKING",
		},
		Actions: map[string]*Action{
			"Servo is know in the setup": {
				Docs: []string{
					"Verify if setup data has any data related to servo-host which mean servo is present in setup.",
				},
				ExecName: "dut_servo_host_present",
			},
			"Servo serial is specified": {
				Docs: []string{
					"Check if root servo serial is present.",
				},
				Conditions: []string{
					"Is not servo_v3",
				},
				ExecName: "dut_servo_has_serial",
			},
			"Device is pingable": {
				Docs: []string{
					"Verify that device is reachable by ping.",
					"Limited to 15 seconds.",
				},
				ExecName: "cros_ping",
				ExecTimeout: &durationpb.Duration{
					Seconds: 15,
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Device is SSHable": {
				Docs: []string{
					"Verify that device is reachable by SSH.",
					"Limited to 15 seconds.",
				},
				ExecTimeout: &durationpb.Duration{Seconds: 15},
				ExecName:    "cros_ssh",
				RunControl:  RunControl_ALWAYS_RUN,
			},
			"Cache latest servod start time": {
				Docs: []string{
					"Cache servod start time based on previous runs.",
					"If we fail all logs will be collected",
				},
				Conditions: []string{
					"Is not servo_v3",
				},
				ExecName:               "cros_register_servod_logs_start",
				AllowFailAfterRecovery: true,
			},
			"Start servod daemon": {
				Docs: []string{
					"Start servod daemon on servo-host",
				},
				ExecName:    "servo_host_servod_init",
				ExecTimeout: &durationpb.Duration{Seconds: 120},
				RecoveryActions: []string{
					"Stop servod",
					"Create request to reboot labstation",
				},
			},
			"Stop servod": {
				Docs: []string{
					"Stop the servod daemon.",
					"Allowed to fail as can be run when servod is not running.",
				},
				ExecName:               "servo_host_servod_stop",
				RunControl:             RunControl_ALWAYS_RUN,
				AllowFailAfterRecovery: true,
			},
			"Initialize docker container": {
				Docs: []string{
					"Initiate docker to have access to the host.",
					"Servod is not needed as on this stage we just verify that servo host is good.",
					"If start container with servod and root servo device is not connected it will fail.",
				},
				Conditions: []string{
					"is_container",
				},
				ExecName: "servo_host_servod_init",
				ExecExtraArgs: []string{
					"no_servod:true",
				},
				ExecTimeout: &durationpb.Duration{Seconds: 360},
				RecoveryActions: []string{
					"Stop servod",
				},
			},
			"Stop servod daemon on servo-host": {
				Docs: []string{
					"Make sure servod daemon is not running on servo-host.",
					"If container then run without daemon.",
					"If daemon is running it will be stopped.",
				},
				ExecName: "servo_host_servod_init",
				ExecExtraArgs: []string{
					"no_servod:true",
				},
				ExecTimeout: &durationpb.Duration{Seconds: 360},
			},
			"Servod port specified": {
				Docs: []string{
					"Verify that servod port is present in servo data.",
					"Port is not expected to be specified for servo_V3.",
				},
				Conditions: []string{
					"Is not servo_v3",
				},
				ExecName: "servo_servod_port_present",
			},
			"Servo_v3 uptime is not long": {
				Docs: []string{
					"If servo_v3 is running longer than 96h it can have some issue.",
				},
				Conditions: []string{
					"Is servo_v3 used",
				},
				ExecName: "cros_validate_uptime",
				ExecExtraArgs: []string{
					"max_duration:96",
				},
				RecoveryActions: []string{
					"Simple reboot and wait",
				},
			},
			"Simple reboot and wait": {
				Docs: []string{
					"Reboot host and wait for it to be up.",
				},
				Dependencies: []string{
					"Simple reboot",
					"Wait to be SSHable (normal boot)",
				},
				ExecName:   "sample_pass",
				RunControl: RunControl_ALWAYS_RUN,
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
			"is_labstation": {
				Docs: []string{
					"Condition to check if the servohost is a labstation.",
				},
				ExecName: "servo_host_is_labstation",
			},
			"is_container": {
				Docs: []string{
					"Condition to check if servo uses servod container.",
				},
				ExecName: "servo_uses_servod_container",
			},
			"Is servo_v3 used": {
				Docs: []string{
					"Condition to check if the servo is v3.",
				},
				ExecName: "is_servo_v3",
			},
			"Mark labstation as servod is in-use": {
				Docs: []string{
					"Create lock file is_in_use.",
				},
				Conditions: []string{
					"is_labstation",
				},
				ExecName: "cros_create_servo_in_use",
				RecoveryActions: []string{
					"Sleep 1s",
					"Create request to reboot labstation",
				},
			},
			"Has enough free disk space": {
				Docs: []string{
					"Check if stateful partition have enough disk space that is at least 0.5GB.",
				},
				Conditions: []string{
					"is_not_container",
				},
				ExecName: "cros_has_enough_storage_space",
				ExecExtraArgs: []string{
					"/mnt/stateful_partition:0.5",
				},
				RecoveryActions: []string{
					"Remove logs and other files",
					"Create request to reboot labstation",
					"Reboot servo_v3",
				},
			},
			"Remove logs and other files": {
				Docs: []string{
					"Clean up the old servod files as well as labstation.",
				},
				Dependencies: []string{
					"servo_labstation_disk_cleanup",
					"Remove logs older 5 days",
				},
				ExecName: "sample_pass",
			},
			"Remove logs older 5 days": {
				Docs: []string{
					"Clean up the old servod logs which older than 5 days.",
				},
				ExecName: "servo_servod_old_logs_cleanup",
				ExecExtraArgs: []string{
					"max_days:5",
				},
			},
			"is_not_container": {
				Conditions: []string{"is_container"},
				ExecName:   "sample_fail",
			},
			"Servo topology": {
				Docs: []string{
					"host.check_diskspace('/mnt/stateful_partition', 0.5)",
				},
				Conditions: []string{
					"Is not servo_v3",
				},
				Dependencies: []string{
					"Device is SSHable",
					"Servo topology min one child",
					"Servo topology min two children",
				},
				ExecName: "sample_pass",
			},
			"Servo topology min one child": {
				Conditions: []string{
					"Is not servo_v3",
				},
				ExecName: "servo_topology_update",
				ExecExtraArgs: []string{
					"min_child:1",
					"persist_topology:true",
				},
				RecoveryActions: []string{
					"Stop servod",
					"Toggle PD once and stop",
					"Toggle PD (5 times) and stop",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reflash Cr50 fw and stop",
					"Reset EC from DUT and stop",
					"Create request to reboot labstation",
				},
			},
			"Servo topology min two children": {
				Conditions: []string{
					"Is not servo_v3",
					"is_dual_setup",
				},
				ExecName: "servo_topology_update",
				ExecExtraArgs: []string{
					"min_child:2",
					"persist_topology:true",
				},
				RecoveryActions: []string{
					"Stop servod",
					"Toggle PD once and stop",
					"Toggle PD (5 times) and stop",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reflash Cr50 fw and stop",
					"Reset EC from DUT and stop",
					"Create request to reboot labstation",
				},
			},
			"Servo_v3 root present": {
				Docs: []string{
					"Check is servo_v3 is present.",
				},
				Conditions: []string{
					"Is servo_v3 used",
				},
				ExecName: "servo_v3_root_present",
				RecoveryActions: []string{
					"Reboot servo_v3",
				},
			},
			"Servo_v4(p1) main present": {
				Docs: []string{
					"Verify that servo_v4(p1) board is present",
				},
				Conditions: []string{
					"Is not servo_v3",
				},
				ExecName: "servo_v4_root_present",
				ExecExtraArgs: []string{
					"update_topology:true",
				},
				RecoveryActions: []string{
					"Create request to reboot labstation",
				},
			},
			"All servo's fw updated": {
				Docs: []string{
					"Check whether servo devices required firmware update.",
					"Check running agains version specified by servo_updater channel.",
				},
				Conditions: []string{
					"Is not servo_v3",
				},
				ExecName:    "servo_fw_need_update",
				ExecTimeout: &durationpb.Duration{Seconds: 300},
				RecoveryActions: []string{
					"Sleep 1s", //first try to re-read
					"Update all servo's firmware",
				},
			},
			"Read servo serial by servod harness": {
				Docs: []string{
					"Try to read servo serial by XMLRPC request to servod.",
				},
				ExecName: "servod_echo",
				RecoveryActions: []string{
					"Stop servod",
					"Toggle PD once and stop",
					// Other actions just in case as we do not expect to run them.
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reflash Cr50 fw and stop",
					"Reset EC from DUT and stop",
				},
			},
			"Read ppdut5_mv value": {
				Docs: []string{
					"Read and print ppdut5_mv control value to logs.",
				},
				ExecExtraArgs: []string{
					"command:ppdut5_mv",
				},
				ExecName: "servo_check_servod_control",
			},
			"Read ppchg5_mv value": {
				Docs: []string{
					"Read and print ppchg5_mv control value to logs.",
				},
				ExecExtraArgs: []string{
					"command:ppchg5_mv",
				},
				ExecName: "servo_check_servod_control",
			},
			"Check if PD is src state": {
				Docs: []string{
					"Verify that PD is src power to the DUT.",
					"Action can fail as not always the power is delivered by servo.",
				},
				Conditions: []string{
					"Is servo_v4(p1) used with type-c connector",
				},
				Dependencies: []string{
					"Read ppdut5_mv value",
					"Read ppchg5_mv value",
				},
				ExecName: "servo_check_servod_control",
				ExecExtraArgs: []string{
					"command:servo_pd_role",
					"expected_string_value:src",
				},
				RecoveryActions: []string{
					"Toggle PD once and stop",
					"Toggle PD (5 times) and stop",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
				},
				AllowFailAfterRecovery: true,
			},
			"Verify Cr50 detected": {
				Docs: []string{
					"Run basic cr50/ti50 detections checks.",
				},
				Conditions: []string{
					"Is not servo_v3",
				},
				Dependencies: []string{
					"Set state:SBU_LOW_VOLTAGE",
					"Servo SBU voltage is good",
					"Set state:DUT_NOT_CONNECTED",
					"Servo Cr50 enumerated",
				},
				ExecName: "sample_pass",
			},
			"Servo SBU voltage is good": {
				Docs: []string{
					"Verify that SBU voltage is in expected range (2500mv).",
				},
				Conditions: []string{
					"Is not servo_v3",
					"Is servo_v4(p1) used with type-c connector",
					"Servod detect voltage issue",
				},
				ExecName: "servo_cr50_low_sbu",
				RecoveryActions: []string{
					"Stop servod",
					"Toggle PD once and stop",
					"Toggle PD (5 times) and stop",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reflash Cr50 fw and stop",
					"Reset EC from DUT and stop",
				},
			},
			"Servod detect voltage issue": {
				Docs: []string{
					"Verify that servod is detected required children.",
				},
				ExecName: "servo_check_servod_control",
				ExecExtraArgs: []string{
					"command:dut_sbu_voltage_float_fault",
					"expected_string_value:on",
				},
				RecoveryActions: []string{
					"Stop servod",
					"Toggle PD once and stop",
					"Toggle PD (5 times) and stop",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
					"Reboot by EC console and stop",
					"Reset EC from DUT and stop",
					"Cold reset the DUT by servod and stop",
					"Reflash Cr50 fw and stop",
				},
			},
			"Servo Cr50 enumerated": {
				Docs: []string{
					"Verify that Cr50/GSC is enumerated or not.",
				},
				Conditions: []string{
					"Is not servo_v3",
					"Is servo_v4(p1) used with type-c connector",
					// If this pass then we have issue.
					"Servod detect voltage issue",
				},
				ExecName: "sample_fail",
				// The action failed when issue detected by conditions.
				// That mean device is not enumerated and we cannot detect it.
				// We need wake it up from DUT side.
				RecoveryActions: []string{
					"Stop servod",
					"Toggle PD once and stop",
					"Toggle PD (5 times) and stop",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
					"Reboot by EC console and stop",
					"Reset EC from DUT and stop",
					"Cold reset the DUT by servod and stop",
					"Reflash Cr50 fw and stop",
				},
			},
			"Servo main device is GSC chip": {
				Docs: []string{
					"Verify that main device is c2d2/cr50/GSC",
				},
				Dependencies: []string{
					"is_servo_v4",
				},
				ExecName: "servo_main_device_is_gcs",
			},
			"Verify cr50 console": {
				Docs: []string{
					"Verify that Cr50 console is responsive.",
				},
				Conditions: []string{
					"Is not servo_v3",
					"Servo main device is GSC chip",
				},
				Dependencies: []string{
					"Initialize DUT part for servo",
				},
				ExecName: "servod_can_read_all",
				ExecExtraArgs: []string{
					"commands:cr50_ccd_level,cr50_testlab,cr50_ccd_state_flags",
					"any_one:true",
				},
				RecoveryActions: []string{
					"Stop servod",
					"Toggle PD once and stop",
					"Toggle PD (5 times) and stop",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reflash Cr50 fw and stop",
					"Reset EC from DUT and stop",
				},
			},
			"Cr50 testlab is enabled": {
				Docs: []string{
					"Verify that testlab flag is enabled in GSC chip.",
					"Expect that cr50/GSC will required to set cr50 testlab is enabled.",
				},
				Conditions: []string{
					"Is not servo_v3",
					"Servo main device is GSC chip",
				},
				ExecName: "servo_check_servod_control",
				ExecExtraArgs: []string{
					"command:cr50_testlab",
					"expected_string_value:on",
				},
				RecoveryActions: []string{
					// TODO: need verify if we can enable testlab.
					"Open gsc testlab",
					"Stop servod",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
				},
			},
			"Open gsc testlab": {
				Docs: []string{
					"If servo uses c2d2/cr50/gsc to control the DUT, open testlab will allowed to work (cr50_reboot, cold_reset, warm_reset)",
				},
				Conditions: []string{
					"Is not servo_v3",
					"Servo main device is GSC chip",
				},
				ExecExtraArgs: []string{
					"command:cr50_testlab",
					"string_value:open",
				},
				ExecName:               "servo_set",
				RunControl:             RunControl_ALWAYS_RUN,
				AllowFailAfterRecovery: true,
			},
			"Initialize DUT part for servo": {
				Docs: []string{
					"Call servod to init dependencies for DUT",
				},
				Conditions: []string{
					"Is not servo_v3",
				},
				Dependencies: []string{
					"Set main servo device",
					"Open gsc testlab",
				},
				ExecName: "init_dut_for_servo",
				RecoveryActions: []string{
					"Stop servod",
					"Toggle PD once and stop",
					"Toggle PD (5 times) and stop",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reflash Cr50 fw and stop",
					"Reset EC from DUT and stop",
				},
			},
			"pwr_button_supported_models": {
				Docs: []string{"power button check is not applicable for these models"},
				ExecExtraArgs: []string{
					"string_values:arkham,gale,mistral,storm,whirlwind",
					"invert_result:true",
				},
				ExecName: "dut_check_model",
			},
			"Verify power button signal": {
				Docs: []string{
					"verify that pwr_button signal is present.",
					"If signal is not present then probably we have issue with servo connection.",
				},
				Conditions: []string{
					"pwr_button_supported_models",
				},
				ExecExtraArgs: []string{
					"command:pwr_button",
					"expected_string_value:release",
				},
				ExecName: "servo_check_servod_control",
				RecoveryActions: []string{
					"Stop servod",
					"Toggle PD once and stop",
					"Toggle PD (5 times) and stop",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reset EC from DUT and stop",
					"Force reflash servo_micro fw and stop",
					"Reflash Cr50 fw and stop",
				},
				AllowFailAfterRecovery: true,
			},
			"Verify servo connected to the DUT": {
				Docs: []string{
					"Verify if servo connected to the DUTand received required voltage from it.",
				},
				Conditions: []string{
					"Is not servo_v3",
					"Is servo_v4(p1) with type-a connector",
					"DUT has CrOS EC",
				},
				ExecName: "servo_low_ppdut5",
				RecoveryActions: []string{
					"Stop servod",
					"Toggle PD (5 times) and stop",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reset EC from DUT and stop",
					"Force reflash servo_micro fw and stop",
				},
			},
			"Servo type-a hub connected": {
				Docs: []string{
					"Verifier to check connection Servo type-a to DUT.",
					"Working only for labstation with servo_micro.",
				},
				Conditions: []string{
					"is_labstation",
					"is_servo_micro",
					"DUT has CrOS EC",
					// Followed is condition to check if voltage is low means servo_micro is not connected.
					"DUT is UP by EC response",
				},
				ExecName: "servo_low_ppdut5",
				RecoveryActions: []string{
					"Stop servod",
					"Try fake disconnect and stop",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reset EC from DUT and stop",
				},
			},
			"DUT is UP by EC response": {
				Docs: []string{
					"Check if DUT is up.",
					"Verification based on EC response.",
				},
				Conditions: []string{
					"DUT has CrOS EC",
				},
				Dependencies: []string{
					"Read servo serial by servod harness",
				},
				ExecName: "servo_check_servod_control",
				ExecExtraArgs: []string{
					"command:ec_system_powerstate",
					"expected_string_value:S0",
				},
			},
			"Verify EC": {
				Conditions: []string{
					"Is not servo_v3",
					"DUT has CrOS EC",
				},
				Dependencies: []string{
					"Set state:EC_BROKEN",
					"Verify EC console",
					"Set state:BAD_RIBBON_CABLE",
					"Verify power button signal",
					"Set state:LID_OPEN_FAILED",
					"Is lid open",
					"servo_battery_charging",
				},
				ExecName: "sample_pass",
			},
			"DUT has CrOS EC": {
				Docs: []string{
					"Verify if DUT has ChromeOS firmware for EC",
				},
				Dependencies: []string{
					"Read servo serial by servod harness",
				},
				ExecExtraArgs: []string{
					"command:supports_cros_ec_communication",
					"expected_string_value:yes",
				},
				ExecName: "servo_check_servod_control",
			},
			"Verify EC console": {
				Conditions: []string{
					"Is not servo_v3",
					"DUT has CrOS EC",
				},
				ExecName: "servod_can_read_all",
				ExecExtraArgs: []string{
					"commands:ec_system_powerstate,ec_board",
					"any_one:true",
				},
				RecoveryActions: []string{
					"Stop servod",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
					"Toggle PD once and stop",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reset EC from DUT and stop",
				},
			},
			"battery_last_charge_readable": {
				ExecExtraArgs: []string{
					"command:battery_full_charge_mah",
				},
				ExecName: "servo_check_servod_control",
			},
			"servo_battery_charging": {
				Conditions: []string{
					"Is not servo_v3",
					"DUT has CrOS EC",
					"battery_last_charge_readable",
				},
				AllowFailAfterRecovery: true,
			},
			"Update USB drive info": {
				Docs: []string{
					"Try to update the information of the servo usbkey in inventory and karte.",
				},
				Conditions: []string{
					"Is not servo_v3",
				},
				Dependencies: []string{
					"Change USB drive direction to servo-host",
				},
				ExecName:               "servo_update_usbkey_history",
				AllowFailAfterRecovery: true,
			},
			"Change USB drive direction to servo-host": {
				Docs: []string{
					"Try to use servod command to point USB drive to servo host.",
				},
				Dependencies: []string{
					"Read servo serial by servod harness",
				},
				ExecName: "servo_check_servod_control",
				ExecExtraArgs: []string{
					"command:image_usbkey_dev",
				},
			},
			"Verify that USB drive is detectable": {
				Docs: []string{
					"Will detect the path to USB Drive on servo-host.",
					"Verify that usb-key is responsive",
				},
				ExecName:               "servo_detect_usbkey",
				ExecTimeout:            &durationpb.Duration{Seconds: 120},
				AllowFailAfterRecovery: true,
			},
			"Audit of USB drive": {
				Docs: []string{
					"This action will detect whether or not the USB drive is in working condition.",
				},
				Dependencies: []string{
					"Verify that USB drive is detectable",
				},
				ExecName:               "servo_audit_usbkey",
				ExecTimeout:            &durationpb.Duration{Seconds: 7300},
				AllowFailAfterRecovery: true,
			},
			"is_servo_v4": {
				Docs: []string{
					"This action will detect whether or not the attached servo device is of type V4.",
				},
				ExecName: "is_servo_v4",
			},
			"Is servo_v4(p1) used with type-c connector": {
				Docs: []string{
					"Verify whether servo_V4(p1) device is connect to DUT using Type-C connection.",
				},
				Conditions: []string{
					"is_servo_v4",
				},
				ExecExtraArgs: []string{
					"command:root.dut_connection_type",
					"expected_string_value:type-c",
				},
				ExecName: "servo_check_servod_control",
			},
			"Is lid open": {
				Docs: []string{
					"Verify lid of the is open",
					"Allowed to fail as check if ont effect the servo functionality.",
				},
				ExecName: "servod_lidopen",
				RecoveryActions: []string{
					"Stop servod",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reflash Cr50 fw and stop",
					"Reset EC from DUT and stop",
				},
				AllowFailAfterRecovery: true,
			},
			"Is not servo_v3": {
				Docs: []string{
					"Verify that servo_v3 isn ot used in setup.",
				},
				Conditions: []string{
					"Is servo_v3 used",
				},
				ExecName: "sample_fail",
			},
			"Is servo_v4(p1) with type-a connector": {
				Docs: []string{
					"Verify whether servo V4(p1) device is connect to DUT using Type-A connection.",
				},
				Conditions: []string{
					"is_servo_v4",
				},
				ExecExtraArgs: []string{
					"command:root.dut_connection_type",
					"expected_string_value:type-a",
				},
				ExecName: "servo_check_servod_control",
			},
			"is_dual_setup": {
				Docs: []string{
					"Check whether the servo device has dual setup. This check only applies to the devices that have the dual setup configured on them.",
				},
				ExecName: "is_dual_setup_configured",
			},
			"is_not_dual_setup": {
				Conditions: []string{
					"is_dual_setup",
				},
				ExecName: "sample_fail",
			},
			"Set main servo device": {
				Docs: []string{
					"Set main device is it not set before.",
					"Applicable if we have more than one child servo device.",
				},
				Conditions: []string{
					"Is not servo_v3",
					"Servod knows about active_dut_controller control",
				},
				ExecName: "servod_set_main_device",
				RecoveryActions: []string{
					"Stop servod",
					"Toggle PD (5 times) and stop",
					"Toggle CC line and stop",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reset EC from DUT and stop",
				},
			},
			"Update all servo's firmware": {
				Docs: []string{
					"Try to update in  normal ways 3 times, if fail allow run force update.",
				},
				Conditions: []string{
					"Is not servo_v3",
				},
				Dependencies: []string{
					"Stop servod daemon on servo-host",
				},
				ExecName: "servo_update_servo_firmware",
				ExecExtraArgs: []string{
					"try_attempt_count:3",
					"try_force_update_after_fail:true",
				},
				ExecTimeout: &durationpb.Duration{Seconds: 600},
				RunControl:  RunControl_RUN_ONCE,
			},
			"Force reflash servo_micro fw and stop": {
				Docs: []string{
					"Try to update servo micro firmware",
				},
				Conditions: []string{
					"is_labstation",
					"is_servo_micro",
					"Is ok to force update servo_micro firmware",
				},
				Dependencies: []string{
					"Force update servo_micro firmware",
					"Stop servod",
				},
				ExecName: "sample_pass",
			},
			"Force update servo_micro firmware": {
				Docs: []string{
					"Try to update servo micro firmware",
				},
				Conditions: []string{
					"is_labstation",
					"is_servo_micro",
					"Is ok to force update servo_micro firmware",
				},
				ExecExtraArgs: []string{
					"force_update:true",
					"ignore_version:true",
					"servo_board:servo_micro",
				},
				ExecTimeout: &durationpb.Duration{Seconds: 180},
				ExecName:    "servo_update_servo_firmware",
			},
			"Is ok to force update servo_micro firmware": {
				Docs: []string{
					"Verify that it is time when we can try to re-flash fw on servo micro.",
					"Re-flashing limited to once per once per 2 weeks to avoid over-flashing the servo device.",
				},
				Conditions: []string{
					"Last servo_micro fw updated within 2 weeks",
				},
				ExecName: "sample_fail",
			},
			"Last servo_micro fw updated within 2 weeks": {
				Docs: []string{
					"Confirm that servo micro fw update action has occurred in the past 2 weeks. (336 hours)",
				},
				ExecExtraArgs: []string{
					"metrics_kind:servo_firmware_update_servo_micro",
					"time_frame_hours:336",
				},
				ExecName: "metrics_found_at_last_time",
			},
			"Warm reset control known by servo": {
				Docs: []string{
					"Verify is servod expected to have warm_reset control",
				},
				ExecName: "servo_check_servod_control",
				ExecExtraArgs: []string{
					"command:warm_reset",
				},
			},
			"Warm reset pin is detected (servo_v3)": {
				// TODO: need monitor before make it critical.
				Docs: []string{
					"Verify that warm_reset pin is detected by servod.",
					"If pin is not present then issue can be related to incorrect connected servo or issue with connector.",
				},
				Conditions: []string{
					"Is servo_v3 used",
					"Warm reset control known by servo",
				},
				ExecName: "servo_check_servod_control",
				ExecExtraArgs: []string{
					"command:warm_reset",
					"expected_string_value:off",
				},
			},
			"Warm reset pin is detected (servo_micro)": {
				// TODO: need monitor before make it critical.
				Docs: []string{
					"Verify that warm_reset pin is detected by servod.",
					"If pin is not present then issue can be related to incorrect connected servo or issue with connector.",
				},
				Conditions: []string{
					"is_servo_micro",
					"Warm reset control known by servo",
				},
				ExecName: "servo_check_servod_control",
				ExecExtraArgs: []string{
					"command:warm_reset",
					"expected_string_value:off",
				},
			},
			"Warm reset pin is detected": {
				Docs: []string{"We need to check for warm reset only for servo micro and V3."},
				Dependencies: []string{
					"Warm reset pin is detected (servo_v3)",
					"Warm reset pin is detected (servo_micro)",
				},
				ExecName:               "sample_pass",
				AllowFailAfterRecovery: true,
			},
			"Cold reset pin is detected": {
				Conditions: []string{
					"Is servo_v3 used",
					"Is servo_v4(p1) with type-a connector",
				},
				ExecName: "servo_check_servod_control",
				ExecExtraArgs: []string{
					"command:cold_reset",
					"expected_string_value:off",
				},
				RecoveryActions: []string{
					"Stop servod",
					"Reboot by EC console and stop",
					"Reset EC from DUT and stop",
				},
			},
			"Servod is responsive to dut-control": {
				Docs: []string{
					"Uses a servod control to check whether the servod daemon is responsive.",
				},
				ExecName:    "servo_servod_echo_host",
				ExecTimeout: &durationpb.Duration{Seconds: 30},
				RecoveryActions: []string{
					"Stop servod",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reset EC from DUT and stop",
					"Create request to reboot labstation",
				},
			},
			"Record good servo type": {
				Docs: []string{
					"Record servo type information.",
				},
				ExecName: "servo_update_servo_type_label",
			},
			"Servod detect all children components": {
				Docs: []string{
					"Check if servod detected all required children components.",
				},
				Conditions: []string{
					"Is not servo_v3",
				},
				ExecName: "servo_check_servod_control",
				ExecExtraArgs: []string{
					"command:dut_controller_missing_fault",
					"expected_string_value:off",
				},
				RecoveryActions: []string{
					"Stop servod",
					"Toggle PD once and stop",
					"Toggle PD (5 times) and stop",
					"Try fake disconnect and stop",
					"Toggle CC line and stop",
					"Reboot by EC console and stop",
					"Cold reset the DUT by servod and stop",
					"Reset EC from DUT and stop",
				},
			},
			"Servod knows about active_dut_controller control": {
				ExecName: "servo_check_servod_control",
				ExecExtraArgs: []string{
					"command:active_dut_controller",
				},
			},
			"servod_restart_dut": {
				ExecName: "sample_pass",
			},
			"Toggle PD once and stop": {
				Docs: []string{
					"Toggle the servod command servo_pd_role only once. And then stop the servod afterwards.",
					"TODO: Add dependency for servo initialize.",
				},
				Dependencies: []string{
					"Toggle PD once",
					"Stop servod",
				},
				ExecName:   "sample_pass",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Toggle PD once": {
				Docs: []string{
					"Toggle the servod command servo_pd_role only once.",
				},
				ExecExtraArgs: []string{
					"toggle_times:1",
					"wait_in_retry:5",
					"wait_before_retry:1",
				},
				RunControl: RunControl_ALWAYS_RUN,
				ExecName:   "servo_servod_toggle_pd_role",
			},
			"Toggle PD (5 times) and stop": {
				Docs: []string{
					"Toggle the servod command servo_pd_role 5 times. And then stop the servod afterwards. TODO: Add dependency for servo initialize.",
				},
				Dependencies: []string{
					"Toggle PD 5 times",
				},
				ExecName:   "servo_host_servod_stop",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Toggle PD 5 times": {
				Docs: []string{
					"Toggle the servod command servo_pd_role 5 times.",
				},
				ExecName: "servo_servod_toggle_pd_role",
				ExecExtraArgs: []string{
					"toggle_times:5",
					"wait_in_retry:5",
					"wait_before_retry:1",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:MISSING_CONFIG": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:MISSING_CONFIG",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:WRONG_CONFIG": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:WRONG_CONFIG",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:NO_SSH": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:NO_SSH",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:BROKEN": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:BROKEN",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:NOT_CONNECTED": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:NOT_CONNECTED",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:NEED_REPLACEMENT": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:NEED_REPLACEMENT",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:TOPOLOGY_ISSUE": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:TOPOLOGY_ISSUE",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:SERVO_UPDATER_ISSUE": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:SERVO_UPDATER_ISSUE",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:SERVOD_ISSUE": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:SERVOD_ISSUE",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:SERVO_HOST_ISSUE": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:SERVO_HOST_ISSUE",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:SERVOD_PROXY_ISSUE": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:SERVOD_PROXY_ISSUE",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:COLD_RESET_PIN_ISSUE": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:COLD_RESET_PIN_ISSUE",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:WARM_RESET_PIN_ISSUE": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:WARM_RESET_PIN_ISSUE",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:SBU_LOW_VOLTAGE": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:SBU_LOW_VOLTAGE",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:DUT_NOT_CONNECTED": {
				ExecExtraArgs: []string{"state:DUT_NOT_CONNECTED"},
				ExecName:      "servo_set_servo_state",
				RunControl:    RunControl_ALWAYS_RUN,
			},
			"Set state:CR50_CONSOLE_MISSING": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:CR50_CONSOLE_MISSING",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:EC_BROKEN": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:EC_BROKEN",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:BAD_RIBBON_CABLE": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:BAD_RIBBON_CABLE",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:LID_OPEN_FAILED": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:LID_OPEN_FAILED",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:CCD_TESTLAB_ISSUE": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:CCD_TESTLAB_ISSUE",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Set state:WORKING": {
				ExecName: "servo_set_servo_state",
				ExecExtraArgs: []string{
					"state:WORKING",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Try fake disconnect and stop": {
				Docs: []string{
					"Try to repair servod by mimic reconnection of servo.",
				},
				Dependencies: []string{
					"Try fake disconnect",
					"Stop servod",
				},
				ExecName:   "sample_pass",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Try fake disconnect": {
				Conditions: []string{
					"is_servo_type_ccd",
				},
				ExecName: "servo_fake_disconnect_dut",
				ExecExtraArgs: []string{
					"delay_in_ms:100",
					"timeout_in_ms:2000",
				},
			},
			"Toggle CC line and stop": {
				Docs: []string{
					"Try to repair servod by toggling cc and stop servod after.",
				},
				Dependencies: []string{
					"Toggle CC lines",
					"Stop servod",
				},
				ExecName:   "sample_pass",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Toggle CC lines": {
				Docs: []string{
					"Toggle cc line connected between servo and DUT to wake up the connection.",
				},
				Conditions: []string{
					"is_servo_type_ccd",
				},
				ExecName: "servo_servod_cc_toggle",
				ExecExtraArgs: []string{
					"cc_off_timeout:10",
					"cc_on_timeout:30",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Reboot by EC console and stop": {
				Docs: []string{
					"Try to reboot DUT's EC by servod UART console and stop servod after that.",
				},
				Dependencies: []string{
					"Reboot by DUT's EC UART",
					"Stop servod",
				},
				ExecName:   "sample_pass",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Reboot by DUT's EC UART": {
				Docs: []string{
					"Try to reboot EC on DUT using servod command.",
					"It reboots just the embedded controllers on the DUT.",
				},
				ExecExtraArgs: []string{
					"wait_timeout:1",
					"value:reboot",
				},
				ExecName:   "servo_set_ec_uart_cmd",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Cold reset the DUT by servod and stop": {
				Docs: []string{
					"Try to reboot DUT by resetting power state command on servod.",
				},
				Dependencies: []string{
					"Cold reset by servod",
					"Stop servod",
				},
				ExecName:   "sample_pass",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Cold reset by servod": {
				Docs: []string{
					"Try to reboot DUT by resetting power state command on servod.",
				},
				ExecName: "servo_power_state_reset",
				ExecExtraArgs: []string{
					"wait_timeout:1",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Create request to reboot labstation": {
				Docs: []string{
					"Try to create reboot flag file request.",
				},
				Conditions: []string{
					"Device is SSHable",
					"is_labstation",
				},
				ExecName:   "cros_create_reboot_request",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Reflash Cr50 fw and stop": {
				Docs: []string{
					"Try to reflash cr50 firmware and reboot AP from DUT side to wake it up.",
				},
				Conditions: []string{
					"is_servo_type_ccd",
					"Is reflash Cr50 was done more 24 hours ago",
				},
				Dependencies: []string{
					"Reflash Cr50 fw on DUT",
					"Stop servod",
				},
				ExecName:   "sample_pass",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Reflash Cr50 fw on DUT": {
				Docs: []string{
					"Try to reflash cr50 firmware and reboot AP from DUT side to wake it up.",
					"Reboot after the fw flash is successful.",
				},
				Dependencies: []string{
					"DUT is SSHable",
				},
				ExecName: "cros_reflash_cr50_fw",
				ExecExtraArgs: []string{
					"flash_timeout:120",
					"wait_timeout:30",
				},
				ExecTimeout: &durationpb.Duration{Seconds: 150},
				RunControl:  RunControl_RUN_ONCE,
			},
			"DUT is SSHable": {
				Docs: []string{
					"verify if DUT is SSH-able",
				},
				ExecName: "cros_ssh_dut",
			},
			"Is reflash Cr50 was done more 24 hours ago": {
				Docs: []string{
					"Verify that it is time when we can try to re-flash fw on cr50 (H1).",
					"Re-flashing limited to once per once per day to avoid over-flashing the device.",
				},
				Conditions: []string{
					"Is reflash Cr50 was done within 24 hours",
				},
				ExecName: "sample_fail",
			},
			"Is reflash Cr50 was done within 24 hours": {
				Docs: []string{
					"Confirm that no cr50 reflash action has occurred in the past 24 hours.",
				},
				ExecExtraArgs: []string{
					"metrics_kind:cr50_flash",
					"time_frame_hours:24",
				},
				ExecName: "metrics_found_at_last_time",
			},
			"Reset EC from DUT and stop": {
				Docs: []string{
					"Try to reset EC from DUT side to wake CR50 up and then stop the servod.",
				},
				Conditions: []string{
					"is_servo_type_ccd",
				},
				Dependencies: []string{
					"cros_reset_ec",
					"Stop servod",
				},
				ExecName:   "sample_pass",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"cros_reset_ec": {
				Docs: []string{
					"Try to reset EC from DUT side by running connads wake up the device as it will trigger recovering ec, cr50, and other fw.",
				},
				Dependencies: []string{
					"DUT is SSHable",
				},
				ExecName: "cros_reset_ec",
				ExecExtraArgs: []string{
					"wait_timeout:30",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"SmartHub is not present in setup": {
				Docs: []string{
					"Specify if smart usbhub is present in setup.",
				},
				ExecName: "servo_is_smarthub_expected",
				ExecExtraArgs: []string{
					"reverse:true",
				},
			},
			"Power-cycle servo-v4p1 network": {
				Docs: []string{
					"Try to reset network controller of servo_v4p1 when smart usbhub not present.",
				},
				Conditions: []string{
					// TODO(gregorynisbet): Can we make the servo v4p1 condition less strict?
					"Is servo_v4(p1) used with type-c connector",
					"Is not servo_v3",
					"SmartHub is not present in setup",
				},
				ExecName:               "servo_v4p1_network_reset",
				RunControl:             RunControl_RUN_ONCE,
				AllowFailAfterRecovery: true,
			},
			"Power-cycle by smart-hub": {
				Docs: []string{
					"Try to reset(power-cycle) the servo via smart usbhub.",
				},
				Conditions: []string{
					"Is not servo_v3",
				},
				ExecName: "servo_power_cycle_root_servo",
				ExecExtraArgs: []string{
					"reset_timeout:60",
					"wait_timeout:20",
				},
				ExecTimeout:            &durationpb.Duration{Seconds: 120},
				RunControl:             RunControl_RUN_ONCE,
				AllowFailAfterRecovery: true,
			},
			"Reboot servo_v3": {
				Docs: []string{
					"Try to reboot servo host v3.",
				},
				Conditions: []string{
					"Is servo_v3 used",
				},
				ExecName: "servo_host_v3_reboot",
				ExecExtraArgs: []string{
					"reboot_timeout:10",
				},
				ExecTimeout: &durationpb.Duration{Seconds: 300},
				RunControl:  RunControl_RUN_ONCE,
			},
			"Sleep 1s": {
				ExecName: "sample_sleep",
				ExecExtraArgs: []string{
					"sleep:1",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
		},
	}
}
