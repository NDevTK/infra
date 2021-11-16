// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package recovery

import (
	"io"
	"strings"
)

// Default cobfiguration with all planes supported by engine.
// WHen you change or add new plan please be sure that is working
// and predictable.
// TODO(otabek@): Add plan for labstation.
// TODO(vkjoshi@): Add plans for Servo and DUT.
const defaultConfig = `
{
	"plans":{
		"labstation_repair":{
			"critical_actions": [
				"dut_state_repair_failed",
				"check_host_info",
				"cros_ping",
				"cros_ssh",
				"cros_is_on_stable_version",
				"update_provision_info",
				"booted_from_right_kernel",
				"reboot_by_request",
				"dut_state_ready"
			],
			"actions": {
				"dut_state_repair_failed":{
					"run_control": 2
				},
				"check_host_info":{
					"docs":[
						"Check basic info for deployment."
					],
					"exec_name": "sample_pass",
					"dependencies":[
						"has_dut_name",
						"has_dut_board_name",
						"has_dut_model_name"
					]
				},
				"cros_is_on_stable_version":{
					"conditions": [
						"has_stable_version_cros_image",
						"cros_kernel_priority_has_not_changed",
						"not_exempted_pool"
					],
					"recovery_actions":[
						"install_stable_os"
					]
				},
				"install_stable_os":{
					"docs":[
						"Install stable OS on the device."
					],
					"conditions": [
						"has_stable_version_cros_image",
						"cros_kernel_priority_has_not_changed"
					],
					"exec_name": "cros_provision",
					"exec_extra_args":[
						"no_reboot"
					],
					"exec_timeout": {
						"seconds": 3600
					}
				},
				"not_exempted_pool":{
					"docs":[
						"There are some labstations we don't want they receive auto-update, e.g. labstations that used for image qualification purpose"
					],
					"exec_name":"dut_not_in_pool",
					"exec_extra_args":[
						"servo_verification",
						"labstation_tryjob",
						"labstation_canary"
					]
				},
				"update_provision_info":{
					"exec_name": "cros_update_provision_os_version"
				},
				"labstation_langid_check":{
					"docs":[
						"This part is not ready.",
						"The action and will validate present of lang_id issue"
					],
					"exec_name": "sample_pass",
					"allow_fail_after_recovery": true
				},
				"cros_stop_powerd":{
					"exec_name":"cros_run_shell_command",
					"exec_extra_args":[
						"stop",
						"powerd"
					],
					"allow_fail_after_recovery": true
				},
				"cros_clean_tmp_owner_request":{
					"docs":[
						"In some cases, the update flow puts the TPM into a state such that it fails verification.",
						"We don't know why. However, this call papers over the problem by clearing the TPM during the reboot.",
						"We ignore failures from 'crossystem'.",
						"Although failure here is unexpected, and could signal a bug, the point of the exercise is to paper over problems."
					],
					"allow_fail_after_recovery": true
				},
				"labstation_uptime_6_hours":{
					"exec_name":"cros_validate_uptime",
					"exec_extra_args":[
						"min_duration:6h"
					]
				},
				"remove_reboot_requests":{
					"exec_name":"cros_remove_reboot_request",
					"allow_fail_after_recovery": true
				},
				"reboot_by_request":{
					"docs":[
						"Some DUTs can request reboot labstation if they has issue with servo-nic or other issues with servo-host."
					],
					"exec_name": "cros_remove_reboot_request",
					"conditions": [
						"cros_has_reboot_request",
						"cros_has_no_servo_in_use",
						"labstation_uptime_6_hours"
					],
					"dependencies":[
						"labstation_reboot",
						"rpm_power_cycle"
					]
				},
				"booted_from_right_kernel":{
					"docs":[
						"Verified if kernel has update and waiting for update.",
						"Kernel can wait for reboot as provisioning is not doing reboot by default for labstations."
					],
					"exec_name": "cros_kernel_priority_has_not_changed",
					"conditions": [
						"remove_reboot_requests",
						"cros_has_no_servo_in_use"
					],
					"recovery_actions":[
						"labstation_reboot",
						"rpm_power_cycle"
					]
				},
				"cros_ping":{
					"recovery_actions":[
						"rpm_power_cycle"
					]
				},
				"cros_ssh":{
					"recovery_actions":[
						"rpm_power_cycle"
					]
				},
				"labstation_reboot":{
					"docs":[
						"Before reboot we need stop powerd service. Assumption if trigger power the labstation will not boot in normal mode."
					],
					"dependencies":[
						"cros_stop_powerd",
						"cros_clean_tmp_owner_request",
						"cros_allowed_reboot",
						"remove_reboot_requests"
					],
					"exec_name": "cros_reboot",
					"exec_timeout": {
						"seconds": 120
					},
					"run_control": 1
				},
				"rpm_power_cycle":{
					"docs":[
						"Action is always runnable."
					],
					"conditions": [
						"has_rpm_info"
					],
					"run_control": 1
				}
			}
		},
		"servo_repair":{
			"critical_actions": [
				"servo_host_info",
				"servod_info",
				"create_host",
				"cros_ssh",
				"servo_v3_uptime",
				"lock_labstation",
				"has_enough_disk_space",
				"servo_components_check",
				"servo_fw_need_update",
				"servo_host_servod_init",
				"servod_echo",
				"servod_get_serialname",
				"servod_servo_pd",
				"servod_read_servo_type",
				"servo_cr50_checks",
				"servo_connection_pins",
				"servo_ppdut5",
				"servod_set_main_device",
				"servo_ec_check",
				"servo_testlab_enabled",
				"servo_detect_usbkey",
				"servo_servod_echo_host"
			],
			"actions": {
				"servo_host_servod_restart": {
					"exec_timeout": {
						"seconds": 120
					}
				},
				"servo_host_servod_init": {
					"exec_timeout": {
						"seconds": 120
					}
				},
				"servo_host_info": {
					"dependencies" : [
						"has_dut_name"
					],
					"exec_name":"sample_pass"
				},
				"servod_info": {
					"exec_name":"sample_pass"
				},
				"create_host": {
					"docs": [
						"Only to create docker"
					],
					"exec_name":"sample_pass"
				},
				"servo_v3_uptime": {
					"exec_name":"cros_validate_uptime",
					"exec_extra_args": [
						"max_duration:96h"
					],
					"conditions": [
						"is_servo_v3"
					],
					"recovery_actions": [
						"reboot"
					]
				},
				"is_labstation": {
					"docs": [
						"condition to check is name contains labstation word"
					],
					"exec_name":"sample_pass"
				},
				"is_container": {
					"docs": [
						"condition to check is name contains labstation word"
					],
					"exec_name":"sample_pass"
				},
				"is_servo_v3": {
					"docs": [
						"Revers from labstation and docker"
					],
					"exec_name":"sample_pass"
				},
				"lock_labstation": {
					"docs": [
						"create lock file is_in_use"
					],
					"conditions": [
						"is_labstation"
					],
					"exec_name":"sample_pass"
				},
				"has_enough_disk_space": {
					"docs":[
						"check the stateful partition have enough disk space that is at least 0.5GB. The storage unit is in GB."
					],
					"conditions": [
						"is_not_container"
					],
					"exec_name":"cros_has_enough_storage_space",
					"exec_extra_args":[
						"/mnt/stateful_partition:0.5"
					]
				},
				"is_not_container": {
					"conditions": [
						"is_servo_v3",
						"is_labstation"
					],
					"exec_name":"sample_fail"
				},
				"servo_components_check": {
					"dependencies": [
						"cros_ssh",
						"root_present_servo_v3",
						"servo_v4_root_present",
						"servo_topology"
					],
					"exec_name":"sample_pass"
				},
				"servo_topology": {
					"docs": [
						"host.check_diskspace('/mnt/stateful_partition', 0.5)"
					],
					"conditions": [
						"is_not_servo_v3"
					],
					"dependencies": [
						"cros_ssh",
						"servo_topology_single_child",
						"servo_topology_dual_setup"
					],
					"exec_name":"sample_pass"
				},
				"servo_topology_single_child": {
					"conditions": [
						"is_not_servo_v3"
					],
					"exec_name":"servo_topology_update",
					"exec_extra_args": [
						"min_child:1",
						"persist_topology:true"
					]
				},
				"servo_topology_dual_setup": {
					"conditions": [
						"is_not_servo_v3",
						"is_dual_setup"
					],
					"exec_name":"servo_topology_update",
					"exec_extra_args": [
						"min_child:2",
						"persist_topology:true"
					],
					"recovery_actions": [
						"servod_restart_dut"
					],
					"allow_fail_after_recovery": true
				},
				"root_present_servo_v3": {
					"dependencies": [
						"cros_ssh"
					],
					"conditions": [
						"is_servo_v3"
					],
					"exec_name":"sample_pass"
				},
				"servo_v4_root_present": {
					"dependencies": [
						"cros_ssh"
					],
					"conditions": [
						"is_not_servo_v3"
					]
				},
				"servo_fw_need_update": {
					"conditions": [
						"is_not_servo_v3"
					],
					"recovery_actions": [
						"servo_fw_update"
					]
				},
				"servod_get_serialname": {
					"docs":[
						"run command from xmlrpc"
					],
					"exec_name":"sample_pass"
				},
				"servod_servo_pd": {
					"docs":[
						"run command from xmlrpc"
					],
					"conditions": [
						"servo_type_c"
					],
					"allow_fail_after_recovery": true,
					"exec_name":"sample_pass",
					"exec_extra_args": [
						"exclude_model:puff,hh"
					]
				},
				"servo_cr50_checks": {
					"conditions": [
						"is_not_servo_v3"
					],
					"dependencies": [
						"servo_cr50_low_sbu",
						"servo_cr50_enumerated",
						"servo_cr50_console"
					],
					"exec_name":"sample_pass"
				},
				"servo_cr50_low_sbu": {
					"conditions": [
						"is_not_servo_v3",
						"servo_type_c"
					],
					"exec_name":"sample_pass"
				},
				"servo_cr50_enumerated": {
					"docs": [
						"prev name servo_cr50_off"
					],
					"conditions": [
						"is_not_servo_v3",
						"servo_type_c"
					],
					"exec_name":"sample_pass"
				},
				"servo_cr50_console": {
					"conditions": [
						"is_not_servo_v3"
					],
					"dependencies": [
						"servod_dut_controller"
					],
					"exec_extra_args": [
						"cr50_ccd_level",
						"cr50_testlab",
						"cr50_ccd_state_flags"
					],
					"exec_name":"sample_pass"
				},
				"cr50_testlab": {
					"exec_extra_args": [
						"command:cr50_testlab",
						"expected_string_value:on"
					],
					"exec_name":"servo_check_servod_control"
				},
				"servod_dut_controller": {
					"conditions": [
						"is_not_servo_v3"
					],
					"exec_name":"sample_pass"
				},
				"servo_connection_pins": {
					"conditions": [
						"is_not_servo_v3"
					],
					"dependencies": [
						"servo_cold_reset_pin",
						"servo_warm_reset_pin"
					],
					"exec_name":"sample_pass"
				},
				"servo_pwr_button_pin": {
					"exec_extra_args": [
						"except:board1,board2"
					],
					"exec_name":"sample_pass"
				},
				"servo_ppdut5": {
					"conditions": [
						"is_not_servo_v3",
						"servo_v4_type_a"
					],
					"exec_name":"servo_low_ppdut5"
				},
				"servo_ec_check": {
					"conditions": [
						"is_not_servo_v3",
						"dut_has_cros_ec"
					],
					"dependencies": [
						"servo_ec_console",
						"servo_pwr_button_pin",
						"servo_lid_open",
						"servo_battery_charging"
					],
					"exec_name":"sample_pass"
				},
				"servo_ec_console": {
					"conditions": [
						"is_not_servo_v3",
						"dut_has_cros_ec"
					],
					"exec_name":"sample_pass"
				},
				"servo_battery_charging": {
					"conditions": [
						"is_not_servo_v3",
						"dut_has_cros_ec"
					],
					"allow_fail_after_recovery": true,
					"exec_name":"sample_pass"
				},
				"servo_testlab_enabled": {
					"conditions": [
						"is_not_servo_v3",
						"servo_v4_type_c"
					],
					"exec_name":"sample_pass"
				},
				"servo_detect_usbkey": {
					"docs": [
						"Will detect the path to USB Drive on servo-host.",
						"Verify that usb-key is responsive"
					],
					"allow_fail_after_recovery": true
				},
				"servo_audit_usbkey": {
					"docs": ["This action will detect whether or not the USB drive is in working condition."],
					"allow_fail_after_recovery": true,
					"dependencies" : ["servo_detect_usbkey"],
					"exec_timeout": {
						"seconds":7300
					}
				},
				"servo_type_c": {
					"exec_name":"sample_pass"
				},
				"servo_lid_open": {
					"exec_name":"sample_pass"
				},
				"is_not_servo_v3": {
					"exec_name":"sample_pass"
				},
				"servo_v4_type_a": {
					"exec_name":"sample_pass"
				},
				"is_dual_setup": {
					"exec_name":"sample_pass"
				},
				"is_not_dual_setup": {
					"exec_name":"sample_pass"
				},
				"servod_set_main_device": {
					"exec_name":"sample_pass"
				},
				"servo_v4_type_c": {
					"exec_name":"sample_pass"
				},
				"servo_fw_update": {
					"docs":[
						"Try to update in  normal ways 3 times",
						"if fail allow run force update"
					],
					"exec_extra_args": [
						"try:3",
						"use_force"
					],
					"exec_name":"sample_pass"
				},
				"servo_warm_reset_pin": {
					"exec_name":"sample_pass"
				},
				"dut_has_cros_ec": {
					"exec_name":"sample_pass"
				},
				"reboot":{
					"exec_name":"sample_pass"
				},
				"servo_cold_reset_pin": {
					"exec_extra_args": [
						"command:cold_reset",
						"expected_string_value:off"
					],
					"exec_name":"servo_check_servod_control"
				},
				"servod_read_servo_type": {
					"exec_name":"sample_pass"
				},
				"servod_restart_dut": {
					"exec_name":"sample_pass"
				},
				"servo_servod_echo_host": {
					"exec_timeout": {
						"seconds": 30
					}
				},
				"servod_echo": {
					"dependencies" : ["servo_servod_echo_host"]
				}
			},
			"allow_fail": true
		},
		"chameleon_repair":{
			"critical_actions": [
				"sample_pass"
			],
			"actions": {},
			"allow_fail": true
		},
		"bluetooth_peer_repair":{
			"critical_actions": [
				"btpeer_state_broken",
				"cros_ping",
				"cros_ssh",
				"btpeer_state_working"
			],
			"actions": {},
			"allow_fail": true
		},
		"cros_repair":{
			"critical_actions": [
				"dut_state_repair_failed",
				"cros_ssh",
				"internal_storage",
				"last_provision_successful",
				"device_system_info",
				"has_python",
				"device_enrollment",
				"power_info",
				"tpm_info",
				"tools_checks",
				"hardware_audit",
				"firmware_check",
				"servo_keyboard"
			],
			"actions": {
				"cros_ssh":{
					"dependencies":[
						"has_dut_name",
						"has_dut_board_name",
						"has_dut_model_name",
						"cros_ping"
					]
				},
				"internal_storage":{
					"dependencies":[
						"cros_storage_writing",
						"cros_storage_file_system",
						"cros_storage_space_check",
						"cros_audit_storage_smart"
					],
					"exec_name":"sample_pass"
				},
				"device_system_info":{
					"dependencies":[
						"cros_default_boot",
						"cros_boot_in_normal_mode",
						"cros_hwid_info",
						"cros_serial_number_info",
						"cros_tpm_fwver_match",
						"cros_tpm_kernver_match"
					],
					"exec_name":"sample_pass"
				},
				"has_python":{
					"dependencies":[
						"cros_storage_writing"
					],
					"exec_name":"cros_has_python_interpreter_working"
				},
				"last_provision_successful":{
					"dependencies":[
						"cros_storage_writing"
					],
					"exec_name":"cros_is_last_provision_successful"
				},
				"device_enrollment":{
					"dependencies":[
						"cros_storage_writing"
					],
					"exec_name":"cros_is_enrollment_in_clean_state"
				},
				"power_info":{
					"conditions":[
						"cros_is_not_virtual_machine"
					],
					"dependencies":[
						"cros_storage_writing",
						"cros_is_ac_power_connected",
						"battery_is_good"
					],
					"exec_name":"sample_pass"
				},
				"tpm_info":{
					"conditions":[
						"cros_is_not_virtual_machine",
						"cros_is_tpm_present"
					],
					"exec_name":"cros_is_tpm_in_good_status"
				},
				"tools_checks":{
					"dependencies":[
						"cros_gsctool"
					],
					"exec_name":"sample_pass"
				},
				"hardware_audit":{
					"dependencies":[
						"wifi_audit",
						"bluetooth_audit"
					],
					"exec_name":"sample_pass"
				},
				"firmware_check":{
					"dependencies":[
						"cros_storage_writing",
						"cros_is_firmware_in_good_state",
						"cros_rw_firmware_stable_verion"
					],
					"exec_name":"sample_pass"
				},
				"servo_keyboard":{
					"conditions":[
						"servo_state_is_working",
						"is_servo_keyboard_image_tool_present"
					],
					"dependencies":[
						"servo_init_usb_keyboard",
						"lufa_keyboard_found"
					],
					"exec_name":"cros_run_shell_command",
					"exec_extra_args":[
						"lsusb -vv -d 03eb:2042 |grep \"Remote Wakeup\""
					],
					"allow_fail_after_recovery": true
				},
				"servo_init_usb_keyboard":{
					"docs":[
						"set servo's 'init_usb_keyboard' command to 'on' value."
					],
					"exec_name":"servo_set",
					"exec_extra_args":[
						"command:init_usb_keyboard",
						"string_value:on"
					]
				},
				"is_servo_keyboard_image_tool_present":{
					"docs":[
						"check if the servo keyboard image specified by the name of dfu-programmer can be found in DUT cli."
					],
					"exec_name":"cros_is_tool_present",
					"exec_extra_args":[
						"tool:dfu-programmer"
					]
				},
				"lufa_keyboard_found":{
					"docs":[
						"check if the lufa keyboard can be found by finding the match of the model information of it."
					],
					"exec_name":"cros_run_shell_command",
					"exec_extra_args":[
						"lsusb -d 03eb:2042 |grep \"LUFA Keyboard Demo\""
					]
				},
				"servo_state_is_working":{
					"docs":[
						"check the servo's state is ServoStateWorking."
					],
					"exec_name":"servo_match_state",
					"exec_extra_args":[
						"state:WORKING"
					]
				},
				"cros_rw_firmware_stable_verion":{
					"dependencies":[
						"cros_storage_writing",
						"cros_is_on_rw_firmware_stable_verion",
						"cros_is_rw_firmware_stable_version_available"
					],
					"exec_name":"sample_pass"
				},
				"cros_gsctool":{
					"exec_name":"sample_pass"
				},
				"battery_is_good":{
					"docs":[
						"Check battery on the DUT is normal and update battery hardware state accordingly."
					],
					"conditions":[
						"cros_is_battery_expected",
						"cros_is_not_virtual_machine",
						"cros_is_battery_present"
					],
					"dependencies":[
						"cros_storage_writing",
						"cros_is_battery_chargable_or_good_level"
					],
					"exec_name":"cros_audit_battery"
				},
				"wifi_audit":{
					"docs":[
						"Check wifi on the DUT is normal and update wifi hardware state accordingly."
					],
					"dependencies":[
						"cros_ssh"
					],
					"exec_name":"cros_audit_wifi",
					"allow_fail_after_recovery": true
				},
				"bluetooth_audit":{
					"docs":[
						"Check bluetooth on the DUT is normal and update bluetooth hardware state accordingly."
					],
					"dependencies":[
						"cros_ssh"
					],
					"exec_name":"cros_audit_bluetooth",
					"allow_fail_after_recovery": true
				},
				"cros_tpm_fwver_match":{
					"dependencies":[
						"cros_storage_writing"
					],
					"exec_name":"cros_match_dev_tpm_firmware_version"
				},
				"cros_tpm_kernver_match":{
					"dependencies":[
						"cros_storage_writing"
					],
					"exec_name":"cros_match_dev_tpm_kernel_version"
				},
				"cros_default_boot":{
					"dependencies":[
						"cros_storage_writing"
					],
					"exec_name":"cros_is_default_boot_from_disk"
				},
				"cros_boot_in_normal_mode":{
					"dependencies":[
						"cros_storage_writing"
					],
					"exec_name":"cros_is_not_in_dev_mode"
				},
				"cros_hwid_info":{
					"conditions":[
						"dut_has_hwid_info"
					],
					"dependencies":[
						"cros_storage_writing"
					],
					"exec_name":"cros_match_hwid_to_inventory"
				},
				"cros_serial_number_info":{
					"conditions":[
						"dut_has_serial_number_info"
					],
					"dependencies":[
						"cros_storage_writing"
					],
					"exec_name":"cros_match_serial_number_inventory"
				},
				"dut_has_hwid_info":{
					"exec_name":"sample_pass"
				},
				"dut_has_serial_number_info":{
					"exec_name":"sample_pass"
				},
				"cros_storage_writing":{
					"dependencies":[
						"cros_ssh"
					],
					"exec_name":"cros_is_file_system_writable"
				},
				"cros_storage_file_system":{
					"dependencies":[
						"cros_ssh"
					],
					"exec_name":"cros_has_critical_kernel_error"
				},
				"cros_storage_space_check":{
					"dependencies":[
						"cros_stateful_partition_has_enough_inodes", 
						"cros_stateful_partition_has_enough_storage_space",
						"cros_encrypted_stateful_partition_has_enough_storage_space"
					],
					"exec_name":"sample_pass"
				},
				"cros_stateful_partition_has_enough_inodes":{
					"docs":[
						"check the stateful partition path has enough inodes"
					],
					"exec_name":"cros_has_enough_inodes",
					"exec_extra_args":[
						"/mnt/stateful_partition:100"
					]
				},
				"cros_stateful_partition_has_enough_storage_space":{
					"docs":[
						"check the stateful partition have enough disk space. The storage unit is in GB."
					],
					"exec_name":"cros_has_enough_storage_space",
					"exec_extra_args":[
						"/mnt/stateful_partition:0.7"
					]
				},
				"cros_encrypted_stateful_partition_has_enough_storage_space":{
					"docs":[
						"check the encrypted stateful partition have enough disk space. The storage unit is in GB."
					],
					"exec_name":"cros_has_enough_storage_space",
					"exec_extra_args":[
						"/mnt/stateful_partition/encrypted:0.1"
					]
				}
			}
		},
		"labstation_deploy":{
			"critical_actions": [
				"dut_state_needs_deploy",
				"check_host_info",
				"cros_ping",
				"cros_ssh",
				"update_inv_info",
				"install_stable_os",
				"remove_reboot_requests",
				"update_provisioned_info",
				"validate_rpm",
				"dut_state_ready"
			],
			"actions": {
				"check_host_info":{
					"docs":[
						"Check basic info for deployment."
					],
					"exec_name": "sample_pass",
					"dependencies":[
						"has_dut_name",
						"has_dut_board_name",
						"has_dut_model_name"
					]
				},
				"update_inv_info":{
					"docs":[
						"Updating device info in inventory."
					],
					"exec_name": "sample_pass",
					"dependencies":[
						"cros_ssh",
						"cros_update_hwid_to_inventory",
						"cros_update_serial_number_inventory"
					]
				},
				"install_stable_os":{
					"docs":[
						"Install stable OS on the device."
					],
					"exec_name": "cros_provision",
					"conditions": [
						"has_stable_version_cros_image",
						"cros_not_on_stable_version"
					],
					"exec_timeout": {
						"seconds": 3600
					}
				},
				"remove_reboot_requests":{
					"docs":[
						"Remove reboot request flag files."
					],
					"exec_name":"cros_remove_reboot_request",
					"allow_fail_after_recovery": true
				},
				"update_provisioned_info":{
					"docs":[
						"Read and update cros-provision label."
					],
					"exec_name": "cros_update_provision_os_version"
				},
				"validate_rpm":{
					"docs":[
						"Validate and update rpm_state.",
						"The execs is not ready yet."
					],
					"exec_name": "rpm_audit",
					"exec_timeout": {
						"seconds": 600
					},
					"conditions": [
						"has_rpm_info"
					]
				}
			}
		},
		"cros_deploy":{
			"critical_actions": [
				"sample_pass"
			],
			"actions": {}
		}
	}
}
 `

// DefaultConfig provides default config for recovery engine.
func DefaultConfig() io.Reader {
	return strings.NewReader(defaultConfig)
}
