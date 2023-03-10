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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
)

func TestCacheServerStartCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{Args: &data.LocalArgs{}}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCacheServerTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCacheServerExecutor(cont)
		cmd := commands.NewCacheServerStartCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestCacheServerStartCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()

	Convey("CacheServerStartCmd extract deps", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{Args: &data.LocalArgs{}}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCacheServerTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCacheServerExecutor(cont)
		cmd := commands.NewCacheServerStartCmd(exec)

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestCacheServerStartCmd_UpdateSKSuccess(t *testing.T) {
	t.Parallel()
	Convey("CacheServerStartCmd update SK", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{Args: &data.LocalArgs{}}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCacheServerTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCacheServerExecutor(cont)
		cmd := commands.NewCacheServerStartCmd(exec)
		cmd.CacheServerAddress = &labapi.IpEndpoint{}

		// Update SK
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.CacheServerAddress, ShouldNotBeNil)
	})
}
