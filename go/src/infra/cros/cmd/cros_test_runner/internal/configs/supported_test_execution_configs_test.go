// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateHwConfigs(t *testing.T) {
	t.Parallel()
	Convey("GenerateHwConfigs", t, func() {
		ctx := context.Background()
		hwConfigs := GenerateHwConfigs(ctx, nil)

		So(hwConfigs, ShouldNotBeNil)
		So(hwConfigs.MainConfigs, ShouldNotBeNil)
		So(hwConfigs.CleanupConfigs, ShouldNotBeNil)
		So(len(hwConfigs.MainConfigs), ShouldBeGreaterThan, 0)
		So(len(hwConfigs.CleanupConfigs), ShouldBeGreaterThan, 0)
	})
}

func TestGeneratePreLocalConfigs(t *testing.T) {
	t.Parallel()
	Convey("GeneratePreLocalConfigs", t, func() {
		ctx := context.Background()
		preLocalConfigs := GeneratePreLocalConfigs(ctx)

		So(preLocalConfigs, ShouldNotBeNil)
		So(preLocalConfigs.MainConfigs, ShouldNotBeNil)
		So(len(preLocalConfigs.MainConfigs), ShouldBeGreaterThan, 0)
	})
}

func TestGenerateLocalConfigs(t *testing.T) {
	t.Parallel()
	Convey("GenerateLocalConfigs", t, func() {
		ctx := context.Background()
		localConfigs := GenerateLocalConfigs(ctx, &data.LocalTestStateKeeper{Args: &data.LocalArgs{}})

		So(localConfigs, ShouldNotBeNil)
		So(localConfigs.MainConfigs, ShouldNotBeNil)
		So(localConfigs.CleanupConfigs, ShouldNotBeNil)
		So(len(localConfigs.MainConfigs), ShouldBeGreaterThan, 0)
		So(len(localConfigs.CleanupConfigs), ShouldBeGreaterThan, 0)
	})
}
