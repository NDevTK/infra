// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	"go.chromium.org/luci/common/errors"

	fleetcostpb "infra/cros/fleetcost/api"
	"infra/cros/fleetcost/internal/costserver/controller"
)

// CreateCostIndicator creates a cost indicator.
func (f *FleetCostFrontend) CreateCostIndicator(_ context.Context, _ *fleetcostpb.CreateCostIndicatorRequest) (*fleetcostpb.CreateCostIndicatorResponse, error) {
	return nil, errors.Reason("not yet implemented").Err()
}

// ListCostIndicators lists the cost indicators in the database satisfying the request.
func (f *FleetCostFrontend) ListCostIndicators(ctx context.Context, request *fleetcostpb.ListCostIndicatorsRequest) (*fleetcostpb.ListCostIndicatorsResponse, error) {
	out, err := controller.ListCostIndicators(ctx, 0)
	if err != nil {
		return nil, err
	}
	return &fleetcostpb.ListCostIndicatorsResponse{
		CostIndicator: out,
	}, nil
}
