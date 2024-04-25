// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testscenarios

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/genproto/googleapis/type/money"

	fleetcostModels "infra/cros/fleetcost/api/models"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver"
	"infra/cros/fleetcost/internal/costserver/testsupport"
	models "infra/unifiedfleet/api/v1/models"
	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// TestFallbackCostResult tests the flow where we fall back to the cost of a DUT.
func TestFallbackCostResult(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	tf.RegisterGetDeviceDataCall(gomock.Any(), &ufsAPI.GetDeviceDataResponse{
		Resource: &ufsAPI.GetDeviceDataResponse_ChromeOsDeviceData{
			ChromeOsDeviceData: &models.ChromeOSDeviceData{
				LabConfig: &models.MachineLSE{
					Lse: &models.MachineLSE_ChromeosMachineLse{
						ChromeosMachineLse: &models.ChromeOSMachineLSE{
							ChromeosLse: &models.ChromeOSMachineLSE_DeviceLse{
								DeviceLse: &models.ChromeOSDeviceLSE{
									Device: &models.ChromeOSDeviceLSE_Dut{
										Dut: &lab.DeviceUnderTest{
											Hostname: "fake-octopus-dut-1",
										},
									},
								},
							},
						},
					},
				},
				Machine: &models.Machine{
					Device: &models.Machine_ChromeosMachine{
						ChromeosMachine: &models.ChromeOSMachine{
							BuildTarget: "build-target",
							Model:       "model",
						},
					},
				},
			},
		},
		ResourceType: ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_CHROMEOS_DEVICE,
	})

	costserver.MustCreateCostIndicator(tf.Ctx, tf.Frontend, &fleetcostModels.CostIndicator{
		Type:     fleetcostModels.IndicatorType_INDICATOR_TYPE_DUT,
		Location: fleetcostModels.Location_LOCATION_ALL,
		Cost: &money.Money{
			CurrencyCode: "USD",
			Units:        1056,
		},
	})

	result, err := tf.Frontend.GetCostResult(tf.Ctx, &fleetcostAPI.GetCostResultRequest{
		Hostname: "fake-octopus-dut-1",
	})

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if cost := result.GetResult().GetDedicatedCost(); cost != 1056 {
		t.Errorf("unexpected cost: %f", cost)
	}
}
