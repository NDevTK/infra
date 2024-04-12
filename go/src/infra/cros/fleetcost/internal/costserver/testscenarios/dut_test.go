// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testscenarios

import (
	"context"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/genproto/googleapis/type/money"

	fleetcostpb "infra/cros/fleetcost/api/models"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver/testsupport"
	models "infra/unifiedfleet/api/v1/models"
	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// TestNonexistentDUT tests trying to get the cost of a DUT that doesn't exist.
func TestNonexistentDUT(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	response := &ufsAPI.GetDeviceDataResponse{}

	tf.MockUFS.EXPECT().GetDeviceData(gomock.Any(), gomock.Any()).Return(response, nil)

	_, err := tf.Frontend.GetCostResult(tf.Ctx, &fleetcostAPI.GetCostResultRequest{
		Hostname: "fake-octopus-dut-1",
	})
	// TODO(gregorynisbet): Make this test a little smarter and look at the
	//                      GCP-reported error status too.
	if ok := (err != nil) && strings.Contains(err.Error(), "find a valid resource type"); !ok {
		t.Errorf("unexpected error: %s", err)
	}
}

// TestDUTWithNoPeripherals tests a device with no peripherals.
func TestDUTWithNoPeripherals(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	if _, err := tf.Frontend.CreateCostIndicator(tf.Ctx, &fleetcostAPI.CreateCostIndicatorRequest{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board:    "build-target",
			Model:    "model",
			Type:     fleetcostpb.IndicatorType_INDICATOR_TYPE_LABSTATION,
			Location: fleetcostpb.Location_LOCATION_ALL,
			Cost: &money.Money{
				Units: 100,
			},
		},
	}); err != nil {
		panic(err)
	}

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

	_, err := tf.Frontend.GetCostResult(tf.Ctx, &fleetcostAPI.GetCostResultRequest{
		Hostname: "fake-octopus-dut-1",
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
