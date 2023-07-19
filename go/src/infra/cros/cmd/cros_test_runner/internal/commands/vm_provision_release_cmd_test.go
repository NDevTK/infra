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
	"go.chromium.org/chromiumos/config/go/test/api"
)

func TestVMProvisionReleaseCmd_NoDeps(t *testing.T) {
	t.Parallel()
	Convey("No deps", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosVMProvisionExecutor(cont)
		cmd := commands.NewVMProvisionReleaseCmd(exec)
		sk := &data.HwTestStateKeeper{}
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
	Convey("No deps - GceRegion", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosVMProvisionExecutor(cont)
		cmd := commands.NewVMProvisionReleaseCmd(exec)
		sk := &data.HwTestStateKeeper{LeaseVMResponse: &api.LeaseVMResponse{LeaseId: "xyz"}}
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
	Convey("No deps - LeaseID", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosVMProvisionExecutor(cont)
		cmd := commands.NewVMProvisionReleaseCmd(exec)
		sk := &data.HwTestStateKeeper{LeaseVMResponse: &api.LeaseVMResponse{Vm: &api.VM{GceRegion: "region1"}}}
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestVMProvisionReleaseCmd_Updates(t *testing.T) {
	t.Parallel()
	Convey("No updates", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosVMProvisionExecutor(cont)
		cmd := commands.NewVMProvisionReleaseCmd(exec)
		sk := &data.HwTestStateKeeper{LeaseVMResponse: &api.LeaseVMResponse{LeaseId: "xyz", Vm: &api.VM{GceRegion: "region1"}}}
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutTopology, ShouldBeNil)
	})
}
