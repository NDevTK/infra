// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"

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
	case commands.BuildInputValidationCmdType:
		cmd = commands.NewBuildInputValidationCmd()

	case commands.ParseEnvInfoCmdType:
		cmd = commands.NewParseEnvInfoCmd()

	case commands.InvServiceStartCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewInvServiceStartCmd(exec)

	case commands.InvServiceStopCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewInvServiceStopCmd(exec)

	case commands.LoadDutTopologyCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewLoadDutTopologyCmd(exec)

	case commands.CtrServiceStartAsyncCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewCtrServiceStartAsyncCmd(exec)

	case commands.CtrServiceStopCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewCtrServiceStopCmd(exec)

	case commands.GcloudAuthCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewGcloudAuthCmd(exec)

	case commands.DutServiceStartCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewDutServiceStartCmd(exec)

	case commands.ProvisionServiceStartCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewProvisionServiceStartCmd(exec)

	case commands.ProvisonInstallCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewProvisionInstallCmd(exec)

	case commands.TestServiceStartCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewTestServiceStartCmd(exec)

	case commands.TestsExecutionCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewTestsExecutionCmd(exec)

	case commands.GcsPublishStartCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewGcsPublishServiceStartCmd(exec)

	case commands.GcsPublishUploadCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewGcsPublishUploadCmd(exec)

	case commands.RdbPublishStartCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewRdbPublishServiceStartCmd(exec)

	case commands.RdbPublishUploadCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewRdbPublishUploadCmd(exec)

	case commands.TkoPublishStartCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewTkoPublishServiceStartCmd(exec)

	case commands.TkoPublishUploadCmdType:
		exec, err := cfg.ExecutorConfig.GetExecutor(execType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting executor for command type %s: ", cmdType).Err()
		}
		cmd = commands.NewTkoPublishUploadCmd(exec)

	case commands.ProcessResultsCmdType:
		cmd = commands.NewProcessResultsCmd()

	case commands.UpdateDutStateCmdType:
		cmd = commands.NewUpdateDutStateCmd()

	default:
		return nil, fmt.Errorf("Command type %s not supported in command configs!", cmdType)
	}

	cfg.commandsMap[cmdType] = cmd
	return cmd, nil
}
