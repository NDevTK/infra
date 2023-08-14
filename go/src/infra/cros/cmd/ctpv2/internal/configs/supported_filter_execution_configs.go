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
var PrepareFilterContainers_NoExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.PrepareFilterContainersCmdType, ExecutorType: common_executors.NoExecutorType}
var ExecuteFilter_FilterExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.FilterExecutionCmdType, ExecutorType: executors.FilterExecutorType}

var CtrStartAsync_CtrExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.CtrServiceStartAsyncCmdType, ExecutorType: common_executors.CtrExecutorType}
var CtrStop_CtrExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.CtrServiceStopCmdType, ExecutorType: common_executors.CtrExecutorType}
var GcloudAuth_CtrExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.GcloudAuthCmdType, ExecutorType: common_executors.CtrExecutorType}
var ContainerStart_ContainerExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.ContainerStartCmdType, ExecutorType: common_executors.ContainerExecutorType}
var ContainerReadLogs_ContainerExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.ContainerReadLogsCmdType, ExecutorType: common_executors.ContainerExecutorType}
var ContainerCloseLogs_ContainerExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.ContainerCloseLogsCmdType, ExecutorType: common_executors.ContainerExecutorType}

// GenerateFilterConfigs generates cmd execution for ctpv2.
func GenerateFilterConfigs(ctx context.Context, totalFilters int) *common_configs.Configs {
	mainConfigs := []*common_configs.CommandExecutorPairedConfig{}
	cleanupConfigs := []*common_configs.CommandExecutorPairedConfig{}

	// Translate request
	mainConfigs = append(mainConfigs,
		TranslateRequest_NoExecutor)

	// Start CTR and do GcloudAuth
	mainConfigs = append(mainConfigs,
		CtrStartAsync_CtrExecutor,
		GcloudAuth_CtrExecutor)

	mainConfigs = append(mainConfigs,
		PrepareFilterContainers_NoExecutor,
		ContainerReadLogs_ContainerExecutor)

	for i := 0; i < totalFilters; i++ {
		mainConfigs = append(mainConfigs,
			ContainerStart_ContainerExecutor,
			ExecuteFilter_FilterExecutor,
		)
	}

	mainConfigs = append(mainConfigs,
		ContainerCloseLogs_ContainerExecutor)

	// Stop CTR
	mainConfigs = append(mainConfigs,
		CtrStop_CtrExecutor)
	cleanupConfigs = append(cleanupConfigs,
		ContainerCloseLogs_ContainerExecutor,
		CtrStop_CtrExecutor)

	return &common_configs.Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}
