// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/cros_test_runner/data"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
)

type UnsupportedStateKeeper struct {
	interfaces.StateKeeper
}

func TestBuildInputValidationCmdDeps_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		cmd := commands.NewBuildInputValidationCmd()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestBuildInputValidationCmdDeps_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		cmd := commands.NewBuildInputValidationCmd()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestBuildInputValidationCmdDeps_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		cmd := commands.NewBuildInputValidationCmd()
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestBuildInputValidationCmdDeps_Execute(t *testing.T) {
	t.Parallel()
	Convey("BuildInputValidationCmd execute", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			CftTestRequest: &skylab_test_runner.CFTTestRequest{ParentBuildId: 12345678},
		}
		cmd := commands.NewBuildInputValidationCmd()

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
