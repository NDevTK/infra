// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/containers"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCpconPublishPublishCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosCpconPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosCpconPublishExecutorType)
		cmd := commands.NewCpconPublishUploadCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

// TODO(b/278760353): Restore when issue fixed.
// func TestCpconPublishPublishCmd_MissingDeps(t *testing.T) {
// 	t.Setenv("SWARMING_TASK_ID", "")

// 	Convey("Cmd missing deps", t, func() {
// 		ctx := context.Background()
// 		sk := &data.HwTestStateKeeper{}
// 		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
// 		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
// 		cont := containers.NewCrosPublishTemplatedContainer(
// 			containers.CrosCpconPublishTemplatedContainerType,
// 			"container/image/path",
// 			ctr)
// 		exec := executors.NewCrosPublishExecutor(
// 			cont,
// 			executors.CrosCpconPublishExecutorType)
// 		cmd := commands.NewCpconPublishUploadCmd(exec)
// 		err := cmd.ExtractDependencies(ctx, sk)
// 		So(err, ShouldNotBeNil)
// 	})
// }

func TestCpconPublishPublishCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosCpconPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosCpconPublishExecutorType)
		cmd := commands.NewCpconPublishUploadCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestCpconPublishPublishCmd_ExtractDepsSuccess(t *testing.T) {
	wantSwarmingTaskId := "123456789abcdef0"
	t.Setenv("SWARMING_TASK_ID", wantSwarmingTaskId)

	Convey("ProvisionStartCmd extract deps", t, func() {

		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosCpconPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosCpconPublishExecutorType)
		cmd := commands.NewCpconPublishUploadCmd(exec)

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.CpconJobName, ShouldEqual, fmt.Sprintf("swarming-%s", wantSwarmingTaskId))

	})

}
