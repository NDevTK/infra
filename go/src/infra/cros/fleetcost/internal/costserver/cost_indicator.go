// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"
	"errors"

	fleetcostpb "infra/cros/fleetcost/api"
	"infra/cros/fleetcost/internal/costserver/controller"
	"infra/cros/fleetcost/internal/costserver/models"
)

// CreateCostIndicator creates a cost indicator.
func (f *FleetCostFrontend) CreateCostIndicator(ctx context.Context, request *fleetcostpb.CreateCostIndicatorRequest) (*fleetcostpb.CreateCostIndicatorResponse, error) {
	// TODO(gregorynisbet): Do some kind of input validation here.
	costIndicator := request.GetCostIndicator()
	entity := models.NewCostIndicator(costIndicator)
	if err := controller.PutCostIndicator(ctx, entity); err != nil {
		return nil, err
	}
	return &fleetcostpb.CreateCostIndicatorResponse{
		CostIndicator: costIndicator,
	}, nil
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

// UpdateCostIndicator updates a CostIndicator.
func (f *FleetCostFrontend) UpdateCostIndicator(ctx context.Context, request *fleetcostpb.UpdateCostIndicatorRequest) (*fleetcostpb.UpdateCostIndicatorResponse, error) {
	return nil, errors.New("not yet implemented")
}
