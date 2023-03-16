// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"
	"fmt"

	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/data"
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
var BuildDutTopology_InvExecutor = &CommandExecutorPairedConfig{CommandType: commands.BuildDutTopologyCmdType, ExecutorType: executors.InvServiceExecutorType}
var CtrStartAsync_CtrExecutor = &CommandExecutorPairedConfig{CommandType: commands.CtrServiceStartAsyncCmdType, ExecutorType: executors.CtrExecutorType}
var CtrStop_CtrExecutor = &CommandExecutorPairedConfig{CommandType: commands.CtrServiceStopCmdType, ExecutorType: executors.CtrExecutorType}
var GcloudAuth_CtrExecutor = &CommandExecutorPairedConfig{CommandType: commands.GcloudAuthCmdType, ExecutorType: executors.CtrExecutorType}
var DutServerStart_CrosDutExecutor = &CommandExecutorPairedConfig{CommandType: commands.DutServiceStartCmdType, ExecutorType: executors.CrosDutExecutorType}
var ProvisionServerStart_CrosProvisionExecutor = &CommandExecutorPairedConfig{CommandType: commands.ProvisionServiceStartCmdType, ExecutorType: executors.CrosProvisionExecutorType}
var ProvisionInstall_CrosProvisionExecutor = &CommandExecutorPairedConfig{CommandType: commands.ProvisonInstallCmdType, ExecutorType: executors.CrosProvisionExecutorType}
var TestServerStart_CrosTestExecutor = &CommandExecutorPairedConfig{CommandType: commands.TestServiceStartCmdType, ExecutorType: executors.CrosTestExecutorType}
var TestsExecution_CrosTestExecutor = &CommandExecutorPairedConfig{CommandType: commands.TestsExecutionCmdType, ExecutorType: executors.CrosTestExecutorType}
var TestFinderServerStart_CrosTestFinderExecutor = &CommandExecutorPairedConfig{CommandType: commands.TestFinderServiceStartCmdType, ExecutorType: executors.CrosTestFinderExecutorType}
var TestFinderExecution_CrosTestFinderExecutor = &CommandExecutorPairedConfig{CommandType: commands.TestFinderExecutionCmdType, ExecutorType: executors.CrosTestFinderExecutorType}
var GcsPublishStart_CrosGcsPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.GcsPublishStartCmdType, ExecutorType: executors.CrosGcsPublishExecutorType}
var GcsPublishUpload_CrosGcsPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.GcsPublishUploadCmdType, ExecutorType: executors.CrosGcsPublishExecutorType}
var RdbPublishStart_CrosRdbPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.RdbPublishStartCmdType, ExecutorType: executors.CrosRdbPublishExecutorType}
var RdbPublishUpload_CrosRdbPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.RdbPublishUploadCmdType, ExecutorType: executors.CrosRdbPublishExecutorType}
var TkoPublishStart_CrosTkoPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.TkoPublishStartCmdType, ExecutorType: executors.CrosTkoPublishExecutorType}
var TkoPublishUpload_CrosTkoPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.TkoPublishUploadCmdType, ExecutorType: executors.CrosTkoPublishExecutorType}
var CpconPublishStart_CrosCpconPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.CpconPublishStartCmdType, ExecutorType: executors.CrosCpconPublishExecutorType}
var CpconPublishUpload_CrosCpconPublishExecutor = &CommandExecutorPairedConfig{CommandType: commands.CpconPublishUploadCmdType, ExecutorType: executors.CrosCpconPublishExecutorType}
var ProcessResults_NoExecutor = &CommandExecutorPairedConfig{CommandType: commands.ProcessResultsCmdType, ExecutorType: executors.NoExecutorType}
var UpdateDutState_NoExecutor = &CommandExecutorPairedConfig{CommandType: commands.UpdateDutStateCmdType, ExecutorType: executors.NoExecutorType}
var TkoDirectUpload_NoExecutor = &CommandExecutorPairedConfig{CommandType: commands.TkoDirectUploadCmdType, ExecutorType: executors.NoExecutorType}
var SshStartTunnel_SshTunnelExecutor = &CommandExecutorPairedConfig{CommandType: commands.SshStartTunnelCmdType, ExecutorType: executors.SshTunnelExecutorType}
var SshStartReverseTunnel_SshTunnelExecutor = &CommandExecutorPairedConfig{CommandType: commands.SshStartReverseTunnelCmdType, ExecutorType: executors.SshTunnelExecutorType}
var SshStopTunnels_SshTunnelExecutor = &CommandExecutorPairedConfig{CommandType: commands.SshStopTunnelsCmdType, ExecutorType: executors.SshTunnelExecutorType}
var CacheServerStart_CacheServerExecutor = &CommandExecutorPairedConfig{CommandType: commands.CacheServerStartCmdType, ExecutorType: executors.CacheServerExecutorType}
var UpdateContainerImagesLocally_NoExecutor = &CommandExecutorPairedConfig{CommandType: commands.UpdateContainerImagesLocallyCmdType, ExecutorType: executors.NoExecutorType}
var FetchContainerMetadata_NoExecutor = &CommandExecutorPairedConfig{CommandType: commands.FetchContainerMetadataCmdType, ExecutorType: executors.NoExecutorType}
var ParseArgs_NoExecutor = &CommandExecutorPairedConfig{CommandType: commands.ParseArgsCmdType, ExecutorType: executors.NoExecutorType}
var DutVmCacheServerStart_CacheServerExecutor = &CommandExecutorPairedConfig{CommandType: commands.DutVmCacheServerStartCmdType, ExecutorType: executors.CacheServerExecutorType}
var DutVmLease_CrosDutVmExecutor = &CommandExecutorPairedConfig{CommandType: commands.DutVmLeaseCmdType, ExecutorType: executors.CrosDutVmExecutorType}
var DutVmRelease_CrosDutVmExecutor = &CommandExecutorPairedConfig{CommandType: commands.DutVmReleaseCmdType, ExecutorType: executors.CrosDutVmExecutorType}
var DutVmGetImage_CrosDutVmExecutor = &CommandExecutorPairedConfig{CommandType: commands.DutVmGetImageCmdType, ExecutorType: executors.CrosDutVmExecutorType}
var DutServiceStart_CrosDutVmExecutor = &CommandExecutorPairedConfig{CommandType: commands.DutServiceStartCmdType, ExecutorType: executors.CrosDutVmExecutorType}

