// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"
	"infra/cros/cmd/common_lib/common_commands"
	"infra/cros/cmd/common_lib/common_configs"
	"infra/cros/cmd/common_lib/common_executors"
	"infra/cros/cmd/ctpv2/internal/commands"
	"infra/cros/cmd/ctpv2/internal/executors"
)

// All currently supported command-executor pairs.
var TranslateRequest_NoExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.TranslateRequestType, ExecutorType: common_executors.NoExecutorType}
var StartFilter_FilterExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.FilterStartCmdType, ExecutorType: executors.FilterExecutorType}
var ExecuteFilter_FilterExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.FilterExecutionCmdType, ExecutorType: executors.FilterExecutorType}

var CtrStartAsync_CtrExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.CtrServiceStartAsyncCmdType, ExecutorType: common_executors.CtrExecutorType}
var CtrStop_CtrExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.CtrServiceStopCmdType, ExecutorType: common_executors.CtrExecutorType}
var GcloudAuth_CtrExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.GcloudAuthCmdType, ExecutorType: common_executors.CtrExecutorType}

// GenerateFilterConfigs generates cmd execution for ctpv2.
func GenerateFilterConfigs(ctx context.Context) *common_configs.Configs {
	mainConfigs := []*common_configs.CommandExecutorPairedConfig{}
	cleanupConfigs := []*common_configs.CommandExecutorPairedConfig{}

	// Translate request
	mainConfigs = append(mainConfigs,
		TranslateRequest_NoExecutor)

	// Start CTR and do GcloudAuth
	mainConfigs = append(mainConfigs,
		CtrStartAsync_CtrExecutor,
		GcloudAuth_CtrExecutor)

	// Execute Test-Finder Filter
	// TODO: Make this more generic to dynamically build configs based on provided
	// container info in inputs.
	mainConfigs = append(mainConfigs,
		StartFilter_FilterExecutor,
		ExecuteFilter_FilterExecutor)

	// Stop CTR
	mainConfigs = append(mainConfigs,
		CtrStop_CtrExecutor)
	cleanupConfigs = append(cleanupConfigs,
		CtrStop_CtrExecutor)

	return &common_configs.Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}
