// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_commands_test

import (
	"context"
	"infra/cros/cmd/common_lib/common_commands"
	"infra/cros/cmd/common_lib/common_executors"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/data"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type UnsupportedStateKeeper struct {
	interfaces.StateKeeper
}

func TestCtrServiceAsyncStartCmd_NoDeps(t *testing.T) {
	t.Parallel()
	Convey("No deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := common_executors.NewCtrExecutor(ctr)
		cmd := common_commands.NewCtrServiceStartAsyncCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestCtrServiceAsyncStartCmd_NoUpdates(t *testing.T) {
	t.Parallel()
	Convey("No updates", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		exec := common_executors.NewCtrExecutor(ctr)
		cmd := common_commands.NewCtrServiceStartAsyncCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}
