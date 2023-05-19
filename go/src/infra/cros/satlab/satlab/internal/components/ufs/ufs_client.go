// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ufs

import (
	"context"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"
	"google.golang.org/grpc"

	"infra/cmd/shivas/site"
	"infra/cmdsupport/cmdlib"
	ufsModels "infra/unifiedfleet/api/v1/models"
	ufsApi "infra/unifiedfleet/api/v1/rpc"
)

// UFSClient interface provides subset of UFS methods relevant to Satlab CLI
// not strictly necessary (could just use proto client directly) but makes it easier to test
// add more methods if additional UFS reqs needed
type UFSClient interface {
	GetDut(context.Context, *ufsApi.GetMachineLSERequest) (*ufsModels.MachineLSE, error)
	GetMachine(context.Context, *ufsApi.GetMachineRequest) (*ufsModels.Machine, error)
	GetMachineLSE(context.Context, *ufsApi.GetMachineLSERequest, ...grpc.CallOption) (*ufsModels.MachineLSE, error)
	UpdateMachineLSE(context.Context, *ufsApi.UpdateMachineLSERequest, ...grpc.CallOption) (*ufsModels.MachineLSE, error)
}

//  implementation of UFS client
type clientImpl struct {
	client ufsApi.FleetClient
}

// GetDut uses the GetMachineLSE method to fetch info about lab setup
func (c *clientImpl) GetDut(ctx context.Context, req *ufsApi.GetMachineLSERequest) (*ufsModels.MachineLSE, error) {
	return c.client.GetMachineLSE(ctx, req)
}

// GetMachine fetches information about the machine we request.
func (c *clientImpl) GetMachine(ctx context.Context, req *ufsApi.GetMachineRequest) (*ufsModels.Machine, error) {
	return c.client.GetMachine(ctx, req)
}

// GetMachine fetches information about the machine we request.
func (c *clientImpl) GetMachineLSE(ctx context.Context, req *ufsApi.GetMachineLSERequest, opts ...grpc.CallOption) (*ufsModels.MachineLSE, error) {
	return c.client.GetMachineLSE(ctx, req)
}

// GetMachine fetches information about the machine we request.
func (c *clientImpl) UpdateMachineLSE(ctx context.Context, req *ufsApi.UpdateMachineLSERequest, opts ...grpc.CallOption) (*ufsModels.MachineLSE, error) {
	return c.client.UpdateMachineLSE(ctx, req)
}

// NewUFSClient creates a new client to access UFS, but only exposing specific methods needed for Satlab CLI
func NewUFSClient(ctx context.Context, ufsService string, authFlags *authcli.Flags) (UFSClient, error) {
	if ufsService == "" {
		return nil, errors.Reason("new ufs client: must provide ufs service hostname").Err()
	}

	httpClient, err := cmdlib.NewHTTPClient(ctx, authFlags)
	if err != nil {
		return nil, err
	}

	client := ufsApi.NewFleetPRPCClient(&prpc.Client{
		C:       httpClient,
		Host:    ufsService,
		Options: site.DefaultPRPCOptions,
	})

	return &clientImpl{client: client}, nil
}
