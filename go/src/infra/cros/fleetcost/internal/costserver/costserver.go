// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	"google.golang.org/grpc"

	fleetcostpb "infra/cros/fleetcost/api"
)

// NewFleetCostFrontend returns a new fleet cost frontend.
func NewFleetCostFrontend() fleetcostpb.FleetCostServer {
	return &FleetCostFrontend{}
}

// FleetCostFrontend is the fleet cost frontend.
type FleetCostFrontend struct{}

// Ping takes a PingRequest which is empty and returns a PingResponse which is empty.
func (f *FleetCostFrontend) Ping(context.Context, *fleetcostpb.PingRequest) (*fleetcostpb.PingResponse, error) {
	return &fleetcostpb.PingResponse{}, nil
}

// InstallServices installs services (such as the prpc server) into the frontend.
func InstallServices(costFrontend fleetcostpb.FleetCostServer, srv grpc.ServiceRegistrar) {
	fleetcostpb.RegisterFleetCostServer(srv, costFrontend)
}
