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

func TestSshStartReverseTunnelCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		exec := executors.NewSshTunnelExecutor()
		cmd := commands.NewSshStartReverseTunnelCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestSshStartReverseTunnelCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{}
		exec := executors.NewSshTunnelExecutor()
		cmd := commands.NewSshStartReverseTunnelCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutCacheServerAddress, ShouldBeNil)
	})
}

func TestSshStartReverseTunnelCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()

	hostname := "DUT1234"
	cacheServerAddress := &labapi.IpEndpoint{Address: "cacheserver", Port: 4321}

	Convey("SshStartReverseTunnelCmd extract deps", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{HwTestStateKeeper: data.HwTestStateKeeper{HostName: hostname}, CacheServerAddress: cacheServerAddress}
		exec := executors.NewSshTunnelExecutor()
		cmd := commands.NewSshStartReverseTunnelCmd(exec)

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.HostName, ShouldEqual, hostname)
		So(cmd.CacheServerPort, ShouldEqual, cacheServerAddress.Port)
	})
}

func TestSshStartReverseTunnelCmd_UpdateSKSuccess(t *testing.T) {
	t.Parallel()

	hostname := "DUT1234"
	cacheServerAddress := &labapi.IpEndpoint{Address: "cacheserver", Port: 4321}

	Convey("SshStartReverseTunnelCmd update SK", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{HwTestStateKeeper: data.HwTestStateKeeper{HostName: hostname}, CacheServerAddress: cacheServerAddress}
		exec := executors.NewSshTunnelExecutor()
		cmd := commands.NewSshStartReverseTunnelCmd(exec)
		cmd.SshReverseTunnelPort = 1234

		// Update SK
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutCacheServerAddress, ShouldNotBeNil)
	})
}
