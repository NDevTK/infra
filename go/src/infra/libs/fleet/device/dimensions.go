// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package device

import (
	"context"
	"fmt"

	"infra/cros/dutstate"
	"infra/libs/fleet/device/attacheddevice"
	"infra/libs/fleet/device/dut"
	"infra/libs/fleet/device/schedulingunit"
	"infra/libs/skylab/inventory/swarming"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// GetOSResourceDims gets the dimensions of a fleet resource to present to the
// scheduling layer.
func GetOSResourceDims(ctx context.Context, client ufsAPI.FleetClient, r swarming.ReportFunc, name string) (swarming.Dimensions, error) {
	deviceData, err := client.GetDeviceData(ctx, &ufsAPI.GetDeviceDataRequest{
		Hostname: name,
	})
	if err != nil {
		return nil, err
	}

	if deviceData.GetResourceType() == ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_SCHEDULING_UNIT {
		return getSUDeviceDims(ctx, client, r, deviceData.GetSchedulingUnit())
	}
	deviceDimensions, err := getBaseResourceDims(ctx, client, r, deviceData)
	if err != nil {
		return nil, err
	}
	return deviceDimensions, nil
}

// getSUDeviceDims gets the dimensions of a SchedulingUnit.
//
// Because Scheduling Units are comprised of multiple devices, this function
// also combines and joins all the dimensions of the other devices.
func getSUDeviceDims(ctx context.Context, client ufsAPI.FleetClient, r swarming.ReportFunc, su *ufspb.SchedulingUnit) (swarming.Dimensions, error) {
	var dutsDims []swarming.Dimensions
	for _, hostname := range su.GetMachineLSEs() {
		deviceData, err := client.GetDeviceData(ctx, &ufsAPI.GetDeviceDataRequest{
			Hostname: hostname,
		})
		if err != nil {
			return nil, err
		}
		deviceDimensions, err := getBaseResourceDims(ctx, client, r, deviceData)
		if err != nil {
			return nil, err
		}
		dutsDims = append(dutsDims, deviceDimensions)
	}
	return schedulingunit.GetSchedulingUnitDimensions(su, dutsDims), nil
}

// getBaseResourceDims gets the dimensions of a base device type (ChromeOS
// Device or Attached Device)
func getBaseResourceDims(ctx context.Context, client ufsAPI.FleetClient, r swarming.ReportFunc, deviceData *ufsAPI.GetDeviceDataResponse) (swarming.Dimensions, error) {
	switch deviceData.GetResourceType() {
	case ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_CHROMEOS_DEVICE:
		dutState := dutstate.Read(ctx, client, deviceData.GetChromeOsDeviceData().GetLabConfig().GetName())
		return dut.GetDUTBotDims(ctx, r, dutState, deviceData.GetChromeOsDeviceData()), nil
	case ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_ATTACHED_DEVICE:
		dutState := dutstate.Read(ctx, client, deviceData.GetAttachedDeviceData().GetLabConfig().GetName())
		return attacheddevice.GetAttachedDeviceBotDims(ctx, r, dutState, deviceData.GetAttachedDeviceData()), nil
	}
	return nil, fmt.Errorf("getBaseResourceDims: invalid device type (%s)", deviceData.GetResourceType())
}
