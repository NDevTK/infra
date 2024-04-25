// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"math"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/gae/service/datastore"

	fleetcostpb "infra/cros/fleetcost/api/models"
	"infra/cros/fleetcost/internal/costserver/entities"
	"infra/cros/fleetcost/internal/utils"
)

// GetCostIndicatorValue gets the value of a cost indicator, potentially falling back/
func GetCostIndicatorValue(ctx context.Context, attribute *IndicatorAttribute, usefallbacks bool) (float64, error) {
	if !usefallbacks {
		return GetCostIndicatorValueDirectly(ctx, attribute)
	}
	sequence, err := GetIndicatorFallbacks(attribute)
	if err != nil {
		return math.NaN(), err
	}
	for _, attribute := range sequence {
		result, err := GetCostIndicatorValueDirectly(ctx, attribute)
		switch {
		case err == nil:
			return result, nil
		case datastore.IsErrNoSuchEntity(err):
			continue
		default:
			return math.NaN(), err
		}

	}
	return math.NaN(), datastore.ErrNoSuchEntity
}

// GetCostIndicatorValueDirectly gets the value of a cost indicator.
func GetCostIndicatorValueDirectly(ctx context.Context, attribute *IndicatorAttribute) (float64, error) {
	entity := attribute.AsEntity()
	if _, err := entities.GetCostIndicatorEntity(ctx, entity); err != nil {
		return 0, errors.Annotate(err, "get cost indicator value").Err()
	}
	return utils.MoneyToFloat(entity.CostIndicator.GetCost()), nil
}

// GetIndicatorFallbacks takes an indicatorAttribute and returns the list of fallback indicator attributes.
func GetIndicatorFallbacks(attribute *IndicatorAttribute) ([]*IndicatorAttribute, error) {
	typ := attribute.IndicatorType
	board := attribute.Board
	model := attribute.Model
	sku := attribute.Sku
	location := attribute.Location

	if location == fleetcostpb.Location_LOCATION_UNKNOWN {
		return nil, errors.New("location cannot be unknown")
	}
	if typ == fleetcostpb.IndicatorType_INDICATOR_TYPE_UNKNOWN {
		return nil, errors.New("type cannot be unknown")
	}

	hasLocationAll := location == fleetcostpb.Location_LOCATION_ALL

	var output []*IndicatorAttribute

	// TODO(gregorynisbet): rework this logic so that it isn't hardcoded.
	if sku != "" {
		output = append(output, NewIndicatorAttribute(typ, board, model, sku, location))
	}
	if sku != "" && !hasLocationAll {
		output = append(output, NewIndicatorAttribute(typ, board, model, sku, fleetcostpb.Location_LOCATION_ALL))
	}
	if model != "" {
		output = append(output, NewIndicatorAttribute(typ, board, model, "", location))
	}
	if model != "" && !hasLocationAll {
		output = append(output, NewIndicatorAttribute(typ, board, model, "", fleetcostpb.Location_LOCATION_ALL))
	}
	if board != "" {
		output = append(output, NewIndicatorAttribute(typ, board, "", "", location))
	}
	if board != "" && !hasLocationAll {
		output = append(output, NewIndicatorAttribute(typ, board, "", "", fleetcostpb.Location_LOCATION_ALL))
	}
	if typ != fleetcostpb.IndicatorType_INDICATOR_TYPE_UNKNOWN {
		output = append(output, NewIndicatorAttribute(typ, "", "", "", location))
	}
	if typ != fleetcostpb.IndicatorType_INDICATOR_TYPE_UNKNOWN && !hasLocationAll {
		output = append(output, NewIndicatorAttribute(typ, "", "", "", fleetcostpb.Location_LOCATION_ALL))
	}

	return output, nil
}
