// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"errors"

	"go.chromium.org/luci/gae/service/datastore"

	fleetcostpb "infra/cros/fleetcost/api"
	"infra/cros/fleetcost/internal/costserver/models"
)

// PutCostIndicatorEntity puts a cost indicator entity into the database.
func PutCostIndicatorEntity(ctx context.Context, entity *models.CostIndicatorEntity) error {
	if entity == nil {
		return errors.New("cost indicator entity cannot be nil")
	}
	return datastore.Put(ctx, entity)
}

// GetCostIndicatorEntity extracts a cost indicator from the database.
func GetCostIndicatorEntity(ctx context.Context, entity *models.CostIndicatorEntity) (*models.CostIndicatorEntity, error) {
	if err := datastore.Get(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

// ListCostIndicators lists the cost indicators in the database, up to a limit (not yet implemented).
func ListCostIndicators(ctx context.Context, limit int) ([]*fleetcostpb.CostIndicator, error) {
	var out []*fleetcostpb.CostIndicator
	query := datastore.NewQuery(models.CostIndicatorKind)
	if err := datastore.Run(ctx, query, func(entity *models.CostIndicatorEntity) {
		out = append(out, entity.CostIndicator)
	}); err != nil {
		return nil, err
	}
	return out, nil
}
