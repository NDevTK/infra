// Copyright 2019 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package frontend implements the drone queen service.
package frontend

import (
	"context"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/server/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"infra/appengine/drone-queen/api"
	"infra/appengine/drone-queen/internal/config"
)

// RegisterServers registers RPC servers.
func RegisterServers(srv grpc.ServiceRegistrar) {
	var q DroneQueenImpl
	api.RegisterDroneServer(srv, &api.DecoratedDrone{
		Service: &q,
		Prelude: checkDroneAccess,
	})
	api.RegisterInventoryProviderServer(srv, &api.DecoratedInventoryProvider{
		Service: &q,
		Prelude: checkInventoryProviderAccess,
	})
	api.RegisterInspectServer(srv, &api.DecoratedInspect{
		Service: &q,
		Prelude: checkInspectAccess,
	})
}

func checkDroneAccess(ctx context.Context, _ string, _ proto.Message) (context.Context, error) {
	g := config.Get(ctx).GetAccessGroups()
	allow, err := auth.IsMember(ctx, g.GetDrones())
	if err != nil {
		return ctx, status.Errorf(codes.Internal, "can't check access group membership: %s", err)
	}
	if !allow {
		return ctx, status.Errorf(codes.PermissionDenied, "permission denied")
	}
	return ctx, nil
}

func checkInventoryProviderAccess(ctx context.Context, _ string, _ proto.Message) (context.Context, error) {
	g := config.Get(ctx).GetAccessGroups()
	allow, err := auth.IsMember(ctx, g.GetInventoryProviders())
	if err != nil {
		return ctx, status.Errorf(codes.Internal, "can't check access group membership: %s", err)
	}
	if !allow {
		return ctx, status.Errorf(codes.PermissionDenied, "permission denied")
	}
	return ctx, nil
}

func checkInspectAccess(ctx context.Context, _ string, _ proto.Message) (context.Context, error) {
	g := config.Get(ctx).GetAccessGroups()
	allow, err := auth.IsMember(ctx, g.GetInspectors())
	if err != nil {
		return ctx, status.Errorf(codes.Internal, "can't check access group membership: %s", err)
	}
	if !allow {
		return ctx, status.Errorf(codes.PermissionDenied, "permission denied")
	}
	return ctx, nil
}
