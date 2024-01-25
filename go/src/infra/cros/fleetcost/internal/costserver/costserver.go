// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"google.golang.org/grpc"

	fleetcostpb "infra/cros/fleetcost/api"
	ufspb "infra/unifiedfleet/api/v1/rpc"
)

// NewFleetCostFrontend returns a new fleet cost frontend.
func NewFleetCostFrontend() fleetcostpb.FleetCostServer {
	return &FleetCostFrontend{}
}

// FleetCostFrontend is the fleet cost frontend.
type FleetCostFrontend struct {
	// Clients.
	fleetClient ufspb.FleetClient
	// Debugging information exposed through admin RPCs.
	ufsHostname string
}

// InstallServices installs services (such as the prpc server) into the frontend.
func InstallServices(costFrontend fleetcostpb.FleetCostServer, srv grpc.ServiceRegistrar) {
	fleetcostpb.RegisterFleetCostServer(srv, costFrontend)
}

// SetUFSClient sets the UFS client on a frontend.
func SetUFSClient(costFrontend *FleetCostFrontend, client ufspb.FleetClient) {
	if costFrontend == nil {
		panic("SetUFSClient: cost frontend cannot be nil")
	}
	if client == nil {
		panic("SetUFSClient: ufs client cannot be nil")
	}
	costFrontend.fleetClient = client
}

// SetUFSHostname sets the UFS hostname on the frontend.
//
// This is used to populate debugging info in the PingUFS RPC.
func SetUFSHostname(costFrontend *FleetCostFrontend, ufsHostname string) {
	costFrontend.ufsHostname = ufsHostname
}
