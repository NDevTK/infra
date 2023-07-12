// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"infra/cros/cmd/common_lib/containers"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTkoPublishStartCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosTkoPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosTkoPublishExecutorType)
		cmd := commands.NewTkoPublishServiceStartCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestTkoPublishStartCmd_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosTkoPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosTkoPublishExecutorType)
		cmd := commands.NewTkoPublishServiceStartCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestTkoPublishStartCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosTkoPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosTkoPublishExecutorType)
		cmd := commands.NewTkoPublishServiceStartCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestTkoPublishStartCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()

	Convey("ProvisionStartCmd extract deps", t, func() {
		ctx := context.Background()
		wantPublishSrcDir := "publish/src/dir"
		sk := &data.HwTestStateKeeper{TkoPublishSrcDir: wantPublishSrcDir}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosTkoPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosTkoPublishExecutorType)
		cmd := commands.NewTkoPublishServiceStartCmd(exec)

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.TkoPublishSrcDir, ShouldEqual, wantPublishSrcDir)
	})
}
