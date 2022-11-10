// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"

	pb "infra/vm_leaser/api"
)

// Prove that Server implements pb.VmLeaserServiceServer by instantiating a Server
var _ pb.VmLeaserServiceServer = (*Server)(nil)

// Server is a struct implements the pb.VmLeaserServiceServer
type Server struct {
	pb.UnimplementedVmLeaserServiceServer
}

// NewServer returns a new Server
func NewServer() *Server {
	return &Server{}
}

// LeaseVm leases a VM defined by LeaseVmRequest
func (s *Server) LeaseVm(ctx context.Context, r *pb.LeaseVmRequest) (*pb.LeaseVmResponse, error) {
	log.Println("[server:LeaseVm] Started")
	if ctx.Err() == context.Canceled {
		return &pb.LeaseVmResponse{}, fmt.Errorf("client cancelled: abandoning")
	}

	return &pb.LeaseVmResponse{
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

// ReleaseVm releases a VM lease
func (s *Server) ReleaseVm(ctx context.Context, r *pb.ReleaseVmRequest) (*pb.ReleaseVmResponse, error) {
	log.Println("[server:ReleaseVm] Started")
	if ctx.Err() == context.Canceled {
		return &pb.ReleaseVmResponse{}, fmt.Errorf("client cancelled: abandoning")
	}

	return &pb.ReleaseVmResponse{}, nil
}
