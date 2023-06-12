// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executor

import (
	"context"
	"fmt"
	managers "infra/cros/cmd/cros_test_platformV2/docker_managers"
)

// CloudContainerExecutor represents executor
// for all crostoolrunner (ctr) related commands.
type CloudContainerExecutor struct {
	manager managers.ContainerManager
}

func NewCloudContainerExecutor(manager managers.ContainerManager) *CloudContainerExecutor {
	return &CloudContainerExecutor{manager: manager}
}

func (ex *CloudContainerExecutor) Execute(ctx context.Context, cmd string) error {
	if cmd == "run" {
		return nil
	} else if cmd == "init" {
		ex.manager.StartContainer(ctx, nil)
		return nil
	} else if cmd == "stop" {
		return nil
	}
	return fmt.Errorf("invalid command given: %s\n", cmd)
}
