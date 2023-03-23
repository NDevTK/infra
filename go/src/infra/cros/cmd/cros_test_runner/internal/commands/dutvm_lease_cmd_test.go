// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"testing"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
	vmlabapi "infra/libs/vmlab/api"

	. "github.com/smartystreets/goconvey/convey"
)

func buildDutVmLeaseCmdForTest() *commands.DutVmLeaseCmd {
	ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
	ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
	exec := executors.NewCtrExecutor(ctr)
	cmd := commands.NewDutVmLeaseCmd(exec)
	return cmd
}

func buildDutVmLeaseCmdStateKeeper() *data.HwTestStateKeeper {
	sk := &data.HwTestStateKeeper{
		DutVmGceImage: &vmlabapi.GceImage{
			Name:    "some-name",
			Project: "some-project",
		},
		CftTestRequest: &skylab_test_runner.CFTTestRequest{
			PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
				DutModel: &labapi.DutModel{
					BuildTarget: "betty",
					ModelName:   "betty",
				},
			},
		},
	}
	return sk
}

func TestDutVmLeaseCmd_MissingDepsDutVmGceImage(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := buildDutVmLeaseCmdStateKeeper()
		sk.DutVmGceImage = nil
		cmd := buildDutVmLeaseCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmLeaseCmd_MissingDepsName(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps name", t, func() {
		ctx := context.Background()
		sk := buildDutVmLeaseCmdStateKeeper()
		sk.DutVmGceImage.Name = ""
		cmd := buildDutVmLeaseCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmLeaseCmd_MissingDepsProject(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps project", t, func() {
		ctx := context.Background()
		sk := buildDutVmLeaseCmdStateKeeper()
		sk.DutVmGceImage.Project = ""
		cmd := buildDutVmLeaseCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmLeaseCmd_MissingDepsCftTestRequest(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps project", t, func() {
		ctx := context.Background()
		sk := buildDutVmLeaseCmdStateKeeper()
		sk.CftTestRequest = nil
		cmd := buildDutVmLeaseCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmLeaseCmd_MissingDepsCftTestRequestPrimaryDut(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps project", t, func() {
		ctx := context.Background()
		sk := buildDutVmLeaseCmdStateKeeper()
		sk.CftTestRequest.PrimaryDut = nil
		cmd := buildDutVmLeaseCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmLeaseCmd_MissingDepsCftTestRequestDutModel(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps project", t, func() {
		ctx := context.Background()
		sk := buildDutVmLeaseCmdStateKeeper()
		sk.CftTestRequest.PrimaryDut.DutModel = nil
		cmd := buildDutVmLeaseCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmLeaseCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()
	Convey("Cmd extract deps success", t, func() {
		ctx := context.Background()
		sk := buildDutVmLeaseCmdStateKeeper()
		cmd := buildDutVmLeaseCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.DutVmGceImage, ShouldEqual, sk.DutVmGceImage)
	})
}

func TestDutVmLeaseCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd update SK", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			DutVm: nil,
		}
		cmd := buildDutVmLeaseCmdForTest()
		cmd.DutVm = &vmlabapi.VmInstance{
			Name: "some-instance",
			Ssh: &vmlabapi.AddressPort{
				Address: "1.2.3.4",
				Port:    22,
			}}

		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutVm, ShouldEqual, cmd.DutVm)
		dutInTopology := sk.DutTopology.GetDuts()[0].GetChromeos()
		So(dutInTopology.GetSsh().GetAddress(), ShouldEqual, cmd.DutVm.GetSsh().GetAddress())
		So(dutInTopology.GetSsh().GetPort(), ShouldEqual, cmd.DutVm.GetSsh().GetPort())
	})
}
