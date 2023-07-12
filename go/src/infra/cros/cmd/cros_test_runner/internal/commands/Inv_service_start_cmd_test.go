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

func TestInvServiceStartCmd_NoDeps(t *testing.T) {
	t.Parallel()
	Convey("No deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		exec := executors.NewInvServiceExecutor("")
		cmd := commands.NewInvServiceStartCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestInvServiceStartCmd_NoUpdates(t *testing.T) {
	t.Parallel()
	Convey("No updates", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		exec := executors.NewInvServiceExecutor("")
		cmd := commands.NewInvServiceStartCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}
