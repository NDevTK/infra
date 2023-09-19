// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"
	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/common_configs"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/data"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestGenerateConfig_UnSupportedConfig(t *testing.T) {
	t.Parallel()
	Convey("Unsupported test execution config type", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := common_configs.NewContainerConfig(ctr, nil, false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)
		sk := &data.HwTestStateKeeper{}
		testExecConfig := NewTrv2ExecutionConfig(UnSupportedTestExecutionConfigType, cmdConfig, sk, nil)
		err := testExecConfig.GenerateConfig(ctx)
		So(err, ShouldNotBeNil)
	})
}

func TestGenerateConfig_SupportedConfig(t *testing.T) {
	t.Parallel()
	Convey("Supported test execution config type", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := common_configs.NewContainerConfig(ctr, nil, false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)
		sk := &data.HwTestStateKeeper{}
		testExecConfig := NewTrv2ExecutionConfig(HwTestExecutionConfigType, cmdConfig, sk, nil)
		err := testExecConfig.GenerateConfig(ctx)
		So(err, ShouldBeNil)
	})
}

func TestExecute_WithoutGeneratedConfig(t *testing.T) {
	t.Parallel()
	Convey("Execute without generating configs", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := common_configs.NewContainerConfig(ctr, nil, false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)
		sk := &data.HwTestStateKeeper{}
		testExecConfig := NewTrv2ExecutionConfig(HwTestExecutionConfigType, cmdConfig, sk, nil)
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
		contConfig := common_configs.NewContainerConfig(ctr, nil, false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)
		sk := &data.HwTestStateKeeper{}
		testExecConfig := NewTrv2ExecutionConfig(HwTestExecutionConfigType, cmdConfig, sk, nil)

		// Generate configs first
		err := testExecConfig.GenerateConfig(ctx)
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
		contConfig := common_configs.NewContainerConfig(ctr, getMockContainerImagesInfo(), false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)
		sk := &data.HwTestStateKeeper{
			CftTestRequest: &skylab_test_runner.CFTTestRequest{
				ParentBuildId: 12345678,
			},
			Injectables: common.NewInjectableStorage(),
		}
		testExecConfig := NewTrv2ExecutionConfig(HwTestExecutionConfigType, cmdConfig, sk, nil)

		// Use mock configs for simplicity
		testExecConfig.Configs = getMockedHwTestConfig()

		// Execute configs
		err := testExecConfig.Execute(ctx)
		So(err, ShouldBeNil)
	})
}

func TestIsAndroidProvisionRequired(t *testing.T) {
	t.Parallel()
	Convey("Execute hw tests successfully", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := common_configs.NewContainerConfig(ctr, getMockContainerImagesInfo(), false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		cmdConfig := NewCommandConfig(execConfig)

		sk := &data.HwTestStateKeeper{
			CftTestRequest: &skylab_test_runner.CFTTestRequest{
				ParentBuildId: 12345678,
			},
			Injectables: common.NewInjectableStorage(),
		}
		testExecConfig := NewTrv2ExecutionConfig(HwTestExecutionConfigType, cmdConfig, sk, nil)

		So(testExecConfig.isAndroidProvisioningRequired(ctx), ShouldEqual, false)
		sk.CftTestRequest.CompanionDuts = getAndroidCompanionDuts()
		testExecConfig = NewTrv2ExecutionConfig(HwTestExecutionConfigType, cmdConfig, sk, nil)
		So(testExecConfig.isAndroidProvisioningRequired(ctx), ShouldEqual, true)
	})
}

func getAndroidCompanionDuts() []*skylab_test_runner.CFTTestRequest_Device {
	cipdPackage := &api.CIPDPackage{
		Name: "gmscore_prodrvc_arm64_alldpi_release_apk",
		VersionOneof: &api.CIPDPackage_InstanceId{
			InstanceId: "pLDpI-z2HEUmNChkoCoc1SS7jj4MzaNFijz7_CawdykC",
		},
	}

	androidOsImage := &api.AndroidOsImage{
		LocationOneof: &api.AndroidOsImage_OsVersion{
			OsVersion: "R97.4356.0.1",
		},
	}

	cipdPackages := []*api.CIPDPackage{cipdPackage}

	provisionMetadata, _ := anypb.New(&api.AndroidProvisionRequestMetadata{
		CipdPackages:   cipdPackages,
		AndroidOsImage: androidOsImage,
	})

	companionDuts := []*skylab_test_runner.CFTTestRequest_Device{
		{
			ProvisionState: &api.ProvisionState{
				ProvisionMetadata: provisionMetadata,
			},
		},
	}
	return companionDuts
}

func getMockedHwTestConfig() *common_configs.Configs {
	mainConfigs := []*common_configs.CommandExecutorPairedConfig{
		InputValidation_NoExecutor,
		ParseEnvInfo_NoExecutor,
	}

	// This should be skipped
	cleanupConfigs := []*common_configs.CommandExecutorPairedConfig{
		ParseEnvInfo_NoExecutor,
	}

	return &common_configs.Configs{MainConfigs: mainConfigs, CleanupConfigs: cleanupConfigs}
}
