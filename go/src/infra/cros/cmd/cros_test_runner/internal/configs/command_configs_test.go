// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"testing"

	"infra/cros/cmd/common_lib/common_commands"
	"infra/cros/cmd/common_lib/common_configs"
	"infra/cros/cmd/common_lib/common_executors"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/executors"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCommand_UnsupportedCmdType(t *testing.T) {
	t.Parallel()
	Convey("Unsupported command type", t, func() {
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := common_configs.NewContainerConfig(ctr, nil, false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)
		cmd, err := cmdConfig.GetCommand(commands.UnSupportedCmdType, executors.NoExecutorType)
		So(cmd, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestGetCommand_SupportedCmdType(t *testing.T) {
	t.Parallel()

	Convey("Supported command type", t, func() {
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := common_configs.NewContainerConfig(ctr, getMockContainerImagesInfo(), false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)

		cmd, err := cmdConfig.GetCommand(commands.BuildInputValidationCmdType, executors.NoExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.ParseEnvInfoCmdType, executors.NoExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.InvServiceStartCmdType, executors.InvServiceExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.InvServiceStopCmdType, executors.InvServiceExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.LoadDutTopologyCmdType, executors.InvServiceExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(common_commands.CtrServiceStartAsyncCmdType, common_executors.CtrExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(common_commands.CtrServiceStopCmdType, common_executors.CtrExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(common_commands.GcloudAuthCmdType, common_executors.CtrExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.DutServiceStartCmdType, executors.CrosDutExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.DutVmGetImageCmdType, executors.CrosDutVmExecutorType)
		So(cmd.GetCommandType(), ShouldEqual, commands.DutVmGetImageCmdType)
		So(err, ShouldBeNil)

		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.DutVmCacheServerStartCmdType, executors.CrosDutVmExecutorType)
		So(cmd.GetCommandType(), ShouldEqual, commands.DutVmCacheServerStartCmdType)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.ProvisionServiceStartCmdType, executors.CrosProvisionExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.ProvisonInstallCmdType, executors.CrosProvisionExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.TestServiceStartCmdType, executors.CrosTestExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.TestsExecutionCmdType, executors.CrosTestExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.RdbPublishStartCmdType, executors.CrosRdbPublishExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.RdbPublishUploadCmdType, executors.CrosRdbPublishExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.TkoPublishStartCmdType, executors.CrosTkoPublishExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.TkoPublishUploadCmdType, executors.CrosTkoPublishExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.GcsPublishStartCmdType, executors.CrosGcsPublishExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.SshStartReverseTunnelCmdType, executors.SshTunnelExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.SshStartTunnelCmdType, executors.SshTunnelExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.SshStopTunnelsCmdType, executors.SshTunnelExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.TestFinderServiceStartCmdType, executors.CrosTestFinderExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.TestFinderExecutionCmdType, executors.CrosTestFinderExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.BuildDutTopologyCmdType, executors.InvServiceExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.CacheServerStartCmdType, executors.CacheServerExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.ParseArgsCmdType, executors.NoExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.UpdateContainerImagesLocallyCmdType, executors.NoExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.VMProvisionServiceStartCmdType, executors.CrosVMProvisionExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.VMProvisionLeaseCmdType, executors.CrosVMProvisionExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.VMProvisionReleaseCmdType, executors.CrosVMProvisionExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.AndroidProvisionServiceStartCmdType, executors.AndroidProvisionExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.AndroidProvisionInstallCmdType, executors.AndroidProvisionExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
