// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseEnvInfoCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		cmd := commands.NewParseEnvInfoCmd()
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestParseEnvInfoCmd_NoDeps(t *testing.T) {
	t.Parallel()
	Convey("No deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		cmd := commands.NewParseEnvInfoCmd()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestParseEnvInfoCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with updates", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		cmd := commands.NewParseEnvInfoCmd()
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestParseEnvInfoCmd_Execute(t *testing.T) {
	hostName := "DUT-1234"
	Convey("ParseEnvInfoCmd execute", t, func() {
		// Set proper env
		t.Setenv("SWARMING_BOT_ID", hostName)
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		cmd := commands.NewParseEnvInfoCmd()

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)

		// Execute cmd
		err = cmd.Execute(ctx)
		So(err, ShouldBeNil)

		// Update SK
		err = cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)

		// Check if SK data updated
		So(sk.HostName, ShouldEqual, hostName)
	})
}
