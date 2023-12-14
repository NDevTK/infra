// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	pb "go.chromium.org/chromiumos/infra/proto/go/satlabrpcserver"
)

const (
	ADDRESS = "localhost:6003"
)

func main() {
	conn, err := grpc.Dial(ADDRESS, grpc.WithInsecure(), grpc.WithBlock())

	if err != nil {
		log.Fatalf("was not able to connect to grpc server: %v", err)
	}

	defer conn.Close()

	ctx := context.Background()

	c := pb.NewSatlabRpcServiceClient(conn)
	res, err := c.ListBuildTargets(ctx, &pb.ListBuildTargetsRequest{})
	log.Printf("%v", res.GetBuildTargets())
}
