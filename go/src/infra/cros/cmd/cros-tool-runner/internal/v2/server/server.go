// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc/codes"
	"infra/cros/cmd/cros-tool-runner/api"
	"infra/cros/cmd/cros-tool-runner/internal/v2/commands"
	"infra/cros/cmd/cros-tool-runner/internal/v2/templates"
)

// ContainerServerImpl implements the gRPC services by running commands and
// mapping errors to proper gRPC status codes.
type ContainerServerImpl struct {
	api.UnimplementedCrosToolRunnerContainerServiceServer
	networks          []string // TODO(mingkong) use a map
	containers        []string
	executor          CommandExecutor
	templateProcessor templates.TemplateProcessor
}

// CreateNetwork creates a new docker network with the given name.
func (s *ContainerServerImpl) CreateNetwork(ctx context.Context, request *api.CreateNetworkRequest) (*api.CreateNetworkResponse, error) {
	if request.Name == "" {
		return nil, utils.invalidArgument("Missing name")
	}
	cmd := commands.NetworkCreate{Name: request.Name}
	_, stderr, err := s.executor.Execute(ctx, &cmd)
	if stderr != "" {
		return nil, utils.toStatusError(stderr)
	}
	if err != nil {
		return nil, err
	}

	id, err := s.getNetworkId(ctx, request.Name)
	if err != nil {
		return nil, err
	}

	log.Println("success: created network", id)
	s.networks = append(s.networks, request.Name)
	return &api.CreateNetworkResponse{Network: &api.Network{Name: request.Name, Id: id, Owned: true}}, nil
}

// GetNetwork retrieves information of given docker network.
func (s *ContainerServerImpl) GetNetwork(ctx context.Context, request *api.GetNetworkRequest) (*api.GetNetworkResponse, error) {
	id, err := s.getNetworkId(ctx, request.Name)
	if err != nil {
		return nil, err
	}
	log.Println("success: found network", id)
	return &api.GetNetworkResponse{Network: &api.Network{Name: request.Name, Id: id, Owned: utils.contains(s.networks, request.Name)}}, nil
}

