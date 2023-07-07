// Copyright 2023 The Chromium Authors
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
	vmlabapi "infra/libs/vmlab/api"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"
)

func TestVMProvisionLeaseCmd_NoDeps(t *testing.T) {
	t.Parallel()
	Convey("No deps", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosVMProvisionExecutor(cont)
		cmd := commands.NewVMProvisionLeaseCmd(exec)
		sk := &data.HwTestStateKeeper{}
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
	Convey("No deps - name", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosVMProvisionExecutor(cont)
		cmd := commands.NewVMProvisionLeaseCmd(exec)
		sk := &data.HwTestStateKeeper{DutVmGceImage: &vmlabapi.GceImage{Project: "project"}}
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
	Convey("No deps - project", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosVMProvisionExecutor(cont)
		cmd := commands.NewVMProvisionLeaseCmd(exec)
		sk := &data.HwTestStateKeeper{DutVmGceImage: &vmlabapi.GceImage{Name: "name"}}
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestVMProvisionLeaseCmd_Updates(t *testing.T) {
	t.Parallel()
	Convey("No updates", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosVMProvisionExecutor(cont)
		cmd := commands.NewVMProvisionLeaseCmd(exec)
		cmd.LeaseVMResponse = &api.LeaseVMResponse{}
		sk := &data.HwTestStateKeeper{DutVmGceImage: &vmlabapi.GceImage{Name: "name", Project: "project"}}
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutTopology, ShouldNotBeNil)
	})
}
