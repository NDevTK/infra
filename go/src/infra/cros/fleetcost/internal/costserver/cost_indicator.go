// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	"go.chromium.org/luci/common/errors"

	fleetcostpb "infra/cros/fleetcost/api"
)

func (f *FleetCostFrontend) CreateCostIndicator(_ context.Context, _ *fleetcostpb.CreateCostIndicatorRequest) (*fleetcostpb.CreateCostIndicatorResponse, error) {
	return nil, errors.Reason("not yet implemented").Err()
}
