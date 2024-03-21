// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package models

import (
	"context"

	"google.golang.org/protobuf/proto"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/gae/service/datastore"

	fleetcostpb "infra/cros/fleetcost/api/models"
	"infra/cros/fleetcost/internal/costserver/maskutils"
)

const CostIndicatorKind = "CostIndicatorKind"

type CostIndicatorEntity struct {
	_kind         string                     `gae:"$kind,CostIndicatorKind"`
	ID            string                     `gae:"$id"`
	Extra         datastore.PropertyMap      `gae:",extra"`
	CostIndicator *fleetcostpb.CostIndicator `gae:"cost_indicator"`
}

// Silence staticcheck warning about unused field.
var _ = CostIndicatorEntity{}._kind

// Clone produces a deep copy of a cost indicator.
//
// This method intentionally takes a non-pointer receiver to perform a
// shallow copy, and then replaces a field to perform a deep copy.
//
// I don't actually know whether I also need to copy the datastore.PropertyMap.
func (indicator CostIndicatorEntity) Clone() *CostIndicatorEntity {
	indicator.CostIndicator = proto.Clone(indicator.CostIndicator).(*fleetcostpb.CostIndicator)
	return &indicator
}

// NewCostIndicatorEntity makes a cost indicator entity from an object extracted from a request.
func NewCostIndicatorEntity(costIndicator *fleetcostpb.CostIndicator) *CostIndicatorEntity {
	return &CostIndicatorEntity{
		ID:            costIndicator.GetName(),
		CostIndicator: costIndicator,
	}
}

// PutCostIndicatorEntity puts a cost indicator entity into the database.
func PutCostIndicatorEntity(ctx context.Context, entity *CostIndicatorEntity) error {
	if entity == nil {
		return errors.New("cost indicator entity cannot be nil")
	}
	return datastore.Put(ctx, entity)
}

// GetCostIndicatorEntity extracts a cost indicator from the database.
func GetCostIndicatorEntity(ctx context.Context, entity *CostIndicatorEntity) (*CostIndicatorEntity, error) {
	if err := datastore.Get(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

// ListCostIndicators lists the cost indicators in the database, up to a limit (not yet implemented).
func ListCostIndicators(ctx context.Context, limit int) ([]*fleetcostpb.CostIndicator, error) {
	var out []*fleetcostpb.CostIndicator
	query := datastore.NewQuery(CostIndicatorKind)
	if err := datastore.Run(ctx, query, func(entity *CostIndicatorEntity) {
		out = append(out, entity.CostIndicator)
	}); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateCostIndicatorEntity extracts a cost indicator entity from the database and updates it.
func UpdateCostIndicatorEntity(ctx context.Context, entity *CostIndicatorEntity, fields []string) (*CostIndicatorEntity, error) {
	oldEntity := entity.Clone()
	newEntity := entity
	if err := datastore.Get(ctx, oldEntity); err != nil {
		return nil, err
	}
	maskutils.UpdateCostIndicatorProto(oldEntity.CostIndicator, newEntity.CostIndicator, fields)
	if err := datastore.Put(ctx, oldEntity); err != nil {
		return nil, err
	}
	return oldEntity, nil
}
