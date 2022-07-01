// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servod

import (
	"context"
	"infra/cros/recovery/docker"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.chromium.org/luci/common/errors"
)

// TestStartServodContainerStartsContainer tests for the execution of `docker.Start()` in various conditions and ensures we use the correct args to do so
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
		dockerArgs := buildServodContainerArgs(ServodContainerOptions{tc.containerName, "board", "model", "serial", true})

		startServodContainer(context.Background(), &tc.fc, tc.containerName, dockerArgs)
		if tc.fc.containerLaunched != tc.expectContainerLaunch {
			t.Errorf("Expected container launch: %t\nActual container launch: %t\n", tc.expectContainerLaunch, tc.fc.containerLaunched)
			t.Errorf("Testcase: %v", tc)
		}

		if tc.fc.containerLaunched {
			if diff := cmp.Diff(tc.fc.runningContainers[tc.containerName], *dockerArgs); diff != "" {
				t.Errorf("Container launched with unexpected args: %s", diff)
			}
		}
	}
}

// TestBuildServodDockerArgs tests that when we build the correct container args
func TestBuildServodDockerArgs(t *testing.T) {
	t.Setenv("SERVOD_CONTAINER_LABEL", "latest") // functionality under test relies on env

	type test struct {
		opts               ServodContainerOptions
		expectedDockerArgs *docker.ContainerArgs
	}

	tests := []test{
		{ServodContainerOptions{"test_container", "board", "model", "serial", false},
			&docker.ContainerArgs{
				Detached:     true,
				ImageName:    "us-docker.pkg.dev/chromeos-partner-moblab/common-core/servod:latest",
				PublishPorts: nil,
				ExposePorts:  nil,
				EnvVar:       []string{"BOARD=board", "MODEL=model", "SERVO_SERIAL=serial", "PORT=9999"},
				Volumes:      []string{"/dev:/dev", "serial_log:/var/log/servod_9999/"},
				Network:      "default_satlab",
				Privileged:   true,
				Exec:         []string{"tail", "-f", "/dev/null"},
			},
		},
		{ServodContainerOptions{"test_container", "board2", "model2", "serial2", true},
			&docker.ContainerArgs{
				Detached:     true,
				ImageName:    "us-docker.pkg.dev/chromeos-partner-moblab/common-core/servod:latest",
				PublishPorts: nil,
				ExposePorts:  nil,
				EnvVar:       []string{"BOARD=board2", "MODEL=model2", "SERVO_SERIAL=serial2", "PORT=9999"},
				Volumes:      []string{"/dev:/dev", "serial2_log:/var/log/servod_9999/"},
				Network:      "default_satlab",
				Privileged:   true,
				Exec:         []string{"bash", "/start_servod.sh"},
			},
		},
	}

	for _, tc := range tests {
		dockerArgs := buildServodContainerArgs(tc.opts)

		if diff := cmp.Diff(dockerArgs, tc.expectedDockerArgs); diff != "" {
			t.Errorf("Unexpected docker args created: %s", diff)
		}
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
