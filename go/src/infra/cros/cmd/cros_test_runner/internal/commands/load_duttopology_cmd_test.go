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

func TestLoadDutTopologyCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		exec := executors.NewInvServiceExecutor("")
		cmd := commands.NewLoadDutTopologyCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestLoadDutTopologyCmd_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		exec := executors.NewInvServiceExecutor("")
		cmd := commands.NewLoadDutTopologyCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestLoadDutTopologyCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		exec := executors.NewInvServiceExecutor("")
		cmd := commands.NewLoadDutTopologyCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestLoadDutTopologyCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()

	hostName := "DUT-1234"

	Convey("BuildInputValidationCmd extract deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{HostName: hostName}
		exec := executors.NewInvServiceExecutor("")
		cmd := commands.NewLoadDutTopologyCmd(exec)

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.HostName, ShouldEqual, hostName)
	})
}

func TestLoadDutTopologyCmd_UpdateSKSuccess(t *testing.T) {
	t.Parallel()
	Convey("BuildInputValidationCmd update SK", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{HostName: "DUT-1234"}
		exec := executors.NewInvServiceExecutor("")
		cmd := commands.NewLoadDutTopologyCmd(exec)
		cmd.DutTopology = &labapi.DutTopology{}

		// Update SK
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.DutTopology, ShouldNotBeNil)
	})
}
