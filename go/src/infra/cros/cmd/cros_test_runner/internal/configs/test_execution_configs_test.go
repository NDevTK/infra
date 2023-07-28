// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
)

func TestGenerateConfig_UnSupportedConfig(t *testing.T) {
	t.Parallel()
	Convey("Unsupported test execution config type", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := NewCftContainerConfig(ctr, nil, false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)
		sk := &data.HwTestStateKeeper{}
		testExecConfig := NewTestExecutionConfig(UnSupportedTestExecutionConfigType, cmdConfig, sk, nil)
		err := testExecConfig.GenerateConfig(ctx, common.BotProviderUnknown)
		So(err, ShouldNotBeNil)
	})
}

func TestGenerateConfig_SupportedConfig(t *testing.T) {
	t.Parallel()
	Convey("Supported test execution config type", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := NewCftContainerConfig(ctr, nil, false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)
		sk := &data.HwTestStateKeeper{}
		testExecConfig := NewTestExecutionConfig(HwTestExecutionConfigType, cmdConfig, sk, nil)
		err := testExecConfig.GenerateConfig(ctx, common.BotProviderUnknown)
		So(err, ShouldBeNil)
	})
}

func TestExecute_WithoutGeneratedConfig(t *testing.T) {
	t.Parallel()
	Convey("Execute without generating configs", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := NewCftContainerConfig(ctr, nil, false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)
		sk := &data.HwTestStateKeeper{}
		testExecConfig := NewTestExecutionConfig(HwTestExecutionConfigType, cmdConfig, sk, nil)
		err := testExecConfig.Execute(ctx)
		So(err, ShouldNotBeNil)
	})
}

func TestExecute_UnsuccesfulHwTestsExecution(t *testing.T) {
	t.Parallel()
	Convey("Execute hw tests with failure", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := NewCftContainerConfig(ctr, nil, false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)
		sk := &data.HwTestStateKeeper{}
		testExecConfig := NewTestExecutionConfig(HwTestExecutionConfigType, cmdConfig, sk, nil)

		// Generate configs first
		err := testExecConfig.GenerateConfig(ctx, common.BotProviderUnknown)
		So(err, ShouldBeNil)

		// Execute configs
		err = testExecConfig.Execute(ctx)
		So(err, ShouldNotBeNil)
	})
}

func TestExecute_SuccesfulHwTestsExecution(t *testing.T) {
	t.Parallel()
	Convey("Execute hw tests successfully", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := NewCftContainerConfig(ctr, getMockContainerImagesInfo(), false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)
		sk := &data.HwTestStateKeeper{
			CftTestRequest: &skylab_test_runner.CFTTestRequest{
				ParentBuildId: 12345678,
			},
		}
		testExecConfig := NewTestExecutionConfig(HwTestExecutionConfigType, cmdConfig, sk, nil)

		// Use mock configs for simplicity
		testExecConfig.configs = getMockedHwTestConfig()

		// Execute configs
		err := testExecConfig.Execute(ctx)
		So(err, ShouldBeNil)
	})
}

func TestExecute_SuccesfulHwTestsExecution_Gce(t *testing.T) {
	t.Parallel()
	Convey("Execute hw tests successfully on Gce", t, func() {
		wantPublishSrcDir := "publish/src/dir"
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := NewCftContainerConfig(ctr, getMockContainerImagesInfo(), false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)
		sk := &data.HwTestStateKeeper{
			GcsPublishSrcDir: wantPublishSrcDir,
			CftTestRequest: &skylab_test_runner.CFTTestRequest{
				ParentBuildId: 12345678,
			},
		}
		testExecConfig := NewTestExecutionConfig(HwTestExecutionConfigType, cmdConfig, sk, nil)

		// Generate configs first
		err := testExecConfig.GenerateConfig(ctx, common.BotProviderGce)
		So(err, ShouldBeNil)

		// Execute configs
		err = testExecConfig.Execute(ctx)
		So(err, ShouldBeNil)
	})
}

// FIXME: add test with failed DutVmGetImage command

func getMockedHwTestConfig() *Configs {
	mainConfigs := []*CommandExecutorPairedConfig{
		InputValidation_NoExecutor,
		ParseEnvInfo_NoExecutor,
	}

	// This should be skipped
	cleanupConfigs := []*CommandExecutorPairedConfig{
		ParseEnvInfo_NoExecutor,
	}

	return &Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}
