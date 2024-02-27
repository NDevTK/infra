// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package maskutils

import (
	fleetcostpb "infra/cros/fleetcost/api"
)

// UpdateCostIndicatorProto updates the cost indicator proto.
//
// The fieldmask argument is probably an expression of the form request.GetFieldMask().GetPaths().
func UpdateCostIndicatorProto(dst *fleetcostpb.CostIndicator, src *fleetcostpb.CostIndicator, fieldmask []string) {
	if dst == nil {
		return
	}
	for _, field := range fieldmask {
		updateCostIndicatorField(dst, src, field)
	}
}

// updateCostIndicatorField updates a single field in a cost indicator proto.
func updateCostIndicatorField(dst *fleetcostpb.CostIndicator, src *fleetcostpb.CostIndicator, field string) {
	switch field {
	case "name":
		dst.Name = src.GetName()
	case "type":
		dst.Type = src.GetType()
	case "board":
		dst.Board = src.GetBoard()
	case "model":
		dst.Model = src.GetModel()
	case "cost":
		dst.Cost = src.GetCost()
	case "cost_cadence":
		dst.CostCadence = src.GetCostCadence()
	case "burnout_rate":
		dst.BurnoutRate = src.GetBurnoutRate()
	case "location":
		dst.Location = src.GetLocation()
	case "description":
		dst.Description = src.GetDescription()
	}
}
