// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/common_commands"
	"infra/cros/cmd/common_lib/common_configs"
	"infra/cros/cmd/common_lib/common_executors"
	"infra/cros/cmd/cros_test_runner/data"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/executors"

	tpcommon "go.chromium.org/chromiumos/infra/proto/go/test_platform/common"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
)

// All currently supported command-executor pairs.
var InputValidation_NoExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.BuildInputValidationCmdType, ExecutorType: executors.NoExecutorType}
var ParseEnvInfo_NoExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.ParseEnvInfoCmdType, ExecutorType: executors.NoExecutorType}
var InvServiceStart_InvExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.InvServiceStartCmdType, ExecutorType: executors.InvServiceExecutorType}
var InvServiceStop_InvExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.InvServiceStopCmdType, ExecutorType: executors.InvServiceExecutorType}
var LoadDutTopology_InvExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.LoadDutTopologyCmdType, ExecutorType: executors.InvServiceExecutorType}
var BuildDutTopology_InvExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.BuildDutTopologyCmdType, ExecutorType: executors.InvServiceExecutorType}
var CtrStartAsync_CtrExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.CtrServiceStartAsyncCmdType, ExecutorType: common_executors.CtrExecutorType}
var CtrStop_CtrExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.CtrServiceStopCmdType, ExecutorType: common_executors.CtrExecutorType}
var GcloudAuth_CtrExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.GcloudAuthCmdType, ExecutorType: common_executors.CtrExecutorType}
var DutServerStart_CrosDutExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.DutServiceStartCmdType, ExecutorType: executors.CrosDutExecutorType}
var AndroidCompanionDutServerStart_AndroidDutExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.AndroidCompanionDutServiceStartCmdType, ExecutorType: executors.AndroidDutExecutorType}
var AndroidProvisionServerStart_AndroidProvisionExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.AndroidProvisionServiceStartCmdType, ExecutorType: executors.AndroidProvisionExecutorType}
var AndroidProvisionInstall_AndroidProvisionExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.AndroidProvisionInstallCmdType, ExecutorType: executors.AndroidProvisionExecutorType}
var ProvisionServerStart_CrosProvisionExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.ProvisionServiceStartCmdType, ExecutorType: executors.CrosProvisionExecutorType}
var ProvisionInstall_CrosProvisionExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.ProvisonInstallCmdType, ExecutorType: executors.CrosProvisionExecutorType}
var TestServerStart_CrosTestExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.TestServiceStartCmdType, ExecutorType: executors.CrosTestExecutorType}
var TestsExecution_CrosTestExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.TestsExecutionCmdType, ExecutorType: executors.CrosTestExecutorType}
var TestFinderServerStart_CrosTestFinderExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.TestFinderServiceStartCmdType, ExecutorType: executors.CrosTestFinderExecutorType}
var TestFinderExecution_CrosTestFinderExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.TestFinderExecutionCmdType, ExecutorType: executors.CrosTestFinderExecutorType}
var GcsPublishStart_CrosGcsPublishExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.GcsPublishStartCmdType, ExecutorType: executors.CrosGcsPublishExecutorType}
var GcsPublishUpload_CrosGcsPublishExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.GcsPublishUploadCmdType, ExecutorType: executors.CrosGcsPublishExecutorType}
var RdbPublishStart_CrosRdbPublishExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.RdbPublishStartCmdType, ExecutorType: executors.CrosRdbPublishExecutorType}
var RdbPublishUpload_CrosRdbPublishExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.RdbPublishUploadCmdType, ExecutorType: executors.CrosRdbPublishExecutorType}
var TkoPublishStart_CrosTkoPublishExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.TkoPublishStartCmdType, ExecutorType: executors.CrosTkoPublishExecutorType}
var TkoPublishUpload_CrosTkoPublishExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.TkoPublishUploadCmdType, ExecutorType: executors.CrosTkoPublishExecutorType}
var CpconPublishStart_CrosCpconPublishExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.CpconPublishStartCmdType, ExecutorType: executors.CrosCpconPublishExecutorType}
var CpconPublishUpload_CrosCpconPublishExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.CpconPublishUploadCmdType, ExecutorType: executors.CrosCpconPublishExecutorType}
var ProcessResults_NoExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.ProcessResultsCmdType, ExecutorType: executors.NoExecutorType}
var UpdateDutState_NoExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.UpdateDutStateCmdType, ExecutorType: executors.NoExecutorType}
var TkoDirectUpload_NoExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.TkoDirectUploadCmdType, ExecutorType: executors.NoExecutorType}
var SshStartTunnel_SshTunnelExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.SshStartTunnelCmdType, ExecutorType: executors.SshTunnelExecutorType}
var SshStartReverseTunnel_SshTunnelExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.SshStartReverseTunnelCmdType, ExecutorType: executors.SshTunnelExecutorType}
var SshStopTunnels_SshTunnelExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.SshStopTunnelsCmdType, ExecutorType: executors.SshTunnelExecutorType}
var CacheServerStart_CacheServerExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.CacheServerStartCmdType, ExecutorType: executors.CacheServerExecutorType}
var UpdateContainerImagesLocally_NoExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.UpdateContainerImagesLocallyCmdType, ExecutorType: executors.NoExecutorType}
var FetchContainerMetadata_NoExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.FetchContainerMetadataCmdType, ExecutorType: executors.NoExecutorType}
var ParseArgs_NoExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.ParseArgsCmdType, ExecutorType: executors.NoExecutorType}
var DutVmCacheServerStart_CacheServerExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.DutVmCacheServerStartCmdType, ExecutorType: executors.CacheServerExecutorType}
var DutVmGetImage_CrosDutVmExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.DutVmGetImageCmdType, ExecutorType: executors.CrosDutVmExecutorType}
var DutServiceStart_CrosDutVmExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.DutServiceStartCmdType, ExecutorType: executors.CrosDutVmExecutorType}
var VMProvisionServerStart_CrosVMProvisionExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.VMProvisionServiceStartCmdType, ExecutorType: executors.CrosVMProvisionExecutorType}
var VMProvisionLease_CrosVMProvisionExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.VMProvisionLeaseCmdType, ExecutorType: executors.CrosVMProvisionExecutorType}
var VMProvisionRelease_CrosVMProvisionExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.VMProvisionReleaseCmdType, ExecutorType: executors.CrosVMProvisionExecutorType}
var ContainerStart_ContainerExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.ContainerStartCmdType, ExecutorType: common_executors.ContainerExecutorType}
var ContainerCloseLogs_ContainerExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.ContainerCloseLogsCmdType, ExecutorType: common_executors.ContainerExecutorType}
var ContainerReadLogs_ContainerExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: common_commands.ContainerReadLogsCmdType, ExecutorType: common_executors.ContainerExecutorType}
var GenericProvision_GenericProvisionExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.GenericProvisionCmdType, ExecutorType: executors.GenericProvisionExecutorType}
var GenericTests_GenericTestsExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.GenericTestsCmdType, ExecutorType: executors.GenericTestsExecutorType}
var GenericPublish_GenericPublishExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.GenericPublishCmdType, ExecutorType: executors.GenericPublishExecutorType}
var ParseDutTopology_NoExecutor = &common_configs.CommandExecutorPairedConfig{CommandType: commands.ParseDutTopologyCmdType, ExecutorType: common_executors.NoExecutorType}

