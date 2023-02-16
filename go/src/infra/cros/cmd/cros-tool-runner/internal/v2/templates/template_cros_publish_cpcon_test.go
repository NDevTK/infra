// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package templates

import (
	"testing"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func TestCrosPublishPopulate(t *testing.T) {
	processor := newCrosCpconPublishProcessor()
	request := getCrosCpconPublishTemplateRequest("mynet")

	convertedRequest, err := processor.Process(request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	check(t, convertedRequest.Name, request.Name)
	check(t, convertedRequest.ContainerImage, request.ContainerImage)
	check(t, convertedRequest.AdditionalOptions.Network, "mynet")
	check(t, convertedRequest.AdditionalOptions.Expose[0], "43146")
	check(t, convertedRequest.AdditionalOptions.Volume[0], "/tmp:/tmp/cpcon-publish")
	check(t, convertedRequest.StartCommand[0], "cpcon-publish")
	check(t, convertedRequest.StartCommand[len(convertedRequest.StartCommand)-1],
		"43146")
}

func TestCrosPublishDiscoverPort_errorPropagated(t *testing.T) {
	processor := newCrosCpconPublishProcessor()
	request := getCrosCpconPublishTemplateRequest("mynet")
	_, err := processor.discoverPort(request)

	if err == nil {
		t.Fatalf("Expected error")
	}
}

func TestCrosCpconPublishPopulate_hostNetwork(t *testing.T) {
	processor := newCrosCpconPublishProcessor()
	request := getCrosCpconPublishTemplateRequest("host")

	convertedRequest, err := processor.Process(request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	check(t, convertedRequest.Name, request.Name)
	check(t, convertedRequest.ContainerImage, request.ContainerImage)
	check(t, convertedRequest.AdditionalOptions.Network, "host")
	check(t, len(convertedRequest.AdditionalOptions.Expose), 0)
	check(t, convertedRequest.AdditionalOptions.Volume[0], "/tmp:/tmp/cpcon-publish")
	check(t, convertedRequest.StartCommand[0], "cpcon-publish")
	check(t, convertedRequest.StartCommand[len(convertedRequest.StartCommand)-1],
		"0")
}

func getCrosCpconPublishTemplateRequest(network string) *api.StartTemplatedContainerRequest {
	return &api.StartTemplatedContainerRequest{
		Name:           "my-container",
		ContainerImage: "gcr.io/image:123",
		Network:        network,
		ArtifactDir:    "/tmp",
		Template: &api.Template{
			Container: &api.Template_CrosPublish{
				CrosPublish: &api.CrosPublishTemplate{
					PublishType: api.CrosPublishTemplate_PUBLISH_CPCON,
				}}}}
}
