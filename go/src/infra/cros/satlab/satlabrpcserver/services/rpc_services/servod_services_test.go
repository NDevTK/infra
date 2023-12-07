// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package rpc_services

import (
	"context"
	"testing"

	api "go.chromium.org/chromiumos/config/go/test/api"
)

// mock request with all needed information. UFS lookup is not needed.
var mockInput1 = &api.StartServodRequest{
	ServodDockerContainerName: "satlab-host1-docker_servod",
	Board:                     "board1",
	Model:                     "model2",
	SerialName:                "SERIAL123ABC",
	ServodPort:                9999,
}

// mock request without the container name.
var mockInput2 = &api.StartServodRequest{
	ServodDockerContainerName: "",
	Board:                     "board1",
	Model:                     "model2",
	SerialName:                "SERIAL123ABC",
	ServodPort:                9999,
}

func TestGetDutNameFromServodDockerContainerName(t *testing.T) {
	t.Parallel()
	s := createMockServer(t)
	r, err := s.getDutNameFromServodDockerContainerName("satlab-host1-docker_servod")
	if err != nil {
		t.Errorf("should not return error, but got an error: %v", err)
	}
	if r != "satlab-host1" {
		t.Errorf("should return `satlab-host1`, but got `%s`", r)
	}
	_, err = s.getDutNameFromServodDockerContainerName("satlab-host1")
	if err == nil {
		t.Errorf("should return error, but got nil instead")
	}
}

func TestStartServod(t *testing.T) {
	t.Parallel()
	s := createMockServer(t)
	ctx := context.Background()
	err := s.validateStartServodRequest(ctx, mockInput1)
	if err != nil {
		t.Errorf("should not return error, but got an error: %v", err)
	}
	err = s.validateStartServodRequest(ctx, mockInput2)
	if err == nil {
		t.Errorf("should return error, but got nil")
	}
}
