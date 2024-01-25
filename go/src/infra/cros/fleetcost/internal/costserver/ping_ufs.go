// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	"google.golang.org/protobuf/types/known/anypb"

	fleetcostpb "infra/cros/fleetcost/api"
	ufspb "infra/unifiedfleet/api/v1/rpc"
)

// PingUFS takes a PingUFSRequest which is empty and pings UFS, returning a description of what it did.
func (f *FleetCostFrontend) PingUFS(ctx context.Context, _ *fleetcostpb.PingUFSRequest) (*fleetcostpb.PingUFSResponse, error) {
	req := &ufspb.ListMachineLSEsRequest{
		KeysOnly: true,
		PageSize: 3,
	}
	ufsError := "no ufs error detected"
	resp, err := f.fleetClient.ListMachineLSEs(ctx, req)
	if err != nil {
		ufsError = err.Error()
	}
	ufsRequest, _ := anypb.New(req)
	ufsResponse, _ := anypb.New(resp)
	return &fleetcostpb.PingUFSResponse{
		UfsRequest:  ufsRequest,
		UfsResponse: ufsResponse,
		UfsHostname: f.ufsHostname,
		UfsError:    ufsError,
	}, nil
}