func (s *ContainerServerImpl) getNetworkId(ctx context.Context, name string) (string, error) {
	getNetworkIdCmd := compatibleLookupNetworkIdCommand(name)
	id, stderr, err := s.executor.Execute(ctx, getNetworkIdCmd)
	if id == "" {
		return "", utils.notFound(fmt.Sprintf("Cannot retrieve network ID with name %s", name))
	}
	if stderr != "" {
		return "", utils.toStatusError(stderr)
	}
	if err != nil {
		return "", err
	}
	return id, nil
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

// StartContainer pulls image and then calls docker run to start a container.
func (s *ContainerServerImpl) StartContainer(ctx context.Context, request *api.StartContainerRequest) (*api.StartContainerResponse, error) {
	if request.Name == "" {
		return nil, utils.invalidArgument("Missing name")
	}
	if request.ContainerImage == "" {
		return nil, utils.invalidArgument("Missing container_image")
	}
	// TODO(mingkong): define behavior of existing name in containers[]
	if request.StartCommand == nil || len(request.StartCommand) == 0 {
		return nil, utils.invalidArgument("Missing start_command")
	}
	if request.AdditionalOptions != nil {
		options := request.AdditionalOptions
		if options.Expose != nil && (len(options.Expose) > 1 || strings.Contains(options.Expose[0], "-")) {
			return nil, utils.unimplemented("Exposing multiple ports are not supported")
		}
	}
	pullErr := s.pullImage(ctx, request.ContainerImage)
	if pullErr != nil {
		return nil, pullErr
	}

	cmd := commands.DockerRun{StartContainerRequest: request}
	id, stderr, err := s.executor.Execute(ctx, &cmd)
	if stderr != "" {
		return nil, utils.toStatusErrorWithMapper(stderr, func(s string) codes.Code {
			switch {
			// docker error
			case strings.Contains(s, fmt.Sprintf("container name \"/%s\" is already in use", request.Name)):
				return codes.AlreadyExists
			// podman error
			case strings.Contains(s, fmt.Sprintf("container name \"%s\" is already in use", request.Name)):
				return codes.AlreadyExists
			default:
				return codes.Unknown
			}
		})
	}
	if err != nil {
		return nil, err
	}
	// TODO(mingkong): handle edge case where id is returned but container has immediately stopped: e.g. cros-dut cannot connect to dut
	log.Println("success: started container", id)
	s.containers = append(s.containers, request.Name)
	return &api.StartContainerResponse{Container: &api.Container{Name: request.Name, Id: id, Owned: true}}, nil
}

// pullImage pulls docker image and handles error mapping specifically
func (s *ContainerServerImpl) pullImage(ctx context.Context, image string) error {
	pullCmd := commands.DockerPull{ContainerImage: image}
	stdout, stderr, _ := s.executor.Execute(ctx, &pullCmd)
	// podman has stderr even when success
	if stdout == "" && stderr != "" {
		return utils.toStatusErrorWithMapper(stderr, func(s string) codes.Code {
			switch {
			// docker error
			case strings.Contains(s, "Permission \"artifactregistry.repositories.downloadArtifacts\" denied on resource"):
				return codes.PermissionDenied
			// podman error
			case strings.Contains(s, "unable to retrieve auth token: invalid username/password: unauthorized: failed authentication"):
				return codes.PermissionDenied
			// common error string
			case strings.Contains(s, "manifest unknown: Failed to fetch"):
				return codes.NotFound
			default:
				return codes.Unknown
			}
		})
	}
	log.Println("success: pulled image", image)
	return nil
}

// StartTemplatedContainer delegates to template processors to populate templates into valid StartContainerRequest,
// and then passes over to the generic endpoint.
func (s *ContainerServerImpl) StartTemplatedContainer(ctx context.Context, request *api.StartTemplatedContainerRequest) (*api.StartContainerResponse, error) {
	processedRequest, err := s.templateProcessor.Process(request)
	if err != nil {
		return nil, err
	}
	return s.StartContainer(ctx, processedRequest)
}

// StackCommands provides a scripting mechanism to execute a series of commands in order.
func (s *ContainerServerImpl) StackCommands(ctx context.Context, request *api.StackCommandsRequest) (*api.StackCommandsResponse, error) {
	outputs := make([]*api.StackCommandsResponse_Stackable, 0)
	for _, r := range request.Requests {
		switch t := r.Command.(type) {
		case *api.StackCommandsRequest_Stackable_CreateNetwork:
			output, err := s.CreateNetwork(ctx, r.GetCreateNetwork())
			if err != nil {
				return &api.StackCommandsResponse{Responses: outputs}, err
			}
			outputs = append(outputs, &api.StackCommandsResponse_Stackable{
				Output: &api.StackCommandsResponse_Stackable_CreateNetwork{
					CreateNetwork: output,
				}})
		case *api.StackCommandsRequest_Stackable_StartContainer:
			output, err := s.StartContainer(ctx, r.GetStartContainer())
			if err != nil {
				return &api.StackCommandsResponse{Responses: outputs}, err
			}
			outputs = append(outputs, &api.StackCommandsResponse_Stackable{
				Output: &api.StackCommandsResponse_Stackable_StartContainer{
					StartContainer: output,
				}})
		case *api.StackCommandsRequest_Stackable_StartTemplatedContainer:
			output, err := s.StartTemplatedContainer(ctx, r.GetStartTemplatedContainer())
			if err != nil {
				return &api.StackCommandsResponse{Responses: outputs}, err
			}
			outputs = append(outputs, &api.StackCommandsResponse_Stackable{
				Output: &api.StackCommandsResponse_Stackable_StartContainer{
					StartContainer: output,
				}})
		default:
			return &api.StackCommandsResponse{Responses: outputs}, utils.unimplemented(fmt.Sprintf("Unimplemented request type %v", t))
		}
	}
	return &api.StackCommandsResponse{Responses: outputs}, nil
}

// stopContainers removes containers that are owned by current CTRv2 service in the reverse order of how they are started.
func (s *ContainerServerImpl) stopContainers() {
	if len(s.containers) == 0 {
		log.Println("no containers to clean up")
		return
	}
	log.Println("stopping containers")
	cmd := commands.ContainerStop{Names: utils.reverse(s.containers)}
	stdout, stderr, _ := s.executor.Execute(context.Background(), &cmd)
	if stdout != "" {
		log.Println("received stdout:", stdout)
	}
	if stderr != "" {
		log.Println("received stderr", stderr)
	}
	// TODO(mingkong) define the behavior of stop container error.
	s.containers = make([]string, 0)
}

// removeNetworks removes networks that were created by current CTRv2 service.
func (s *ContainerServerImpl) removeNetworks() {
	if len(s.networks) == 0 {
		log.Println("no networks to clean up")
		return
	}
	log.Println("removing networks")
	cmd := commands.NetworkRemove{Names: s.networks}
	stdout, stderr, _ := s.executor.Execute(context.Background(), &cmd)
	if stdout != "" {
		log.Println("received stdout:", stdout)
	}
	if stderr != "" {
		log.Println("received stderr", stderr)
	}
	// TODO(mingkong) define the behavior of remove network error.
	s.networks = make([]string, 0)
}

// cleanup removes containers and networks in order to allow graceful shutdown of the CTRv2 service.
func (s *ContainerServerImpl) cleanup() {
	s.stopContainers()
	s.removeNetworks()
}
