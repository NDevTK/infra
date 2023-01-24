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
var InvServiceStart_InvExecutor = &CommandExecutorPairedConfig{CommandType: commands.InvServiceStartCmdType, ExecutorType: executors.InvServiceExecutorType}
var InvServiceStop_InvExecutor = &CommandExecutorPairedConfig{CommandType: commands.InvServiceStopCmdType, ExecutorType: executors.InvServiceExecutorType}
var LoadDutTopology_InvExecutor = &CommandExecutorPairedConfig{CommandType: commands.LoadDutTopologyCmdType, ExecutorType: executors.InvServiceExecutorType}
var CtrStartAsync_CtrExecutor = &CommandExecutorPairedConfig{CommandType: commands.CtrServiceStartAsyncCmdType, ExecutorType: executors.CtrExecutorType}
var CtrStop_CtrExecutor = &CommandExecutorPairedConfig{CommandType: commands.CtrServiceStopCmdType, ExecutorType: executors.CtrExecutorType}
var GcloudAuth_CtrExecutor = &CommandExecutorPairedConfig{CommandType: commands.GcloudAuthCmdType, ExecutorType: executors.CtrExecutorType}

// GenerateHwConfigs generates hw tests execution for lab environment.
func GenerateHwConfigs(ctx context.Context) *Configs {
	mainConfigs := []*CommandExecutorPairedConfig{
		InputValidation_NoExecutor,
		ParseEnvInfo_NoExecutor,
		InvServiceStart_InvExecutor,
		LoadDutTopology_InvExecutor,
		InvServiceStop_InvExecutor,
		CtrStartAsync_CtrExecutor,
		GcloudAuth_CtrExecutor,
		CtrStop_CtrExecutor,
	}

	// Clean up configs. They will be executed if any failures occurs in main configs.
	// If any of the cleanup cmd is already executed, they will be skipped.
	cleanupConfigs := []*CommandExecutorPairedConfig{
		CtrStop_CtrExecutor,
	}

	return &Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}
