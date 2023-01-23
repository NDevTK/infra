// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package interfaces

import (
	"context"
)

// Command type
type CommandType string

// CommandInterface defines the contract a command will have to satisfy.
type CommandInterface interface {
	// GetCommandType returns the command type.
	GetCommandType() CommandType

	// ExtractDependencies extracts command dependencies from the provided state keeper before command execution.
	ExtractDependencies(context.Context, StateKeeperInterface) error

	// Execute executes the command.
	Execute(context.Context) error

	// UpdateStateKeeper updates the provided state keeper after command execution.
	UpdateStateKeeper(context.Context, StateKeeperInterface) error
}

// AbstractCmd satisfies the command requirement that is common to all.
type AbstractCmd struct {
	CommandInterface

	cmdType CommandType
}

func NewAbstractCmd(cmdType CommandType) *AbstractCmd {
	return &AbstractCmd{cmdType: cmdType}
}

func (cmd *AbstractCmd) GetCommandType() CommandType {
	return cmd.cmdType
}

func (cmd *AbstractCmd) ExtractDependencies(ctx context.Context, ski StateKeeperInterface) error {
	// No deps (Sub classes will overwrite if necessary)
	return nil
}

func (cmd *AbstractCmd) UpdateStateKeeper(ctx context.Context, ski StateKeeperInterface) error {
	// Nothing to update (Sub classes will overwrite if necessary)
	return nil
}

// AbstractSingleCmdByNoExecutor represents a single command with no executor.
type AbstractSingleCmdByNoExecutor struct {
	*AbstractCmd
}
