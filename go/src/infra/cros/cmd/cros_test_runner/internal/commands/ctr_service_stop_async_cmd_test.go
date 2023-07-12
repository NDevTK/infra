// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCtrServiceStopCmd_NoDeps(t *testing.T) {
	t.Parallel()
	Convey("No deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := executors.NewCtrExecutor(ctr)
		cmd := commands.NewCtrServiceStopCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestCtrServiceStopCmd_NoUpdates(t *testing.T) {
	t.Parallel()
	Convey("No updates", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := executors.NewCtrExecutor(ctr)
		cmd := commands.NewCtrServiceStopCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}
