// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"fmt"
	"infra/cros/cmd/common_lib/containers"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/data"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/executors"
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
		cont := containers.NewGenericTemplatedContainer(
			containers.CrosPublishTemplatedContainerType,
			"container/image/path",
			"cros-publish",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosPublishExecutorType)
		cmd := commands.NewCpconPublishUploadCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestCpconPublishPublishCmd_MissingDeps(t *testing.T) {
	t.Setenv("SWARMING_TASK_ID", "")

	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewGenericTemplatedContainer(
			containers.CrosPublishTemplatedContainerType,
			"container/image/path",
			"cros-publish",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosPublishExecutorType)
		cmd := commands.NewCpconPublishUploadCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestCpconPublishPublishCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewGenericTemplatedContainer(
			containers.CrosPublishTemplatedContainerType,
			"container/image/path",
			"cros-publish",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosPublishExecutorType)
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
		wantGcsURL := "gs://this-is-a-gcs-path/results"
		sk := &data.HwTestStateKeeper{
			GcsURL:             wantGcsURL,
			CpconPublishSrcDir: "this/is/a/fake/path",
		}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewGenericTemplatedContainer(
			containers.CrosPublishTemplatedContainerType,
			"container/image/path",
			"cros-publish",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosPublishExecutorType)
		cmd := commands.NewCpconPublishUploadCmd(exec)

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.CpconJobName, ShouldEqual, fmt.Sprintf("swarming-%s", wantSwarmingTaskId))
		So(cmd.GcsURL, ShouldEqual, wantGcsURL)
	})

}
