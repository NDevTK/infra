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

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/grpc/codes"
	"infra/cros/cmd/cros-tool-runner/internal/v2/commands"
	"infra/cros/cmd/cros-tool-runner/internal/v2/templates"
)

// ContainerServerImpl implements the gRPC services by running commands and
// mapping errors to proper gRPC status codes.
type ContainerServerImpl struct {
	api.UnimplementedCrosToolRunnerContainerServiceServer
	networks          ownershipRecorder
	containers        ownershipRecorder
	executor          CommandExecutor
	templateProcessor templates.TemplateProcessor
	containerLookuper templates.ContainerLookuper
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
	s.networks.recordOwnership(request.Name, id)
	return &api.CreateNetworkResponse{Network: &api.Network{Name: request.Name, Id: id, Owned: true}}, nil
}

// GetNetwork retrieves information of given docker network.
func (s *ContainerServerImpl) GetNetwork(ctx context.Context, request *api.GetNetworkRequest) (*api.GetNetworkResponse, error) {
	id, err := s.getNetworkId(ctx, request.Name)
	if err != nil {
		return nil, err
	}
	log.Println("success: found network", id)
	return &api.GetNetworkResponse{Network: &api.Network{Name: request.Name, Id: id, Owned: s.networks.hasOwnership(request.Name, id)}}, nil
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
		err = p.Signal(os.Interrupt)
	}
	log.Println("interrupt signal sent")
	return &api.ShutdownResponse{}, err
}

// LoginRegistry logs in a docker image registry server
func (s *ContainerServerImpl) LoginRegistry(ctx context.Context, request *api.LoginRegistryRequest) (*api.LoginRegistryResponse, error) {
	if request.Username == "" {
		return nil, utils.invalidArgument("Missing username")
	}
	if request.Password == "" {
		return nil, utils.invalidArgument("Missing password")
	}
	if request.Registry == "" {
		return nil, utils.invalidArgument("Missing registry")
	}
	extensionOutput := s.handleLoginRegistryExtension(ctx, request)
	if request.Password == "$(gcloud auth print-access-token)" {
		password, stderr, err := s.executor.Execute(ctx, &commands.GcloudAuthTokenPrint{})
		if err != nil {
			return nil, errors.Annotate(err, stderr).Err()
		}
		request.Password = password
	}
	cmd := commands.DockerLogin{LoginRegistryRequest: request}
	stdout, stderr, err := s.executor.Execute(ctx, &cmd)
	// docker always has stderr warning
	if stdout == "" && stderr != "" {
		return nil, utils.toStatusErrorWithMapper(stderr, func(s string) codes.Code {
			switch {
			// docker error
			case strings.Contains(s, "unauthorized: failed authentication"):
				return codes.PermissionDenied
			// podman error
			case strings.Contains(s, "invalid username/password"):
				return codes.PermissionDenied
			default:
				return codes.Unknown
			}
		})
	}
	if err != nil {
		return nil, err
	}
	return &api.LoginRegistryResponse{Message: stdout, ExtensionsOutput: extensionOutput}, nil
}

// handleLoginRegistryExtension processes extensions with best-effort support
// and never throws errors.
// If there are more extensions, the handling should be moved to a new module.
func (s *ContainerServerImpl) handleLoginRegistryExtension(ctx context.Context, request *api.LoginRegistryRequest) []string {
	var extensionOutput []string
	if request.Extensions != nil && len(request.Extensions.GcloudAuthServiceAccountArgs) > 0 {
		stdout, stderr, err := s.executor.Execute(ctx, &commands.GcloudAuthServiceAccount{Args: request.Extensions.GcloudAuthServiceAccountArgs})
		if stdout != "" {
			extensionOutput = append(extensionOutput, stdout)
		}
		if err != nil || stderr != "" {
			log.Printf("warning: gcloud activate service account exit with error: %s", stderr)
			if stderr != "" {
				extensionOutput = append(extensionOutput, stderr)
			}
		}
	}
	return extensionOutput
}

