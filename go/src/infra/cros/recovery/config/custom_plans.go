// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"fmt"

	"google.golang.org/protobuf/types/known/durationpb"
)

// DownloadImageToServoUSBDrive creates configuration to download image to USB-drive connected to the servo.
func DownloadImageToServoUSBDrive(gsImagePath, imageName string) *Configuration {
	rc := CrosRepairConfig()
	rc.PlanNames = []string{
		PlanServo,
		PlanCrOS,
	}
	// Servo plan is not critical we just care to start servod.
	rc.Plans[PlanServo].AllowFail = true
	// Remove closing plan as we do not collect any logs or update states kin spacial ways.
	delete(rc.Plans, PlanClosing)
	cp := rc.Plans[PlanCrOS]
	const targetAction = "Download stable image to USB-key"
	cp.CriticalActions = []string{targetAction}
	var newArgs []string
	if gsImagePath != "" {
		newArgs = append(newArgs, fmt.Sprintf("os_image_path:%s", gsImagePath))
	} else if imageName != "" {
		newArgs = append(newArgs, fmt.Sprintf("os_name:%s", imageName))
	}
	cp.GetActions()[targetAction].ExecExtraArgs = newArgs
	return rc
}

// ReserveDutConfig creates configuration to reserve a dut
func ReserveDutConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanCrOS,
		},
		Plans: map[string]*Plan{
			PlanCrOS: {
				CriticalActions: []string{
					"dut_state_reserved",
				},
			},
		},
	}
}

// DeepRepairConfig creates configuration to perform deep repair.
//
// Please look to the configuration to see all steps.
// Configuration is not critical and do not update the state of the DUT.
// Please do not apply close plan from repair to avoid unexpected state changes.
func DeepRepairConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanServo,
			PlanCrOS,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanCrOS: {
				CriticalActions: []string{
					"Flash EC (FW) by servo",
					"Flash AP (FW) and set GBB to 0x18 from fw-image by servo (without reboot)",
					"Download stable image to USB-key",
					"Install OS in DEV mode by USB-drive",
				},
				Actions: crosRepairActions(),
			},
			PlanServo: {
				CriticalActions: []string{
					"Servo is know in the setup",
					"Servod port specified",
					"Servo serial is specified",
					"Device is pingable",
					"Device is SSHable",
					"servo_power_cycle_root_servo",
					"Mark labstation as servod is in-use",
					"Start servod daemon without recovery",
					"Servod is responsive to dut-control",
				},
				Actions: map[string]*Action{
					"Servo is know in the setup": {
						Docs: []string{
							"Verify if setup data has any data related to servo-host which mean servo is present in setup.",
						},
						ExecName: "dut_servo_host_present",
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
					"servo_power_cycle_root_servo": {
						Docs: []string{
							"Try to reset(power-cycle) the servo via smart usbhub.",
						},
						Conditions: []string{
							"Is not servo_v3",
						},
						ExecExtraArgs: []string{
							"reset_timeout:60",
							"wait_timeout:20",
						},
						ExecTimeout:            &durationpb.Duration{Seconds: 120},
						RunControl:             RunControl_RUN_ONCE,
						AllowFailAfterRecovery: true,
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
					"Start servod daemon without recovery": {
						Docs: []string{
							"Start servod daemon on servo-host without recovery mode.",
							"Start servod in recovery mode can cause return value change in control like ec_chip.",
							"Which can cause recover action like flash_ec fail.",
						},
						ExecName: "servo_host_servod_init",
						ExecExtraArgs: []string{
							"recovery_mode:false",
						},
						ExecTimeout: &durationpb.Duration{Seconds: 120},
						RecoveryActions: []string{
							"Stop servod",
							"Create request to reboot labstation",
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
						},
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
					"Stop servod": {
						Docs: []string{
							"Stop the servod daemon.",
							"Allowed to fail as can be run when servod is not running.",
						},
						ExecName:               "servo_host_servod_stop",
						RunControl:             RunControl_ALWAYS_RUN,
						AllowFailAfterRecovery: true,
					},
					"Is servo_v3 used": {
						Docs: []string{
							"Condition to check if the servo is v3.",
						},
						ExecName: "is_servo_v3",
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
					"is_labstation": {
						Docs: []string{
							"Condition to check if the servohost is a labstation.",
						},
						ExecName: "servo_host_is_labstation",
					},
					"Sleep 1s": {
						ExecName: "sample_sleep",
						ExecExtraArgs: []string{
							"sleep:1",
						},
						RunControl: RunControl_ALWAYS_RUN,
					},
				},
			},
			PlanClosing: {
				CriticalActions: []string{
					"Remove in-use flag on servo-host",
					"Stop servod",
				},
				Actions: map[string]*Action{
					"Remove in-use flag on servo-host": {
						Conditions: []string{
							"Is not Flex device",
							"dut_servo_host_present",
						},
						ExecName:               "cros_remove_servo_in_use",
						AllowFailAfterRecovery: true,
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
					"Is not Flex device": {
						Docs: []string{"Verify that device is belong Reven models"},
						ExecExtraArgs: []string{
							"string_values:x1c",
							"invert_result:true",
						},
						ExecName: "dut_check_model",
					},
				},
			},
		},
	}
}
