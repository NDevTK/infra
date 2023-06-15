// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"fmt"
	"strings"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"google.golang.org/protobuf/types/known/durationpb"
	"infra/cros/recovery/tlw"
)

func wifiRouterRepairPlan() *Plan {
	return &Plan{
		CriticalActions: []string{
			// Initial actions to prepare for device-specific actions.
			"Set WifiRouter state to BROKEN",
			"Device is ping-able",
			"Identify and set WifiRouter device_type",

			// Actions only executed for specific device types.
			"Check ChromeOS Gale device",
			"Check OpenWrt device",
			"Check AsusWrt device",

			// General final actions done for all device types.
			"Reboot device",
			"Set WifiRouter state to WORKING",
		},
		Actions: map[string]*Action{
			// Generic AP actions.
			"Set WifiRouter state to BROKEN": {
				ExecName: "wifi_router_set_state",
				ExecExtraArgs: []string{
					fmt.Sprintf("state:%d", tlw.WifiRouterHost_BROKEN),
				},
			},
			"Set WifiRouter state to WORKING": {
				ExecName: "wifi_router_set_state",
				ExecExtraArgs: []string{
					fmt.Sprintf("state:%d", tlw.WifiRouterHost_WORKING),
				},
			},
			"Device is ping-able": {
				ExecName:    "cros_ping",
				ExecTimeout: &durationpb.Duration{Seconds: 15},
			},
			"Identify and set WifiRouter device_type": {
				Docs: []string{
					"Identifies the device type of the test AP by probing the device over ssh.",
					"APs that fail to be identified as one of the supported device types will be left in a broken state.",
				},
				ExecName: "wifi_router_identify_device_type",
			},
			"Reboot device": {
				Docs: []string{
					"Reboots the device over ssh and waits for it to be ssh-able again.",
				},
				ExecName:    "wifi_router_reboot",
				ExecTimeout: &durationpb.Duration{Seconds: 200},
			},

			// Device type conditions.
			"Is ChromeOS Gale": {
				ExecName: "wifi_router_device_type_in_list",
				ExecExtraArgs: []string{
					buildWifiRouterDeviceTypesArg(labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_CHROMEOS_GALE),
				},
			},
			"Is OpenWrt": {
				ExecName: "wifi_router_device_type_in_list",
				ExecExtraArgs: []string{
					buildWifiRouterDeviceTypesArg(labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_OPENWRT),
				},
			},
			"Is AsusWrt": {
				ExecName: "wifi_router_device_type_in_list",
				ExecExtraArgs: []string{
					buildWifiRouterDeviceTypesArg(labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_ASUSWRT),
				},
			},

			// ChromeOS Gale actions.
			"Check ChromeOS Gale device": {
				Conditions: []string{
					"Is ChromeOS Gale",
				},
				Dependencies: []string{
					"Device is on stable-version",
					"Device has required wifi tools",
					"Device has 50 percent tmp diskspace",
					"Device has 50 percent stateful partition diskspace",
					"Set WifiRouter model and features to hardcoded values for Gales",
				},
				ExecName: "sample_pass",
			},
			"Device is on stable-version": {
				ExecName: "cros_is_on_stable_version",
				ExecExtraArgs: []string{
					"device_type:wifi_router",
				},
				RecoveryActions: []string{
					"Provision Gale device to stable version",
				},
			},
			"Set WifiRouter model and features to hardcoded values for Gales": {
				Docs: []string{
					"All ChromeOS Gale devices are expected to have the same model and ",
					"features, so we can set these as hardcoded values.",
				},
				ExecName: "wifi_router_update_model_and_features",
			},
			"Device has required wifi tools": {
				Docs: []string{
					"check whether wifirouter critical tools present: ",
					"tcpdump, hostapd, dnsmasq, netperf, iperf, iw",
				},
				Dependencies: []string{
					"cros_ssh",
				},
				ExecName: "cros_is_tool_present",
				ExecExtraArgs: []string{
					"tools:tcpdump,hostapd,dnsmasq,netperf,iperf,iw",
				},
				RecoveryActions: []string{
					"Provision Gale device to stable version",
				},
			},
			"Provision Gale device to stable version": {
				Docs: []string{
					"Install wifirouter stable os.",
					"Currently only has one version",
				},
				ExecName: "cros_provision",
				ExecExtraArgs: []string{
					"device_type:wifi_router",
				},
				ExecTimeout: &durationpb.Duration{Seconds: 3600},
			},
			"Device has 50 percent tmp diskspace": {
				Docs: []string{
					"Check if there are more than 50 percent of diskspace in /tmp",
				},
				ExecName: "cros_has_enough_storage_space_percentage",
				ExecExtraArgs: []string{
					"path:/tmp",
					"expected:50",
				},
				RecoveryActions: []string{
					"Clean up tmp space",
					"Provision Gale device to stable version",
				},
			},
			"Clean up tmp space": {
				Docs: []string{
					"Clean up tmp space",
				},
				ExecName: "cros_run_shell_command",
				ExecExtraArgs: []string{
					"rm -Rf /tmp/*",
				},
			},
			"Device has 50 percent stateful partition diskspace": {
				Docs: []string{
					"Check if there are more than 50 percent of diskspace in /mnt/stateful_partition",
				},
				ExecName: "cros_has_enough_storage_space_percentage",
				ExecExtraArgs: []string{
					"path:/mnt/stateful_partition",
					"expected:50",
				},
				RecoveryActions: []string{
					"Clean up stateful sub space",
					"Provision Gale device to stable version",
				},
			},
			"Clean up stateful sub space": {
				Docs: []string{
					"Clean up  /mnt/stateful_partition/home/.shadow ,/mnt/stateful_partition/dev_image/telemetry space",
				},
				ExecName: "cros_run_shell_command",
				ExecExtraArgs: []string{
					"rm -Rf /mnt/stateful_partition/home/.shadow /mnt/stateful_partition/dev_image/telemetry",
				},
			},

			// OpenWrt actions.
			"Check OpenWrt device": {
				Conditions: []string{
					"Is OpenWrt",
				},
				Dependencies: []string{
					"Fetch OpenWrt OS image build info from device",
					"Fetch OpenWrt image config from GCS",
					"Identify expected OS image for this OpenWrt device",
					"Device has expected OpenWrt OS image",
					"Set WifiRouter model and features based on this OpenWrt device",
				},
				ExecName: "sample_pass",
			},
			"Device has expected OpenWrt OS image": {
				RecoveryActions: []string{
					"Update OpenWrt OS image with expected image",
				},
				ExecName: "wifi_router_openwrt_has_expected_image",
			},
			"Set WifiRouter model and features based on this OpenWrt device": {
				Docs: []string{
					"The model and features are retrieved from the image build info file ",
					"included on every OpenWrt OS image built for ChromeOS testing.",
				},
				ExecName: "wifi_router_update_model_and_features",
			},
			"Identify expected OS image for this OpenWrt device": {
				ExecName: "wifi_router_openwrt_identify_expected_image",
			},
			"Fetch OpenWrt OS image build info from device": {
				ExecName: "wifi_router_openwrt_fetch_build_info",
			},
			"Fetch OpenWrt image config from GCS": {
				ExecName: "wifi_router_openwrt_fetch_config",
			},
			"Update OpenWrt OS image with expected image": {
				ExecName:    "wifi_router_openwrt_update_to_expected_image",
				ExecTimeout: &durationpb.Duration{Seconds: 600},
			},

			// AsusWrt actions.
			"Check AsusWrt device": {
				Conditions: []string{
					"Is AsusWrt",
				},
				Dependencies: []string{
					"Fetch model from AsusWrt device",
					"Update model and features based on this AsusWrt device",
				},
				ExecName: "sample_pass",
			},
			"Update model and features based on this AsusWrt device": {
				Docs: []string{
					"Sets model based on data read from the AsusWrt device and sets the ",
					"features based on known, hardcoded values based on model.",
				},
				ExecName: "wifi_router_update_model_and_features",
			},
			"Fetch model from AsusWrt device": {
				ExecName: "wifi_router_asuswrt_fetch_model",
			},
		},
	}
}

func buildWifiRouterDeviceTypesArg(deviceTypes ...labapi.WifiRouterDeviceType) string {
	var argValue []string
	for _, dt := range deviceTypes {
		argValue = append(argValue, fmt.Sprintf("%d", dt))
	}
	return fmt.Sprintf("device_types:%s", strings.Join(argValue, ","))
}
