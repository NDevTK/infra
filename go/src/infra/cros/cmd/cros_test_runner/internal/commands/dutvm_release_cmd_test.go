// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"

	"infra/cros/cmd/common_lib/common_executors"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/data"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	vmlabapi "infra/libs/vmlab/api"
)

func buildDutVmReleaseCmdForTest() *commands.DutVmReleaseCmd {
	ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
	ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
	exec := common_executors.NewCtrExecutor(ctr)
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
	ctx := context.Background()
	Convey("Cmd missing deps name", t, func() {
		Convey("Cmd missing deps name - gcloud backend", func() {
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
		Convey("Cmd missing deps name - vm leaser backend", func() {
			sk := &data.HwTestStateKeeper{
				DutVm: &vmlabapi.VmInstance{
					Config: &vmlabapi.Config{
						Backend: &vmlabapi.Config_VmLeaserBackend_{
							VmLeaserBackend: &vmlabapi.Config_VmLeaserBackend{
								VmRequirements: &api.VMRequirements{
									GceProject: "test-project",
								},
							},
						},
					},
					GceRegion: "test-region",
				},
			}
			cmd := buildDutVmReleaseCmdForTest()
			err := cmd.ExtractDependencies(ctx, sk)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestDutVmReleaseCmd_MissingDepsProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Cmd missing deps project", t, func() {
		Convey("Cmd missing deps project - gcloud backend", func() {
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
		Convey("Cmd missing deps project - vm leaser backend", func() {
			sk := &data.HwTestStateKeeper{
				DutVm: &vmlabapi.VmInstance{
					Name: "instance1",
					Config: &vmlabapi.Config{
						Backend: &vmlabapi.Config_VmLeaserBackend_{
							VmLeaserBackend: &vmlabapi.Config_VmLeaserBackend{
								VmRequirements: &api.VMRequirements{},
							},
						},
					},
					GceRegion: "test-region",
				},
			}
			cmd := buildDutVmReleaseCmdForTest()
			err := cmd.ExtractDependencies(ctx, sk)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestDutVmReleaseCmd_MissingDepsZone(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Cmd missing deps zone", t, func() {
		Convey("Cmd missing deps zone - gcloud backend", func() {
			sk := &data.HwTestStateKeeper{
				DutVm: &vmlabapi.VmInstance{
					Name: "instance1",
					Config: &vmlabapi.Config{
						Backend: &vmlabapi.Config_GcloudBackend{
							GcloudBackend: &vmlabapi.Config_GCloudBackend{
								Project: "test-project",
							},
						},
					},
				},
			}
			cmd := buildDutVmReleaseCmdForTest()
			err := cmd.ExtractDependencies(ctx, sk)
			So(err, ShouldNotBeNil)
		})
		Convey("Cmd missing deps zone - vm leaser backend", func() {
			sk := &data.HwTestStateKeeper{
				DutVm: &vmlabapi.VmInstance{
					Name: "instance1",
					Config: &vmlabapi.Config{
						Backend: &vmlabapi.Config_VmLeaserBackend_{
							VmLeaserBackend: &vmlabapi.Config_VmLeaserBackend{
								VmRequirements: &api.VMRequirements{
									GceProject: "test-project",
								},
							},
						},
					},
				},
			}
			cmd := buildDutVmReleaseCmdForTest()
			err := cmd.ExtractDependencies(ctx, sk)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestDutVmReleaseCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Cmd extract deps success", t, func() {
		Convey("Cmd extract deps success - gcloud backend", func() {
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
		Convey("Cmd extract deps success - vm leaser backend", func() {
			sk := &data.HwTestStateKeeper{
				DutVm: &vmlabapi.VmInstance{
					Name: "instance1",
					Config: &vmlabapi.Config{
						Backend: &vmlabapi.Config_VmLeaserBackend_{
							VmLeaserBackend: &vmlabapi.Config_VmLeaserBackend{
								VmRequirements: &api.VMRequirements{
									GceProject: "test-project",
								},
							},
						},
					},
					GceRegion: "test-region",
				},
			}
			cmd := buildDutVmReleaseCmdForTest()
			err := cmd.ExtractDependencies(ctx, sk)
			So(err, ShouldBeNil)
			So(cmd.DutVm, ShouldEqual, sk.DutVm)
		})
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
