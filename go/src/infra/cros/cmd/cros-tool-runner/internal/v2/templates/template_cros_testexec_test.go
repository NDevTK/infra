// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package templates

import (
	"strings"
	"testing"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func TestCrosTestPopulate(t *testing.T) {
	processor := newCrosTestProcessor()
	request := getCrosTestTemplateRequest("mynet")

	convertedRequest, err := processor.Process(request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	check(t, convertedRequest.Name, request.Name)
	check(t, convertedRequest.ContainerImage, request.ContainerImage)
	check(t, convertedRequest.AdditionalOptions.Network, "mynet")
	check(t, convertedRequest.AdditionalOptions.Expose[0], "8001")
	check(t, len(convertedRequest.AdditionalOptions.Volume), 2)
	if !strings.Contains(strings.Join(convertedRequest.StartCommand, " "), "cros-test") {
		t.Fatalf("cros-test is not part of start command")
	}
	if !strings.Contains(strings.Join(convertedRequest.StartCommand, " "), "-port 8001") {
		t.Fatalf("-port 8001 is not part of start command")
	}
}

func TestCrosTestDiscoverPort_errorPropagated(t *testing.T) {
	processor := &crosTestProcessor{
		defaultPortDiscoverer: getMockPortDiscovererWithError("error when discover port"),
	}
	request := getCrosTestTemplateRequest("mynet")
	_, err := processor.discoverPort(request)

	if err == nil {
		t.Fatalf("Expected error")
	}
}

func TestCrosTestPopulate_hostNetwork(t *testing.T) {
	processor := newCrosTestProcessor()
	request := getCrosTestTemplateRequest("host")

	convertedRequest, err := processor.Process(request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	check(t, convertedRequest.Name, request.Name)
	check(t, convertedRequest.ContainerImage, request.ContainerImage)
	check(t, convertedRequest.AdditionalOptions.Network, "host")
	check(t, len(convertedRequest.AdditionalOptions.Expose), 0)
	check(t, len(convertedRequest.AdditionalOptions.Volume), 2)
	if !strings.Contains(strings.Join(convertedRequest.StartCommand, " "), "-port 0") {
		t.Fatalf("-port 0 is not part of start command")
	}
}

func TestCrosTestDiscoverPort_bridgeNetwork_populateProtocolOnly(t *testing.T) {
	expected := &api.Container_PortBinding{
		ContainerPort: int32(42),
		Protocol:      protocolTcp,
	}
	processor := &crosTestProcessor{
		defaultPortDiscoverer: getMockPortDiscovererWithSuccess(expected.ContainerPort),
	}
	request := getCrosTestTemplateRequest("mynet")
	binding, err := processor.discoverPort(request)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	check(t, binding.String(), expected.String())
}

func TestCrosTestDiscoverPort_hostNetwork_populateAllFields(t *testing.T) {
	expected := &api.Container_PortBinding{
		ContainerPort: int32(42),
		Protocol:      protocolTcp,
		HostIp:        localhostIp,
		HostPort:      int32(42),
	}
	processor := &crosTestProcessor{
		defaultPortDiscoverer: getMockPortDiscovererWithSuccess(expected.ContainerPort),
	}
	request := getCrosTestTemplateRequest("host")
	binding, err := processor.discoverPort(request)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	check(t, binding.String(), expected.String())
}

func getCrosTestTemplateRequest(network string) *api.StartTemplatedContainerRequest {
	return &api.StartTemplatedContainerRequest{
		Name:           "my-container",
		ContainerImage: "gcr.io/image:123",
		Network:        network,
		ArtifactDir:    "/tmp/unit-tests",
		Template: &api.Template{
			Container: &api.Template_CrosTest{
				CrosTest: &api.CrosTestTemplate{}}}}
}
