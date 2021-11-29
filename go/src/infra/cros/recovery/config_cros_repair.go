// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package recovery

const crosRepairPlanBody = `
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
	"rw_vpd",
	"servo_keyboard",
	"servo_mac_address",
	"device_labels"
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
	"rw_vpd":{
		"docs":[
			"Verify that keys: 'should_send_rlz_ping', 'gbind_attribute', 'ubind_attribute' are present in vpd RW_VPD partition."
		],
		"exec_name":"cros_are_required_rw_vpd_keys_present",
		"allow_fail_after_recovery": true
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
	"servo_mac_address":{
		"conditions":[
			"is_not_servo_v3",
			"servod_control_exist_for_mac_address"
		],
		"exec_name":"servo_audit_nic_mac_address",
		"allow_fail_after_recovery": true
	},
	"is_not_servo_v3": {
		"conditions":[
			"servo_is_v3"
		],
		"exec_name":"sample_fail"
	},
	"servod_control_exist_for_mac_address":{
		"exec_name":"servo_check_servod_control",
		"exec_extra_args":[
			"command:macaddr"
		]
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
	},
	"device_labels":{
		"dependencies":[
			"device_sku",
			"cr_50"
		 ],
		 "exec_name":"sample_pass"
	},
	"cr_50":{
		"docs":[
			"Update the cr_50 label on the cros Device."
		],
		"conditions":[
			"cros_is_cr_50_firmware_exist"
		],
		"exec_name":"cros_update_cr_50",
		"allow_fail_after_recovery": true
	},
	"device_sku":{
		"docs":[
			"Update the device_sku label from the device if not present in inventory data."
		],
		"conditions":[
			"dut_does_not_have_device_sku"
		],
		"exec_name":"cros_update_device_sku",
		"allow_fail_after_recovery": true
	},
	"dut_does_not_have_device_sku":{
		"docs":[
			"Confirm that the DUT itself does not have device_sku label."
		],
		"conditions":[
			"has_dut_device_sku"
		],
		"exec_name":"sample_fail"
	}
}
`
