// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "infra/appengine/weetbix/proto/v1"
)

// Rules implements pb.TestHistoryServer.
type testHistoryServer struct {
}

// NewTestHistoryServer returns a new pb.TestHistoryServer.
func NewTestHistoryServer() pb.TestHistoryServer {
	return &pb.DecoratedTestHistory{
		Service:  &testHistoryServer{},
		Postlude: gRPCifyAndLogPostlude,
	}
}

// Retrieves test verdicts for a given test ID in a given project and in a given
// range of time.
func (*testHistoryServer) Query(ctx context.Context, req *pb.QueryTestHistoryRequest) (*pb.QueryTestHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Query not implemented")
}

// Retrieves a summary of test verdicts for a given test ID in a given project
// and in a given range of times.
func (*testHistoryServer) QueryStats(ctx context.Context, req *pb.QueryTestHistoryStatsRequest) (*pb.QueryTestHistoryStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryStats not implemented")
}

// Retrieves variants for a given test ID in a given project that were recorded
// in the past 90 days.
func (*testHistoryServer) QueryVariants(ctx context.Context, req *pb.QueryVariantsRequest) (*pb.QueryVariantsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryVariantsRequest not implemented")
}
