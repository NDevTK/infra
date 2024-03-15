// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver/controller"
	"infra/cros/fleetcost/internal/costserver/models"
)

// CreateCostIndicator creates a cost indicator.
func (f *FleetCostFrontend) CreateCostIndicator(ctx context.Context, request *fleetcostAPI.CreateCostIndicatorRequest) (*fleetcostAPI.CreateCostIndicatorResponse, error) {
	// TODO(gregorynisbet): Do some kind of input validation here.
	costIndicator := request.GetCostIndicator()
	entity := models.NewCostIndicatorEntity(costIndicator)
	if err := controller.PutCostIndicatorEntity(ctx, entity); err != nil {
		return nil, err
	}
	return &fleetcostAPI.CreateCostIndicatorResponse{
		CostIndicator: costIndicator,
	}, nil
}

// ListCostIndicators lists the cost indicators in the database satisfying the request.
func (f *FleetCostFrontend) ListCostIndicators(ctx context.Context, request *fleetcostAPI.ListCostIndicatorsRequest) (*fleetcostAPI.ListCostIndicatorsResponse, error) {
	out, err := controller.ListCostIndicators(ctx, 0)
	if err != nil {
		return nil, err
	}
	return &fleetcostAPI.ListCostIndicatorsResponse{
		CostIndicator: out,
	}, nil
}

// UpdateCostIndicator updates a CostIndicator.
func (f *FleetCostFrontend) UpdateCostIndicator(ctx context.Context, request *fleetcostAPI.UpdateCostIndicatorRequest) (*fleetcostAPI.UpdateCostIndicatorResponse, error) {
	entity := models.NewCostIndicatorEntity(request.GetCostIndicator())
	out, err := controller.UpdateCostIndicatorEntity(ctx, entity, request.GetUpdateMask().GetPaths())
	if err != nil {
		return nil, err
	}
	return &fleetcostAPI.UpdateCostIndicatorResponse{
		CostIndicator: out.CostIndicator,
	}, nil
}
