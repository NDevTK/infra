// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package server

import (
	"context"
	"log"
	"os"

	"infra/cros/cmd/cros-tool-runner/api"
	"infra/cros/cmd/cros-tool-runner/internal/v2/commands"
)

// ContainerServerImpl implements the gRPC services by running commands and
// mapping errors to proper gRPC status codes.
type ContainerServerImpl struct {
	api.UnimplementedCrosToolRunnerContainerServiceServer
	networks []string // TODO(mingkong) use a map
}

// CreateNetwork creates a new docker network with the given name.
func (s *ContainerServerImpl) CreateNetwork(ctx context.Context, request *api.CreateNetworkRequest) (*api.CreateNetworkResponse, error) {
	cmd := commands.NetworkCreate{Name: request.Name}
	stdout, stderr, err := cmd.Execute(ctx)
	if stderr != "" {
		return nil, utils.toStatusError(stderr)
	}
	if err != nil {
		return nil, err
	}
	id := utils.firstLine(stdout)
	log.Println("success: created network", id)
	s.networks = append(s.networks, request.Name)
	return &api.CreateNetworkResponse{Network: &api.Network{Name: request.Name, Id: id, Owned: true}}, nil
}

// GetNetwork retrieves information of given docker network.
func (s *ContainerServerImpl) GetNetwork(ctx context.Context, request *api.GetNetworkRequest) (*api.GetNetworkResponse, error) {
	cmd := commands.NetworkInspect{Names: []string{request.Name}, Format: "{{.Id}}"}
	stdout, stderr, err := cmd.Execute(ctx)
	if stderr != "" {
		return nil, utils.toStatusError(stderr)
	}
	if err != nil {
		return nil, err
	}
	id := utils.firstLine(stdout)
	log.Println("success: found network", id)
	return &api.GetNetworkResponse{Network: &api.Network{Name: request.Name, Id: id, Owned: utils.contains(s.networks, request.Name)}}, nil
}

// Shutdown signals to shut down the CTRv2 gRPC server.
func (s *ContainerServerImpl) Shutdown(context.Context, *api.ShutdownRequest) (*api.ShutdownResponse, error) {
	log.Println("processing shutdown request")
	p, err := os.FindProcess(os.Getpid())
	if err == nil {
		p.Signal(os.Interrupt)
	}
	log.Println("interrupt signal sent")
	return &api.ShutdownResponse{}, err
}

// removeNetworks removes networks that were created by current CTRv2 service.
func (s *ContainerServerImpl) removeNetworks() {
	if len(s.networks) == 0 {
		log.Println("no networks to clean up")
		return
	}
	log.Println("removing networks")
	cmd := commands.NetworkRemove{Names: s.networks}
	stdout, stderr, err := cmd.Execute(context.Background())
	if stdout != "" {
		log.Println("received stdout:", stdout)
	}
	if stderr != "" {
		log.Println("received stderr", stderr)
	}
	if err != nil {
		log.Println("received error", err)
	}
}
