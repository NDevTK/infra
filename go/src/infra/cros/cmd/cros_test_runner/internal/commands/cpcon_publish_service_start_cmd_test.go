// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/containers"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCpconPublishStartCmd_UnsupportedSK(t *testing.T) {
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
		cmd := commands.NewCpconPublishServiceStartCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestCpconPublishStartCmd_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
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
		cmd := commands.NewCpconPublishServiceStartCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestCpconPublishStartCmd_UpdateSK(t *testing.T) {
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
		cmd := commands.NewCpconPublishServiceStartCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestCpconPublishStartCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()

	Convey("ProvisionStartCmd extract deps", t, func() {
		ctx := context.Background()
		wantPublishSrcDir := "publish/src/dir"
		sk := &data.HwTestStateKeeper{CpconPublishSrcDir: wantPublishSrcDir}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosCpconPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosCpconPublishExecutorType)
		cmd := commands.NewCpconPublishServiceStartCmd(exec)

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.CpconPublishSrcDir, ShouldEqual, wantPublishSrcDir)
	})
}
