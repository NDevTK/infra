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

// PutCostIndicator puts a cost indicator into the database.
func PutCostIndicator(ctx context.Context, costIndicator *models.CostIndicator) error {
	if costIndicator == nil {
		return errors.New("costIndicator cannot be nil")
	}
	return datastore.Put(ctx, costIndicator)
}

// GetCostIndicator extracts a cost indicator from the database.
func GetCostIndicator(ctx context.Context, costIndicator *models.CostIndicator) (*models.CostIndicator, error) {
	if err := datastore.Get(ctx, costIndicator); err != nil {
		return nil, err
	}
	return costIndicator, nil
}

// ListCostIndicators lists the cost indicators in the database, up to a limit (not yet implemented).
func ListCostIndicators(ctx context.Context, limit int) ([]*fleetcostpb.CostIndicator, error) {
	var out []*fleetcostpb.CostIndicator
	query := datastore.NewQuery(models.CostIndicatorKind)
	if err := datastore.Run(ctx, query, func(entity *models.CostIndicator) {
		out = append(out, entity.CostIndicator)
	}); err != nil {
		return nil, err
	}
	return out, nil
}
