// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"infra/cros/cmd/common_lib/containers"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/data"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	_go "go.chromium.org/chromiumos/config/go"
	buildapi "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/config/go/test/api/metadata"
	"go.chromium.org/chromiumos/config/go/test/artifact"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	. "go.chromium.org/luci/common/testing/assertions"
)

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

	Convey("ProvisionStartCmd extract deps with TestResultForRdb", t, func() {
		ctx := context.Background()
		wantInvId := "Inv-1234"
		wantTesthausUrl := "www.testhaus.com"
		sk := &data.HwTestStateKeeper{
			CurrentInvocationId: wantInvId,
			TesthausUrl:         wantTesthausUrl,
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
			TestResultForRdb: &artifact.TestResult{Version: 1234},
		}
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

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.CurrentInvocationId, ShouldEqual, wantInvId)
		So(cmd.TesthausUrl, ShouldEqual, wantTesthausUrl)
		So(cmd.Sources, ShouldResembleProto, &metadata.PublishRdbMetadata_Sources{
			GsPath:            "gs://some-bucket/builder/build-12345/metadata/sources.jsonpb",
			IsDeploymentDirty: false,
		})
	})

	Convey("ProvisionStartCmd extract deps without TestResultForRdb", t, func() {
		ctx := context.Background()
		wantInvId := "Inv-1234"
		wantTesthausUrl := "www.testhaus.com"
		sk := &data.HwTestStateKeeper{
			CurrentInvocationId: wantInvId,
			TesthausUrl:         wantTesthausUrl,
			CftTestRequest: &skylab_test_runner.CFTTestRequest{
				PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
					ProvisionState: &api.ProvisionState{},
				},
			},
		}

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

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldErrLike, "missing dependency: BuildState")
	})
}
