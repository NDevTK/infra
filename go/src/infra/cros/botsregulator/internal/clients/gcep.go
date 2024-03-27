// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clients

import (
	"context"

	"google.golang.org/grpc"

	gcepAPI "go.chromium.org/luci/gce/api/config/v1"
)

// MockGCEPClientKey is used for testing.
var MockGCEPClientKey contextKey = "used in tests only for setting the mock GCEPClient"

// GCEPClient is GCE Provider API wrapper for BotsRegulator specific usage.
// It is used for mocking and testing.
type GCEPClient interface {
	Get(ctx context.Context, in *gcepAPI.GetRequest, opts ...grpc.CallOption) (*gcepAPI.Config, error)
	Update(ctx context.Context, in *gcepAPI.UpdateRequest, opts ...grpc.CallOption) (*gcepAPI.Config, error)
}

func NewGCEPClient(ctx context.Context, host string) (GCEPClient, error) {
	if mockGCEPClient, ok := ctx.Value(MockGCEPClientKey).(GCEPClient); ok {
		return mockGCEPClient, nil
	}
	pc, err := rawPRPCClient(ctx, host)
	if err != nil {
		return nil, err
	}
	ic := gcepAPI.NewConfigurationPRPCClient(pc)
	return &gcepService{
		client: ic,
	}, nil
}

// UFSService is used in non-test environments.
type gcepService struct {
	client gcepAPI.ConfigurationClient
}

func (g *gcepService) Get(ctx context.Context, in *gcepAPI.GetRequest, opts ...grpc.CallOption) (*gcepAPI.Config, error) {
	return g.client.Get(ctx, in, opts...)
}

func (g *gcepService) Update(ctx context.Context, in *gcepAPI.UpdateRequest, opts ...grpc.CallOption) (*gcepAPI.Config, error) {
	return g.client.Update(ctx, in, opts...)
}
