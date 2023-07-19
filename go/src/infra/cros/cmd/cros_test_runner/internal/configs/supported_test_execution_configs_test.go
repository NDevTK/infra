// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"
	"infra/cros/cmd/cros_test_runner/data"
	"testing"

	"infra/cros/cmd/common_lib/common"

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

	Convey("hwConfigsForPlatform for VM", t, func() {
		hwConfigs := hwConfigsForPlatform(nil, common.BotProviderGce)

		So(hwConfigs.MainConfigs, ShouldContain, VMProvisionRelease_CrosVMProvisionExecutor)
		So(hwConfigs.CleanupConfigs, ShouldContain, VMProvisionRelease_CrosVMProvisionExecutor)
		So(hwConfigs.MainConfigs, ShouldNotContain, DutServerStart_CrosDutExecutor)
		So(hwConfigs.MainConfigs, ShouldNotContain, UpdateDutState_NoExecutor)
		So(hwConfigs.CleanupConfigs, ShouldNotContain, UpdateDutState_NoExecutor)
	})

	Convey("hwConfigsForPlatform for HW", t, func() {
		hwConfigs := hwConfigsForPlatform(nil, common.BotProviderDrone)

		So(hwConfigs.MainConfigs, ShouldContain, DutServerStart_CrosDutExecutor)
		So(hwConfigs.MainConfigs, ShouldContain, UpdateDutState_NoExecutor)
		So(hwConfigs.CleanupConfigs, ShouldContain, UpdateDutState_NoExecutor)
		So(hwConfigs.MainConfigs, ShouldNotContain, DutVmRelease_CrosDutVmExecutor)
		So(hwConfigs.CleanupConfigs, ShouldNotContain, DutVmRelease_CrosDutVmExecutor)
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
