// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"

	"go.chromium.org/luci/gae/service/datastore"

	"infra/cros/fleetcost/internal/costserver/models"
)

// PutCostIndicator puts a cost indicator into the database.
func PutCostIndicator(ctx context.Context, costIndicator *models.CostIndicator) error {
	return datastore.Put(ctx, costIndicator)
}

// GetCostIndicator extracts a cost indicator from the database.
func GetCostIndicator(ctx context.Context, costIndicator *models.CostIndicator) (*models.CostIndicator, error) {
	if err := datastore.Get(ctx, costIndicator); err != nil {
		return nil, err
	}
	return costIndicator, nil
}
