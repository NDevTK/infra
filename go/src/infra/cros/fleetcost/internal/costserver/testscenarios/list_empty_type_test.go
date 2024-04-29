// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testscenarios

import (
	"context"
	"testing"

	"google.golang.org/genproto/googleapis/type/money"

	fleetcostModels "infra/cros/fleetcost/api/models"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver"
	"infra/cros/fleetcost/internal/costserver/testsupport"
)

// TestListEmptyType tests that using an empty type correctly imposes
// no constraints on the location field when listing cost indicators.
//
// At time of writing, an empty string is what the command line tool uses to
// signal that we are not looking for cost records by type.
func TestListEmptyType(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	costserver.MustCreateCostIndicator(tf.Ctx, tf.Frontend, &fleetcostModels.CostIndicator{
		Type:     fleetcostModels.IndicatorType_INDICATOR_TYPE_DUT,
		Location: fleetcostModels.Location_LOCATION_SFO36,
		Board:    "octopus",
		Cost: &money.Money{
			CurrencyCode: "USD",
			Units:        3456789,
		},
	})

	costserver.MustCreateCostIndicator(tf.Ctx, tf.Frontend, &fleetcostModels.CostIndicator{
		Type:     fleetcostModels.IndicatorType_INDICATOR_TYPE_DUT,
		Location: fleetcostModels.Location_LOCATION_SFO36,
		Board:    "not-octopus",
		Cost: &money.Money{
			CurrencyCode: "USD",
			Units:        70,
		},
	})

	resp, err := tf.Frontend.ListCostIndicators(tf.Ctx, &fleetcostAPI.ListCostIndicatorsRequest{
		Filter: &fleetcostAPI.ListCostIndicatorsFilter{
			Location: fleetcostModels.Location_LOCATION_SFO36.String(),
			Type:     "",
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if n := len(resp.GetCostIndicator()); n != 2 {
		t.Errorf("wrong number of responses %d != 2", n)
	}
}
