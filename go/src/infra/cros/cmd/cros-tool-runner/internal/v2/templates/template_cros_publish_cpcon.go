// Copyright 2022 The Chromium OS Authors. All rights reserved.
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

const DockerCpconPublishLogsDir = "/tmp/cpcon-publish"
const DockerCpconPublishTestArtifactsDir = "/tmp/cpcon-publish-test-artifacts/"
const DockerCpconPublishPort = "43146"

type crosCpconPublishProcessor struct {
	TemplateProcessor
	cmdExecutor             cmdExecutor
	defaultServerPort       string // Default port used
	dockerArtifactDirName   string // Path on the docker where service put the logs by default
	dockerPublishSrcDirName string // Path on the docker where publish src dir will be mounted to
}

func newCrosCpconPublishProcessor() *crosCpconPublishProcessor {
	return &crosCpconPublishProcessor{
		cmdExecutor:             &commands.ContextualExecutor{},
		defaultServerPort:       DockerCpconPublishPort,
		dockerArtifactDirName:   DockerCpconPublishLogsDir,
		dockerPublishSrcDirName: DockerCpconPublishTestArtifactsDir,
	}
}

// CpconPublish processes commands to Cros-Publish for test results that
// require upload to GCS and display in the CPCon test results console
func (p *crosCpconPublishProcessor) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	t := request.GetTemplate().GetCrosPublish()
	if t == nil {

		return nil, status.Error(codes.Internal, "unable to process")
	}
	volumes := []string{}
	volumes = append(volumes, fmt.Sprintf("%s:%s", request.GetArtifactDir(), p.dockerArtifactDirName))
	volumes = append(volumes, fmt.Sprintf("%s:%s", t.GetPublishSrcDir(), p.dockerPublishSrcDirName))

	port := portZero
	expose := make([]string, 0)
	if request.GetNetwork() != hostNetworkName {
		port = p.defaultServerPort
		expose = append(expose, port)
	}
	additionalOptions := &api.StartContainerRequest_Options{
		Network: request.Network,
		Expose:  expose,
		Volume:  volumes,
	}
	startCommand := []string{
		"cpcon-publish",
		"server",
		"-port", port,
	}
	return &api.StartContainerRequest{Name: request.Name, ContainerImage: request.ContainerImage, AdditionalOptions: additionalOptions, StartCommand: startCommand}, nil
}

func (p *crosCpconPublishProcessor) discoverPort(request *api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error) {
	// delegate to default impl, any template-specific logic should be implemented here.
	return defaultDiscoverPort(p.cmdExecutor, request)
}
