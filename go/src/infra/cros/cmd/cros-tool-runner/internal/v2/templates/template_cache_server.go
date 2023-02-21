// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package templates

import (
	"fmt"

	"infra/cros/cmd/cros-tool-runner/internal/v2/commands"

	"go.chromium.org/chromiumos/config/go/test/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const DockerCacheServerPort = "43150"
const DockerCacheServerLogsDir = "/tmp/cacheserver"

type cacheServerProcessor struct {
	TemplateProcessor
	cmdExecutor           cmdExecutor
	defaultServerPort     string // Default port used
	dockerArtifactDirName string // Path on the docker where service put the logs by default
}

func newCacheServerProcessor() *cacheServerProcessor {
	return &cacheServerProcessor{
		cmdExecutor:           &commands.ContextualExecutor{},
		defaultServerPort:     DockerCacheServerPort,
		dockerArtifactDirName: DockerCacheServerLogsDir,
	}
}

func (p *cacheServerProcessor) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	t := request.GetTemplate().GetCacheServer()
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
		Volume: []string{
			volume,
		},
	}
	startCommand := []string{
		"cacheserver",
		"-location", DockerCacheServerLogsDir,
		"-port", port,
	}
	return &api.StartContainerRequest{Name: request.Name, ContainerImage: request.ContainerImage, AdditionalOptions: additionalOptions, StartCommand: startCommand}, nil
}

func (p *cacheServerProcessor) discoverPort(request *api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error) {
	// delegate to default impl, any template-specific logic should be implemented here.
	return defaultDiscoverPort(p.cmdExecutor, request)
}
