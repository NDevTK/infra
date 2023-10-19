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

	"infra/cros/satlab/common/services"
	"infra/cros/satlab/common/services/build_service"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/satlabrpcserver/platform/cpu_temperature"
	pb "infra/cros/satlab/satlabrpcserver/proto"
	"infra/cros/satlab/satlabrpcserver/services/bucket_services"
	"infra/cros/satlab/satlabrpcserver/services/dut_services"
	"infra/cros/satlab/satlabrpcserver/services/rpc_services"
	m "infra/cros/satlab/satlabrpcserver/utils/monitor"
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

	monitor := m.New()
	defer monitor.Stop()

	bucketService, err := bucket_services.New(ctx, site.GetGCSImageBucket())

	if err != nil {
		log.Fatalf("Failed to create a bucket connector %v", err)
	}
	buildService, err := build_service.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create a build connector %v", err)
	}
	dutService, err := dut_services.New()
	if err != nil {
		log.Fatalf("Failed to create a DUT service")
	}
	swarmingService, err := services.NewSwarmingService(ctx)
	if err != nil {
		// We don't want to fatal if user doesn't login
		log.Printf("Failed to create a swarming service")
	}

	// Register a CPU temperature orchestrator if we can find the temperature
	// on a platform
	var cpuTemperatureOrchestrator *cpu_temperature.CPUTemperatureOrchestrator
	cpuTemperature, err := cpu_temperature.NewCPUTemperature()
	if err != nil {
		log.Printf("This platform doesn't support getting the temperature, got an error: %v", err)
	} else {
		cpuTemperatureOrchestrator = cpu_temperature.NewOrchestrator(cpuTemperature, 30)
		monitor.Register(cpuTemperatureOrchestrator, time.Minute)
	}

	server := rpc_services.New(
		false,
		buildService,
		bucketService,
		dutService,
		cpuTemperatureOrchestrator,
		swarmingService,
	)

	defer server.Close()
	pb.RegisterSatlabRpcServiceServer(s, server)

	// Register reflection service on gRPC server.
	reflection.Register(s)

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}
