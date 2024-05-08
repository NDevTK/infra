// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clients

import (
	"context"

	"google.golang.org/grpc"

	apipb "go.chromium.org/luci/swarming/proto/api_v2"
)

// MockSwarmingClientKey is used for testing.
var MockSwarmingClientKey contextKey = "used in tests only for setting the mock SwarmingClient"

type SwarmingClient interface {
	ListBots(ctx context.Context, in *apipb.BotsRequest, opts ...grpc.CallOption) (*apipb.BotInfoListResponse, error)
}

func NewSwarmingClient(ctx context.Context, host string) (SwarmingClient, error) {
	if mockClient, ok := ctx.Value(MockSwarmingClientKey).(SwarmingClient); ok {
		return mockClient, nil
	}
	pc, err := rawPRPCClient(ctx, host)
	if err != nil {
		return nil, err
	}
	ic := apipb.NewBotsClient(pc)
	return &swarmingService{
		client: ic,
	}, nil
}

// ufsService is used in non-test environments.
type swarmingService struct {
	client apipb.BotsClient
}

func (u *swarmingService) ListBots(ctx context.Context, in *apipb.BotsRequest, opts ...grpc.CallOption) (*apipb.BotInfoListResponse, error) {
	return u.client.ListBots(ctx, in, opts...)
}
