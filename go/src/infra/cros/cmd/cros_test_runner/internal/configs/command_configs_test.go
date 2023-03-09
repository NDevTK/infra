// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"testing"

	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCommand_UnsupportedCmdType(t *testing.T) {
	t.Parallel()
	Convey("Unsupported command type", t, func() {
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := NewCftContainerConfig(ctr, nil)
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
		contConfig := NewCftContainerConfig(ctr, getMockContainerImagesInfo())
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

		cmd, err = cmdConfig.GetCommand(commands.CtrServiceStartAsyncCmdType, executors.CtrExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.CtrServiceStopCmdType, executors.CtrExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.GcloudAuthCmdType, executors.CtrExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.DutServiceStartCmdType, executors.CrosDutExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.DutVmGetImageCmdType, executors.CrosDutVmExecutorType)
		So(cmd.GetCommandType(), ShouldEqual, commands.DutVmGetImageCmdType)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.DutVmLeaseCmdType, executors.CrosDutVmExecutorType)
		So(cmd.GetCommandType(), ShouldEqual, commands.DutVmLeaseCmdType)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.DutVmReleaseCmdType, executors.CrosDutVmExecutorType)
		So(cmd.GetCommandType(), ShouldEqual, commands.DutVmReleaseCmdType)
		So(err, ShouldBeNil)

		cmd, err = cmdConfig.GetCommand(commands.CacheServerStartCmdType, executors.CrosDutVmExecutorType)
		So(cmd.GetCommandType(), ShouldEqual, commands.CacheServerStartCmdType)
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

		cmd, err = cmdConfig.GetCommand(commands.GcsPublishUploadCmdType, executors.CrosGcsPublishExecutorType)
		So(cmd, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
