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
var DutServerStart_CrosDutExecutor = &CommandExecutorPairedConfig{CommandType: commands.DutServiceStartCmdType, ExecutorType: executors.CrosDutExecutorType}
var ProvisionServerStart_CrosProvisionExecutor = &CommandExecutorPairedConfig{CommandType: commands.ProvisionServiceStartCmdType, ExecutorType: executors.CrosProvisionExecutorType}
var ProvisionInstall_CrosProvisionExecutor = &CommandExecutorPairedConfig{CommandType: commands.ProvisonInstallCmdType, ExecutorType: executors.CrosProvisionExecutorType}

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
		DutServerStart_CrosDutExecutor,
		ProvisionServerStart_CrosProvisionExecutor,
		ProvisionInstall_CrosProvisionExecutor,
		CtrStop_CtrExecutor,
	}

	// Clean up configs. They will be executed if any failures occurs in main configs.
	// If any of the cleanup cmd is already executed, they will be skipped.
	cleanupConfigs := []*CommandExecutorPairedConfig{
		CtrStop_CtrExecutor,
	}

	return &Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}

// GetHwConfigsEnvVars gets all env vars that are required for hw test execution configs.
func GetHwConfigsEnvVars() []string {
	return []string{
		"LUCI_CONTEXT",
		"GCE_METADATA_HOST",
		"GCE_METADATA_IP",
		"GCE_METADATA_ROOT",
		"CONTAINER_CACHE_SERVICE_PORT",
		"CONTAINER_CACHE_SERVICE_HOST",
		"DRONE_AGENT_BOT_BLKIO_READ_BPS",
		"DRONE_AGENT_BOT_BLKIO_WRITE_BPS"}
}
