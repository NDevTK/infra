// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package templates

import (
	"fmt"

	"go.chromium.org/chromiumos/config/go/test/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type crosDutProcessor struct {
	defaultPortDiscoverer portDiscoverer
	defaultServerPort     string // Default port used in cros-provision
	dockerArtifactDirName string // Path on the drone where service put the logs by default
}

func newCrosDutProcessor() TemplateProcessor {
	return &crosDutProcessor{
		defaultPortDiscoverer: &defaultPortDiscoverer{},
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
	t := request.GetTemplate().GetCrosDut()
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
