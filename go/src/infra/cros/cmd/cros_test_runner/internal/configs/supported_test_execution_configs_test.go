// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/chromiumos/config/go/test/api"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/cros_test_runner/data"
)

func TestGenerateHwConfigs(t *testing.T) {
	Convey("GenerateHwConfigs", t, func() {
		ctx := context.Background()
		hwConfigs := GenerateHwConfigs(ctx, nil, nil, false)

		So(hwConfigs, ShouldNotBeNil)
		So(hwConfigs.MainConfigs, ShouldNotBeNil)
		So(len(hwConfigs.MainConfigs), ShouldBeGreaterThan, 0)
	})

	Convey("GenerateHwConfigs with CrosTestRunnerRequest", t, func() {
		ctx := context.Background()
		req := &api.CrosTestRunnerDynamicRequest{
			OrderedTasks: []*api.CrosTestRunnerDynamicRequest_Task{
				{
					OrderedContainerRequests: []*api.ContainerRequest{
						{
							DynamicIdentifier: "container",
						},
					},
					Task: &api.CrosTestRunnerDynamicRequest_Task_Provision{
						Provision: &api.ProvisionTask{},
					},
				},
			},
		}
		hwConfigs := GenerateHwConfigs(ctx, nil, req, false)

		So(hwConfigs, ShouldNotBeNil)
		So(hwConfigs.MainConfigs, ShouldNotBeNil)
		So(len(hwConfigs.MainConfigs), ShouldBeGreaterThan, 0)
		So(hwConfigs.MainConfigs, ShouldContain, GenericProvision_GenericProvisionExecutor)
	})

	Convey("hwConfigsForPlatform for VM", t, func() {
		hwConfigs := hwConfigsForPlatform(nil, common.BotProviderGce, false)

		So(hwConfigs.MainConfigs, ShouldContain, VMProvisionRelease_CrosVMProvisionExecutor.WithRequired(true))
		So(hwConfigs.MainConfigs, ShouldNotContain, DutServerStart_CrosDutExecutor)
		So(hwConfigs.MainConfigs, ShouldNotContain, UpdateDutState_NoExecutor.WithRequired(true))
	})

	Convey("hwConfigsForPlatform for HW", t, func() {
		hwConfigs := hwConfigsForPlatform(nil, common.BotProviderDrone, false)

		So(hwConfigs.MainConfigs, ShouldContain, DutServerStart_CrosDutExecutor)
		So(hwConfigs.MainConfigs, ShouldContain, UpdateDutState_NoExecutor.WithRequired(true))
	})
}

func TestGeneratePreLocalConfigs(t *testing.T) {
	Convey("GeneratePreLocalConfigs", t, func() {
		ctx := context.Background()
		preLocalConfigs := GeneratePreLocalConfigs(ctx)

		So(preLocalConfigs, ShouldNotBeNil)
		So(preLocalConfigs.MainConfigs, ShouldNotBeNil)
		So(len(preLocalConfigs.MainConfigs), ShouldBeGreaterThan, 0)
	})
}

func TestGenerateLocalConfigs(t *testing.T) {
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
