// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package models

import (
	"google.golang.org/protobuf/proto"

	"go.chromium.org/luci/gae/service/datastore"

	fleetcostpb "infra/cros/fleetcost/api"
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
