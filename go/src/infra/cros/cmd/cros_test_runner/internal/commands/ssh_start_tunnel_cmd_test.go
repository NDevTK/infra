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
)

func TestSshStartTunnelCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		exec := executors.NewSshTunnelExecutor()
		cmd := commands.NewSshStartTunnelCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestSshStartTunnelCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{}
		exec := executors.NewSshTunnelExecutor()
		cmd := commands.NewSshStartTunnelCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutSshAddress, ShouldBeNil)
	})
}

func TestSshStartTunnelCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()

	hostname := "DUT1234"

	Convey("SshStartTunnelCmd extract deps", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{HwTestStateKeeper: data.HwTestStateKeeper{HostName: hostname}}
		exec := executors.NewSshTunnelExecutor()
		cmd := commands.NewSshStartTunnelCmd(exec)

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.HostName, ShouldEqual, hostname)
	})
}

func TestSshStartTunnelCmd_UpdateSKSuccess(t *testing.T) {
	t.Parallel()

	hostname := "DUT1234"

	Convey("SshStartTunnelCmd update SK", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{HwTestStateKeeper: data.HwTestStateKeeper{HostName: hostname}}
		exec := executors.NewSshTunnelExecutor()
		cmd := commands.NewSshStartTunnelCmd(exec)
		cmd.SshTunnelPort = 1234

		// Update SK
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutSshAddress, ShouldNotBeNil)
	})
}
