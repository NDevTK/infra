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
	TemplateProcessor
	serverPort            string // Default port used in cros-provision
	dockerArtifactDirName string // Path on the drone where service put the logs by default
}

func newCrosDutProcessor() TemplateProcessor {
	return &crosDutProcessor{
		serverPort:            "80",
		dockerArtifactDirName: "/tmp/cros-dut",
	}
}

func (p *crosDutProcessor) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	t := request.Template.GetCrosDut()
	if t == nil {
		return nil, status.Error(codes.Internal, "unable to process")
	}
	volume := fmt.Sprintf("%s:%s", t.ArtifactDir, p.dockerArtifactDirName)
	additionalOptions := &api.StartContainerRequest_Options{
		Network: t.Network,
		Expose:  []string{p.serverPort},
		Volume:  []string{volume},
	}
	startCommand := []string{
		"cros-dut",
		"-dut_address", TemplateUtils.endpointToAddress(t.DutAddress),
		"-cache_address", TemplateUtils.endpointToAddress(t.CacheServer),
		"-port", p.serverPort,
	}
	return &api.StartContainerRequest{Name: request.Name, ContainerImage: request.ContainerImage, AdditionalOptions: additionalOptions, StartCommand: startCommand}, nil
}