var RequiredCmdExecPairMap = map[*common_configs.CommandExecutorPairedConfig]*common_configs.CommandExecutorPairedConfig{}

func GetCmdExecPair(pair_base *common_configs.CommandExecutorPairedConfig, required bool) *common_configs.CommandExecutorPairedConfig {
	if !required {
		return pair_base
	}
	if pair_required, ok := RequiredCmdExecPairMap[pair_base]; ok {
		return pair_required
	}

	pair_required := &common_configs.CommandExecutorPairedConfig{
		CommandType:  pair_base.CommandType,
		ExecutorType: pair_base.ExecutorType,
	}
	pair_required.SetRequired(true)

	RequiredCmdExecPairMap[pair_base] = pair_required

	return pair_required
}

// GenerateHwConfigs generates hw tests execution for lab environment.
func GenerateHwConfigs(ctx context.Context, cftHwStepsConfig *tpcommon.HwTestConfig, inputV2 *skylab_test_runner.CrosTestRunnerRequest, isAndroidProvisionRequired bool) *common_configs.Configs {
	platform := common.GetBotProvider()
	if inputV2 != nil {
		return hwConfigsForPlatformV2(cftHwStepsConfig, inputV2, platform)
	}
	return hwConfigsForPlatform(cftHwStepsConfig, platform, isAndroidProvisionRequired)
}

