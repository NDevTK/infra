// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
)

func TestBuildDutTopologyCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		exec := executors.NewInvServiceExecutor("")
		cmd := commands.NewBuildDutTopologyCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestBuildDutTopologyCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{}
		exec := executors.NewInvServiceExecutor("")
		cmd := commands.NewBuildDutTopologyCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.DutTopology, ShouldBeNil)
	})
}

func TestBuildDutTopologyCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()

	board := "kevin"
	dutSshAddress := &labapi.IpEndpoint{Address: "dutssh", Port: 1234}
	cacheServerAddress := &labapi.IpEndpoint{Address: "cacheserver", Port: 4321}

	Convey("BuildInputValidationCmd extract deps", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{Args: &data.LocalArgs{BuildBoard: board}, DutSshAddress: dutSshAddress, DutCacheServerAddress: cacheServerAddress}
		exec := executors.NewInvServiceExecutor("")
		cmd := commands.NewBuildDutTopologyCmd(exec)

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.Board, ShouldEqual, board)
		So(cmd.DutSshAddress, ShouldEqual, dutSshAddress)
		So(cmd.CacheServerAddress, ShouldEqual, cacheServerAddress)
	})
}

func TestBuildDutTopologyCmd_UpdateSKSuccess(t *testing.T) {
	t.Parallel()

	board := "kevin"
	dutSshAddress := &labapi.IpEndpoint{Address: "dutssh", Port: 1234}
	cacheServerAddress := &labapi.IpEndpoint{Address: "cacheserver", Port: 4321}

	Convey("BuildInputValidationCmd update SK", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{Args: &data.LocalArgs{BuildBoard: board}, DutSshAddress: dutSshAddress, DutCacheServerAddress: cacheServerAddress}
		exec := executors.NewInvServiceExecutor("")
		cmd := commands.NewBuildDutTopologyCmd(exec)
		cmd.DutTopology = &labapi.DutTopology{}

		// Update SK
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutTopology, ShouldNotBeNil)
	})
}
