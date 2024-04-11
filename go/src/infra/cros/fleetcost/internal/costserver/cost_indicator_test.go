// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver_test

import (
	"context"
	"testing"

	"google.golang.org/genproto/googleapis/type/money"

	models "infra/cros/fleetcost/api/models"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	testsupport "infra/cros/fleetcost/internal/costserver/testsupport"
)

func TestCreateCostIndicator(t *testing.T) {
	t.Parallel()
	tf := testsupport.NewFixture(context.Background(), t)

	_, err := tf.Frontend.CreateCostIndicator(tf.Ctx, &fleetcostAPI.CreateCostIndicatorRequest{
		CostIndicator: &models.CostIndicator{
			Board: "board",
			Model: "model",
			Cost: &money.Money{
				CurrencyCode: "USD",
				Units:        12,
			},
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
