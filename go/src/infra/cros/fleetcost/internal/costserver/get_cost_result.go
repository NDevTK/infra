// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	"go.chromium.org/luci/common/errors"

	// TODO, move shared util to a standalone directory.
	shivasUtil "infra/cmd/shivas/utils"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver/controller"
	ufsUtil "infra/unifiedfleet/app/util"
)

// GetCostResult gets cost result of a fleet resource(DUT, scheduling unit).
func (f *FleetCostFrontend) GetCostResult(ctx context.Context, req *fleetcostAPI.GetCostResultRequest) (*fleetcostAPI.GetCostResultResponse, error) {
	// Handling OS namespace request only at MVP.
	ctx = shivasUtil.SetupContext(ctx, ufsUtil.OSNamespace)
	if f.fleetClient == nil {
		return nil, errors.New("fleet client must exist")
	}
	res, err := controller.CalculateCostForOsResource(ctx, f.fleetClient, req.GetHostname())
	if err != nil {
		return nil, errors.Annotate(err, "get cost result").Err()
	}
	return &fleetcostAPI.GetCostResultResponse{Result: res}, nil
}
