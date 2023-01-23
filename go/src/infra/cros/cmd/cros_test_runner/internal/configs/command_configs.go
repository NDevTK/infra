// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// CommandConfig represents command config.
type CommandConfig struct {
	ExecutorConfig interfaces.ExecutorConfigInterface

	commandsMap map[interfaces.CommandType]interfaces.CommandInterface
}

func NewCommandConfig(execConfig interfaces.ExecutorConfigInterface) interfaces.CommandConfigInterface {
	return &CommandConfig{ExecutorConfig: execConfig, commandsMap: make(map[interfaces.CommandType]interfaces.CommandInterface)}
}

// GetCommand returns the concrete command based on provided command and executor type.
func (cfg *CommandConfig) GetCommand(cmdType interfaces.CommandType, execType interfaces.ExecutorType) (interfaces.CommandInterface, error) {
	// Return cmd if already created.
	if savedCmd, ok := cfg.commandsMap[cmdType]; ok {
		return savedCmd, nil
	}

	var cmd interfaces.CommandInterface

	// Get cmd based on cmd type.
	switch cmdType {
	case commands.BuildInputValidationCmdType:
		cmd = commands.NewBuildInputValidationCmd()

	case commands.ParseEnvInfoCmdType:
		cmd = commands.NewParseEnvInfoCmd()

	default:
		return nil, fmt.Errorf("Command type %s not supported in command configs!", cmdType)
	}

	cfg.commandsMap[cmdType] = cmd
	return cmd, nil
}
