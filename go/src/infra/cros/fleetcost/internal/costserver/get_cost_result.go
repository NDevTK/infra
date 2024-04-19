// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"

	// TODO, move shared util to a standalone directory.
	shivasUtil "infra/cmd/shivas/utils"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver/controller"
	ufsUtil "infra/unifiedfleet/app/util"
)

// GetCostResult gets cost result of a fleet resource(DUT, scheduling unit).
func (f *FleetCostFrontend) GetCostResult(ctx context.Context, req *fleetcostAPI.GetCostResultRequest) (*fleetcostAPI.GetCostResultResponse, error) {
	if req.GetForceUpdate() {
		return f.getCostResultImpl(ctx, req)
	}
	readResult, readErr := controller.ReadCachedCostResult(ctx, req.GetHostname())
	if readErr == nil {
		return &fleetcostAPI.GetCostResultResponse{Result: readResult}, nil
	}
	if !datastore.IsErrNoSuchEntity(readErr) {
		return nil, readErr
	}
	return f.getCostResultImpl(ctx, req)
}

// Function getCostResultImpl calculates a cost result and saves it to the database.
//
// We assume that either GetForceUpdate has been applied or that there's no cache entry that's recent enough to use instead.
func (f *FleetCostFrontend) getCostResultImpl(ctx context.Context, req *fleetcostAPI.GetCostResultRequest) (*fleetcostAPI.GetCostResultResponse, error) {
	// Handling OS namespace request only at MVP.
	logging.Infof(ctx, "begin cost result request for dut %q Id %q", req.GetHostname(), req.GetDeviceId())
	ctx = shivasUtil.SetupContext(ctx, ufsUtil.OSNamespace)
	if f.fleetClient == nil {
		return nil, errors.New("fleet client must exist")
	}
	res, err := controller.CalculateCostForOsResource(ctx, f.fleetClient, req.GetHostname())
	if err != nil {
		return nil, errors.Annotate(err, "get cost result").Err()
	}
	if err := controller.StoreCachedCostResult(ctx, req.GetHostname(), res); err != nil {
		logging.Errorf(ctx, "%s\n", errors.Annotate(err, "caching get cost result").Err())
	}
	return &fleetcostAPI.GetCostResultResponse{Result: res}, nil
}
