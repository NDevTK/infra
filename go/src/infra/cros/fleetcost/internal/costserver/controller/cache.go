// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"

	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver/entities"
	"infra/cros/fleetcost/internal/utils"
)

// CacheTTL is how long a cache entity lives.
var CacheTTL = 4 * time.Hour

// StoreCachedCostResult stores a cached cost result.
func StoreCachedCostResult(ctx context.Context, hostname string, currentTime time.Time, costResult *fleetcostAPI.GetCostResultResponse) error {
	return datastore.Put(ctx, &entities.CachedCostResultEntity{
		Hostname:       hostname,
		CostResult:     costResult.GetResult(),
		ExpirationTime: currentTime.Add(CacheTTL),
	})
}

// StoreCachedCostResultDefer stores a cached cost result.
//
// This function returns nothing because it is intended to be called in a defer block.
// By default, it will log the error it encountered (if any). To change the behavior on error,
// specify a non-nil value for the onError parameter.
func StoreCachedCostResultDefer(ctx context.Context, hostname string, currentTime time.Time, costResult *fleetcostAPI.GetCostResultResponse, parentFunctionErr error, onError func(error)) {
	if parentFunctionErr != nil {
		return
	}
	if costResult == nil {
		return
	}
	if err := StoreCachedCostResult(ctx, hostname, currentTime, costResult); err != nil {
		if onError == nil {
			logging.Errorf(ctx, "%s\n", err)
		} else {
			onError(err)
		}
	}
}

// ReadCachedCostResult reads a cached cost result if it's before the deadline.
func ReadCachedCostResult(ctx context.Context, req *fleetcostAPI.GetCostResultRequest, deadline time.Time) (*fleetcostAPI.GetCostResultResponse, error) {
	utils.Assert(!req.GetForceUpdate(), "ReadCachedCostResult should never be called when forceUpdate is given")
	query := datastore.NewQuery(entities.CachedCostResultKind).Eq("hostname", req.GetHostname())
	var ents []*entities.CachedCostResultEntity
	if err := datastore.GetAll(ctx, query, &ents); err != nil {
		return nil, err
	}
	if len(ents) == 0 {
		return nil, datastore.ErrNoSuchEntity
	}
	if len(ents) != 1 {
		return nil, fmt.Errorf("cache inconsitency: found %d records for hostname %q", len(ents), req.GetHostname())
	}
	if ents[0].ExpirationTime.After(deadline) {
		return &fleetcostAPI.GetCostResultResponse{
			Result: ents[0].CostResult,
		}, nil
	}
	return nil, fmt.Errorf("expiration time is too early: %s is after %s", ents[0].ExpirationTime.String(), deadline.String())
}
