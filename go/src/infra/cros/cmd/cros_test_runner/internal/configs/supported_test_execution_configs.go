// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"
	"fmt"

	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"

	tpcommon "go.chromium.org/chromiumos/infra/proto/go/test_platform/common"
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
var TestServerStart_CrosTestExecutor = &CommandExecutorPairedConfig{CommandType: commands.TestServiceStartCmdType, ExecutorType: executors.CrosTestExecutorType}
var TestsExecution_CrosTestExecutor = &CommandExecutorPairedConfig{CommandType: commands.TestsExecutionCmdType, ExecutorType: executors.CrosTestExecutorType}
var GcsPublishStart_CrosGcsPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.GcsPublishStartCmdType, ExecutorType: executors.CrosGcsPublishExecutorType}
var GcsPublishUpload_CrosGcsPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.GcsPublishUploadCmdType, ExecutorType: executors.CrosGcsPublishExecutorType}
var RdbPublishStart_CrosRdbPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.RdbPublishStartCmdType, ExecutorType: executors.CrosRdbPublishExecutorType}
var RdbPublishUpload_CrosRdbPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.RdbPublishUploadCmdType, ExecutorType: executors.CrosRdbPublishExecutorType}
var TkoPublishStart_CrosTkoPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.TkoPublishStartCmdType, ExecutorType: executors.CrosTkoPublishExecutorType}
var TkoPublishUpload_CrosTkoPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.TkoPublishUploadCmdType, ExecutorType: executors.CrosTkoPublishExecutorType}
var ProcessResults_NoExecutor = &CommandExecutorPairedConfig{CommandType: commands.ProcessResultsCmdType, ExecutorType: executors.NoExecutorType}
var UpdateDutState_NoExecutor = &CommandExecutorPairedConfig{CommandType: commands.UpdateDutStateCmdType, ExecutorType: executors.NoExecutorType}

// GenerateHwConfigs generates hw tests execution for lab environment.
func GenerateHwConfigs(ctx context.Context, cftHwStepsConfig *tpcommon.HwTestConfig) *Configs {
	mainConfigs := []*CommandExecutorPairedConfig{}
	cleanupConfigs := []*CommandExecutorPairedConfig{}

	// Input validation and parse env commands
	mainConfigs = append(mainConfigs,
		InputValidation_NoExecutor,
		ParseEnvInfo_NoExecutor)

	// Dut topology commands
	if !cftHwStepsConfig.GetSkipLoadingDutTopology() {
		mainConfigs = append(mainConfigs,
			InvServiceStart_InvExecutor,
			LoadDutTopology_InvExecutor,
			InvServiceStop_InvExecutor)
	}

	// Start CTR and gcloud auth commands
	mainConfigs = append(mainConfigs,
		CtrStartAsync_CtrExecutor,
		GcloudAuth_CtrExecutor)

	// Start dut server command
	if !cftHwStepsConfig.GetSkipStartingDutService() {
		mainConfigs = append(mainConfigs,
			DutServerStart_CrosDutExecutor)
	}

	// Provision commands
	if !cftHwStepsConfig.GetSkipProvision() {
		mainConfigs = append(mainConfigs,
			ProvisionServerStart_CrosProvisionExecutor,
			ProvisionInstall_CrosProvisionExecutor)
	}

	// Test execution commands
	if !cftHwStepsConfig.GetSkipTestExecution() {
		mainConfigs = append(mainConfigs,
			TestServerStart_CrosTestExecutor,
			TestsExecution_CrosTestExecutor)
	}

	// Publish commands
	if !cftHwStepsConfig.GetSkipAllResultPublish() {
		// Rdb publish commands
		if !cftHwStepsConfig.GetSkipRdbPublish() {
			mainConfigs = append(mainConfigs,
				RdbPublishStart_CrosRdbPublishExecutor,
				RdbPublishUpload_CrosRdbPublishExecutor)
		}

		// TKO publish commands
		if !cftHwStepsConfig.GetSkipTkoPublish() {
			// TODO (b/241155482): Enable TKO publish after tko publish issues are fixed.
			// mainConfigs = append(mainConfigs,
			// 	TkoPublishStart_CrosTkoPublishExecutor,
			// 	TkoPublishUpload_CrosTkoPublishExecutor)
		}

		// Gcs publish commands
		if !cftHwStepsConfig.GetSkipGcsPublish() {
			mainConfigs = append(mainConfigs,
				GcsPublishStart_CrosGcsPublishExecutor,
				GcsPublishUpload_CrosGcsPublishExecutor)
			cleanupConfigs = append(cleanupConfigs,
				GcsPublishStart_CrosGcsPublishExecutor,
				GcsPublishUpload_CrosGcsPublishExecutor)
		}
	}

	// Stop CTR and result processing commands
	mainConfigs = append(mainConfigs,
		CtrStop_CtrExecutor,
		UpdateDutState_NoExecutor,
		ProcessResults_NoExecutor)
	cleanupConfigs = append(cleanupConfigs,
		CtrStop_CtrExecutor,
		UpdateDutState_NoExecutor,
		ProcessResults_NoExecutor)

	return &Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}

// GetHwConfigsEnvVars gets all env vars that are required
// for hw test execution configs.
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
