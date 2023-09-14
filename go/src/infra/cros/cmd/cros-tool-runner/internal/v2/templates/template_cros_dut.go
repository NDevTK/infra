// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package templates

import (
	"fmt"

	"go.chromium.org/chromiumos/config/go/test/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"infra/cros/cmd/cros-tool-runner/internal/v2/commands"
)

type crosDutProcessor struct {
	cmdExecutor           cmdExecutor
	defaultServerPort     string // Default port used in cros-provision
	dockerArtifactDirName string // Path on the drone where service put the logs by default
}

func newCrosDutProcessor() *crosDutProcessor {
	return &crosDutProcessor{
		cmdExecutor:           &commands.ContextualExecutor{},
		defaultServerPort:     "80",
		dockerArtifactDirName: "/tmp/cros-dut",
	}
}

func (p *crosDutProcessor) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	t := request.GetTemplate().GetCrosDut()
	if t == nil {

		return nil, status.Error(codes.Internal, "unable to process")
	}
	volume := fmt.Sprintf("%s:%s", request.ArtifactDir, p.dockerArtifactDirName)
	port := portZero
	expose := make([]string, 0)
	if request.Network != hostNetworkName {
		port = p.defaultServerPort
		expose = append(expose, port)
	}
	additionalOptions := &api.StartContainerRequest_Options{
		Network: request.Network,
		Expose:  expose,
		Volume:  []string{volume},
	}
	startCommand := []string{
		"cros-dut",
		"-dut_address", TemplateUtils.endpointToAddress(t.DutAddress),
		"-cache_address", TemplateUtils.endpointToAddress(t.CacheServer),
		"-port", port,
	}
	return &api.StartContainerRequest{Name: request.Name, ContainerImage: request.ContainerImage, AdditionalOptions: additionalOptions, StartCommand: startCommand}, nil
}

func (p *crosDutProcessor) discoverPort(request *api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error) {
	// delegate to default impl, any template-specific logic should be implemented here.
	return defaultDiscoverPort(p.cmdExecutor, request)
}
