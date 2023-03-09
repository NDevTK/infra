// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"testing"

	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
	vmlabapi "infra/libs/vmlab/api"

	. "github.com/smartystreets/goconvey/convey"
)

func buildDutVmReleaseCmdForTest() *commands.DutVmReleaseCmd {
	ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
	ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
	exec := executors.NewCtrExecutor(ctr)
	cmd := commands.NewDutVmReleaseCmd(exec)
	return cmd
}

func TestDutVmReleaseCmd_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		cmd := buildDutVmReleaseCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmReleaseCmd_MissingDepsName(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps name", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			DutVm: &vmlabapi.VmInstance{
				Config: &vmlabapi.Config{
					Backend: &vmlabapi.Config_GcloudBackend{
						GcloudBackend: &vmlabapi.Config_GCloudBackend{
							Project: "vmlab-project",
							Zone:    "us-west-2",
						},
					},
				},
			},
		}
		cmd := buildDutVmReleaseCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmReleaseCmd_MissingDepsProject(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps project", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			DutVm: &vmlabapi.VmInstance{
				Name: "instance1",
				Config: &vmlabapi.Config{
					Backend: &vmlabapi.Config_GcloudBackend{
						GcloudBackend: &vmlabapi.Config_GCloudBackend{
							Zone: "us-west-2",
						},
					},
				},
			},
		}
		cmd := buildDutVmReleaseCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmReleaseCmd_MissingDepsZone(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps project", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			DutVm: &vmlabapi.VmInstance{
				Name: "instance1",
				Config: &vmlabapi.Config{
					Backend: &vmlabapi.Config_GcloudBackend{
						GcloudBackend: &vmlabapi.Config_GCloudBackend{
							Project: "vmlab-project",
						},
					},
				},
			},
		}
		cmd := buildDutVmReleaseCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmReleaseCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()
	Convey("Cmd extract deps success", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			DutVm: &vmlabapi.VmInstance{
				Name: "instance1",
				Config: &vmlabapi.Config{
					Backend: &vmlabapi.Config_GcloudBackend{
						GcloudBackend: &vmlabapi.Config_GCloudBackend{
							Project: "vmlab-project",
							Zone:    "us-west-2",
						},
					},
				},
			},
		}
		cmd := buildDutVmReleaseCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.DutVm, ShouldEqual, sk.DutVm)
	})
}

func TestDutVmReleaseCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd update SK", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			DutVm: nil,
		}
		cmd := buildDutVmReleaseCmdForTest()
		cmd.DutVm = &vmlabapi.VmInstance{Name: "some-instance"}

		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutVm, ShouldBeNil)
	})
}
