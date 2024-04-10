// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/fleetcost/internal/costserver/models"
	"infra/cros/fleetcost/internal/utils"
)

// GetCostIndicatorValue gets the value of a cost indicator.
func GetCostIndicatorValue(ctx context.Context, attribute IndicatorAttribute) (float64, error) {
	entity := attribute.AsEntity()
	if _, err := models.GetCostIndicatorEntity(ctx, entity); err != nil {
		return 0, errors.Annotate(err, "get cost indicator value").Err()
	}
	return utils.MoneyToFloat(entity.CostIndicator.GetCost()), nil
}
