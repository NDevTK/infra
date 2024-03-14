// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"
	"errors"

	fleetcostpb "infra/cros/fleetcost/api"
)

// GetCostResult gets information about a device.
func (f *FleetCostFrontend) GetCostResult(ctx context.Context, _ *fleetcostpb.GetCostResultRequest) (*fleetcostpb.GetCostResultResponse, error) {
	return nil, errors.New("not yet implemented")
}
