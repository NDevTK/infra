// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"google.golang.org/grpc"

	fleetcostpb "infra/cros/fleetcost/api"
)

// NewFleetCostFrontend returns a new fleet cost frontend.
func NewFleetCostFrontend() fleetcostpb.FleetCostServer {
	return &FleetCostFrontend{}
}

// FleetCostFrontend is the fleet cost frontend.
type FleetCostFrontend struct{}

// InstallServices installs services (such as the prpc server) into the frontend.
func InstallServices(costFrontend fleetcostpb.FleetCostServer, srv grpc.ServiceRegistrar) {
	fleetcostpb.RegisterFleetCostServer(srv, costFrontend)
}
