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

func TestParseArgsCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		cmd := commands.NewParseArgsCmd()
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestParseArgsCmd_NoDeps(t *testing.T) {
	t.Parallel()
	Convey("No deps", t, func() {
		ctx := context.Background()
		sk := &data.PreLocalTestStateKeeper{}
		cmd := commands.NewParseArgsCmd()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestParseArgsCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with updates", t, func() {
		ctx := context.Background()
		sk := &data.PreLocalTestStateKeeper{}
		cmd := commands.NewParseArgsCmd()
		cmd.Tests = []string{"test"}
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.Tests, ShouldResemble, cmd.Tests)
	})
}

func TestParseArgsCmd_Execute(t *testing.T) {
	Convey("ParseArgsCmd execute", t, func() {
		ctx := context.Background()
		sk := &data.PreLocalTestStateKeeper{Args: &data.LocalArgs{Tests: "test1,test2"}}
		cmd := commands.NewParseArgsCmd()

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)

		// Execute cmd
		err = cmd.Execute(ctx)
		So(err, ShouldBeNil)

		// Update SK
		err = cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)

		// Check if SK data is expected
		So(sk.Tests, ShouldResemble, []string{"test1", "test2"})
	})
}
