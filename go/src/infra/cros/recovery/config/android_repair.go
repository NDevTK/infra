// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"path/filepath"

	"google.golang.org/protobuf/types/known/durationpb"
)

// Below are vendor keys used for testing purpose and they're already public visible.
const (
	adbPrivateVendorKeyFile = "/var/lib/android_keys/arc.adb_key"
	adbPublicVendorKeyFile  = "/data/misc/adb/adb_keys"
	adbPrivateVendorKey     = `-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCnHNzujonYRLoI
F2pyJX1SSrqmiT/3rTRCP1X0pj1V/sPGwgvIr+3QjZehLUGRQL0wneBNXd6EVrST
drO4cOPwSxRJjCf+/PtS1nwkz+o/BGn5yhNppdSro7aPoQxEVM8qLtN5Ke9tx/zE
ggxpF8D3XBC6Los9lAkyesZI6xqXESeofOYu3Hndzfbz8rAjC0X+p6Sx561Bt1dn
T7k2cP0mwWfITjW8tAhzmKgL4tGcgmoLhMHl9JgScFBhW2Nd0QAR4ACyVvryJ/Xa
2L6T2YpUjqWEDbiJNEApFb+m+smIbyGz0H/Kj9znoRs84z3/8rfyNQOyf7oqBpr2
52XG4totAgMBAAECggEARisKYWicXKDO9CLQ4Uj4jBswsEilAVxKux5Y+zbqPjeR
AN3tkMC+PHmXl2enRlRGnClOS24ExtCZVenboLBWJUmBJTiieqDC7o985QAgPYGe
9fFxoUSuPbuqJjjbK73olq++v/tpu1Djw6dPirkcn0CbDXIJqTuFeRqwM2H0ckVl
mVGUDgATckY0HWPyTBIzwBYIQTvAYzqFHmztcUahQrfi9XqxnySI91no8X6fR323
R8WQ44atLWO5TPCu5JEHCwuTzsGEG7dEEtRQUxAsH11QC7S53tqf10u40aT3bXUh
XV62ol9Zk7h3UrrlT1h1Ae+EtgIbhwv23poBEHpRQQKBgQDeUJwLfWQj0xHO+Jgl
gbMCfiPYvjJ9yVcW4ET4UYnO6A9bf0aHOYdDcumScWHrA1bJEFZ/cqRvqUZsbSsB
+thxa7gjdpZzBeSzd7M+Ygrodi6KM/ojSQMsen/EbRFerZBvsXimtRb88NxTBIW1
RXRPLRhHt+VYEF/wOVkNZ5c2eQKBgQDAbwNkkVFTD8yQJFxZZgr1F/g/nR2IC1Yb
ylusFztLG998olxUKcWGGMoF7JjlM6pY3nt8qJFKek9bRJqyWSqS4/pKR7QTU4Nl
a+gECuD3f28qGFgmay+B7Fyi9xmBAsGINyVxvGyKH95y3QICw1V0Q8uuNwJW2feo
3+UD2/rkVQKBgFloh+ljC4QQ3gekGOR0rf6hpl8D1yCZecn8diB8AnVRBOQiYsX9
j/XDYEaCDQRMOnnwdSkafSFfLbBrkzFfpe6viMXSap1l0F2RFWhQW9yzsvHoB4Br
W7hmp73is2qlWQJimIhLKiyd3a4RkoidnzI8i5hEUBtDsqHVHohykfDZAoGABNhG
q5eFBqRVMCPaN138VKNf2qon/i7a4iQ8Hp8PHRr8i3TDAlNy56dkHrYQO2ULmuUv
Erpjvg5KRS/6/RaFneEjgg9AF2R44GrREpj7hP+uWs72GTGFpq2+v1OdTsQ0/yr0
RGLMEMYwoY+y50Lnud+jFyXHZ0xhkdzhNTGqpWkCgYBigHVt/p8uKlTqhlSl6QXw
1AyaV/TmfDjzWaNjmnE+DxQfXzPi9G+cXONdwD0AlRM1NnBRN+smh2B4RBeU515d
x5RpTRFgzayt0I4Rt6QewKmAER3FbbPzaww2pkfH1zr4GJrKQuceWzxUf46K38xl
yee+dcuGhs9IGBOEEF7lFA==
-----END PRIVATE KEY-----`
	adbPublicVendorKey = "QAAAAFt6z0Mt2uLGZef2mgYqun+yAzXyt/L/PeM8G6Hn3I/Kf9CzIW+IyfqmvxUpQDSJuA2EpY5UitmTvtja9Sfy+layAOARANFdY1thUHASmPTlwYQLaoKc0eILqJhzCLS8NU7IZ8Em/XA2uU9nV7dBreexpKf+RQsjsPLz9s3dedwu5nyoJxGXGutIxnoyCZQ9iy66EFz3wBdpDILE/Mdt7yl50y4qz1REDKGPtqOr1KVpE8r5aQQ/6s8kfNZS+/z+J4xJFEvw43C4s3aTtFaE3l1N4J0wvUCRQS2hl43Q7a/IC8LGw/5VPab0VT9CNK33P4mmukpSfSVyahcIukTYiY7u3Byn0Nc9qhPPbSQYNQiofN7w91BWzW46V8CgWzBCKZoKhF7YmTdAm48qmaV0rqMGaf1AtRz5QY0a47seRYCgk9lMx7BeMgIuAZDmYPsUG+mAG+IiQYfvJMIEMBowtc8IlfZv9A7bwLKcs4rRhxFdCzJ7odPgFdgUv7MEAYF+HhnQg6DYEhoqe7YkB98Pb8VbU4f/ZTNkHYtIOxMIb53saW09zop5MlQrR6E7hBeZ5FwMNOK7+yc20ulUlqq38iB6QoHx7lli8dfGpD47J1ETHw7m9uAuxMu75MD4bIxYgmj2Ud1TvmWqXtmg75+E+B1I3osGcw9a2Qxo2ypV1Nkq8b1lmgEAAQA= root@localhost"
	wifiSSID           = "nearbysharing_1"
	wifiSecurity       = "wpa2"
	wifiPassword       = "password"
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
		},
		"Set state: repair_failed": {
			Docs: []string{
				"Initial state of Android DUT before repair to indicate failure if recovery fails by any reason.",
			},
			ExecName:      "dut_set_state",
			ExecExtraArgs: []string{"state:repair_failed"},
			RunControl:    RunControl_RUN_ONCE,
		},
		"Set state: ready": {
			Docs: []string{
				"Final state of Android DUT indicating successful repair.",
			},
			ExecName:      "dut_set_state",
			ExecExtraArgs: []string{"state:ready"},
			RunControl:    RunControl_RUN_ONCE,
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
				"Associated host FS is writeable",
			},
			ExecName: "sample_pass",
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
		"Associated host FS is writeable": {
			Docs: []string{
				"This verifier checks whether associated host FS is writeable.",
			},
			ExecName: "android_associated_host_fs_is_writable",
			RecoveryActions: []string{
				"Schedule associated host reboot and fail",
			},
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
				"This verifier checks whether adb and vendor key are properly provisioned on associated host of the DUT.",
			},
			Dependencies: []string{
				"Validate associated host",
				"Associated host has vendor key",
				"Associated host has adb",
				"Adb server is running",
			},
			ExecName: "sample_pass",
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
		"Adb server is running": {
			Docs: []string{
				"This verifier ensures that Adb server is running on associated host of the DUT.",
			},
			Dependencies: []string{
				"Adb server is stopped",
			},
			ExecName:      "android_associated_host_start_adb",
			ExecExtraArgs: []string{"adb_vendor_key:" + filepath.Dir(adbPrivateVendorKeyFile)},
			RecoveryActions: []string{
				"Schedule associated host reboot and fail",
			},
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
			ExecName: "android_dut_is_accessible",
			RecoveryActions: []string{
				"Reboot device if in fastboot mode",
				"Schedule associated host reboot and fail",
			},
		},
		"Schedule associated host reboot and fail": {
			Docs: []string{
				"Schedules reboot of the DUT associated host and fails repair till the next run.",
			},
			Dependencies: []string{
				"Validate associated host",
				"android_associated_host_schedule_reboot",
			},
			ExecName:   "sample_fail",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Reset DUT": {
			Docs: []string{"Resets DUT to factory settings."},
			Conditions: []string{
				"DUT is rooted",
			},
			Dependencies: []string{
				"Validate associated host",
				"Validate adb",
				"DUT is accessible over adb",
				"Reset public key",
				"android_dut_reset",
				"Wait for DUT",
				"Connect to WiFi network",
			},
			ExecName: "sample_pass",
		},
		"DUT is rooted": {
			Docs: []string{"This verifier checks whether the DUT is rooted."},
			Dependencies: []string{
				"Validate associated host",
				"Validate adb",
				"DUT is accessible over adb",
			},
			ExecName: "android_dut_is_rooted",
		},
		"Wait for DUT": {
			Docs: []string{"Waits for DUT to become available."},
			Dependencies: []string{
				"android_wait_for_offline_dut",
				"Sleep 60s",
			},
			ExecName:    "android_wait_for_online_dut",
			ExecTimeout: &durationpb.Duration{Seconds: 300},
		},
		"Sleep 60s": {
			ExecName:      "sample_sleep",
			ExecExtraArgs: []string{"sleep:60"},
			ExecTimeout:   &durationpb.Duration{Seconds: 90},
			RunControl:    RunControl_ALWAYS_RUN,
		},
		"Connect to WiFi network": {
			Docs: []string{"Connects DUT to WiFi network."},
			Conditions: []string{
				"DUT is rooted",
			},
			Dependencies: []string{
				"Validate associated host",
				"Validate adb",
				"DUT is accessible over adb",
				"Ensure adbd runs as root",
				"android_enable_wifi",
			},
			ExecName: "android_connect_wifi_network",
			ExecExtraArgs: []string{
				"wifi_ssid:" + wifiSSID,
				"wifi_security:" + wifiSecurity,
				"wifi_password:" + wifiPassword,
			},
		},
		"Ensure adbd runs as root": {
			Docs:       []string{"Restart adbd with root permission."},
			ExecName:   "android_restart_adbd_as_root",
			RunControl: RunControl_ALWAYS_RUN,
		},
		"Reset public key": {
			Docs: []string{"Validates and restores ADB public vendor key."},
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
				"This action will also restart adb server as we may need to re-auth.",
			},
			Dependencies: []string{
				"android_device_in_fastboot_mode",
				"android_reboot_device_via_fastboot",
				"Sleep 60s",
				"Stop ADB server",
				"Start ADB server",
				"android_wait_for_online_dut",
			},
			ExecName: "sample_pass",
		},
		"Start ADB server": {
			Docs:          []string{"Start adb server, this action will always run."},
			ExecExtraArgs: []string{"adb_vendor_key:" + filepath.Dir(adbPrivateVendorKeyFile)},
			ExecName:      "android_associated_host_start_adb",
			RunControl:    RunControl_ALWAYS_RUN,
		},
		"Stop ADB server": {
			Docs:       []string{"Stop adb server, this action will always run."},
			ExecName:   "android_associated_host_stop_adb",
			RunControl: RunControl_ALWAYS_RUN,
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
