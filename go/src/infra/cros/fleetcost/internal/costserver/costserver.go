// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"cloud.google.com/go/bigquery"
	"google.golang.org/grpc"

	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	ufspb "infra/unifiedfleet/api/v1/rpc"
)

// NewFleetCostFrontend returns a new fleet cost frontend.
func NewFleetCostFrontend() fleetcostAPI.FleetCostServer {
	return &FleetCostFrontend{}
}

// FleetCostFrontend is the fleet cost frontend.
type FleetCostFrontend struct {
	// Clients.
	fleetClient ufspb.FleetClient
	// Debugging information exposed through admin RPCs.
	ufsHostname string
	// bqClient is a BigQuery client.
	bqClient *bigquery.Client
}

// InstallServices installs services (such as the prpc server) into the frontend.
func InstallServices(costFrontend fleetcostAPI.FleetCostServer, srv grpc.ServiceRegistrar) {
	fleetcostAPI.RegisterFleetCostServer(srv, costFrontend)
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

// SetBQClient sets the bigquery client.
func SetBQClient(costFrontend *FleetCostFrontend, client *bigquery.Client) {
	costFrontend.bqClient = client
}
