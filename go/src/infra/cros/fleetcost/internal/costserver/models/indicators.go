// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package models

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/gae/service/datastore"

	fleetcostpb "infra/cros/fleetcost/api/models"
	"infra/cros/fleetcost/internal/costserver/maskutils"
)

const CostIndicatorKind = "CostIndicatorKind"

type CostIndicatorEntity struct {
	_kind         string                     `gae:"$kind,CostIndicatorKind"`
	Extra         datastore.PropertyMap      `gae:",extra"`
	CostIndicator *fleetcostpb.CostIndicator `gae:"cost_indicator"`
	// Indexed fields for improved query performance.
	Board    string `gae:"board"`
	Model    string `gae:"model"`
	Sku      string `gae:"sku"`
	Type     string `gae:"type"`
	Location string `gae:"location"`
}

// Silence staticcheck warning about unused field.
var _ = CostIndicatorEntity{}._kind

// Save saves an entity.
func (indicator *CostIndicatorEntity) Save(withMeta bool) (datastore.PropertyMap, error) {
	// TODO(gregorynisbet): extract normalization logic to helper function.
	indicator.Board = indicator.CostIndicator.GetBoard()
	indicator.Model = indicator.CostIndicator.GetModel()
	indicator.Sku = indicator.CostIndicator.GetSku()
	if int(indicator.CostIndicator.GetType()) != 0 {
		indicator.Type = indicator.CostIndicator.GetType().String()
	}
	if int(indicator.CostIndicator.GetLocation()) != 0 {
		indicator.Location = indicator.CostIndicator.GetLocation().String()
	}
	return datastore.GetPLS(indicator).Save(withMeta)
}

// Load loads an entity.
func (indicator *CostIndicatorEntity) Load(propertyMap datastore.PropertyMap) error {
	return datastore.GetPLS(indicator).Load(propertyMap)
}

var _ datastore.PropertyLoadSaver = &CostIndicatorEntity{}

var _ datastore.MetaGetterSetter = &CostIndicatorEntity{}

// GetAllMeta transfers control to the default implementation of GetAllMeta.
// We need this function so that we can compute the ID.
func (indicator *CostIndicatorEntity) GetAllMeta() datastore.PropertyMap {
	return datastore.GetPLS(indicator).GetAllMeta()
}

// SetMeta always returns false because we do not allow meta keys to be changed and false communicates this to the LUCI datastore library.
func (indicator *CostIndicatorEntity) SetMeta(key string, value any) bool {
	return false
}

// GetMeta gets meta-values. The id ("$id") is computed. The other things (like $kind, for instance) get their default values.
func (indicator *CostIndicatorEntity) GetMeta(key string) (any, bool) {
	if key == "id" {
		costIndicator := indicator.CostIndicator
		return fmt.Sprintf("v1-%s-%s-%s", costIndicator.GetBoard(), costIndicator.GetModel(), costIndicator.GetSku()), true
	}
	return datastore.GetPLS(indicator).GetMeta(key)
}

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
