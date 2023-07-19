// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_configs

import (
	"context"
	"fmt"
	"strings"

	"infra/cros/cmd/common_lib/interfaces"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// CommandExecutorPairedConfig represents command and executor pair
type CommandExecutorPairedConfig struct {
	CommandType  interfaces.CommandType
	ExecutorType interfaces.ExecutorType
}

// ToString returns string representation of the object.
func (cepc *CommandExecutorPairedConfig) ToString() string {
	if cepc == nil {
		return ""
	}

	return fmt.Sprintf("%s_%s", cepc.CommandType, cepc.ExecutorType)
}

// Config types
const (
	HwTestExecutionConfigType       interfaces.ConfigType = "HwTest"
	PreLocalTestExecutionConfigType interfaces.ConfigType = "PreLocalTest"
	LocalTestExecutionConfigType    interfaces.ConfigType = "LocalTest"

	// For unit tests purposes only
	UnSupportedTestExecutionConfigType interfaces.ConfigType = "UnsupportedTest"
)

// Configs represents main and cleanup configs.
// Cleanup configs are executed if any error occurs with main config execution.
// If any command of cleanup config is already executed,
// they will be skipped during cleanup.
type Configs struct {
	MainConfigs    []*CommandExecutorPairedConfig
	CleanupConfigs []*CommandExecutorPairedConfig
}

// ToString returns string representation of the object.
func (configs *Configs) ToString() string {
	if configs == nil {
		return ""
	}

	retStringList := []string{}

	// Get main configs
	if len(configs.MainConfigs) > 0 {
		retStringList = append(retStringList, "Main Configs:")

		for _, mainConfig := range configs.MainConfigs {
			retStringList = append(retStringList, mainConfig.ToString())
		}
	}

	// Extra space
	if len(retStringList) > 0 {
		retStringList = append(retStringList, "\n")
	}

	// Get cleanup configs
	if len(configs.CleanupConfigs) > 0 {
		retStringList = append(retStringList, "Cleanup Configs:")

		for _, cleanupConfig := range configs.CleanupConfigs {
			retStringList = append(retStringList, cleanupConfig.ToString())
		}
	}

	if len(retStringList) == 0 {
		return ""
	}

	return strings.Join(retStringList, "\n")
}

// TestExecutionConfig represents the configuration for any test execution.
type CmdExecutionConfig struct {
	*interfaces.AbstractTestExecutionConfig

	CommandConfig interfaces.CommandConfigInterface
	StateKeeper   interfaces.StateKeeperInterface
	Configs       *Configs

	executedCommands map[interfaces.CommandType]bool
}

func NewCmdExecutionConfig(
	configType interfaces.ConfigType,
	cmdConfig interfaces.CommandConfigInterface,
	ski interfaces.StateKeeperInterface) *CmdExecutionConfig {

	executedCmdMap := make(map[interfaces.CommandType]bool)
	abstractConfig := interfaces.NewAbstractTestExecutionConfig(configType)
	return &CmdExecutionConfig{
		AbstractTestExecutionConfig: abstractConfig,
		CommandConfig:               cmdConfig,
		StateKeeper:                 ski,
		executedCommands:            executedCmdMap}
}

func (tecfg *CmdExecutionConfig) Execute(ctx context.Context) error {
	if tecfg.Configs == nil || len(tecfg.Configs.MainConfigs) == 0 {
		return fmt.Errorf(
			"Cannot execute nil or empty conifgs. Please generate configs first.",
		)
	}

	// Process main configs
	cmds, err := tecfg.processCommandConfig(ctx, tecfg.Configs.MainConfigs)
	if err != nil {
		return errors.Annotate(
			err,
			"error during processing main configs for config type %s: ", tecfg.GetConfigType()).Err()
	}
	if len(cmds) == 0 {
		return fmt.Errorf("No valid commands found for config type %s.", tecfg.GetConfigType())
	}

	// Process cleanup configs
	cleanupCmds, err := tecfg.processCommandConfig(ctx, tecfg.Configs.CleanupConfigs)
	if err != nil {
		return errors.Annotate(err, "error during processing clean up configs for config type %s: ", tecfg.GetConfigType()).Err()
	}

	err = tecfg.executeCommands(ctx, cmds, false)
	if err != nil {
		// execute clean up commands
		cleanupErr := tecfg.executeCommands(ctx, cleanupCmds, true)
		if cleanupErr != nil {
			err = fmt.Errorf("main error: %w; cleanup error: %s", err, cleanupErr)
		}
		return errors.Annotate(err, "error during executing config type %s: ", tecfg.GetConfigType()).Err()
	}

	return nil
}

// processCommandConfig processes paired configs to specific commands.
func (tecfg *CmdExecutionConfig) processCommandConfig(
	ctx context.Context,
	pairedConfig []*CommandExecutorPairedConfig) ([]interfaces.CommandInterface, error) {

	cmds := []interfaces.CommandInterface{}
	for _, cmdConfig := range pairedConfig {
		cmd, err := tecfg.CommandConfig.GetCommand(cmdConfig.CommandType, cmdConfig.ExecutorType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting command for cmd type %s and executor type %s: ", cmdConfig.CommandType, cmdConfig.ExecutorType).Err()
		}
		logging.Infof(ctx, "Processing cmd: %T", cmd)
		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

// executeCommands executes given commands.
// Assuming the given commands are correct, it will execute repeated ones.
func (tecfg *CmdExecutionConfig) executeCommands(
	ctx context.Context,
	cmds []interfaces.CommandInterface,
	executeAllCmds bool) error {
	var allErr error
	var singleErr error
	for _, cmd := range cmds {
		cmdType := cmd.GetCommandType()
		logging.Infof(ctx, "Executing cmd: %T", cmd)
		if _, ok := tecfg.executedCommands[cmdType]; ok {
			logging.Warningf(ctx, "Command type %s already executed as part of current config. Executing again...", cmdType)
		}

		if singleErr = cmd.ExtractDependencies(ctx, tecfg.StateKeeper); singleErr != nil {
			if executeAllCmds {
				allErr = errors.Append(allErr, singleErr)
			} else {
				return singleErr
			}
		}
		if singleErr = cmd.Execute(ctx); singleErr != nil {
			if executeAllCmds {
				allErr = errors.Append(allErr, singleErr)
			} else {
				logging.Infof(ctx, "Command type %s execution failed. Attempting to update state keeper.", cmdType)
				if innerErr := cmd.UpdateStateKeeper(ctx, tecfg.StateKeeper); innerErr != nil {
					logging.Infof(ctx, "Command type %s could not update state keeper: %s", cmdType, innerErr)
				}
				return singleErr
			}
		}
		logging.Infof(ctx, "Command type %s execution completed. Marking as completed.", cmdType)
		tecfg.executedCommands[cmdType] = true
		if singleErr = cmd.UpdateStateKeeper(ctx, tecfg.StateKeeper); singleErr != nil {
			if executeAllCmds {
				allErr = errors.Append(allErr, singleErr)
			} else {
				return singleErr
			}
		}
	}

	return allErr
}
