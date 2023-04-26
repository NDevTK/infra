// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"google.golang.org/protobuf/types/known/durationpb"
)

// DeepRepairConfig creates configuration to perform deep repair.
// Configuration is not critical and do not update the state of the DUT.
func DeepRepairConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanServo,
			PlanCrOS,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanServo:   deepRepairServoPlan(),
			PlanCrOS:    deepRepairCrosPlan(),
			PlanClosing: setAllowFail(crosClosePlan(), true),
		},
	}
}

func deepRepairCrosPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Flash EC (FW) by servo",
			"Sleep 60 seconds",
			"Flash AP (FW) and set GBB to 0x18 from fw-image by servo (without reboot)",
			"Download stable version OS image to servo usbkey if necessary (allow fail)",
			"Install OS in DEV mode by USB-drive",
		},
		Actions: crosRepairActions(),
	}
}

// deepRepairServoPlan returns a plan contains prepare with minimum requirements which is prerequisite to run deep repair cros plan.
func deepRepairServoPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Servo is know in the setup",
			"Servod port specified",
			"Servo serial is specified",
			"Device is SSHable",
			"Power-cycle by smart-hub",
			"Mark labstation as servod is in-use",
			"Start servod daemon without recovery",
			"Servod is responsive to dut-control",
		},
		Actions: map[string]*Action{
			"Servo is know in the setup": {
				Docs: []string{
					"Verify if setup data has any data related to servo-host which mean servo is present in setup.",
				},
				ExecName:   "dut_servo_host_present",
				RunControl: RunControl_RUN_ONCE,
			},
			"Servod port specified": {
				Docs: []string{
					"Verify that servod port is present in servo data.",
				},
				ExecName:   "servo_servod_port_present",
				RunControl: RunControl_RUN_ONCE,
			},
			"Servo serial is specified": {
				Docs: []string{
					"Check if root servo serial is present.",
				},
				ExecName:   "dut_servo_has_serial",
				RunControl: RunControl_RUN_ONCE,
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
			"Power-cycle by smart-hub": {
				Docs: []string{
					"Try to reset(power-cycle) the servo via smart usbhub.",
				},
				Conditions: []string{
					// Disable power-cycle by smart hub for v4p1 due to b/243042046,
					// The built-in reboot and ethernet power control in v4p1 also
					// makes power-cycle the entire device unnecessary.
					"Serial number is not servo_v4p1",
					// We try restart only if we lost network to the dut.
					"DUT is not SSHable",
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
			"DUT is SSHable": {
				Docs: []string{
					"verify if DUT is SSH-able",
				},
				ExecName:    "cros_ssh_dut",
				ExecTimeout: &durationpb.Duration{Seconds: 15},
				RunControl:  RunControl_ALWAYS_RUN,
			},
			"DUT is not SSHable": {
				Docs: []string{
					"Verify if DUT is not SSH-able",
				},
				Conditions: []string{
					"DUT is SSHable",
				},
				ExecName:   "sample_fail",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Serial number is not servo_v4p1": {
				Docs: []string{
					"Verify that the servo serial number is not a servo_v4p1 serial number",
				},
				Conditions: []string{
					"is_servo_v4p1_by_serial_number",
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
				RunControl:             RunControl_ALWAYS_RUN,
				AllowFailAfterRecovery: true,
				MetricsConfig:          &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
			},
		},
	}
}