// GenerateHwConfigs generates hw tests execution for lab environment.
func GenerateHwConfigs(ctx context.Context, cftHwStepsConfig *tpcommon.HwTestConfig) *Configs {
	platform := common.GetBotProvider()
	return hwConfigsForPlatform(cftHwStepsConfig, platform)
}

// hwConfigsForPlatform generates platform-specific configs.
// GCE platform will get configs for VM test on GCE.
// Non-GCE platforms (Drone and Unknown) will get configs for HW test on Drone.
func hwConfigsForPlatform(cftHwStepsConfig *tpcommon.HwTestConfig, platform common.SwarmingBotProvider) *Configs {
	// Overwrite configs that don't apply to VM test
	if platform == common.BotProviderGce {
		if cftHwStepsConfig == nil {
			cftHwStepsConfig = &tpcommon.HwTestConfig{}
		}
		// Skip DutTopology and Provision steps, as those are done in the
		// non-skippable Starting Dut Service step
		cftHwStepsConfig.SkipLoadingDutTopology = true
		cftHwStepsConfig.SkipProvision = true
		cftHwStepsConfig.SkipStartingDutService = false
	}
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
		if platform == common.BotProviderGce {
			// Prepare image, lease VM, start cache server before finally start Dut service
			mainConfigs = append(mainConfigs,
				DutVmGetImage_CrosDutVmExecutor,
				DutVmLease_CrosDutVmExecutor,
				DutVmCacheServerStart_CacheServerExecutor,
				DutServiceStart_CrosDutVmExecutor)
		} else {
			mainConfigs = append(mainConfigs,
				DutServerStart_CrosDutExecutor)
		}
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

			// Meanwhile, we will invoke the script directly
			mainConfigs = append(mainConfigs,
				TkoDirectUpload_NoExecutor)
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

		// Cpcon publish commands
		if cftHwStepsConfig.GetRunCpconPublish() {
			mainConfigs = append(mainConfigs,
				CpconPublishStart_CrosCpconPublishExecutor,
				CpconPublishUpload_CrosCpconPublishExecutor)
			cleanupConfigs = append(cleanupConfigs,
				CpconPublishStart_CrosCpconPublishExecutor,
				CpconPublishUpload_CrosCpconPublishExecutor)
		}
	}

	// Recycle Dut is either update state to need repair (HW) or release (VM)
	dutRecycle := UpdateDutState_NoExecutor
	if platform == common.BotProviderGce {
		dutRecycle = DutVmRelease_CrosDutVmExecutor
	}

	// Stop CTR and result processing commands
	mainConfigs = append(mainConfigs,
		CtrStop_CtrExecutor,
		dutRecycle,
		ProcessResults_NoExecutor)
	cleanupConfigs = append(cleanupConfigs,
		CtrStop_CtrExecutor,
		dutRecycle,
		ProcessResults_NoExecutor)

	return &Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}

