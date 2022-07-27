// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import "google.golang.org/protobuf/types/known/durationpb"

// CrosAuditUSBConfig audits the USB storage for a servo associated with a ChromeOS DUT.
func CrosAuditUSBConfig() *Configuration {
	// This plan is taken from the custom plan for deep repair.
	// We first check to make sure that the servo is healthy and servod can start
	// before we manipulate the DUT.
	servoPlan := &Plan{
		CriticalActions: []string{
			"Servo is known in the setup",
			"Servod port specified",
			"Servo serial is specified",
			"Device is pingable",
			"Device is SSHable",
			"Power-cycle by smart-hub",
			"Mark labstation as servod is in-use",
			"Start servod daemon without recovery",
			"Servod is responsive to dut-control",
		},
		Actions: map[string]*Action{
			"Servo is known in the setup": {
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
		AllowFail: false,
	}

	// We need the CrOS repair actions to recover from cases where the device is unexpectedly down.
	crosPlanActions := crosRepairActions()
	// The AuditUSB action is the main point of this plan.
	// It comes last sequentially in the plan below so we do not bother to add conditions like
	// "working servo" or similar to its conditions.
	crosPlanActions["Audit USB"] = &Action{
		Docs: []string{
			"Audit the USB drive",
		},
		Conditions:      nil,
		Dependencies:    nil,
		ExecName:        "audit_usb_from_dut_side",
		RecoveryActions: nil,
		ExecTimeout:     &durationpb.Duration{Seconds: 2 * 60 * 60},
	}

	// The crosPlan for a USB audit checks to make sure that the device is SSHable, then
	// it sets the USB direction, SSH's into a DUT, and initiates a check.
	crosPlan := &Plan{
		CriticalActions: []string{
			// We defensively set the state to needs repair before every task so that we force
			// a repair once the audit task is complete.
			"Set state: needs_repair",
			// Check that we can SSH to the DUT in question.
			"Device is SSHable",
			// Attempt to audit the USB from the DUT side.
			"Audit USB",
		},
		Actions:   crosPlanActions,
		AllowFail: false,
	}

	return &Configuration{
		PlanNames: []string{
			// First we run the servo plan, make sure that the servo is in a good state
			// and servod is up.
			PlanServo,
			// Then we run the CrOS plan, the main thing that this contains is the ability to
			// run "audit_usb_from_dut_side".
			PlanCrOS,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanServo:   servoPlan,
			PlanCrOS:    crosPlan,
			PlanClosing: setAllowFail(crosClosePlan(), true),
		},
	}
}
