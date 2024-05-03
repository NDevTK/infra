// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clients

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

type contextKey string

// MockUFSClientKey is used for testing.
var MockUFSClientKey contextKey = "used in tests only for setting the mock UFSClient"

// UFSClient is UFS API	wrapper for BotsRegulator specific usage.
// It is used for mocking and testing.
type UFSClient interface {
	ListMachineLSEs(ctx context.Context, in *ufsAPI.ListMachineLSEsRequest, opts ...grpc.CallOption) (*ufsAPI.ListMachineLSEsResponse, error)
	ListSchedulingUnits(ctx context.Context, in *ufsAPI.ListSchedulingUnitsRequest, opts ...grpc.CallOption) (*ufsAPI.ListSchedulingUnitsResponse, error)
}

func NewUFSClient(ctx context.Context, host string) (UFSClient, error) {
	if mockClient, ok := ctx.Value(MockUFSClientKey).(UFSClient); ok {
		return mockClient, nil
	}
	pc, err := rawPRPCClient(ctx, host)
	if err != nil {
		return nil, err
	}
	ic := ufsAPI.NewFleetPRPCClient(pc)
	return &ufsService{
		client: ic,
	}, nil
}

// ufsService is used in non-test environments.
type ufsService struct {
	client ufsAPI.FleetClient
}

func (u *ufsService) ListMachineLSEs(ctx context.Context, in *ufsAPI.ListMachineLSEsRequest, opts ...grpc.CallOption) (*ufsAPI.ListMachineLSEsResponse, error) {
	return u.client.ListMachineLSEs(ctx, in, opts...)
}

func (u *ufsService) ListSchedulingUnits(ctx context.Context, in *ufsAPI.ListSchedulingUnitsRequest, opts ...grpc.CallOption) (*ufsAPI.ListSchedulingUnitsResponse, error) {
	return u.client.ListSchedulingUnits(ctx, in, opts...)
}

// SetUFSNamespace is a helper function to set UFS namespace in context.
func SetUFSNamespace(ctx context.Context, namespace string) context.Context {
	md := metadata.Pairs("namespace", namespace)
	return metadata.NewOutgoingContext(ctx, md)
}