func GeneratePreLocalConfigs(ctx context.Context) *Configs {
	mainConfigs := []*CommandExecutorPairedConfig{
		ParseArgs_NoExecutor,
		FetchContainerMetadata_NoExecutor,
		UpdateContainerImagesLocally_NoExecutor,
	}

	// Clean up configs. They will be executed if any failures occurs
	// in main configs. If any of the cleanup cmd is already executed,
	// they will be skipped.
	cleanupConfigs := []*CommandExecutorPairedConfig{}

	return &Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}

func GenerateLocalConfigs(ctx context.Context, sk *data.LocalTestStateKeeper) *Configs {
	mainConfigs := []*CommandExecutorPairedConfig{
		CtrStartAsync_CtrExecutor,
		GcloudAuth_CtrExecutor,
	}

	if !sk.Args.SkipCacheServer {
		mainConfigs = append(mainConfigs,
			CacheServerStart_CacheServerExecutor)
	}

	if !sk.Args.SkipSshTunnel {
		mainConfigs = append(mainConfigs,
			SshStartTunnel_SshTunnelExecutor)
	}

	if !sk.Args.SkipSshReverseTunnel {
		mainConfigs = append(mainConfigs,
			SshStartReverseTunnel_SshTunnelExecutor)
	}

	if !sk.Args.SkipBuildDutTopology {
		mainConfigs = append(mainConfigs,
			BuildDutTopology_InvExecutor)
	}

	if !sk.Args.SkipDutServer {
		mainConfigs = append(mainConfigs,
			DutServerStart_CrosDutExecutor)
	}
	if !sk.Args.SkipProvision {
		mainConfigs = append(mainConfigs,
			ProvisionServerStart_CrosProvisionExecutor,
			ProvisionInstall_CrosProvisionExecutor)
	}
	if !sk.Args.SkipTestFinder {
		mainConfigs = append(mainConfigs,
			TestFinderServerStart_CrosTestFinderExecutor,
			TestFinderExecution_CrosTestFinderExecutor)
	}
	if !sk.Args.SkipTest {
		mainConfigs = append(mainConfigs,
			TestServerStart_CrosTestExecutor,
			TestsExecution_CrosTestExecutor)
	}

	mainConfigs = append(mainConfigs,
		CtrStop_CtrExecutor,
		SshStopTunnels_SshTunnelExecutor,
		ProcessResults_NoExecutor)

	// Clean up configs. They will be executed if any failures occurs
	// in main configs. If any of the cleanup cmd is already executed,
	// they will be skipped.
	cleanupConfigs := []*CommandExecutorPairedConfig{
		CtrStop_CtrExecutor,
		SshStopTunnels_SshTunnelExecutor,
		ProcessResults_NoExecutor,
	}

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
