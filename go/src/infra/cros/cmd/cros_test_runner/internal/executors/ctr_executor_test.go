// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executors

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
)

func TestCtrServiceStartAsync(t *testing.T) {
	t.Parallel()

	Convey("ctr initialization error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := NewCtrExecutor(ctr)
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
		exec := NewCtrExecutor(ctr)
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
		exec := NewCtrExecutor(ctr)
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
		exec := NewCtrExecutor(ctr)
		err := exec.ExecuteCommand(ctx, NewUnsupportedCmd())
		So(err, ShouldNotBeNil)
	})

	Convey("start async cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := NewCtrExecutor(ctr)
		cmd := commands.NewCtrServiceStartAsyncCmd(exec)
		err := exec.ExecuteCommand(ctx, cmd)
		So(err, ShouldNotBeNil)
	})

	Convey("stop cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := NewCtrExecutor(ctr)
		cmd := commands.NewCtrServiceStopCmd(exec)
		err := exec.ExecuteCommand(ctx, cmd)
		So(err, ShouldBeNil)
	})

	Convey("gcloud auth cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := NewCtrExecutor(ctr)
		cmd := commands.NewGcloudAuthCmd(exec)
		err := exec.ExecuteCommand(ctx, cmd)
		So(err, ShouldNotBeNil)
	})
}
