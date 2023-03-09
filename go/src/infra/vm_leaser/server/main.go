// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/config/server/cfgmodule"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/cron"
	"go.chromium.org/luci/server/gaeemulation"
	"go.chromium.org/luci/server/module"
	"google.golang.org/grpc"

	pb "infra/vm_leaser/api/v1"
)

// InstallServices takes a VM Leaser service server and exposes it to a
// LUCI prpc.Server.
func InstallServices(s *Server, srv grpc.ServiceRegistrar) {
	pb.RegisterVMLeaserServiceServer(srv, s)
}

func main() {
	modules := []module.Module{
		gaeemulation.NewModuleFromFlags(),
		cfgmodule.NewModuleFromFlags(),
		cron.NewModuleFromFlags(),
	}

	// TODO(justinsuen): Temporarily use localhost endpoint. Need to add endpoint
	// to configs and dynamically determine GRPCAddr.
	options := server.Options{
		GRPCAddr: "127.0.0.1:50051",
	}

	server.Main(&options, modules, func(srv *server.Server) error {
		srv.Context = gologger.StdConfig.Use(srv.Context)
		srv.Context = logging.SetLevel(srv.Context, logging.Debug)

		logging.Infof(srv.Context, "Starting server.")
		logging.Infof(srv.Context, "Installing Services.")
		InstallServices(NewServer(), srv)
		logging.Infof(srv.Context, "Initialization finished.")
		return nil
	})
}
