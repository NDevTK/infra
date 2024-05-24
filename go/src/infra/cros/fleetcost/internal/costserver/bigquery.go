// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	"go.chromium.org/luci/common/errors"

	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver/inventory/ufs"
)

// PersistToBigquery persists the current cost indicators to BigQuery.
//
// Or rather, it would, if it were implemented, which it is not.
func (f *FleetCostFrontend) PersistToBigquery(ctx context.Context, request *fleetcostAPI.PersistToBigqueryRequest) (*fleetcostAPI.PersistToBigqueryResponse, error) {
	return nil, errors.New("not yet implemented")
}

// RepopulateCache repopulates the datastore cache ahead of persisting to bigquery.
func (f *FleetCostFrontend) RepopulateCache(ctx context.Context, request *fleetcostAPI.RepopulateCacheRequest) (*fleetcostAPI.RepopulateCacheResponse, error) {
	machines, err := ufs.GetAllMachineLSEs(ctx, f.fleetClient, true, nil)
	if err != nil {
		return nil, errors.Annotate(err, "reading machine LSEs from UFS").Err()
	}

	for _, machineLSE := range machines {
		hostname := machineLSE.GetName()
		if _, err := f.GetCostResult(ctx, &fleetcostAPI.GetCostResultRequest{
			Hostname:    hostname,
			ForceUpdate: true,
		}); err != nil {
			return nil, err
		}
	}

	return &fleetcostAPI.RepopulateCacheResponse{}, nil
}
