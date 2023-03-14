// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package templates

import (
	"strings"
	"testing"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func TestCacheServerPopulate_host(t *testing.T) {
	processor := newCacheServerProcessor()
	request := getCacheServerTemplateRequest("host")

	convertedRequest, err := processor.Process(request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	check(t, convertedRequest.Name, request.Name)
	check(t, convertedRequest.ContainerImage, request.ContainerImage)
	check(t, convertedRequest.AdditionalOptions.Network, "host")
	check(t, len(convertedRequest.AdditionalOptions.Expose), 0)
	check(t, len(convertedRequest.AdditionalOptions.Volume), 1)
	if !checkStartCommandContains(convertedRequest.StartCommand, "-port 0") {
		t.Fatalf("-port 0 is not part of start command")
	}
}

func TestCacheServerPopulate_bridge(t *testing.T) {
	processor := newCacheServerProcessor()
	request := getCacheServerTemplateRequest("bridge")

	convertedRequest, err := processor.Process(request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	check(t, convertedRequest.AdditionalOptions.Network, "bridge")
	check(t, len(convertedRequest.AdditionalOptions.Expose), 1)
	if !checkStartCommandContains(convertedRequest.StartCommand, DockerCacheServerPort) {
		t.Fatalf("default port is not part of start command")
	}
}

func TestCacheServerPopulate_serviceAccountKey(t *testing.T) {
	processor := newCacheServerProcessor()
	request := getCacheServerTemplateRequest("host")
	decorateRequestWithServiceAccountKey(request, "/creds/service_account/service-account-chromeos.json")

	convertedRequest, err := processor.Process(request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	check(t, len(convertedRequest.AdditionalOptions.Volume), 2)
	check(t, convertedRequest.AdditionalOptions.Volume[1], "/creds/service_account/:/usr/mount/creds")
	if !checkStartCommandContains(convertedRequest.StartCommand,
		"export GOOGLE_APPLICATION_CREDENTIALS=/usr/mount/creds/service-account-chromeos.json") {
		t.Fatalf("env var is not exported")
	}
}

func TestCacheServerPopulate_serviceAccountKey_invalidFileOnly(t *testing.T) {
	processor := newCacheServerProcessor()
	request := getCacheServerTemplateRequest("host")
	decorateRequestWithServiceAccountKey(request, "service-account-chromeos.json")

	_, err := processor.Process(request)

	if err == nil {
		t.Errorf("expect invalid argument error")
	}
}

func TestCacheServerPopulate_serviceAccountKey_invalidDirOnly(t *testing.T) {
	processor := newCacheServerProcessor()
	request := getCacheServerTemplateRequest("host")
	decorateRequestWithServiceAccountKey(request, "/creds/service_account/")

	_, err := processor.Process(request)

	if err == nil {
		t.Errorf("expect invalid argument error")
	}
}

func decorateRequestWithServiceAccountKey(request *api.StartTemplatedContainerRequest, keyFile string) {
	adc := api.CacheServerTemplate_ServiceAccountKeyfile{ServiceAccountKeyfile: keyFile}
	request.GetTemplate().GetCacheServer().ApplicationDefaultCredentials = &adc
}

func getCacheServerTemplateRequest(network string) *api.StartTemplatedContainerRequest {
	return &api.StartTemplatedContainerRequest{
		Name:           "my-container",
		ContainerImage: "gcr.io/image:123",
		Network:        network,
		ArtifactDir:    "/tmp/unit-tests",
		Template: &api.Template{
			Container: &api.Template_CacheServer{
				CacheServer: &api.CacheServerTemplate{}}}}
}

func checkStartCommandContains(startCommand []string, substr string) bool {
	return strings.Contains(strings.Join(startCommand, " "), substr)
}
