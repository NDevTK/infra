// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "infra/cros/satlab/satlabrpcserver/proto"
	"infra/cros/satlab/satlabrpcserver/services/bucket_services"
	"infra/cros/satlab/satlabrpcserver/services/build_services"
	"infra/cros/satlab/satlabrpcserver/services/rpc_services"
	"infra/cros/satlab/satlabrpcserver/utils"
	"infra/cros/satlab/satlabrpcserver/utils/constants"
)

const (
	// PORT for gRPC server to listen to
	PORT = ":6003"
)

func main() {
	lis, err := net.Listen("tcp", PORT)

	if err != nil {
		log.Fatalf("failed connection: %v", err)
	}

	s := grpc.NewServer()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*20)
	defer cancel()

	bucketService, err := bucket_services.New(ctx, constants.BucketName)

	if err != nil {
		log.Fatalf("Failed to create a bucket connector %v", err)
	}
	buildService, err := build_services.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create a build connector %v", err)
	}
	labelParser, err := utils.NewLabelParser()
	if err != nil {
		log.Fatalf("Failed to create a label parser %v", err)
	}

	server := rpc_services.New(buildService, bucketService, labelParser)
	pb.RegisterSatlabRpcServiceServer(s, server)

	// Register reflection service on gRPC server.
	reflection.Register(s)

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}

	defer server.Close()
}
