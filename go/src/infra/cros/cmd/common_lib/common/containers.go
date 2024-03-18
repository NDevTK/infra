// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"fmt"

	buildapi "go.chromium.org/chromiumos/config/go/build/api"
)

const (
	DefaultCrosFwProvisionSha = "36c32627ae54429d861d0df0fbc4d883170f01d83976ef5be9a0b4e719111aa0"
)

// Create container with provided name and digest, setting repository
// to the hostname `us-docker.pkg.dev` and project `cros-registry/test-services`.
func CreateTestServicesContainer(name, digest string) *buildapi.ContainerImageInfo {
	return &buildapi.ContainerImageInfo{
		Repository: &buildapi.GcrRepository{
			Hostname: "us-docker.pkg.dev",
			Project:  "cros-registry/test-services",
		},
		Name:   name,
		Digest: fmt.Sprintf("sha256:%s", digest),
		Tags:   []string{"prod"},
	}
}

// Set name within images to be a test service container with given name and digest.
func AddTestServiceContainerToImages(images map[string]*buildapi.ContainerImageInfo, name, digest string) {
	images[name] = CreateTestServicesContainer(name, digest)
}
