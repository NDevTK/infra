// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"infra/cros/cmd/common_lib/containers"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/data"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	api "go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestAndroidCompanionDutServiceStartCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewAndroidDutTemplatedContainer("container/image/path", ctr)
		exec := executors.NewAndroidDutExecutor(cont)
		cmd := commands.NewAndroidCompanionDutServiceStartCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestAndroidCompanionDutServiceStartCmd_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewAndroidDutTemplatedContainer("container/image/path", ctr)
		exec := executors.NewAndroidDutExecutor(cont)
		cmd := commands.NewAndroidCompanionDutServiceStartCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestAndroidCompanionDutServiceStartCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewAndroidDutTemplatedContainer("container/image/path", ctr)
		exec := executors.NewAndroidDutExecutor(cont)
		cmd := commands.NewAndroidCompanionDutServiceStartCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestAndroidCompanionDutServiceStartCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()

	Convey("AndroidCompanionDutServiceStartCmd extract deps", t, func() {
		ctx := context.Background()
		dutTopo := &labapi.DutTopology{
			Duts: []*labapi.Dut{
				{
					CacheServer: &labapi.CacheServer{Address: &labapi.IpEndpoint{}},
					DutType: &labapi.Dut_Chromeos{
						Chromeos: &labapi.Dut_ChromeOS{
							Ssh: &labapi.IpEndpoint{},
						},
					},
				},
				{
					DutType: &labapi.Dut_Android_{
						Android: &labapi.Dut_Android{
							DutModel: &labapi.DutModel{
								BuildTarget: "pixel6",
							},
						},
					},
				},
			},
		}
		androidProvisionRequestMetadata := &api.AndroidProvisionRequestMetadata{
			AndroidOsImage: &api.AndroidOsImage{
				LocationOneof: &api.AndroidOsImage_OsVersion{
					OsVersion: "R98.3451.0.1",
				},
			},
		}
		companionDevices := []*api.CrosTestRequest_Device{
			{
				Dut: &labapi.Dut{
					DutType: &labapi.Dut_Android_{
						Android: &labapi.Dut_Android{
							DutModel: &labapi.DutModel{
								BuildTarget: "pixel6",
							},
						},
					},
				},
			},
		}
		primaryDevice := &api.CrosTestRequest_Device{
			Dut: dutTopo.Duts[0],
		}

		provisionMetadata, _ := anypb.New(androidProvisionRequestMetadata)

		cftTestReq := &skylab_test_runner.CFTTestRequest{
			CompanionDuts: []*skylab_test_runner.CFTTestRequest_Device{
				{
					DutModel: &labapi.DutModel{
						BuildTarget: "pixel6",
					},
					ProvisionState: &api.ProvisionState{
						ProvisionMetadata: provisionMetadata,
					},
				},
			},
		}
		sk := &data.HwTestStateKeeper{CftTestRequest: cftTestReq, CompanionDevices: companionDevices, PrimaryDevice: primaryDevice}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewAndroidDutTemplatedContainer("container/image/path", ctr)
		exec := executors.NewAndroidDutExecutor(cont)
		cmd := commands.NewAndroidCompanionDutServiceStartCmd(exec)

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestAndroidCompanionDutServiceStartCmd_UpdateSKSuccess(t *testing.T) {
	t.Parallel()
	Convey("AndroidCompanionDutServiceStartCmd update SK", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{HostName: "DUT-1234"}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewAndroidDutTemplatedContainer("container/image/path", ctr)
		exec := executors.NewAndroidDutExecutor(cont)
		cmd := commands.NewAndroidCompanionDutServiceStartCmd(exec)
		cmd.AndroidDutServerAddress = &labapi.IpEndpoint{}

		// Update SK
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.AndroidDutServerAddress, ShouldNotBeNil)
	})
}
