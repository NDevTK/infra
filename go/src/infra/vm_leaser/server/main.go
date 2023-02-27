// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "infra/vm_leaser/api/v1"
)

func main() {
	ctx := gologger.StdConfig.Use(context.Background())
	ctx = logging.SetLevel(ctx, logging.Debug)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	grpcEndpoint := fmt.Sprintf(":%s", port)
	logging.Infof(ctx, "gRPC endpoint [%s]", grpcEndpoint)

	grpcServer := grpc.NewServer()
	pb.RegisterVMLeaserServiceServer(grpcServer, NewServer())

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	listen, err := net.Listen("tcp", grpcEndpoint)
	if err != nil {
		logging.Errorf(ctx, "failed to listen: %v", err)
		os.Exit(1)
	}

	logging.Infof(ctx, "Starting: gRPC Listener [%s]\n", grpcEndpoint)
	grpcServer.Serve(listen)
}
