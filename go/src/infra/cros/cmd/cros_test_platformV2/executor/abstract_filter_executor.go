// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executor

import (
	"context"
	"fmt"
	managers "infra/cros/cmd/cros_test_platformV2/docker_managers"

	"go.chromium.org/chromiumos/config/go/test/api"
	"google.golang.org/grpc"
)

type FilterExecutor struct {
	Ctr  managers.ContainerManager
	resp *api.TestSuite

	conn *grpc.ClientConn
}

func (ex *FilterExecutor) Execute(ctx context.Context, cmd string) error {
	if cmd == "run" {
		return nil // Call the (running) binary inside the executing container.
	} else if cmd == "init" {
		// TODO, consider moving this to "process", and adjusting process to meet our needs.
		return nil
	} else if cmd == "stop" {
		return nil // Stop containers.
	}
	return fmt.Errorf("invalid command given: %s\n", cmd)
}
func (ex *FilterExecutor) run() error {
	// Call binary, put the resp in ex.resp = ...
	return nil
}

// init starts the container, creates a client.
func (ex *FilterExecutor) init(binary_name string, container_path string) error {
	fmt.Println("Starting container")
	// Call build the genericTempaltedContainer interface.
	return nil
}
