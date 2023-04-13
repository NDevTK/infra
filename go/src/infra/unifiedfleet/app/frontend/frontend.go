// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/router"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/acl"
	"infra/unifiedfleet/app/untrusted"
)

// InstallServices installs ...
func InstallServices(apiServer *server.Server) {
	apiServer.ConfigurePRPC(func(p *prpc.Server) { p.AccessControl = prpc.AllowOriginAll })
	api.RegisterFleetServer(apiServer, &api.DecoratedFleet{
		Service: &FleetServerImpl{},
		Prelude: checkAccess,
	})
}

// InstallHandlers installs non PRPC handlers
func InstallHandlers(r *router.Router, mc router.MiddlewareChain) {
	mc = mc.Extend(func(ctx *router.Context, next router.Handler) {
		context, err := checkAccess(ctx.Context, ctx.HandlerPath, nil)
		ctx.Context = context
		if err != nil {
			logging.Errorf(ctx.Context, "Failed authorization %v", err)
			return
		}
		next(ctx)
	})
	r.POST("/pubsub/hart", mc, HaRTPushHandler)
	r.POST(untrusted.VerifierEndpoint, mc, untrusted.DeploymentVerifier)
}

// checkAccess verifies that the request is from an authorized user.
func checkAccess(ctx context.Context, rpcName string, _ proto.Message) (context.Context, error) {
	logging.Debugf(ctx, "Check access for %s", rpcName)
	// Everyone can call the RPC to check fleet test policy
	if rpcName == "CheckFleetTestsPolicy" {
		return ctx, nil
	}
	group := acl.Resolve(rpcName)
	allow, err := auth.IsMember(ctx, group...)
	if err != nil {
		logging.Errorf(ctx, "Check group '%s' membership failed: %s", group, err.Error())
		return ctx, status.Errorf(codes.Internal, "can't check access group membership: %s", err)
	}
	if !allow {
		return ctx, status.Errorf(codes.PermissionDenied, "%s is not a member of %s", auth.CurrentIdentity(ctx), group)
	}
	logging.Infof(ctx, "%s is a member of %s", auth.CurrentIdentity(ctx), group)
	return ctx, nil
}
