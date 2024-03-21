// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	"go.chromium.org/luci/common/errors"

	fleetcostAPI "infra/cros/fleetcost/api/rpc"
)

// GetCostResult gets information about a device.
func (f *FleetCostFrontend) GetCostResult(ctx context.Context, _ *fleetcostAPI.GetCostResultRequest) (*fleetcostAPI.GetCostResultResponse, error) {
	return nil, errors.New("not yet implemented")
}
