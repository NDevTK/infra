// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"

	"go.chromium.org/luci/common/errors"

	fleetcostpb "infra/cros/fleetcost/api/models"
	"infra/cros/fleetcost/internal/costserver/entities"
	ufsFetcher "infra/cros/fleetcost/internal/costserver/inventory/ufs"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// IndicatorAttribute is the information that's necessary to look up a datastore record.
//
// TODO(gregorynisbet): Remove this type. It duplicates the functionality of the datastore entity and protos.
type IndicatorAttribute struct {
	IndicatorType fleetcostpb.IndicatorType
	Board         string
	Model         string
	Sku           string
	Location      fleetcostpb.Location
}

// AsEntity converts an IndicatorAttribute to a datastore Entity.
func (attribute *IndicatorAttribute) AsEntity() *entities.CostIndicatorEntity {
	if attribute == nil {
		return nil
	}
	return &entities.CostIndicatorEntity{
		CostIndicator: &fleetcostpb.CostIndicator{
			Type:     attribute.IndicatorType,
			Board:    attribute.Board,
			Model:    attribute.Model,
			Sku:      attribute.Sku,
			Location: attribute.Location,
		},
	}
}

// CalculateCostForOsResource calculates the cost for an OS resource.
//
// So far, only ChromeOS devices are supported.
func CalculateCostForOsResource(ctx context.Context, ic ufsAPI.FleetClient, hostname string) (*fleetcostpb.CostResult, error) {
	res, err := ic.GetDeviceData(ctx, &ufsAPI.GetDeviceDataRequest{Hostname: hostname})
	if err != nil {
		return nil, errors.Annotate(err, "calculate cost for os resource").Err()
	}
	switch res.GetResourceType() {
	case ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_CHROMEOS_DEVICE:
		resp, err := CalculateCostForSingleChromeosDut(ctx, ic, res.GetChromeOsDeviceData())
		return resp, errors.Annotate(err, "calculate ChromeOS device cost").Err()
	case ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_ATTACHED_DEVICE:
		return nil, errors.Reason("%s is an attached device, support is not implemented yet.", hostname).Err()
	case ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_SCHEDULING_UNIT:
		return nil, errors.Reason("%s is an scheduling unit, support is not implemented yet.", hostname).Err()
	default:
		return nil, errors.Reason("Cannot find a valid resource type for %s: %s", hostname, res.GetResourceType()).Err()
	}
}

// CalculateCostForSingleChromeosDut calculates the cost of a ChromeOS DUT.
func CalculateCostForSingleChromeosDut(ctx context.Context, ic ufsAPI.FleetClient, data *ufspb.ChromeOSDeviceData) (*fleetcostpb.CostResult, error) {
	dut := data.GetLabConfig().GetChromeosMachineLse().GetDeviceLse().GetDut()
	peripherals := dut.GetPeripherals()
	servo := peripherals.GetServo()
	// TODO: add a map that convert UFS location to cost indicator location. Hardcode to all for now.
	location := fleetcostpb.Location_LOCATION_ALL
	if dut == nil {
		return nil, errors.Reason("%s is not a valid ChromeOS DUT", data.GetLabConfig().GetHostname()).Err()
	}
	var dedicateCost, sharedCost, cloudCost float64
	// Cost for DUT hardware.
	dutCost, err := GetDutHardwareCost(ctx, data.GetMachine().GetChromeosMachine(), location)
	if err != nil {
		return nil, errors.Annotate(err, "calculate cost for single chromeos dut").Err()
	}
	dedicateCost = dedicateCost + dutCost
	// Cost for servo related items.
	if servo.GetServoHostname() != "" {
		servoCost, err := GetServoCost(ctx, servo.GetServoType(), location)
		if err != nil {
			return nil, errors.Annotate(err, "calculate cost for single chromeos dut").Err()
		}
		dedicateCost = dedicateCost + float64(servoCost)
		labstationCost, err := getLabstationCost(ctx, ic, servo.GetServoHostname(), location)
		sharedCost = sharedCost + labstationCost
		if err != nil {
			return nil, errors.Annotate(err, "calculate cost for single chromeos dut").Err()
		}
	}
	return &fleetcostpb.CostResult{
		DedicatedCost:    dedicateCost,
		SharedCost:       sharedCost,
		CloudServiceCost: cloudCost,
	}, nil
}

func GetServoCost(ctx context.Context, servoType string, location fleetcostpb.Location) (float64, error) {
	v, err := GetCostIndicatorValue(ctx, IndicatorAttribute{
		IndicatorType: fleetcostpb.IndicatorType_INDICATOR_TYPE_SERVO,
		Board:         servoType,
		Location:      location,
	})
	if err != nil {
		return 0, errors.Annotate(err, "get servo cost").Err()
	}
	return v, nil
}

// GetDutHardwareCost gets the hardware cost for a single DUT.
func GetDutHardwareCost(ctx context.Context, m *ufspb.ChromeOSMachine, location fleetcostpb.Location) (float64, error) {
	v, err := GetCostIndicatorValue(ctx, IndicatorAttribute{
		IndicatorType: fleetcostpb.IndicatorType_INDICATOR_TYPE_DUT,
		Board:         m.GetBuildTarget(),
		Model:         m.GetModel(),
		Sku:           m.GetSku(),
		Location:      location,
	})
	if err != nil {
		return 0, errors.Annotate(err, "get dut hardware cost").Err()
	}
	return v, nil
}

func getLabstationCost(ctx context.Context, ic ufsAPI.FleetClient, hostname string, location fleetcostpb.Location) (float64, error) {
	data, err := ufsFetcher.GetChromeosDeviceData(ctx, ic, hostname)
	if err != nil {
		return 0, errors.Annotate(err, "get labstation cost").Err()
	}
	m := data.GetMachine().GetChromeosMachine()
	v, err := GetCostIndicatorValue(ctx, IndicatorAttribute{
		IndicatorType: fleetcostpb.IndicatorType_INDICATOR_TYPE_LABSTATION,
		Board:         m.GetBuildTarget(),
		Model:         m.GetModel(),
		Sku:           m.GetSku(),
		Location:      location,
	})
	if err != nil {
		return 0, errors.Annotate(err, "get labstation cost").Err()
	}
	labMap, err := ufsFetcher.GetLabstationDutMapping(ctx, ic, []string{hostname})
	if err != nil {
		return 0, errors.Annotate(err, "get labstation cost").Err()
	}
	if l, ok := labMap[hostname]; ok {
		if len(l) > 0 {
			return v / float64(len(l)), nil
		}
	}
	return 0, errors.Reason("Unable to get number of DUTs under %s", hostname).Err()
}
