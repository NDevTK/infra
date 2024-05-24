// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/genproto/googleapis/type/money"

	fleetcostModels "infra/cros/fleetcost/api/models"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver"
	"infra/cros/fleetcost/internal/costserver/fakeufsdata"
	testsupport "infra/cros/fleetcost/internal/costserver/testsupport"
	models "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

func TestRepopulateCache(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	costserver.MustCreateCostIndicator(tf.Ctx, tf.Frontend, &fleetcostModels.CostIndicator{
		Type:     fleetcostModels.IndicatorType_INDICATOR_TYPE_DUT,
		Board:    "build-target",
		Model:    "model",
		Sku:      "",
		Location: fleetcostModels.Location_LOCATION_ALL,
		Cost: &money.Money{
			CurrencyCode: "USD",
			Units:        134,
		},
	})

	tf.RegisterListMachineLSEs(gomock.Any(), &ufsAPI.ListMachineLSEsResponse{
		MachineLSEs: []*models.MachineLSE{
			{
				Hostname: "fake-octopus-dut-1",
			},
		},
	})

	tf.RegisterGetDeviceDataCall(gomock.Any(), fakeufsdata.FakeOctopusDUTDeviceDataResponse)

	_, err := tf.Frontend.RepopulateCache(tf.Ctx, &fleetcostAPI.RepopulateCacheRequest{})

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
