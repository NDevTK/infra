// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"
	"time"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/gae/service/datastore"

	// TODO, move shared util to a standalone directory.
	shivasUtil "infra/cmd/shivas/utils"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver/controller"
	"infra/cros/fleetcost/internal/utils"
	ufsUtil "infra/unifiedfleet/app/util"
)

// GetCostResult gets cost result of a fleet resource(DUT, scheduling unit).
func (f *FleetCostFrontend) GetCostResult(ctx context.Context, req *fleetcostAPI.GetCostResultRequest) (costResult *fleetcostAPI.GetCostResultResponse, costResultErr error) {
	now := time.Now().UTC()
	defer controller.StoreCachedCostResultDefer(ctx, req.GetHostname(), now, costResult, costResultErr, nil)
	if !req.GetForceUpdate() {
		readResult, readErr := controller.ReadCachedCostResult(ctx, req, now)
		if readResult != nil {
			utils.Assert(readErr == nil, "readErr must be nil")
			return readResult, nil
		}
		if !datastore.IsErrNoSuchEntity(readErr) {
			utils.Assert(readErr != nil, "readErr cannot be nil")
			return nil, readErr
		}
	}
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
