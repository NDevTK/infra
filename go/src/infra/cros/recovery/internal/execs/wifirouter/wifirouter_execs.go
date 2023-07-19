// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package wifirouter

import (
	"context"
	"strconv"
	"strings"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"infra/cros/recovery/internal/execs/wifirouter/controller"
	"infra/cros/recovery/internal/log"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/tlw"
)

func setStateExec(ctx context.Context, info *execs.ExecInfo) error {
	const stateArgName = "state"
	stateArgValue := info.GetActionArgs(ctx).AsInt(ctx, stateArgName, -1)
	if stateArgValue == -1 {
		return errors.Reason("missing required action argument %q", stateArgName).Err()
	}
	if _, ok := tlw.WifiRouterHost_State_name[int32(stateArgValue)]; !ok {
		return errors.Reason("action argument %q (%d) does not match a known state", stateArgName, stateArgValue).Err()
	}
	wifiRouterHost, err := activeHost(info)
	if err != nil {
		return err
	}
	wifiRouterHost.State = tlw.WifiRouterHost_State(stateArgValue)
	return nil
}

func updatePeripheralWifiStateExec(ctx context.Context, info *execs.ExecInfo) error {
	chromeos := info.GetChromeos()
	if chromeos == nil {
		return errors.Reason("update peripheral wifi state: chromeos is not present").Err()
	}
	routers := chromeos.GetWifiRouters()
	pws := tlw.ChromeOS_PERIPHERAL_WIFI_STATE_NOT_APPLICABLE
	if len(routers) > 0 {
		pws = tlw.ChromeOS_PERIPHERAL_WIFI_STATE_WORKING
		for _, routerHost := range chromeos.GetWifiRouters() {
			if routerHost.GetState() != tlw.WifiRouterHost_WORKING {
				pws = tlw.ChromeOS_PERIPHERAL_WIFI_STATE_BROKEN
				break
			}
		}
	}
	chromeos.PeripheralWifiState = pws
	return nil
}

// updateWifiRouterFeaturesExec updates the overall testbed WifiRouterFeatures
// to only include features that are common among all routers in the testbed.
func updateWifiRouterFeaturesExec(ctx context.Context, info *execs.ExecInfo) error {
	chromeos := info.GetChromeos()
	if chromeos == nil {
		return errors.Reason("update peripheral wifi state: chromeos is not present").Err()
	}
	commonFeatures := controller.CollectOverallTestbedWifiRouterFeatures(chromeos.WifiRouters)
	controller.SortWifiRouterFeaturesByName(commonFeatures)
	chromeos.WifiRouterFeatures = commonFeatures
	return nil
}

func identifyDeviceTypeExec(ctx context.Context, info *execs.ExecInfo) error {
	host, err := activeHost(info)
	if err != nil {
		return errors.Annotate(err, "failed to get active wifi router host").Err()
	}
	deviceType, err := controller.IdentifyRouterDeviceType(ctx, info.GetAccess(), info.GetActiveResource())
	if err != nil {
		return errors.Annotate(err, "failed to analyze device type of wifirouter").Err()
	}
	host.DeviceType = deviceType
	if host.DeviceType == labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_INVALID {
		return errors.Reason("router not identified as a valid device type").Err()
	}
	log.Debugf(ctx, "WifiRouter %q device type identified as %q", host.Name, host.DeviceType)
	return nil
}

func deviceTypeInListExec(ctx context.Context, info *execs.ExecInfo) error {
	host, err := activeHost(info)
	if err != nil {
		return errors.Annotate(err, "failed to get active wifi router host").Err()
	}
	const deviceTypesArgName = "device_types"
	actionArgs := info.GetActionArgs(ctx)
	if !actionArgs.Has(deviceTypesArgName) {
		return errors.Reason("missing required action argument %q", deviceTypesArgName).Err()
	}
	deviceTypesArgValue := actionArgs.AsStringSlice(ctx, deviceTypesArgName, nil)
	if len(deviceTypesArgValue) == 0 {
		return errors.Reason("action argument %q must not be empty", deviceTypesArgName).Err()
	}
	var deviceTypes []labapi.WifiRouterDeviceType
	hostHasMatchingDeviceType := false
	for i, deviceTypeIntStr := range deviceTypesArgValue {
		deviceTypeInt, err := strconv.Atoi(deviceTypeIntStr)
		if err != nil {
			errors.Annotate(err, "failed to parse action argument %q", deviceTypes)
		}
		if _, ok := labapi.WifiRouterDeviceType_name[int32(deviceTypeInt)]; !ok {
			return errors.Reason("action argument %q[%d] (%s) does not match a known device type", deviceTypesArgName, i, deviceTypeIntStr).Err()
		}
		dt := labapi.WifiRouterDeviceType(deviceTypeInt)
		deviceTypes = append(deviceTypes, dt)
		if host.DeviceType == dt {
			hostHasMatchingDeviceType = true
		}
	}
	if !hostHasMatchingDeviceType {
		var deviceTypeNames []string
		for _, deviceType := range deviceTypes {
			deviceTypeNames = append(deviceTypeNames, deviceType.String())
		}
		return errors.Reason("wifi router type %q does match any of the desired device types [%s]", host.DeviceType.String(), strings.Join(deviceTypeNames, ",")).Err()
	}
	return nil
}

// rebootDeviceExec reboots the router host. Can take up to 4 minutes, as it
// waits for the host to come back up before completing.
func rebootDeviceExec(ctx context.Context, info *execs.ExecInfo) error {
	c, err := activeHostRouterController(ctx, info)
	if err != nil {
		return err
	}
	if err := c.Reboot(ctx); err != nil {
		return errors.Annotate(err, "failed to reboot device").Err()
	}
	return nil
}

func updateModelAndFeaturesExec(ctx context.Context, info *execs.ExecInfo) error {
	c, err := activeHostRouterController(ctx, info)
	if err != nil {
		return err
	}

	model, err := c.Model()
	if err != nil {
		return errors.Annotate(err, "failed to get device model").Err()
	}
	features, err := c.Features()
	if err != nil {
		return errors.Annotate(err, "failed to get device features").Err()
	}
	features = controller.CleanWifiRouterFeatures(features)

	wifiRouterHost := c.WifiRouterHost()
	wifiRouterHost.Model = model
	wifiRouterHost.Features = features

	return nil
}

func init() {
	execs.Register("update_peripheral_wifi_state", updatePeripheralWifiStateExec)
	execs.Register("update_wifi_router_features", updateWifiRouterFeaturesExec)
	execs.Register("wifi_router_set_state", setStateExec)
	execs.Register("wifi_router_identify_device_type", identifyDeviceTypeExec)
	execs.Register("wifi_router_device_type_in_list", deviceTypeInListExec)
	execs.Register("wifi_router_update_model_and_features", updateModelAndFeaturesExec)
	execs.Register("wifi_router_reboot", rebootDeviceExec)
}
