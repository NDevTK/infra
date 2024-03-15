// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package testsupport provides the `NewFixture` function, which produces
// a context and a frontend that's suitable for testing.
package testsupport

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"

	fleetcostpb "infra/cros/fleetcost/api/models"
	"infra/cros/fleetcost/internal/costserver"
	"infra/cros/fleetcost/internal/costserver/controller"
	"infra/cros/fleetcost/internal/costserver/models"
	mockufs "infra/unifiedfleet/api/v1/rpc/mock"
)

type Fixture struct {
	Ctx      context.Context
	Frontend *costserver.FleetCostFrontend
	MockUFS  *mockufs.MockFleetClient
}

// NewFixture creates a basic fixture with fake versions of datastore and UFS with properties
// that are convenient for unit tests.
func NewFixture(ctx context.Context, t *testing.T) *Fixture {
	mc := gomock.NewController(t)
	var out Fixture
	out.Ctx = memory.Use(ctx)
	datastore.GetTestable(out.Ctx).Consistent(true)
	out.Frontend = costserver.NewFleetCostFrontend().(*costserver.FleetCostFrontend)
	out.MockUFS = mockufs.NewMockFleetClient(mc)
	costserver.SetUFSClient(out.Frontend, out.MockUFS)
	return &out
}

// NewFixtureWithData returns a fixture with test data.
func NewFixtureWithData(ctx context.Context, t *testing.T) *Fixture {
	tf := NewFixture(ctx, t)
	err := controller.PutCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
		ID: "fake-cost-indicator",
		CostIndicator: &fleetcostpb.CostIndicator{
			Name:  "a",
			Board: "e",
		},
	})
	if err != nil {
		panic(err)
	}
	return tf
}
