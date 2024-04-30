// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package clients defines the PRPC clients
package clients

import (
	"context"
	"net/http"
	"time"

	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/server/auth"
	apipb "go.chromium.org/luci/swarming/proto/api_v2"
)

// SwarmingClient is a Swarming API wrapper for drone-queen specific usage.
//
// In prod, a SwarmingClient for interacting with the Swarming service will be
// used. Tests should use a fake implementation.
type SwarmingClient interface {
	ListBots(ctx context.Context, in *apipb.BotsRequest) (*apipb.BotInfoListResponse, error)
}

// swarmingServiceImpl for use in real production envs.
type swarmingServiceImpl struct {
	BotsClient apipb.BotsClient
}

// Ensure swarmingClientImpl implements SwarmingClient.
var _ SwarmingClient = &swarmingServiceImpl{}

var MockSwarmingClientKey = "used in tests only for setting the mock SwarmingClient"

// NewSwarmingClient returns a new SwarmingClient to interact with Swarming APIs.
func NewSwarmingClient(ctx context.Context, host string, project string) (SwarmingClient, error) {
	if mockClient, ok := ctx.Value(&MockSwarmingClientKey).(*MockSwarmingClient); ok {
		return mockClient, nil
	}

	t, err := auth.GetRPCTransport(ctx, auth.AsSelf)
	if err != nil {
		return nil, err
	}

	prpcClient := prpc.Client{
		C:    &http.Client{Transport: t},
		Host: host,
	}
	return &swarmingServiceImpl{
		BotsClient: apipb.NewBotsClient(&prpcClient),
	}, nil
}

// ListBots calls `apipb.BotsClient.ListBots` to get the bots of a task via a task id.
func (s *swarmingServiceImpl) ListBots(ctx context.Context, req *apipb.BotsRequest) (*apipb.BotInfoListResponse, error) {
	subCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	return s.BotsClient.ListBots(subCtx, req)
}
