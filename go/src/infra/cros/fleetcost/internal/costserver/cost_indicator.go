// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/gae/service/datastore"

	fleetcostModels "infra/cros/fleetcost/api/models"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver/models"
	"infra/cros/fleetcost/internal/utils"
	"infra/cros/fleetcost/internal/validation"
)

// MustCreateCostIndicator is a helper function for tests that ergonomically creates a
// CostIndicator so that the database can be brought to a known state.
func MustCreateCostIndicator(ctx context.Context, f *FleetCostFrontend, costIndicator *fleetcostModels.CostIndicator) {
	_, err := f.CreateCostIndicator(ctx, &fleetcostAPI.CreateCostIndicatorRequest{
		CostIndicator: costIndicator,
	})
	if err != nil {
		panic(err)
	}
}

// CreateCostIndicator creates a cost indicator.
func (f *FleetCostFrontend) CreateCostIndicator(ctx context.Context, request *fleetcostAPI.CreateCostIndicatorRequest) (*fleetcostAPI.CreateCostIndicatorResponse, error) {
	if err := validation.ValidateCreateCostIndicatorRequest(request); err != nil {
		return nil, err
	}
	costIndicator := request.GetCostIndicator()
	entity := models.NewCostIndicatorEntity(costIndicator)
	if err := utils.InsertOneWithoutReplacement(ctx, true, entity, nil); err != nil {
		return nil, errors.Annotate(err, "create cost indicator").Err()
	}
	return &fleetcostAPI.CreateCostIndicatorResponse{
		CostIndicator: costIndicator,
	}, nil
}

// ListCostIndicators lists the cost indicators in the database satisfying the request.
func (f *FleetCostFrontend) ListCostIndicators(ctx context.Context, request *fleetcostAPI.ListCostIndicatorsRequest) (*fleetcostAPI.ListCostIndicatorsResponse, error) {
	out, err := models.ListCostIndicators(ctx, 0, request.GetFilter())
	if err != nil {
		return nil, errors.Annotate(err, "list cost indicators").Err()
	}
	return &fleetcostAPI.ListCostIndicatorsResponse{
		CostIndicator: out,
	}, nil
}

// UpdateCostIndicator updates a CostIndicator.
func (f *FleetCostFrontend) UpdateCostIndicator(ctx context.Context, request *fleetcostAPI.UpdateCostIndicatorRequest) (*fleetcostAPI.UpdateCostIndicatorResponse, error) {
	entity := models.NewCostIndicatorEntity(request.GetCostIndicator())
	out, err := models.UpdateCostIndicatorEntity(ctx, entity, request.GetUpdateMask().GetPaths())
	if err != nil {
		return nil, errors.Annotate(err, "update cost indicator").Err()
	}
	return &fleetcostAPI.UpdateCostIndicatorResponse{
		CostIndicator: out.CostIndicator,
	}, nil
}

// DeleteCostIndicator deletes a CostIndicator.
func (f *FleetCostFrontend) DeleteCostIndicator(ctx context.Context, request *fleetcostAPI.DeleteCostIndicatorRequest) (*fleetcostAPI.DeleteCostIndicatorResponse, error) {
	entity := models.NewCostIndicatorEntity(request.GetCostIndicator())
	if err := datastore.Delete(ctx, entity); err != nil {
		return nil, errors.Annotate(err, "delete cost indicator").Err()
	}
	return &fleetcostAPI.DeleteCostIndicatorResponse{
		CostIndicator: entity.CostIndicator,
	}, nil
}
