// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"fmt"

	"infra/cros/dutstate"
	"infra/libs/skylab/inventory/swarming"
	"infra/libs/skylab/inventory/swarming/attacheddevice"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// GetDUTBotDims gets all Swarming dimensions of a DUT bot.
func GetDUTBotDims(ctx context.Context, r swarming.ReportFunc, ds dutstate.Info, deviceData *ufspb.ChromeOSDeviceData) swarming.Dimensions {
	c := deviceData.GetDutV1().GetCommon()
	dims := swarming.Convert(c.GetLabels())
	dims["dut_id"] = []string{c.GetId()}
	dims["dut_name"] = []string{c.GetHostname()}
	if v := c.GetHwid(); v != "" {
		dims["hwid"] = []string{v}
	}
	if v := c.GetSerialNumber(); v != "" {
		dims["serial_number"] = []string{v}
	}
	if v := c.GetLocation(); v != nil {
		location := fmt.Sprintf("%s-row%d-rack%d-host%d",
			v.GetLab().GetName(), v.GetRow(), v.GetRack(), v.GetHost())
		dims["location"] = []string{location}
	}
	dims["dut_state"] = []string{string(ds.State)}
	dims["ufs_zone"] = []string{deviceData.GetLabConfig().GetZone()}

	// Only expose the label when there is a valid value for logical zone.
	if deviceData.GetLabConfig().GetLogicalZone() != ufspb.LogicalZone_LOGICAL_ZONE_UNSPECIFIED {
		dims["logical_zone"] = []string{deviceData.GetLabConfig().GetLogicalZone().String()}
	}
	swarming.Sanitize(dims, r)
	return dims
}

// GetAttachedDeviceBotDims gets all Swarming dimensions of an AttachedDevice bot.
func GetAttachedDeviceBotDims(ctx context.Context, r swarming.ReportFunc, ds dutstate.Info, deviceData *ufsAPI.AttachedDeviceData) swarming.Dimensions {
	dims := attacheddevice.Convert(deviceData)
	if v := deviceData.GetMachine().GetLocation(); v != nil {
		location := fmt.Sprintf("%s-aisle%s-row%s-rack%s-racknumber%s-shelf%s-position%s",
			v.GetAisle(), v.GetZone(), v.GetRow(), v.GetRack(), v.GetRackNumber(), v.GetShelf(), v.GetPosition())
		dims["location"] = []string{location}
	}
	dims["dut_state"] = []string{string(ds.State)}
	dims["ufs_zone"] = []string{deviceData.GetLabConfig().GetZone()}
	swarming.Sanitize(dims, r)
	return dims
}
