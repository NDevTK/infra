// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	shivasUtil "infra/cmd/shivas/utils"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver/controller"
	"infra/cros/fleetcost/internal/costserver/inventory/ufs"
	ufsUtil "infra/unifiedfleet/app/util"
)

// PersistToBigquery persists the current cost indicators to BigQuery.
//
// Or rather, it would, if it were implemented, which it is not.
func (f *FleetCostFrontend) PersistToBigquery(ctx context.Context, request *fleetcostAPI.PersistToBigqueryRequest) (*fleetcostAPI.PersistToBigqueryResponse, error) {
	logging.Infof(ctx, "begin PersistToBigquery")
	err := controller.PersistToBigquery(ctx, f.projectID, f.bqClient, request.GetReadonly())
	if err != nil {
		return nil, err
	}
	logging.Infof(ctx, "PersistToBigquery was successful")
	return &fleetcostAPI.PersistToBigqueryResponse{}, nil
}

// RepopulateCache repopulates the datastore cache ahead of persisting to bigquery.
func (f *FleetCostFrontend) RepopulateCache(ctx context.Context, request *fleetcostAPI.RepopulateCacheRequest) (*fleetcostAPI.RepopulateCacheResponse, error) {
	ctx = shivasUtil.SetupContext(ctx, ufsUtil.OSNamespace)
	machines, err := ufs.GetAllMachineLSEs(ctx, f.fleetClient, false, nil)
	if err != nil {
		return nil, errors.Annotate(err, "reading machine LSEs from UFS").Err()
	}

	var tally int32
	for _, machineLSE := range machines {
		hostname := machineLSE.GetHostname()
		if hostname == "" {
			logging.Debugf(ctx, "machine %q has empty hostname, skipping. Full record %v", machineLSE.GetName(), machineLSE)
			continue
		}
		if _, err := f.GetCostResult(ctx, &fleetcostAPI.GetCostResultRequest{
			Hostname:              hostname,
			ForceUpdate:           true,
			ForgiveMissingEntries: request.GetForgiveMissingEntries(),
		}); err != nil {
			return nil, err
		}
		tally++
	}

	logging.Debugf(ctx, "processed %d records total", tally)

	return &fleetcostAPI.RepopulateCacheResponse{
		ProcessedRecords: tally,
	}, nil
}
