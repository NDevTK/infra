// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"

	"google.golang.org/protobuf/types/known/anypb"

	"go.chromium.org/luci/common/errors"

	fleetcostpb "infra/cros/fleetcost/api"
	ufspb "infra/unifiedfleet/api/v1/rpc"
)

// PingUFS takes a PingUFSRequest which is empty and pings UFS, returning a description of what it did.
func (f *FleetCostFrontend) PingUFS(ctx context.Context, _ *fleetcostpb.PingUFSRequest) (*fleetcostpb.PingUFSResponse, error) {
	req := &ufspb.ListMachineLSEsRequest{
		PageSize: 3,
	}
	resp, err := f.fleetClient.ListMachineLSEs(ctx, req)
	if err != nil {
		return nil, errors.Annotate(err, "error calling UFS from cost server").Err()
	}
	ufsRequest, _ := anypb.New(req)
	ufsResponse, _ := anypb.New(resp)
	return &fleetcostpb.PingUFSResponse{
		UfsRequest:  ufsRequest,
		UfsResponse: ufsResponse,
	}, nil
}
