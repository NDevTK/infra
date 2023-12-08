// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	fleetcostpb "infra/cros/fleetcost/api"
)

// Ping takes a PingRequest which is empty and returns a PingResponse which is empty.
func (f *FleetCostFrontend) Ping(context.Context, *fleetcostpb.PingRequest) (*fleetcostpb.PingResponse, error) {
	return &fleetcostpb.PingResponse{}, nil
}
