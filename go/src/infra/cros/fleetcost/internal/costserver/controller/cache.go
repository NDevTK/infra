// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/luci/gae/service/datastore"

	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver/entities"
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

// ReadCachedCostResult reads a cached cost result if it's before the deadline.
func ReadCachedCostResult(ctx context.Context, hostname string, deadline time.Time) (*fleetcostAPI.GetCostResultResponse, error) {
	query := datastore.NewQuery(entities.CachedCostResultKind).Eq("hostname", hostname)
	var ents []*entities.CachedCostResultEntity
	if err := datastore.GetAll(ctx, query, &ents); err != nil {
		return nil, err
	}
	if len(ents) == 0 {
		return nil, datastore.ErrNoSuchEntity
	}
	if len(ents) != 1 {
		return nil, fmt.Errorf("cache inconsitency: found %d records for hostname %q", len(ents), hostname)
	}
	if ents[0].ExpirationTime.After(deadline) {
		return &fleetcostAPI.GetCostResultResponse{
			Result: ents[0].CostResult,
		}, nil
	}
	return nil, fmt.Errorf("expiration time is too early: %s is after %s", ents[0].ExpirationTime.String(), deadline.String())
}
