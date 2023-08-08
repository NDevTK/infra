// Copyright 2022 The Chromium Authors
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
