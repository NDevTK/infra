// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testsupport

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"infra/cros/fleetcost/internal/costserver"
	mockufs "infra/unifiedfleet/api/v1/rpc/mock"
)

type Fixture struct {
	Ctx      context.Context
	Frontend *costserver.FleetCostFrontend
	MockUFS  *mockufs.MockFleetClient
}

func NewFixture(ctx context.Context, t *testing.T) *Fixture {
	mc := gomock.NewController(t)
	var out Fixture
	out.Ctx = ctx
	out.Frontend = costserver.NewFleetCostFrontend().(*costserver.FleetCostFrontend)
	out.MockUFS = mockufs.NewMockFleetClient(mc)
	costserver.SetUFSClient(out.Frontend, out.MockUFS)
	return &out
}
