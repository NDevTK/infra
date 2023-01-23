// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"

	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// CommandExecutorPairedConfig represents command and executor pair
type CommandExecutorPairedConfig struct {
	CommandType  interfaces.CommandType
	ExecutorType interfaces.ExecutorType
}

// All currently supported command-executor pairs.
var InputValidation_NoExecutor = &CommandExecutorPairedConfig{CommandType: commands.BuildInputValidationCmdType, ExecutorType: executors.NoExecutorType}
var ParseEnvInfo_NoExecutor = &CommandExecutorPairedConfig{CommandType: commands.ParseEnvInfoCmdType, ExecutorType: executors.NoExecutorType}

// GenerateHwConfigs generates hw tests execution for lab environment.
func GenerateHwConfigs(ctx context.Context) *Configs {
	mainConfigs := []*CommandExecutorPairedConfig{
		InputValidation_NoExecutor,
		ParseEnvInfo_NoExecutor,
	}

	// Clean up configs. They will be executed if any failures occurs in main configs.
	// If any of the cleanup cmd is already executed, they will be skipped.
	cleanupConfigs := []*CommandExecutorPairedConfig{}

	return &Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}
