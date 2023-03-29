// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"google.golang.org/protobuf/types/known/durationpb"
)

// setAllowFail updates allowFail property and return plan.
func setAllowFail(p *Plan, allowFail bool) *Plan {
	p.AllowFail = allowFail
	return p
}

// CrosRepairConfig provides config for repair cros setup in the lab task.
func CrosRepairConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanServo,
			PlanCrOS,
			PlanChameleon,
			PlanBluetoothPeer,
			PlanWifiRouter,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanServo:         setAllowFail(servoRepairPlan(), true),
			PlanCrOS:          setAllowFail(crosRepairPlan(), false),
			PlanChameleon:     setAllowFail(chameleonPlan(), true),
			PlanBluetoothPeer: setAllowFail(btpeerRepairPlan(), true),
			PlanWifiRouter:    setAllowFail(wifiRouterRepairPlan(), true),
			PlanClosing:       setAllowFail(crosClosePlan(), true),
		}}
}

// CrosRepairWithDeepRepairConfig provides config for combination of deep repair + normal repair.
func CrosRepairWithDeepRepairConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanServoDeepRepair,
			PlanCrOSDeepRepair,
			PlanServo,
			PlanCrOS,
			PlanChameleon,
			PlanBluetoothPeer,
			PlanWifiRouter,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanServoDeepRepair: setAllowFail(deepRepairServoPlan(), true),
			// We allow CrOSDeepRepair to fail(so the task continue) as some of actions in it may result to a later normal repair success.
			PlanCrOSDeepRepair: setAllowFail(deepRepairCrosPlan(), true),
			PlanServo:          setAllowFail(servoRepairPlan(), true),
			PlanCrOS:           setAllowFail(crosRepairPlan(), false),
			PlanChameleon:      setAllowFail(chameleonPlan(), true),
			PlanBluetoothPeer:  setAllowFail(btpeerRepairPlan(), true),
			PlanWifiRouter:     setAllowFail(wifiRouterRepairPlan(), true),
			PlanClosing:        setAllowFail(crosClosePlan(), true),
		}}
}

// CrosDeployConfig provides config for deploy cros setup in the lab task.
func CrosDeployConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanServo,
			PlanCrOS,
			PlanChameleon,
			PlanBluetoothPeer,
			PlanWifiRouter,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanServo:         setAllowFail(servoRepairPlan(), false),
			PlanCrOS:          setAllowFail(crosDeployPlan(), false),
			PlanChameleon:     setAllowFail(chameleonPlan(), true),
			PlanBluetoothPeer: setAllowFail(btpeerRepairPlan(), true),
			PlanWifiRouter:    setAllowFail(wifiRouterRepairPlan(), true),
			PlanClosing:       setAllowFail(crosClosePlan(), true),
		},
	}
}

