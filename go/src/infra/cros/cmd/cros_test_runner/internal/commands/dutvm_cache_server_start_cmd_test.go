// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"testing"

	"infra/cros/cmd/common_lib/common_executors"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/data"
	"infra/cros/cmd/cros_test_runner/internal/commands"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
)

func buildDutVmCacheServerStartCmdForTest() *commands.DutVmCacheServerStartCmd {
	ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
	ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
	exec := common_executors.NewCtrExecutor(ctr)
	cmd := commands.NewDutVmCacheServerStartCmd(exec)
	return cmd
}

func TestDutVmCacheServerStartCmd_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		cmd := buildDutVmCacheServerStartCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmCacheServerStartCmd_MissingDepsPrimaryDut(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps primary dut", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			DutTopology: &labapi.DutTopology{},
		}
		cmd := buildDutVmCacheServerStartCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmCacheServerStartCmd_MissingDepsSsh(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps primary dut", t, func() {
		ctx := context.Background()
		duts := []*labapi.Dut{{
			Id: &labapi.Dut_Id{Value: "VM"},
			DutType: &labapi.Dut_Chromeos{
				Chromeos: &labapi.Dut_ChromeOS{},
			}}}
		sk := &data.HwTestStateKeeper{
			PrimaryDevice: &api.CrosTestRequest_Device{
				Dut: duts[0],
			},
			DutTopology: &labapi.DutTopology{
				Duts: duts,
			},
		}
		cmd := buildDutVmCacheServerStartCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestDutVmCacheServerStartCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()
	Convey("Cmd extract deps success", t, func() {
		ctx := context.Background()
		duts := []*labapi.Dut{{
			Id: &labapi.Dut_Id{Value: "VM"},
			DutType: &labapi.Dut_Chromeos{
				Chromeos: &labapi.Dut_ChromeOS{
					Ssh: &labapi.IpEndpoint{
						Address: "1.2.3.4",
						Port:    22,
					},
				},
			}}}
		sk := &data.HwTestStateKeeper{
			PrimaryDevice: &api.CrosTestRequest_Device{
				Dut: duts[0],
			},
			DutTopology: &labapi.DutTopology{
				Duts: duts,
			},
		}
		cmd := buildDutVmCacheServerStartCmdForTest()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.DutTopology, ShouldEqual, sk.DutTopology)
	})
}

func TestDutVmCacheServerStartCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd update SK", t, func() {
		ctx := context.Background()
		duts := []*labapi.Dut{{
			Id: &labapi.Dut_Id{Value: "VM"},
			DutType: &labapi.Dut_Chromeos{
				Chromeos: &labapi.Dut_ChromeOS{
					Ssh: &labapi.IpEndpoint{
						Address: "1.2.3.4",
						Port:    22,
					},
				},
			}}}
		sk := &data.HwTestStateKeeper{
			PrimaryDevice: &api.CrosTestRequest_Device{
				Dut: duts[0],
			},
			DutTopology: &labapi.DutTopology{
				Duts: duts,
			},
		}
		cmd := buildDutVmCacheServerStartCmdForTest()
		cmd.CacheServerAddress = &labapi.IpEndpoint{
			Address: "4.3.2.1",
			Port:    8080,
		}

		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutTopology.Duts[0].CacheServer.Address, ShouldEqual, cmd.CacheServerAddress)
	})
}

func TestDutVmCacheServerStartCmd_UpdateSKMissingDutTopology(t *testing.T) {
	t.Parallel()
	Convey("Cmd update SK Missing Deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			DutTopology: nil,
		}
		cmd := buildDutVmCacheServerStartCmdForTest()
		cmd.CacheServerAddress = &labapi.IpEndpoint{
			Address: "4.3.2.1",
			Port:    8080,
		}

		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutTopology, ShouldBeNil)
	})
}
