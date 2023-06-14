// Copyright 2023 The Chromium Authors
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

type genericProcessor struct {
	cmdExecutor cmdExecutor
}

func newGenericProcessor() *genericProcessor {
	return &genericProcessor{
		cmdExecutor: &commands.ContextualExecutor{},
	}
}

func (p *genericProcessor) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	t := request.GetTemplate().GetGeneric()
	if t == nil {
		return nil, status.Error(codes.Internal, "unable to process")
	}

	artifactVolume := fmt.Sprintf("%s:%s", request.ArtifactDir, t.DockerArtifactDir)
	volumes := []string{}
	volumes = append(volumes, artifactVolume)
	volumes = append(volumes, t.AdditionalVolumes...)
	additionalOptions := &api.StartContainerRequest_Options{
		Network: request.Network,
		Expose:  t.Expose,
		Volume:  volumes,
		Env:     t.Env,
	}
	startCommand := []string{
		t.BinaryName,
	}
	startCommand = append(startCommand, t.BinaryArgs...)
	return &api.StartContainerRequest{Name: request.Name, ContainerImage: request.ContainerImage, AdditionalOptions: additionalOptions, StartCommand: startCommand}, nil
}

func (p *genericProcessor) discoverPort(request *api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error) {
	// delegate to default impl, any template-specific logic should be implemented here.
	return defaultDiscoverPort(p.cmdExecutor, request)
}
