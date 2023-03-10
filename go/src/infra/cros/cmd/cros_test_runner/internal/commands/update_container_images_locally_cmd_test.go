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
	"go.chromium.org/chromiumos/config/go/build/api"
)

func TestUpdateContainerImagesLocallyCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		cmd := commands.NewUpdateContainerImagesLocallyCmd()
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestUpdateContainerImagesLocallyCmd_NoDeps(t *testing.T) {
	t.Parallel()
	Convey("No deps", t, func() {
		ctx := context.Background()
		sk := &data.PreLocalTestStateKeeper{Args: &data.LocalArgs{}}
		cmd := commands.NewUpdateContainerImagesLocallyCmd()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestUpdateContainerImagesLocallyCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with updates", t, func() {
		ctx := context.Background()
		sk := &data.PreLocalTestStateKeeper{}
		cmd := commands.NewUpdateContainerImagesLocallyCmd()
		cmd.Containers = map[string]*api.ContainerImageInfo{
			"test": {
				Name: "testcontainer",
			},
		}
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.ContainerImages, ShouldEqual, cmd.Containers)
	})
}
