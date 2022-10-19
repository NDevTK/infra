// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

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
			"Servo-host logs",
			"Remove in-use flag on servo-host",
			"Remove request to reboot if servo is good",
			"Update DUT state for failures more than threshold",
			"Update DUT state based on servo state",
			"Turn off servo usbkey power",
			"Stop servod",
		},
		Actions: map[string]*Action{
			"Is servo_state:working": {
				Docs: []string{
					"check the servo's state is ServoStateWorking.",
				},
				ExecName:      "servo_match_state",
				ExecExtraArgs: []string{"state:WORKING"},
			},
			"Remove request to reboot if servo is good": {
				Conditions: []string{
					"Is not Flex device",
					"dut_servo_host_present",
					"Is servo_state:working",
				},
				ExecName:               "cros_remove_reboot_request",
				AllowFailAfterRecovery: true,
			},
			"Remove in-use flag on servo-host": {
				Conditions: []string{
					"Is not Flex device",
					"dut_servo_host_present",
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
			"Servo-host logs": {
				Dependencies: []string{
					"Try copy messages from servo-host",
					"Try to collect servod logs",
				},
				ExecName: "sample_pass",
			},
			"Try to collect servod logs": {
				Docs: []string{
					"Try to collect all servod logs since latest start time.",
				},
				Conditions: []string{
					"dut_servo_host_present",
					"Is not servo_v3",
				},
				ExecName:               "cros_collect_servod_logs",
				AllowFailAfterRecovery: true,
			},
			"Try copy messages from servo-host": {
				Docs: []string{
					"Try to collect /var/log/messages from servo-host.",
				},
				Conditions: []string{
					"dut_servo_host_present",
					"Is not servo_v3",
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
			"Is not servo_v3": {
				Docs: []string{
					"Verify that servo_v3 isn ot used in setup.",
				},
				Conditions: []string{
					"is_servo_v3",
				},
				ExecName: "sample_fail",
				MetricsConfig: &MetricsConfig{
					UploadPolicy: MetricsConfig_SKIP_ALL,
				},
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
			"Failure count above threshold": {
				Docs: []string{
					"Check if the number of times the recovery task ",
					"has failed is greater than a threshold value or ",
					"not.",
				},
				ExecName: "metrics_check_task_failures",
				ExecExtraArgs: []string{
					"task_name:recovery",
					"repair_failed_count:49",
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
					"dut_servo_host_present",
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
					"Is not servo_v3",
					"dut_servo_host_present",
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
				ExecName:               "servo_host_servod_stop",
				RunControl:             RunControl_ALWAYS_RUN,
				AllowFailAfterRecovery: true,
			},
		},
	}
}
