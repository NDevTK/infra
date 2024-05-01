// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"

	"infra/cros/cmd/common_lib/common_executors"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/data"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	vmlabapi "infra/libs/vmlab/api"
)

func buildDutVmGetImageCmdForTest() *commands.DutVmGetImageCmd {
	ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
	ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
	exec := common_executors.NewCtrExecutor(ctr)
	cmd := commands.NewDutVmGetImageCmd(exec)
	return cmd
}

func TestDutVmGetImageCmd_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		cmd := buildDutVmGetImageCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmGetImageCmd_MissingDepsBuild(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps name", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: &skylab_test_runner.CFTTestRequest{}}
		cmd := buildDutVmGetImageCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmGetImageCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()
	Convey("Cmd extract deps success", t, func() {
		ctx := context.Background()
		keyVals := make(map[string]string, 0)
		keyVals["build"] = "betty/R101"
		sk := &data.HwTestStateKeeper{CftTestRequest: &skylab_test_runner.CFTTestRequest{
			AutotestKeyvals: keyVals,
		}}
		cmd := buildDutVmGetImageCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.DutVmGceImage, ShouldEqual, sk.DutVmGceImage)
	})
}

func TestDutVmGetImageCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd update SK", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			DutVm: nil,
		}
		cmd := buildDutVmGetImageCmdForTest()
		cmd.DutVmGceImage = &vmlabapi.GceImage{
			Name:    "some-name",
			Project: "some-project",
		}

		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutVmGceImage, ShouldEqual, cmd.DutVmGceImage)
	})
}
