// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package commands hosts all CLI commands CTRv2 interacts with. Each command is
// a struct that represents a CLI command, holds caller-specified arguments, and
// implements an API to uniformly execute itself.
package commands

import (
	"context"
	"os/exec"
	"time"

	"infra/cros/cmd/cros-tool-runner/internal/common"
)

// dockerCmd is the name of docker command. To mimic drone environment locally
// the value can be changed to podman, which is the underlying command on drones
// (docker is an alias to podman on drones).
const dockerCmd = "docker"

// Command is the interface of the command pattern. Only support blocking
// execution for now.
type Command interface {
	Execute(context.Context) (string, string, error)
}

// argumentsComposer is the interface to be implemented by more complicated
// commands to separate composing command from execution.
type argumentsComposer interface {
	// compose returns an array of arguments, error is returned if not composable
	compose() ([]string, error)
}

// execute runs blocking command and returns stdout and stderr as strings.
func execute(ctx context.Context, name string, args []string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	// TODO(mingkong) update RunWithTimeout since timeout is part of ctx
	return common.RunWithTimeout(ctx, cmd, time.Minute, true)
}

// ContextualExecutor executes a command using the provided context.
type ContextualExecutor struct{}

func (*ContextualExecutor) Execute(ctx context.Context, cmd Command) (string, string, error) {
	return cmd.Execute(ctx)
}
