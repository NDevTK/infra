// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	"go.chromium.org/luci/common/errors"

	fleetcostAPI "infra/cros/fleetcost/api/rpc"
)

// PersistToBigquery persists the current cost indicators to BigQuery.
//
// Or rather, it would, if it were implemented, which it is not.
func (f *FleetCostFrontend) PersistToBigquery(ctx context.Context, request *fleetcostAPI.PersistToBigqueryRequest) (*fleetcostAPI.PersistToBigqueryResponse, error) {
	return nil, errors.New("not yet implemented")
}
