// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"testing"
	"time"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func TestGetContainerImageNameFrom_LocationWithImplicitTag(t *testing.T) {
	expected := "cacheserver"
	location := "us-docker.pkg.dev/cros-registry/test-services/cacheserver"

	actual := getContainerImageNameFrom(location)

	if actual != expected {
		t.Errorf("getContainerImageNameFrom(%q) = %q, want %q", location, actual, expected)
	}
}

func TestGetContainerImageNameFrom_LocationWithLatestTag(t *testing.T) {
	expected := "cros-test"
	location := "us-docker.pkg.dev/cros-registry/test-services/cros-test:latest"

	actual := getContainerImageNameFrom(location)

	if actual != expected {
		t.Errorf("getContainerImageNameFrom(%q) = %q, want %q", location, actual, expected)
	}
}

func TestGetContainerImageNameFrom_LocationWithBuildIdAsTag(t *testing.T) {
	expected := "cros-test"
	location := "us-docker.pkg.dev/cros-registry/test-services/cros-test:8811903382633993457"

	actual := getContainerImageNameFrom(location)

	if actual != expected {
		t.Errorf("getContainerImageNameFrom(%q) = %q, want %q", location, actual, expected)
	}
}

func TestGetContainerImageNameFrom_LocationWithDigest(t *testing.T) {
	expected := "cros-dut"
	location := "us-docker.pkg.dev/cros-registry/test-services/cros-dut@sha256:43622ab765c7ba0"

	actual := getContainerImageNameFrom(location)

	if actual != expected {
		t.Errorf("getContainerImageNameFrom(%q) = %q, want %q", location, actual, expected)
	}
}

func TestGetContainerImageNameFrom_UnsupportedLocation(t *testing.T) {
	expected := "UNKNOWN"
	location := "ubuntu"

	actual := getContainerImageNameFrom(location)

	if actual != expected {
		t.Errorf("getContainerImageNameFrom(%q) = %q, want %q", location, actual, expected)
	}
}

func TestGetContainerImageNameFrom_InvalidLocation(t *testing.T) {
	expected := "UNKNOWN"
	location := "@:se//"

	actual := getContainerImageNameFrom(location)

	if actual != expected {
		t.Errorf("getContainerImageNameFrom(%q) = %q, want %q", location, actual, expected)
	}
}

func TestGetContainerImageNameFrom_EmptyLocation(t *testing.T) {
	expected := "UNKNOWN"
	location := ""

	actual := getContainerImageNameFrom(location)

	if actual != expected {
		t.Errorf("getContainerImageNameFrom(%q) = %q, want %q", location, actual, expected)
	}
}

// TODO(mingkong): add real test coverage when refactoring metrics after v1 deprecation
func TestMonitorTime_NotBroken(t *testing.T) {
	cmdPull := &DockerPull{
		ContainerImage: "us-docker.pkg.dev/cros-registry/test-services/cros-test",
	}
	cmdRun := &DockerRun{&api.StartContainerRequest{
		ContainerImage: "us-docker.pkg.dev/cros-registry/test-services/cros-test",
	}}
	cmdAuth := &DockerLogin{}
	monitorTime(cmdPull, time.Now())
	monitorTime(cmdRun, time.Now())
	monitorTime(cmdAuth, time.Now())
}

// TODO(mingkong): add real test coverage when refactoring metrics after v1 deprecation
func TestMonitorStatus_NotBroken(t *testing.T) {
	cmdRun := &DockerRun{&api.StartContainerRequest{
		ContainerImage: "us-docker.pkg.dev/cros-registry/test-services/cros-test",
	}}
	monitorStatus(cmdRun, statusFail)
	monitorStatus(cmdRun, statusPass)
}
