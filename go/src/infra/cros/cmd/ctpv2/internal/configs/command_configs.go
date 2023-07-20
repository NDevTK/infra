// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"fmt"

	"infra/cros/cmd/common_lib/common_commands"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/ctpv2/internal/commands"

	"go.chromium.org/luci/common/errors"
)

// CommandConfig represents command config.
type CommandConfig struct {
	ExecutorConfig interfaces.ExecutorConfigInterface

	commandsMap map[interfaces.CommandType]interfaces.CommandInterface
}

func NewCommandConfig(execConfig interfaces.ExecutorConfigInterface) interfaces.CommandConfigInterface {
	cmdMap := make(map[interfaces.CommandType]interfaces.CommandInterface)
	return &CommandConfig{ExecutorConfig: execConfig, commandsMap: cmdMap}
}

// GetCommand returns the concrete command based on provided command and executor type.
func (cfg *CommandConfig) GetCommand(
	cmdType interfaces.CommandType,
	execType interfaces.ExecutorType) (interfaces.CommandInterface, error) {

	// Return cmd if already created.
	if savedCmd, ok := cfg.commandsMap[cmdType]; ok {
		return savedCmd, nil
	}

	var cmd interfaces.CommandInterface

	// Get cmd based on cmd type.
	switch cmdType {
	case commands.TranslateRequestType:
		cmd = commands.NewTranslateRequestCmd()

	case commands.FilterStartCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewFilterStartCmd(exec)

	case commands.FilterExecutionCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewFilterExecutionCmd(exec)

	case common_commands.CtrServiceStartAsyncCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = common_commands.NewCtrServiceStartAsyncCmd(exec)

	case common_commands.CtrServiceStopCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = common_commands.NewCtrServiceStopCmd(exec)

	case common_commands.GcloudAuthCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = common_commands.NewGcloudAuthCmd(exec)

	default:
		return nil, fmt.Errorf("Command type %s not supported in command configs!", cmdType)
	}

	cfg.commandsMap[cmdType] = cmd
	return cmd, nil
}
