// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_executors_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"infra/cros/cmd/common_lib/common_commands"
	"infra/cros/cmd/common_lib/common_executors"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
)

type UnsupportedCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor
}

func NewUnsupportedCmd() interfaces.CommandInterface {
	absCmd := interfaces.NewAbstractCmd(common_commands.UnSupportedCmdType)
	absSingleCmdByNoExec := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: absCmd}
	return &UnsupportedCmd{AbstractSingleCmdByNoExecutor: absSingleCmdByNoExec}
}

func TestCtrServiceStartAsync(t *testing.T) {
	t.Parallel()

	Convey("ctr initialization error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := common_executors.NewCtrExecutor(ctr)
		err := exec.StartAsync(ctx)
		So(err, ShouldNotBeNil)
	})
}

func TestGcloudAuth(t *testing.T) {
	t.Parallel()

	Convey("gcloud auth error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := common_executors.NewCtrExecutor(ctr)
		err := exec.GcloudAuth(ctx, "", false)
		So(err, ShouldNotBeNil)
	})
}

func TestCtrStop(t *testing.T) {
	t.Parallel()

	Convey("gcloud auth error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := common_executors.NewCtrExecutor(ctr)
		err := exec.Stop(ctx)
		So(err, ShouldBeNil)
	})
}

func TestCtrExecuteCommand(t *testing.T) {
	t.Parallel()

	Convey("unsupported cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := common_executors.NewCtrExecutor(ctr)
		err := exec.ExecuteCommand(ctx, NewUnsupportedCmd())
		So(err, ShouldNotBeNil)
	})

	Convey("start async cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := common_executors.NewCtrExecutor(ctr)
		cmd := common_commands.NewCtrServiceStartAsyncCmd(exec)
		err := exec.ExecuteCommand(ctx, cmd)
		So(err, ShouldNotBeNil)
	})

	Convey("stop cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := common_executors.NewCtrExecutor(ctr)
		cmd := common_commands.NewCtrServiceStopCmd(exec)
		err := exec.ExecuteCommand(ctx, cmd)
		So(err, ShouldBeNil)
	})

	Convey("gcloud auth cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := common_executors.NewCtrExecutor(ctr)
		cmd := common_commands.NewGcloudAuthCmd(exec)
		err := exec.ExecuteCommand(ctx, cmd)
		So(err, ShouldNotBeNil)
	})
}
