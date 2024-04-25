// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package external

import (
	"context"
	"net/http"

	"google.golang.org/grpc"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/server/auth"

	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// UFSClient interface provides a subset of UFS methods relevant to Device
// Manager. This provides an interface for testing and to add additional
// UFS methods as requirements evolve.
type UFSClient interface {
	GetDeviceData(context.Context, *ufsAPI.GetDeviceDataRequest, ...grpc.CallOption) (*ufsAPI.GetDeviceDataResponse, error)
	ListMachines(context.Context, *ufsAPI.ListMachinesRequest, ...grpc.CallOption) (*ufsAPI.ListMachinesResponse, error)
	ListMachineLSEs(context.Context, *ufsAPI.ListMachineLSEsRequest, ...grpc.CallOption) (*ufsAPI.ListMachineLSEsResponse, error)
}

// NewUFSClient creates a new client to access UFS.
func NewUFSClient(ctx context.Context, ufsHostname string) (ufsAPI.FleetClient, error) {
	if ufsHostname == "" {
		return nil, errors.Reason("NewUFSClient: must provide ufs service hostname").Err()
	}
	t, err := auth.GetRPCTransport(ctx, auth.AsSelf, auth.WithScopes(auth.CloudOAuthScopes...))
	if err != nil {
		return nil, errors.Annotate(err, "NewUFSClient: failed to get RPC transport to UFS service").Err()
	}
	return ufsAPI.NewFleetPRPCClient(&prpc.Client{
		C: &http.Client{
			Transport: t,
		},
		Host: ufsHostname,
	}), nil
}
