// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/dutstate"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdateDutStateCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		cmd := commands.NewUpdateDutStateCmd()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestUpdateDutStateCmd_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		cmd := commands.NewUpdateDutStateCmd()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestUpdateDutStateCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Update SK", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CurrentDutState: dutstate.Ready}
		cmd := commands.NewUpdateDutStateCmd()
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.CurrentDutState, ShouldNotBeNil)
	})
}

func TestUpdateDutStateCmd_Execute(t *testing.T) {
	t.Parallel()
	Convey("TestUpdateDutState execute", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			HostName: "host",
		}
		cmd := commands.NewUpdateDutStateCmd()

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)

		// Execute cmd
		err = cmd.Execute(ctx)
		So(err, ShouldBeNil)

		// Update SK
		err = cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})

}
