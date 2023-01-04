// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package templates

import (
	"fmt"
	"os"
	"path/filepath"

	"go.chromium.org/chromiumos/config/go/test/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const DockerRdbPublishLogsDir = "/tmp/rdb-publish/"
const DockerRdbPublishLuciContextDir = "/tmp/rdb-luci-context/"
const DockerRdbLuciContextDir = "/tmp/rdb-luci-context/"
const DockerRdbPublishPort = "43149"

const LuciContext = "LUCI_CONTEXT"

type crosRdbPublishProcessor struct {
	TemplateProcessor
	defaultPortDiscoverer    portDiscoverer
	defaultServerPort        string // Default port used
	dockerArtifactDirName    string // Path on the docker where service put the logs by default
	dockerPublishLuciDirName string // Path on the docker where publish src dir will be mounted to

}

func newCrosRdbPublishProcessor() TemplateProcessor {
	return &crosRdbPublishProcessor{
		defaultPortDiscoverer:    &defaultPortDiscoverer{},
		defaultServerPort:        DockerRdbPublishPort,
		dockerArtifactDirName:    DockerRdbPublishLogsDir,
		dockerPublishLuciDirName: DockerRdbLuciContextDir,
	}
}

func (p *crosRdbPublishProcessor) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	t := request.GetTemplate().GetCrosPublish()
	if t == nil {

		return nil, status.Error(codes.Internal, "unable to process")
	}

	volumes := []string{}
	envVars := []string{}

	volumes = append(volumes, fmt.Sprintf("%s:%s", request.GetArtifactDir(), p.dockerArtifactDirName))
	// Set LUCI_CONTEXT inside container
	if luciContextLoc, present := os.LookupEnv(LuciContext); present == true {
		luciContextParentDir := filepath.Dir(luciContextLoc)
		luciContextBase := filepath.Base(luciContextLoc)
		envVars = append(envVars, fmt.Sprintf("%s=%s", LuciContext, filepath.Join(p.dockerPublishLuciDirName, luciContextBase)))
		volumes = append(volumes, fmt.Sprintf("%s:%s", luciContextParentDir, p.dockerPublishLuciDirName))
	}

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
		Env:     envVars,
	}
	startCommand := []string{
		"rdb-publish",
		"server",
		"-port", port,
	}
	return &api.StartContainerRequest{Name: request.Name, ContainerImage: request.ContainerImage, AdditionalOptions: additionalOptions, StartCommand: startCommand}, nil
}

func (p *crosRdbPublishProcessor) discoverPort(request *api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error) {
	t := request.GetTemplate().GetCrosPublish()
	if t == nil {
		return nil, status.Error(codes.Internal, "unable to process")
	}
	portBinding, err := p.defaultPortDiscoverer.discoverPort(request)
	if err != nil {
		return portBinding, err
	}
	if request.Network == hostNetworkName {
		portBinding.HostPort = portBinding.ContainerPort
		portBinding.HostIp = localhostIp
	}
	portBinding.Protocol = protocolTcp
	return portBinding, nil
}
