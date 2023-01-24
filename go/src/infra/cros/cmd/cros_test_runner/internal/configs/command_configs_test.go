// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCommand_UnsupportedCmdType(t *testing.T) {
	t.Parallel()
	Convey("Unsupported command type", t, func() {
		execConfig := NewExecutorConfig()
		cmdConfig := NewCommandConfig(execConfig)
		cmd, err := cmdConfig.GetCommand(commands.UnSupportedCmdType, executors.NoExecutorType)
		So(cmd, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestGetCommand_SupportedCmdType(t *testing.T) {
	t.Parallel()

	Convey("Supported command type", t, func() {
		execConfig := NewExecutorConfig()
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
	})
}
