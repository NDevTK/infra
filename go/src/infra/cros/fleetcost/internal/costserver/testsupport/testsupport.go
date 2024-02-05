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
	out.Ctx = memory.Use(ctx)
	out.Frontend = costserver.NewFleetCostFrontend().(*costserver.FleetCostFrontend)
	out.MockUFS = mockufs.NewMockFleetClient(mc)
	costserver.SetUFSClient(out.Frontend, out.MockUFS)
	return &out
}
