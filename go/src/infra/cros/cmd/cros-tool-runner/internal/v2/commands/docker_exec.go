// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
)

// DockerExec represents `docker exec`.
// Example of ExecCommand: ["bash", "-c", "echo $HOME && echo $PATH"] prints the
// value of environment variables $HOME and $PATH inside the container.
type DockerExec struct {
	Name        string   // name of container
	ExecCommand []string // command to be executed
}

func (c *DockerExec) Execute(ctx context.Context) (string, string, error) {
	args := []string{"exec", c.Name}
	args = append(args, c.ExecCommand...)
	return execute(ctx, dockerCmd, args)
}
