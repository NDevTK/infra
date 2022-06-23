// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servod

import (
	"context"
	"infra/cros/recovery/docker"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.chromium.org/luci/common/errors"
)

// TestStartServodContainerStartsContainer tests solely for the execution of `docker.Start()` in various conditions
func TestStartServodContainerStartsContainer(t *testing.T) {
	t.Parallel()

	type test struct {
		fc                    FakeDockerClient
		containerName         string
		expectContainerLaunch bool
	}

	tests := []test{
		{NewFakeDockerClient(false, false), "testContainer", true},                       // no errs, should launch
		{NewFakeDockerClient(true, false), "testContainer", false},                       // error when calling docker.Up()
		{NewFakeDockerClient(false, true), "testContainer", false},                       // error when calling docker.Start()
		{NewFakeDockerClient(false, false, "testContainer"), "testContainer", false},     // testContainer is already running
		{NewFakeDockerClient(false, false, "testContainer"), "differentContainer", true}, // a container, but not the container, is running
	}

	for _, tc := range tests {
		StartServodContainer(&tc.fc, context.Background(), tc.containerName, "board", "model", "serial")
		if tc.fc.containerLaunched != tc.expectContainerLaunch {
			t.Errorf("Expected container launch: %t\nActual container launch: %t\n", tc.expectContainerLaunch, tc.fc.containerLaunched)
			t.Errorf("Testcase: %v", tc)
		}
	}
}

// TestStartServodContainerArgs tests that when we start a docker container it does so with the expected ContainerArgs
// Could ~likely~ do this in the same test as above but this keeps the test cases simpler
func TestStartServodContainerArgs(t *testing.T) {
	fc := NewFakeDockerClient(false, false)

	// intentionally set to non-default value to test that functionality
	os.Setenv("SERVOD_CONTAINER_LABEL", "latest")

	StartServodContainer(&fc, context.Background(), "testContainer", "board", "model", "serial")

	expectedArgs := docker.ContainerArgs{
		Detached:     true,
		ImageName:    "us-docker.pkg.dev/chromeos-partner-moblab/common-core/servod:latest",
		PublishPorts: nil,
		ExposePorts:  nil,
		EnvVar:       []string{"BOARD=board", "MODEL=model", "SERVO_SERIAL=serial", "PORT=9999"},
		Volumes:      []string{"/dev:/dev", "serial_log:/var/log/servod_9999/"},
		Network:      "default_satlab",
		Privileged:   true,
		Exec:         []string{"bash", "/start_servod.sh"},
	}

	if diff := cmp.Diff(fc.runningContainers["testContainer"], expectedArgs); diff != "" {
		t.Errorf("unexpected diff: %s", diff)
	}
}

// Implementation of docker client. Allows specifying whether docker.Up or docker.Start calls will error
// Also keeps a map of "running" containers, and notes if a container was launched after initialization
type FakeDockerClient struct {
	runningContainers map[string]docker.ContainerArgs
	isUpErr           bool // whether the d.Up call errs
	startContainerErr bool // whether the d.Start call errs
	containerLaunched bool // whether a call to Start launched the container
}

// NewFakeDockerClient is the recommended way of creating a FakeDockerClient because it handles the map initialization in runningContainers
func NewFakeDockerClient(isUpErr bool, startContainerErr bool, containerNames ...string) FakeDockerClient {
	var fc FakeDockerClient

	fc.isUpErr = isUpErr
	fc.startContainerErr = startContainerErr
	fc.runningContainers = make(map[string]docker.ContainerArgs)

	for _, cn := range containerNames {
		fc.runningContainers[cn] = docker.ContainerArgs{}
	}

	return fc
}

// Fake implementation of `Start` which adds a container to `runningContainers` and sets `containerLaunched` to be true
// Will error if FakeClient.startContainerErr is true
func (fc *FakeDockerClient) Start(ctx context.Context, containerName string, req *docker.ContainerArgs, timeout time.Duration) (*docker.StartResponse, error) {
	if fc.startContainerErr {
		return nil, errors.Reason("Error during Start execution").Err()
	}

	fc.runningContainers[containerName] = *req
	fc.containerLaunched = true
	return &docker.StartResponse{}, nil
}

// Fake implementation of `IsUp` which uses the `runningContainers` map's keys to decide if a given container "isUp"
// Will error if FakeClient.isUpErr is true
func (fc *FakeDockerClient) IsUp(ctx context.Context, containerName string) (bool, error) {
	if fc.isUpErr {
		return false, errors.Reason("Error during IsUp execution").Err()
	}

	_, keyFoundOk := fc.runningContainers[containerName]
	return keyFoundOk, nil
}
