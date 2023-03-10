// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
)

func buildDutVmCacheServerStartCmdForTest() *commands.DutVmCacheServerStartCmd {
	ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
	ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
	exec := executors.NewCtrExecutor(ctr)
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
