// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package models

import (
	"go.chromium.org/luci/gae/service/datastore"

	fleetcostpb "infra/cros/fleetcost/api"
)

const CostIndicatorKind = "CostIndicatorKind"

type CostIndicator struct {
	_kind         string                     `gae:"$kind,CostIndicatorKind"`
	ID            string                     `gae:"$id"`
	Extra         datastore.PropertyMap      `gae:",extra"`
	CostIndicator *fleetcostpb.CostIndicator `gae:"cost_indicator"`
}

// Silence staticcheck warning about unused field.
var _ = CostIndicator{}._kind
