// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/containers"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/data"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	. "github.com/smartystreets/goconvey/convey"
	_go "go.chromium.org/chromiumos/config/go"
	configpb "go.chromium.org/chromiumos/config/go"
	buildapi "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/config/go/test/api/metadata"
	"go.chromium.org/chromiumos/config/go/test/artifact"
	artifactpb "go.chromium.org/chromiumos/config/go/test/artifact"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func parseTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05.99Z", s)
	return t
}

func TestRdbPublishPublishCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosRdbPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosRdbPublishExecutorType)
		cmd := commands.NewRdbPublishUploadCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestRdbPublishPublishCmd_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosRdbPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosRdbPublishExecutorType)
		cmd := commands.NewRdbPublishUploadCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestRdbPublishPublishCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosRdbPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := executors.NewCrosPublishExecutor(
			cont,
			executors.CrosRdbPublishExecutorType)
		cmd := commands.NewRdbPublishUploadCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestRdbPublishPublishCmd_ExtractSources(t *testing.T) {
	t.Parallel()
	Convey("With CFT Test Request", t, func() {
		request := &skylab_test_runner.CFTTestRequest{
			PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
				ProvisionState: &api.ProvisionState{
					SystemImage: &api.ProvisionState_SystemImage{
						SystemImagePath: &_go.StoragePath{
							HostType: _go.StoragePath_GS,
							Path:     "gs://some-bucket/builder/build-12345",
						},
					},
				},
			},
		}
		expectedSources := &metadata.PublishRdbMetadata_Sources{
			GsPath:            "gs://some-bucket/builder/build-12345/metadata/sources.jsonpb",
			IsDeploymentDirty: false,
		}
		Convey("Base case", func() {
			sources, err := commands.SourcesFromCFTTestRequest(request)
			So(err, ShouldBeNil)
			So(sources, ShouldResembleProto, expectedSources)
		})
		Convey("Invalid input", func() {
			Convey("No gs:// prefix", func() {
				request.PrimaryDut.ProvisionState.SystemImage.SystemImagePath.Path = "/a/b/c"
				_, err := commands.SourcesFromCFTTestRequest(request)
				So(err, ShouldErrLike, "system_image_path.path: must start with gs://")
			})
			Convey("Trailing slash", func() {
				request.PrimaryDut.ProvisionState.SystemImage.SystemImagePath.Path = "gs://some-bucket/builder/build-12345/"
				_, err := commands.SourcesFromCFTTestRequest(request)
				So(err, ShouldErrLike, "system_image_path.path: must not have trailing '/'")
			})
		})
		Convey("Local testing", func() {
			request.PrimaryDut.ProvisionState.SystemImage.SystemImagePath = &_go.StoragePath{
				HostType: _go.StoragePath_LOCAL,
				Path:     "/builds/build-12345",
			}
			sources, err := commands.SourcesFromCFTTestRequest(request)
			So(err, ShouldBeNil)
			So(sources, ShouldBeNil)
		})
		Convey("Lacros testing", func() {
			request.PrimaryDut.ProvisionState.Packages = []*api.ProvisionState_Package{
				{
					PortagePackage: &buildapi.Portage_Package{},
				},
			}
			expectedSources.IsDeploymentDirty = true

			sources, err := commands.SourcesFromCFTTestRequest(request)
			So(err, ShouldBeNil)
			So(sources, ShouldResembleProto, expectedSources)
		})
		Convey("Firmware testing", func() {
			request.PrimaryDut.ProvisionState.Firmware = &buildapi.FirmwareConfig{
				MainRoPayload: &buildapi.FirmwarePayload{},
			}
			expectedSources.IsDeploymentDirty = true

			sources, err := commands.SourcesFromCFTTestRequest(request)
			So(err, ShouldBeNil)
			So(sources, ShouldResembleProto, expectedSources)
		})
	})
}

func TestRdbPublishPublishCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()

	// Common setup
	ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
	ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
	cont := containers.NewCrosPublishTemplatedContainer(
		containers.CrosRdbPublishTemplatedContainerType,
		"container/image/path",
		ctr)
	exec := executors.NewCrosPublishExecutor(
		cont,
		executors.CrosRdbPublishExecutorType)
	cmd := commands.NewRdbPublishUploadCmd(exec)

	Convey("Populate TestResultForRdb with full info", t, func() {
		ctx := context.Background()
		createTime := timestamppb.New(parseTime("2022-09-07T18:53:33.983328614Z"))
		startedTime := timestamppb.New(parseTime("2022-09-07T20:53:33.983328614Z"))
		duration := &duration.Duration{Seconds: 60}
		dut := &labapi.Dut{
			Id: &labapi.Dut_Id{Value: "0wgtfqin2033834d-ecghcra"},
			DutType: &labapi.Dut_Chromeos{
				Chromeos: &labapi.Dut_ChromeOS{
					Name: "0wgtfqin2033834d-ecghcra",
					DutModel: &labapi.DutModel{
						ModelName: "nipperkin",
					},
					Chameleon: &labapi.Chameleon{
						Types: []labapi.Chameleon_Type{
							labapi.Chameleon_V2,
							labapi.Chameleon_V3,
						},
					},
					ModemInfo: &labapi.ModemInfo{
						Type: labapi.ModemType_MODEM_TYPE_FIBOCOMM_L850GL,
					},
				},
			},
		}
		wantTestResult := &artifactpb.TestResult{
			TestInvocation: &artifactpb.TestInvocation{
				PrimaryExecutionInfo: &artifactpb.ExecutionInfo{
					BuildInfo: &artifactpb.BuildInfo{
						Name:        "hatch-cq/R106-15048.0.0",
						Board:       "hatch",
						BuildTarget: "hatch",
						BuildMetadata: &artifactpb.BuildMetadata{
							Sku: &artifactpb.BuildMetadata_Sku{
								HwidSku:  "CRAASK-HULX D4B-F4E-F3F-B2K-L3I-Q6I",
								DlmSkuId: "16968",
							},
							Chipset: &artifactpb.BuildMetadata_Chipset{
								WifiChip:         "INTEL_GFP2_AX211",
								WifiRouterModels: "gale",
							},
							Cellular: &artifactpb.BuildMetadata_Cellular{
								Carrier: "CARRIER_ESIM",
							},
							Firmware: &artifactpb.BuildMetadata_Firmware{},
							Kernel:   &artifactpb.BuildMetadata_Kernel{},
							Lacros:   &artifactpb.BuildMetadata_Lacros{},
							Chameleon: &labapi.Chameleon{
								Types: []labapi.Chameleon_Type{
									labapi.Chameleon_V2,
									labapi.Chameleon_V3,
								},
							},
							ModemInfo: &labapi.ModemInfo{
								Type: labapi.ModemType_MODEM_TYPE_FIBOCOMM_L850GL,
							},
						},
					},
					DutInfo: &artifactpb.DutInfo{
						Dut: dut,
						ProvisionState: &testapi.ProvisionState{
							SystemImage: &testapi.ProvisionState_SystemImage{
								SystemImagePath: &_go.StoragePath{
									HostType: _go.StoragePath_GS,
									Path:     "gs://some-bucket/builder/build-12345",
								},
							},
						},
					},
					EnvInfo: &artifactpb.ExecutionInfo_SkylabInfo{
						SkylabInfo: &artifactpb.SkylabInfo{
							DroneInfo: &artifactpb.DroneInfo{
								Drone:       "skylab-drone-deployment-prod-6dc79d4f9-czjlj",
								DroneServer: "chromeos4-row4-rack1-drone8",
							},
							BuildbucketInfo: &artifactpb.BuildbucketInfo{
								Id:          100,
								AncestorIds: []int64{98, 99},
								Builder: &artifactpb.BuilderID{
									Project: "chromeos",
									Bucket:  "test_runner",
									Builder: "test_runner-dev",
								},
							},
							SwarmingInfo: &artifactpb.SwarmingInfo{
								TaskId:      "taskId0",
								SuiteTaskId: "parentId0",
								TaskName:    "bb-100-chromeos/test_runner/test_runner-dev",
								Pool:        "ChromeOSSkylab",
								LabelPool:   "DUT_POOL_QUOTA",
							},
						},
					},
					InventoryInfo: &artifactpb.InventoryInfo{
						UfsZone: "ZONE_SFO36_OS",
					},
				},
				DutTopology: &labapi.DutTopology{
					Id:   &labapi.DutTopology_Id{Value: "0wgtfqin2033834d-ecghcra"},
					Duts: []*labapi.Dut{dut},
				},
				SchedulingMetadata: &artifactpb.SchedulingMetadata{
					SchedulingArgs: map[string]string{
						"display_name":      "hatch-cq/R102-14632.0.0-62834-8818718496810023809/wificell-cq/tast.wificell-cq",
						"analytics_name":    "Bluetooth_Sa_Perbuild",
						"ctp-fwd-task-name": "Bluetooth_Sa_Perbuild",
						"qs_account":        "unmanaged_p2",
						"parent_task_id":    "parentId1",
						"branch-trigger":    "DEV",
					},
				},
			},
			TestRuns: []*artifactpb.TestRun{
				{
					TestCaseInfo: &artifactpb.TestCaseInfo{
						TestCaseResult: &testapi.TestCaseResult{
							TestHarness: &testapi.TestHarness{
								TestHarnessType: &testapi.TestHarness_Tast_{
									Tast: &testapi.TestHarness_Tast{},
								},
							},
							TestCaseId: &testapi.TestCase_Id{
								Value: "tast.rlz_CheckPing",
							},
							Verdict:   &testapi.TestCaseResult_Pass_{},
							StartTime: startedTime,
							Duration:  duration,
						},
						DisplayName:     "hatch-cq/R102-14632.0.0-62834-8818718496810023809/wificell-cq/tast.wificell-cq",
						Suite:           "arc-cts-vm",
						Branch:          "main",
						MainBuilderName: "main-release",
						Channel:         "DEV",
					},
					LogsInfo: []*configpb.StoragePath{
						{
							HostType: configpb.StoragePath_GS,
							Path:     "gs://some-bucket/builder/build-12345",
						},
					},
					TestHarness: &testapi.TestHarness{
						TestHarnessType: &testapi.TestHarness_Tast_{
							Tast: &testapi.TestHarness_Tast{},
						},
					},
					TimeInfo: &artifactpb.TimingInfo{
						QueuedTime:  createTime,
						StartedTime: startedTime,
						Duration:    duration,
					},
				},
			},
		}

		// Sets up the build info.
		buildPb := &bbpb.Build{
			Id:     100,
			Status: bbpb.Status_SUCCESS,
			Builder: &bbpb.BuilderID{
				Project: "chromeos",
				Bucket:  "test_runner",
				Builder: "test_runner-dev",
			},
			AncestorIds: []int64{98, 99},
			Tags: []*buildbucketpb.StringPair{
				{Key: "display_name", Value: "hatch-cq/R102-14632.0.0-62834-8818718496810023809/wificell-cq/tast.wificell-cq"},
				{Key: "analytics_name", Value: "Bluetooth_Sa_Perbuild"},
				{Key: "ctp-fwd-task-name", Value: "Bluetooth_Sa_Perbuild"},
				{Key: "qs_account", Value: "unmanaged_p2"},
				{Key: "parent_task_id", Value: "parentId1"},
				{Key: "branch-trigger", Value: "DEV"},
			},
			CreateTime: createTime,
			Infra: &bbpb.BuildInfra{Swarming: &bbpb.BuildInfra_Swarming{
				TaskId:      "taskId1",
				ParentRunId: "parentId1",
				BotDimensions: []*buildbucketpb.StringPair{
					{Key: "label-wifi_chip", Value: "INTEL_GFP2_AX211"},
					{Key: "label-wifi_router_models", Value: "gale"},
					{Key: "label-hwid_sku", Value: "CRAASK-HULX D4B-F4E-F3F-B2K-L3I-Q6I"},
					{Key: "label-dlm_sku_id", Value: "16968"},
					{Key: "label-carrier", Value: "CARRIER_ESIM"},
					{Key: "drone", Value: "skylab-drone-deployment-prod-6dc79d4f9-czjlj"},
					{Key: "drone_server", Value: "chromeos4-row4-rack1-drone8"},
					{Key: "pool", Value: "ChromeOSSkylab"},
					{Key: "label-pool", Value: "DUT_POOL_QUOTA"},
					{Key: "ufs_zone", Value: "ZONE_SFO36_OS"},
				},
			}},
		}
		buildState, ctx, err := build.Start(ctx, buildPb)
		defer func() { buildState.End(err) }()

		sk := &data.HwTestStateKeeper{
			CurrentInvocationId: "Inv-1234",
			TesthausURL:         "www.testhaus.com",
			BaseVariant: map[string]string{
				"board":        "asurada",
				"model":        "hayato",
				"build_target": "asurada",
			},
			CftTestRequest: &skylab_test_runner.CFTTestRequest{
				PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
					DutModel: &labapi.DutModel{
						BuildTarget: "hatch",
					},
					ProvisionState: &api.ProvisionState{
						SystemImage: &api.ProvisionState_SystemImage{
							SystemImagePath: &_go.StoragePath{
								HostType: _go.StoragePath_GS,
								Path:     "gs://some-bucket/builder/build-12345",
							},
						},
					},
				},
				AutotestKeyvals: map[string]string{
					"build_target":        "hatch",
					"build":               "hatch-cq/R106-15048.0.0",
					"suite":               "arc-cts-vm",
					"branch":              "main",
					"master_build_config": "main-release",
				},
			},
			Devices: map[string]*testapi.CrosTestRequest_Device{
				common.Primary: {
					Dut: dut,
				},
			},
			GcsURL:     "gs://some-bucket/builder/build-12345",
			BuildState: buildState,
			DutTopology: &labapi.DutTopology{
				Id:   &labapi.DutTopology_Id{Value: "0wgtfqin2033834d-ecghcra"},
				Duts: []*labapi.Dut{dut},
			},
			TestResponses: &testapi.CrosTestResponse{
				TestCaseResults: []*testapi.TestCaseResult{
					{
						TestHarness: &testapi.TestHarness{
							TestHarnessType: &testapi.TestHarness_Tast_{
								Tast: &testapi.TestHarness_Tast{},
							},
						},
						TestCaseId: &testapi.TestCase_Id{
							Value: "tast.rlz_CheckPing",
						},
						Verdict:   &testapi.TestCaseResult_Pass_{},
						StartTime: startedTime,
						Duration:  duration,
					},
				},
			},
		}

		// Extract deps first
		err = cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.TestResultForRdb, ShouldResembleProto, wantTestResult)
	})

	Convey("ProvisionStartCmd extract deps with TestResultForRdb", t, func() {
		ctx := context.Background()
		wantInvId := "Inv-1234"
		wantTesthausURL := "www.testhaus.com"
		wantBaseVariant := map[string]string{
			"board":        "asurada",
			"model":        "hayato",
			"build_target": "asurada",
		}
		sk := &data.HwTestStateKeeper{
			CurrentInvocationId: wantInvId,
			TesthausURL:         wantTesthausURL,
			CftTestRequest: &skylab_test_runner.CFTTestRequest{
				PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
					ProvisionState: &api.ProvisionState{
						SystemImage: &api.ProvisionState_SystemImage{
							SystemImagePath: &_go.StoragePath{
								HostType: _go.StoragePath_GS,
								Path:     "gs://some-bucket/builder/build-12345",
							},
						},
					},
				},
			},
			BaseVariant:      wantBaseVariant,
			TestResultForRdb: &artifact.TestResult{Version: 1234},
		}

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.CurrentInvocationId, ShouldEqual, wantInvId)
		So(cmd.TesthausURL, ShouldEqual, wantTesthausURL)
		So(cmd.Sources, ShouldResembleProto, &metadata.PublishRdbMetadata_Sources{
			GsPath:            "gs://some-bucket/builder/build-12345/metadata/sources.jsonpb",
			IsDeploymentDirty: false,
		})
		So(cmd.BaseVariant, ShouldEqual, wantBaseVariant)
	})

	Convey("ProvisionStartCmd extract deps without TestResultForRdb", t, func() {
		ctx := context.Background()
		wantInvId := "Inv-1234"
		wantTesthausURL := "www.testhaus.com"
		wantBaseVariant := map[string]string{
			"board":        "asurada",
			"model":        "hayato",
			"build_target": "asurada",
		}
		sk := &data.HwTestStateKeeper{
			CurrentInvocationId: wantInvId,
			TesthausURL:         wantTesthausURL,
			BaseVariant:         wantBaseVariant,
			CftTestRequest: &skylab_test_runner.CFTTestRequest{
				PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
					ProvisionState: &api.ProvisionState{},
				},
			},
		}

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(cmd.CurrentInvocationId, ShouldEqual, wantInvId)
		So(cmd.TesthausURL, ShouldEqual, wantTesthausURL)
		So(cmd.BaseVariant, ShouldEqual, wantBaseVariant)
		So(err, ShouldErrLike, "missing dependency: BuildState")
	})
}
