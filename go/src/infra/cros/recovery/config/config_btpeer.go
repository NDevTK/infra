// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

func btpeerRepairPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			"Set state: BROKEN",
			"Device is pingable",
			"Device is SSHable",
			"Chameleond version is up to date",
			"Chameleond service is running",
			"Device has been rebooted recently",
			"Set state: WORKING",
		},
		Actions: map[string]*Action{
			"Set state: BROKEN": {
				Docs: []string{
					"The device state BROKEN.",
				},
				ExecName:      "btpeer_state_broken",
				RunControl:    RunControl_RUN_ONCE,
				MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
			},
			"Set state: WORKING": {
				Docs: []string{
					"The device state WORKING.",
				},
				ExecName:      "btpeer_state_working",
				RunControl:    RunControl_RUN_ONCE,
				MetricsConfig: &MetricsConfig{UploadPolicy: MetricsConfig_SKIP_ALL},
			},
			"Device is pingable": {
				Docs: []string{
					"Wait device to be pingable.",
					"Waiting time 15 seconds.",
				},
				ExecName:    "cros_ping",
				ExecTimeout: &durationpb.Duration{Seconds: 15},
				RunControl:  RunControl_ALWAYS_RUN,
			},
			"Device is SSHable": {
				Docs: []string{
					"Try to wait device to be sshable.",
					"Waiting time 150 seconds.",
				},
				ExecName:   "cros_ssh",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Device has been rebooted recently": {
				Docs: []string{
					"Checks the device's uptime and fails if it is not less than one day.",
					"Recovers by rebooting the device.",
				},
				ExecName: "btpeer_assert_uptime_is_less_than_duration",
				ExecExtraArgs: []string{
					"duration_min:1440",
				},
				RecoveryActions: []string{
					"Reboot device",
				},
			},
			"Reboot device": {
				Docs: []string{
					"Reboots the device over ssh and waits for the device to become ssh-able again.",
				},
				ExecName:    "btpeer_reboot",
				ExecTimeout: durationpb.New(5 * time.Minute),
				RunControl:  RunControl_ALWAYS_RUN,
			},
			"Chameleond version is up to date": {
				Docs: []string{
					"Checks if the chameleond version on the btpeer is outdated and updates it if it is.",
				},
				Dependencies: []string{
					"Fetch btpeer chameleond release config from GCS",
					"Identify expected chameleond release bundle",
					"Fetch installed chameleond bundle commit from btpeer",
					"Btpeer has expected chameleond bundle installed",
				},
				ExecName: "sample_pass",
			},
			"Fetch btpeer chameleond release config from GCS": {
				Docs: []string{
					"Retrieves the production btpeer chameleond config from GCS and stores it in the exec state for later reference.",
				},
				ExecName:   "btpeer_fetch_btpeer_chameleond_release_config",
				RunControl: RunControl_RUN_ONCE,
			},
			"Identify expected chameleond release bundle": {
				Docs: []string{
					"Identifies the expected chameleond release bundle based off of the config and DUT host.",
					"Note: For now this step ignores the DUT host and always selects the latest, non-next bundle.",
				},
				ExecName:   "btpeer_identify_expected_chameleond_release_bundle",
				RunControl: RunControl_RUN_ONCE,
			},
			"Fetch installed chameleond bundle commit from btpeer": {
				Docs: []string{
					"Retrieves the chameleond commit of the currently installed chameleond version from a log file on the btpeer.",
					"If it fails to retrieve the commit we assume the chameleond version is too old to have the needed log file and we recover by installing the expected chameleond bundle.",
				},
				ExecName: "btpeer_fetch_installed_chameleond_bundle_commit",
				RecoveryActions: []string{
					"Install expected chameleond release bundle and then reboot device",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Btpeer has expected chameleond bundle installed": {
				Docs: []string{
					"Checks if the installed chameleond commit matches the expected chameleond bundle commit.",
					"If the check fails, it attempts to recover by updating chameleond with the expected bundle.",
				},
				ExecName: "btpeer_assert_btpeer_has_expected_chameleond_release_bundle_installed",
				RecoveryActions: []string{
					"Install expected chameleond release bundle and then reboot device",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Install expected chameleond release bundle and then reboot device": {
				Docs: []string{
					"Installs/updates chameleond on the btpeer and then reboots the device.",
				},
				Dependencies: []string{
					"Install expected chameleond release bundle",
					"Reboot device",
				},
				ExecName:   "sample_pass",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Install expected chameleond release bundle": {
				Docs: []string{
					"Installs/updates chameleond on the btpeer with the expected chameleond bundle.",
					"The expected bundle archive is downloaded from GCS to the btpeer through the cache, extracted, and installed via make.",
				},
				ExecName:    "btpeer_install_expected_chameleond_release_bundle",
				ExecTimeout: durationpb.New(15 * time.Minute),
				RecoveryActions: []string{
					"Reboot device",
				},
			},
			"Chameleond service is running": {
				Docs: []string{
					"Checks the status of the chameleond service on the device to see if it is running.",
					"Fails if the service is not running and attempts recovery by rebooting.",
				},
				ExecName: "btpeer_assert_chameleond_service_is_running",
				RecoveryActions: []string{
					"Reboot device",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
		},
	}
}
