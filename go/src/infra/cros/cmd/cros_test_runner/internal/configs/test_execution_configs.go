// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"
	"fmt"

	"infra/cros/cmd/cros_test_runner/internal/interfaces"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// Config types
const (
	HwTestExecutionConfigType interfaces.ConfigType = "HwTest"

	// For unit tests purposes only
	UnSupportedTestExecutionConfigType interfaces.ConfigType = "UnsupportedTest"
)

// Configs represents main and cleanup configs.
// Cleanup configs are executed if any error occurs with main config execution.
// If any command of cleanup config is already executed, they will be skipped during cleanup.
type Configs struct {
	MainConfigs    []*CommandExecutorPairedConfig
	CleanupConfigs []*CommandExecutorPairedConfig
}

// TestExecutionConfig represents the configuration for any test execution.
type TestExecutionConfig struct {
	interfaces.AbstractTestExecutionConfig

	commandConfig    interfaces.CommandConfigInterface
	stateKeeper      interfaces.StateKeeperInterface
	configs          *Configs
	executedCommands map[interfaces.CommandType]bool
}

func NewTestExecutionConfig(configType interfaces.ConfigType, cmdConfig interfaces.CommandConfigInterface, ski interfaces.StateKeeperInterface) TestExecutionConfig {
	executedCmdMap := make(map[interfaces.CommandType]bool)
	abstractConfig := interfaces.NewAbstractTestExecutionConfig(configType)
	return TestExecutionConfig{AbstractTestExecutionConfig: abstractConfig, commandConfig: cmdConfig, stateKeeper: ski, executedCommands: executedCmdMap}
}

func (tecfg *TestExecutionConfig) GenerateConfig(ctx context.Context) error {
	switch configType := tecfg.GetConfigType(); configType {
	case HwTestExecutionConfigType:
		tecfg.configs = GenerateHwConfigs(ctx)
	default:
		return fmt.Errorf("Config type %s is not supported!", configType)
	}

	return nil
}

func (tecfg *TestExecutionConfig) Execute(ctx context.Context) error {
	if tecfg.configs == nil || len(tecfg.configs.MainConfigs) == 0 {
		return fmt.Errorf("Cannot execute nil or empty conifgs. Please generate configs first.")
	}

	// Process main configs
	cmds, err := tecfg.processCommandConfig(ctx, tecfg.configs.MainConfigs)
	if err != nil {
		return errors.Annotate(err, "error during processing main configs for config type %s: ", tecfg.GetConfigType()).Err()
	}
	if len(cmds) == 0 {
		return fmt.Errorf("No valid commands found for config type %s.", tecfg.GetConfigType())
	}

	// Process cleanup configs
	cleanupCmds, err := tecfg.processCommandConfig(ctx, tecfg.configs.CleanupConfigs)
	if err != nil {
		return errors.Annotate(err, "error during processing clean up configs for config type %s: ", tecfg.GetConfigType()).Err()
	}

	err = tecfg.executeCommands(ctx, cmds)
	if err != nil {
		// execute clean up commands
		cleanupErr := tecfg.executeCommands(ctx, cleanupCmds)
		if cleanupErr != nil {
			err = fmt.Errorf("main error: %w; cleanup error: %s", err, cleanupErr)
		}
		return errors.Annotate(err, "error during executing config type %s: ", tecfg.GetConfigType()).Err()
	}

	return nil
}

// processCommandConfig processes paired configs to specific commands.
func (tecfg *TestExecutionConfig) processCommandConfig(ctx context.Context, pairedConfig []*CommandExecutorPairedConfig) ([]interfaces.CommandInterface, error) {
	cmds := []interfaces.CommandInterface{}
	for _, cmdConfig := range pairedConfig {
		cmd, err := tecfg.commandConfig.GetCommand(cmdConfig.CommandType, cmdConfig.ExecutorType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting command for cmd type %s and executor type %s: ", cmdConfig.CommandType, cmdConfig.ExecutorType).Err()
		}
		logging.Infof(ctx, "Processing cmd: %T", cmd)
		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

// executeCommands executes given commands.
// It will skip any commands that are already executed.
func (tecfg *TestExecutionConfig) executeCommands(ctx context.Context, cmds []interfaces.CommandInterface) error {
	var err error
	for _, cmd := range cmds {
		cmdType := cmd.GetCommandType()
		logging.Infof(ctx, "Executing cmd: %T", cmd)
		if _, ok := tecfg.executedCommands[cmdType]; ok {
			logging.Infof(ctx, "Command type %s already executed as part of current config. Skipping...", cmdType)
			continue
		}

		if err = cmd.ExtractDependencies(ctx, tecfg.stateKeeper); err != nil {
			return err
		}
		if err = cmd.Execute(ctx); err != nil {
			return err
		}
		tecfg.executedCommands[cmdType] = true
		if err = cmd.UpdateStateKeeper(ctx, tecfg.stateKeeper); err != nil {
			return err
		}
	}

	return nil
}
