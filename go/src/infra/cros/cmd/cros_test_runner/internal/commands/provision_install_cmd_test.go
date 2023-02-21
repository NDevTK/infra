// Copyright 2023 The Chromium OS Authors. All rights reserved.
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
	_go "go.chromium.org/chromiumos/config/go"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
)

func TestProvisionInstallCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosProvisionExecutor(cont)
		cmd := commands.NewProvisionInstallCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestProvisionInstallCmd_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosProvisionExecutor(cont)
		cmd := commands.NewProvisionInstallCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestProvisionInstallCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosProvisionExecutor(cont)
		cmd := commands.NewProvisionInstallCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})

	Convey("Cmd with updates", t, func() {
		ctx := context.Background()
		wantProvisionResp := &api.InstallResponse{Status: api.InstallResponse_STATUS_OK}
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosProvisionExecutor(cont)
		cmd := commands.NewProvisionInstallCmd(exec)
		cmd.ProvisionResp = wantProvisionResp
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.ProvisionResp, ShouldEqual, wantProvisionResp)
	})
}

func TestProvisionInstallCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()

	Convey("ProvisionInstallCmd extract deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			CftTestRequest: &skylab_test_runner.CFTTestRequest{
				PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
					ProvisionState: &api.ProvisionState{
						SystemImage: &api.ProvisionState_SystemImage{
							SystemImagePath: &_go.StoragePath{},
						},
					},
				},
			},
		}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosProvisionExecutor(cont)
		cmd := commands.NewProvisionInstallCmd(exec)

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
	})
}