// hwConfigsForPlatform generates platform-specific configs.
// GCE platform will get configs for VM test on GCE.
// Non-GCE platforms (Drone and Unknown) will get configs for HW test on Drone.
func hwConfigsForPlatform(cftHwStepsConfig *tpcommon.HwTestConfig, platform common.SwarmingBotProvider, isAndroidProvisionRequired bool) *common_configs.Configs {
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
	mainConfigs := []*common_configs.CommandExecutorPairedConfig{}
	cleanupConfigs := []*common_configs.CommandExecutorPairedConfig{}

	// Input validation and parse env commands
	mainConfigs = append(mainConfigs,
		InputValidation_NoExecutor,
		ParseEnvInfo_NoExecutor)

	// Dut topology commands
	if !cftHwStepsConfig.GetSkipLoadingDutTopology() {
		mainConfigs = append(mainConfigs,
			InvServiceStart_InvExecutor,
			LoadDutTopology_InvExecutor,
			InvServiceStop_InvExecutor,
			ParseDutTopology_NoExecutor)
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
				VMProvisionServerStart_CrosVMProvisionExecutor,
				VMProvisionLease_CrosVMProvisionExecutor,
				DutVmCacheServerStart_CacheServerExecutor,
				DutServiceStart_CrosDutVmExecutor)
		} else {
			mainConfigs = append(mainConfigs,
				DutServerStart_CrosDutExecutor)
			if isAndroidProvisionRequired {
				mainConfigs = append(mainConfigs,
					AndroidCompanionDutServerStart_AndroidDutExecutor)
			}
		}
	}

	// Provision commands
	if !cftHwStepsConfig.GetSkipProvision() {
		mainConfigs = append(mainConfigs,
			ProvisionServerStart_CrosProvisionExecutor,
			ProvisionInstall_CrosProvisionExecutor)
		if isAndroidProvisionRequired {
			mainConfigs = append(mainConfigs,
				AndroidProvisionServerStart_AndroidProvisionExecutor,
				AndroidProvisionInstall_AndroidProvisionExecutor)
		}
	}

	// Test execution commands
	if !cftHwStepsConfig.GetSkipTestExecution() {
		mainConfigs = append(mainConfigs,
			TestServerStart_CrosTestExecutor,
			TestsExecution_CrosTestExecutor)
	}

	// Publish commands
	if !cftHwStepsConfig.GetSkipAllResultPublish() {
		// Re-auth as long test execution can expire previous auth
		mainConfigs = append(mainConfigs, GcloudAuth_CtrExecutor)

		// Rdb publish commands
		if !cftHwStepsConfig.GetSkipRdbPublish() {
			mainConfigs = append(mainConfigs,
				RdbPublishStart_CrosRdbPublishExecutor,
				RdbPublishUpload_CrosRdbPublishExecutor)
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

	// Stop CTR and result processing commands
	if platform == common.BotProviderGce {
		mainConfigs = append(mainConfigs,
			VMProvisionRelease_CrosVMProvisionExecutor,
			CtrStop_CtrExecutor,
			ProcessResults_NoExecutor)
		cleanupConfigs = append(cleanupConfigs,
			VMProvisionRelease_CrosVMProvisionExecutor,
			CtrStop_CtrExecutor,
			ProcessResults_NoExecutor)
	} else {
		mainConfigs = append(mainConfigs,
			CtrStop_CtrExecutor,
			UpdateDutState_NoExecutor,
			ProcessResults_NoExecutor)
		cleanupConfigs = append(cleanupConfigs,
			CtrStop_CtrExecutor,
			UpdateDutState_NoExecutor,
			ProcessResults_NoExecutor)
	}

	return &common_configs.Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}

func GeneratePreLocalConfigs(ctx context.Context) *common_configs.Configs {
	mainConfigs := []*common_configs.CommandExecutorPairedConfig{
		ParseArgs_NoExecutor,
		FetchContainerMetadata_NoExecutor,
		UpdateContainerImagesLocally_NoExecutor,
	}

	// Clean up configs. They will be executed if any failures occurs
	// in main configs. If any of the cleanup cmd is already executed,
	// they will be skipped.
	cleanupConfigs := []*common_configs.CommandExecutorPairedConfig{}

	return &common_configs.Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}

// hwConfigsForPlatformV2 generates the command/executor pair configs for the CrosTestRunnerRequest.
func hwConfigsForPlatformV2(cftHwStepsConfig *tpcommon.HwTestConfig, inputV2 *skylab_test_runner.CrosTestRunnerRequest, platform common.SwarmingBotProvider) *common_configs.Configs {
	mainConfigs := []*common_configs.CommandExecutorPairedConfig{}
	cleanupConfigs := []*common_configs.CommandExecutorPairedConfig{}

	// TODO(cdelagarza): provide check for whether platform is VM and handle how DUTs get loaded.

	mainConfigs = append(mainConfigs,
		InputValidation_NoExecutor,
		ParseEnvInfo_NoExecutor)

	if !cftHwStepsConfig.GetSkipLoadingDutTopology() {
		mainConfigs = append(mainConfigs,
			InvServiceStart_InvExecutor,
			LoadDutTopology_InvExecutor,
			InvServiceStop_InvExecutor,
			ParseDutTopology_NoExecutor)
	}

	// Start CTR and gcloud auth commands
	mainConfigs = append(mainConfigs,
		CtrStartAsync_CtrExecutor,
		GcloudAuth_CtrExecutor,
		ContainerReadLogs_ContainerExecutor)

	// Add task configs
	mainConfigs = append(mainConfigs, generateTaskConfigs(inputV2).MainConfigs...)

	// Stop CTR and result processing commands
	mainConfigs = append(mainConfigs,
		GetCmdExecPair(ContainerCloseLogs_ContainerExecutor, true),
		GetCmdExecPair(CtrStop_CtrExecutor, true),
		GetCmdExecPair(UpdateDutState_NoExecutor, true),
		GetCmdExecPair(ProcessResults_NoExecutor, true))

	return &common_configs.Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}

func GenerateLocalConfigs(ctx context.Context, sk *data.LocalTestStateKeeper) *common_configs.Configs {
	mainConfigs := []*common_configs.CommandExecutorPairedConfig{
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

	// Cpcon publish commands
	if sk.Args.RunCpconPublish {
		mainConfigs = append(mainConfigs,
			CpconPublishStart_CrosCpconPublishExecutor,
			CpconPublishUpload_CrosCpconPublishExecutor)
	}

	mainConfigs = append(mainConfigs,
		CtrStop_CtrExecutor,
		SshStopTunnels_SshTunnelExecutor,
		ProcessResults_NoExecutor)

	// Clean up configs. They will be executed if any failures occurs
	// in main configs. If any of the cleanup cmd is already executed,
	// they will be skipped.
	cleanupConfigs := []*common_configs.CommandExecutorPairedConfig{
		CtrStop_CtrExecutor,
		SshStopTunnels_SshTunnelExecutor,
		ProcessResults_NoExecutor,
	}

	return &common_configs.Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}

func generateTaskConfigs(inputV2 *skylab_test_runner.CrosTestRunnerRequest) *common_configs.Configs {
	mainConfigs := []*common_configs.CommandExecutorPairedConfig{}

	for _, task := range inputV2.GetOrderedTasks() {
		for range task.GetOrderedContainerRequests() {
			mainConfigs = append(mainConfigs,
				GetCmdExecPair(ContainerStart_ContainerExecutor, task.Required))
		}
		switch task.Task.(type) {
		case *skylab_test_runner.CrosTestRunnerRequest_Task_Provision:
			mainConfigs = append(mainConfigs,
				GetCmdExecPair(GenericProvision_GenericProvisionExecutor, task.Required))
		case *skylab_test_runner.CrosTestRunnerRequest_Task_PreTest:
		case *skylab_test_runner.CrosTestRunnerRequest_Task_Test:
			mainConfigs = append(mainConfigs,
				GetCmdExecPair(GenericTests_GenericTestsExecutor, task.Required))
		case *skylab_test_runner.CrosTestRunnerRequest_Task_PostTest:
		case *skylab_test_runner.CrosTestRunnerRequest_Task_Publish:
			mainConfigs = append(mainConfigs,
				GetCmdExecPair(GcloudAuth_CtrExecutor, task.Required),
				GetCmdExecPair(GenericPublish_GenericPublishExecutor, task.Required))
		default:
		}
	}

	return &common_configs.Configs{MainConfigs: mainConfigs}
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
		"DRONE_AGENT_BOT_BLKIO_WRITE_BPS",
		"SWARMING_TASK_ID",
		"LOGDOG_STREAM_PREFIX",
		"DOCKER_CONFIG"}
}
