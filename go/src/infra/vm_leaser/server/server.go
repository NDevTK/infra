// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"

	pb "infra/vm_leaser/api/v1"
)

// Prove that Server implements pb.VMLeaserServiceServer by instantiating a Server
var _ pb.VMLeaserServiceServer = (*Server)(nil)

// Server is a struct implements the pb.VMLeaserServiceServer
type Server struct {
	pb.UnimplementedVMLeaserServiceServer
}

// NewServer returns a new Server
func NewServer() *Server {
	return &Server{}
}

// LeaseVM leases a VM defined by LeaseVMRequest
func (s *Server) LeaseVM(ctx context.Context, r *pb.LeaseVMRequest) (*pb.LeaseVMResponse, error) {
	log.Println("[server:LeaseVM] Started")
	if ctx.Err() == context.Canceled {
		return &pb.LeaseVMResponse{}, fmt.Errorf("client cancelled: abandoning")
	}

	return &pb.LeaseVMResponse{
		LeaseId: "Test ID",
	}, nil
}

// ExtendLease extends a VM lease
func (s *Server) ExtendLease(ctx context.Context, r *pb.ExtendLeaseRequest) (*pb.ExtendLeaseResponse, error) {
	log.Println("[server:ExtendLease] Started")
	if ctx.Err() == context.Canceled {
		return &pb.ExtendLeaseResponse{}, fmt.Errorf("client cancelled: abandoning")
	}

	return &pb.ExtendLeaseResponse{}, nil
}

// ReleaseVM releases a VM lease
func (s *Server) ReleaseVM(ctx context.Context, r *pb.ReleaseVMRequest) (*pb.ReleaseVMResponse, error) {
	log.Println("[server:ReleaseVM] Started")
	if ctx.Err() == context.Canceled {
		return &pb.ReleaseVMResponse{}, fmt.Errorf("client cancelled: abandoning")
	}

	return &pb.ReleaseVMResponse{}, nil
}