// StartContainer pulls image and then calls docker run to start a container.
func (s *ContainerServerImpl) StartContainer(ctx context.Context, request *api.StartContainerRequest) (*api.StartContainerResponse, error) {
	if request.Name == "" {
		return nil, utils.invalidArgument("Missing name")
	}
	if request.ContainerImage == "" {
		return nil, utils.invalidArgument("Missing container_image")
	}
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
		log.Printf("warning: error when pulling image: %s", pullErr)
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
	log.Println("success: started container", id)
	s.containers.recordOwnership(request.Name, id)
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
		case *api.StackCommandsRequest_Stackable_LoginRegistry:
			output, err := s.LoginRegistry(ctx, r.GetLoginRegistry())
			if err != nil {
				return &api.StackCommandsResponse{Responses: outputs}, err
			}
			outputs = append(outputs, &api.StackCommandsResponse_Stackable{
				Output: &api.StackCommandsResponse_Stackable_LoginRegistry{
					LoginRegistry: output,
				}})
		default:
			return &api.StackCommandsResponse{Responses: outputs}, utils.unimplemented(fmt.Sprintf("Unimplemented request type %v", t))
		}
	}
	return &api.StackCommandsResponse{Responses: outputs}, nil
}

// GetContainer retrieves information of a container.
func (s *ContainerServerImpl) GetContainer(ctx context.Context, request *api.GetContainerRequest) (*api.GetContainerResponse, error) {
	id, err := s.getContainerId(ctx, request.Name)
	if err != nil {
		return nil, err
	}
	portBindings, err := s.getPortBindings(ctx, request.Name)
	if err != nil {
		return nil, err
	}
	log.Println("success: found container", id)
	return &api.GetContainerResponse{Container: &api.Container{Name: request.Name, Id: id, Owned: s.containers.hasOwnership(request.Name, id), PortBindings: portBindings}}, nil
}

func (s *ContainerServerImpl) getContainerId(ctx context.Context, name string) (string, error) {
	getContainerIdCmd := commands.ContainerInspect{Names: []string{name}, Format: "{{.Id}}"}
	id, stderr, err := s.executor.Execute(ctx, &getContainerIdCmd)
	if id == "" {
		return "", utils.notFound(fmt.Sprintf("Cannot retrieve container ID with name %s", name))
	}
	if stderr != "" {
		return "", utils.toStatusError(stderr)
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *ContainerServerImpl) getPortBindings(ctx context.Context, name string) ([]*api.Container_PortBinding, error) {
	return s.containerLookuper.LookupContainerPortBindings(name)
}

// stopContainers removes containers that are owned by current CTRv2 service in the reverse order of how they are started.
func (s *ContainerServerImpl) stopContainers() {
	containerIds := s.containers.getIdsToClearOwnership()
	if len(containerIds) == 0 {
		log.Println("no containers to clean up")
		return
	}
	log.Printf("stopping containers: %v", s.containers.getMapping())

	// Need to stop container one by one because podman doesn't process a bulk if one of them is dead.
	for _, id := range containerIds {
		log.Printf("stopping container: %s", id)
		cmd := commands.ContainerStop{Names: []string{id}}
		stdout, stderr, _ := cmd.Execute(context.Background())
		if stdout != "" {
			log.Printf("received stdout: %s", stdout)
		}
		if stderr != "" {
			log.Printf("received stderr: %s", stderr)
		}
	}
	s.containers.clear()
}

// removeNetworks removes networks that were created by current CTRv2 service.
func (s *ContainerServerImpl) removeNetworks() {
	networkIds := s.networks.getIdsToClearOwnership()
	if len(networkIds) == 0 {
		log.Println("no networks to clean up")
		return
	}
	log.Printf("removing networks: %v", s.networks.getMapping())
	cmd := commands.NetworkRemove{Names: networkIds}
	stdout, stderr, _ := cmd.Execute(context.Background())
	if stdout != "" {
		log.Printf("received stdout: %s", stdout)
	}
	if stderr != "" {
		log.Printf("received stderr: %s", stderr)
	}
	s.networks.clear()
}

// cleanup removes containers and networks in order to allow graceful shutdown of the CTRv2 service.
func (s *ContainerServerImpl) cleanup() {
	s.stopContainers()
	s.removeNetworks()
}
