// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	mockufs "infra/unifiedfleet/api/v1/rpc/mock"
)

type fixture struct {
	ctx      context.Context
	frontend *FleetCostFrontend
	mockUFS  *mockufs.MockFleetClient
}

func newFixture(ctx context.Context, t *testing.T) *fixture {
	mc := gomock.NewController(t)
	var out fixture
	out.ctx = ctx
	out.frontend = NewFleetCostFrontend().(*FleetCostFrontend)
	out.mockUFS = mockufs.NewMockFleetClient(mc)
	SetUFSClient(out.frontend, out.mockUFS)
	return &out
}