// crosClosePlan provides plan to close cros repair/deploy tasks.
func crosClosePlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Update peripheral wifi state",
			"Update chameleon state for chameleonless dut",
			"Update DUT state for failures more than threshold",
			"Update DUT state based on servo state",
			"Update cellular modem state for non-cellular pools",
			"Close Servo-host",
		},
		Actions: map[string]*Action{
			"Is servo_state:working": {
				Docs: []string{
					"check the servo's state is ServoStateWorking.",
				},
				ExecName:      "servo_match_state",
				ExecExtraArgs: []string{"state:WORKING"},
				MetricsConfig: &MetricsConfig{
					UploadPolicy: MetricsConfig_SKIP_ALL,
				},
			},
			"Servo-host known": {
				ExecName: "dut_servo_host_present",
				MetricsConfig: &MetricsConfig{
					UploadPolicy: MetricsConfig_SKIP_ALL,
				},
			},
			"Remove request to reboot if servo is good": {
				Conditions: []string{
					"Is not Flex device",
					"Servo-host known",
					"Is servo_state:working",
				},
				ExecName:               "cros_remove_reboot_request",
				AllowFailAfterRecovery: true,
			},
			"Close Servo-host": {
				Conditions: []string{
					"Servo-host known",
					"Is not Flex device",
					"Servo-host is sshable",
				},
				Dependencies: []string{
					"Try copy messages from servo-host",
					"Try to collect servod logs",
					"Remove in-use flag on servo-host",
					"Remove request to reboot if servo is good",
					"Turn off servo usbkey power",
					"Stop servod",
				},
				ExecName: "sample_pass",
				MetricsConfig: &MetricsConfig{
					UploadPolicy: MetricsConfig_SKIP_ALL,
				},
				AllowFailAfterRecovery: true,
			},
			"Remove in-use flag on servo-host": {
				Conditions: []string{
					"Servo-host known",
				},
				ExecName:               "cros_remove_servo_in_use",
				AllowFailAfterRecovery: true,
			},
			"Is not Flex device": {
				Docs: []string{"Verify that device is belong Reven models"},
				ExecExtraArgs: []string{
					"string_values:x1c",
					"invert_result:true",
				},
				ExecName: "dut_check_model",
				MetricsConfig: &MetricsConfig{
					UploadPolicy: MetricsConfig_SKIP_ALL,
				},
			},
			"Try to collect servod logs": {
				Docs: []string{
					"Try to collect all servod logs since latest start time.",
				},
				Conditions: []string{
					"Servo-host known",
				},
				ExecName:               "cros_collect_servod_logs",
				AllowFailAfterRecovery: true,
			},
			"Try copy messages from servo-host": {
				Docs: []string{
					"Try to collect /var/log/messages from servo-host.",
				},
				Conditions: []string{
					"Servo-host known",
				},
				ExecName: "cros_copy_to_logs",
				ExecExtraArgs: []string{
					"src_host_type:servo_host",
					"src_path:/var/log/messages",
					"src_type:file",
					"use_host_dir:true",
				},
				AllowFailAfterRecovery: true,
			},
			"Is not in cellular pool": {
				Docs: []string{
					"Verify that DUT is not in a cellular pool.",
				},
				ExecName: "dut_not_in_pool_regex",
				ExecExtraArgs: []string{
					"regex:(?i)^cellular",
				},
				MetricsConfig: &MetricsConfig{
					UploadPolicy: MetricsConfig_SKIP_ALL,
				},
			},
			"Update cellular modem state for non-cellular pools": {
				Docs: []string{
					"Set cellular modem state for DUTs in non-cellular pools.",
				},
				Conditions: []string{
					"Is not in cellular pool",
					"has_cellular_info",
				},
				ExecName: "set_cellular_modem_state",
				ExecExtraArgs: []string{
					"state:hardware_not_detected",
				},
				AllowFailAfterRecovery: true,
			},
			"Update peripheral wifi state": {
				Docs: []string{
					"Update peripheral wifi state based on wifi router states",
				},
				Conditions: []string{
					"wifi_router_host_present",
				},
				ExecName:               "update_peripheral_wifi_state",
				AllowFailAfterRecovery: true,
			},
			"Update chameleon state for chameleonless dut": {
				Docs: []string{
					"Update chameleon state to not applicable for chameleonless dut",
				},
				Conditions: []string{
					"chameleon_not_present",
				},
				ExecName:               "chameleon_state_not_applicable",
				AllowFailAfterRecovery: true,
			},
			"Failure count above threshold": {
				Docs: []string{
					"Check if the number of times the recovery task ",
					"has failed is greater than a threshold value or ",
					"not.",
				},
				ExecName: "metrics_check_task_failures",
				ExecExtraArgs: []string{
					"task_name:recovery",
					"repair_failed_count:6",
				},
			},
			"Update DUT state for failures more than threshold": {
				Docs: []string{
					"Set the DUT state to the value passed in the ",
					"extra args.",
				},
				Conditions: []string{
					"DUT state is repair_failed",
					"Failure count above threshold",
					// Apply conditions in separate CL.
					// "DUT state is not ready",
					// "DUT state is not needs_replacement",
				},
				ExecName: "dut_set_state",
				ExecExtraArgs: []string{
					"state:needs_manual_repair",
				},
				AllowFailAfterRecovery: true,
			},
			"Failure count are not above threshold": {
				Docs: []string{
					"Check if the number of times the recovery task ",
					"has failed is not greater than a threshold value",
				},
				Dependencies: []string{
					"Failure count above threshold",
				},
				ExecName: "sample_fail",
			},
			"DUT state is not ready": {
				Docs: []string{
					`Check if the DUT state is anything other than "ready".`,
				},
				ExecName: "dut_state_match",
				ExecExtraArgs: []string{
					"invert:true",
					"state:ready",
				},
			},
			"DUT state is not needs_replacement": {
				Docs: []string{
					`Check if the DUT state is anything other than "needs_replacement".`,
				},
				ExecName: "dut_state_match",
				ExecExtraArgs: []string{
					"invert:true",
					"state:needs_replacement",
				},
			},
			"DUT state is repair_failed": {
				Docs: []string{
					"Check if the DUT's state is in repair_failed state, if not then fail.",
				},
				ExecName: "dut_state_match",
				ExecExtraArgs: []string{
					"state:repair_failed",
				},
			},
			"Servo state demands manual repair": {
				Docs: []string{
					"Check whether the state of servo mandates a ",
					"manual repair on the DUT.",
				},
				ExecName: "dut_servo_state_required_manual_attention",
				ExecExtraArgs: []string{
					"servo_states:NEED_REPLACEMENT,DUT_NOT_CONNECTED",
				},
			},
			"Update DUT state based on servo state": {
				Docs: []string{
					"Set the DUT state based on the state of servo. ",
					"This applies only if the count of failures is ",
					"not above a threshold amount",
				},
				Conditions: []string{
					"DUT state is repair_failed",
					"Failure count are not above threshold",
					"Servo-host known",
					"Servo state demands manual repair",
				},
				ExecName: "dut_set_state",
				ExecExtraArgs: []string{
					"state:needs_manual_repair",
				},
				AllowFailAfterRecovery: true,
			},
			"Turn off servo usbkey power": {
				Docs: []string{
					"Ensure that servo usbkey power is in off state.",
				},
				Conditions: []string{
					"Servo-host known",
				},
				ExecName: "servo_set",
				ExecExtraArgs: []string{
					"command:image_usbkey_pwr",
					"string_value:off",
				},
				RunControl:             RunControl_ALWAYS_RUN,
				AllowFailAfterRecovery: true,
			},
			"Stop servod": {
				Docs: []string{
					"Stop the servod daemon.",
					"Allowed to fail as can be run when servod is not running.",
				},
				Dependencies: []string{
					"Save UART capture",
				},
				ExecName:               "servo_host_servod_stop",
				RunControl:             RunControl_ALWAYS_RUN,
				AllowFailAfterRecovery: true,
			},
			"Servo-host is sshable": {
				Docs: []string{
					"Stop the servod daemon.",
					"Allowed to fail as can be run when servod is not running.",
				},
				ExecName: "cros_ssh",
				ExecExtraArgs: []string{
					"device_type:servo",
				},
				ExecTimeout: &durationpb.Duration{
					Seconds: 15,
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Save UART capture": {
				Dependencies: []string{
					"Stop UART capture",
				},
				ExecName:               "servod_save_uart_capture",
				AllowFailAfterRecovery: true,
				MetricsConfig:          &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
			},
			"Stop UART capture": {
				ExecName:               "servod_stop_uart_capture",
				AllowFailAfterRecovery: true,
				MetricsConfig:          &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
			},
		},
	}
}
