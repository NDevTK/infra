// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executors

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/longrunning"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"infra/cros/cmd/common_lib/containers"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/mocked_services"
)

func TestVMProvisionServiceStart(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("VM Provision service start fails without starting ctr", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosVMProvisionExecutor(cont)
		err := exec.Start(ctx)
		So(err, ShouldNotBeNil)
	})

	Convey("VM Provision service start fails on failing StartTemplatedContainer", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		mocked_client := mocked_services.NewMockCrosToolRunnerContainerServiceClient(ctrl)
		ctr.CtrClient = mocked_client
		getMockedStartTemplatedContainer(mocked_client).Return(nil, fmt.Errorf("some error"))
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosVMProvisionExecutor(cont)
		err := exec.Start(ctx)
		So(err, ShouldNotBeNil)
	})
}

func TestVMProvisionServiceLeaseDutVM(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("VM Provision service LeaseDutVM fails with nil install request", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosVMProvisionExecutor(cont)
		resp, err := exec.LeaseDutVM(ctx, nil)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("VM Provision service LeaseDutVM fails with no established client", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosVMProvisionExecutor(cont)
		resp, err := exec.LeaseDutVM(ctx, &testapi.InstallRequest{})
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("VM Provision service LeaseDutVM fails with install error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosVMProvisionExecutor(cont)
		mocked_client := mocked_services.NewMockGenericProvisionServiceClient(ctrl)
		exec.CrosVMProvisionServiceClient = mocked_client
		getMockedVMProvisionInstall(mocked_client).Return(nil, fmt.Errorf("some_error"))
		resp, err := exec.LeaseDutVM(ctx, &testapi.InstallRequest{})
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("VM Provision service LeaseDutVM fails with empty lro response", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosVMProvisionExecutor(cont)
		mocked_client := mocked_services.NewMockGenericProvisionServiceClient(ctrl)
		exec.CrosVMProvisionServiceClient = mocked_client
		getMockedVMProvisionInstall(mocked_client).Return(nil, nil)
		resp, err := exec.LeaseDutVM(ctx, &testapi.InstallRequest{})
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("VM Provision service LeaseDutVM success", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosVMProvisionExecutor(cont)
		mocked_client := mocked_services.NewMockGenericProvisionServiceClient(ctrl)
		exec.CrosVMProvisionServiceClient = mocked_client
		wantResp := &testapi.InstallResponse{}
		wantRespAnypb, _ := anypb.New(wantResp)
		getMockedVMProvisionInstall(mocked_client).Return(&longrunning.Operation{
			Done: true,
			Result: &longrunning.Operation_Response{
				Response: wantRespAnypb,
			},
		},
			nil)
		resp, err := exec.LeaseDutVM(ctx, &testapi.InstallRequest{})
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		//So(resp, ShouldEqual, wantResp)
		So(proto.Equal(resp, wantResp), ShouldBeTrue)
	})
}

func TestVMProvisionServiceExecuteCommand(t *testing.T) {
	t.Parallel()

	Convey("VM Provision service unsupported cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosVMProvisionExecutor(cont)
		err := exec.ExecuteCommand(ctx, NewUnsupportedCmd())
		So(err, ShouldNotBeNil)
	})

	Convey("VM Provision service start cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosVMProvisionExecutor(cont)
		err := exec.ExecuteCommand(ctx, commands.NewVMProvisionServiceStartCmd(exec))
		So(err, ShouldNotBeNil)
	})

	Convey("VM Provision service install cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosVMProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosVMProvisionExecutor(cont)
		err := exec.ExecuteCommand(ctx, commands.NewVMProvisionLeaseCmd(exec))
		So(err, ShouldNotBeNil)
	})
}

func TestGetDiskSizeByBoard(t *testing.T) {
	Convey("Given an image name", t, func() {
		Convey("When the image contains 'reven-vmtest'", func() {
			image := "reven-vmtest"
			expectedSize := int64(20)

			size := getDiskSizeByBoard(image)

			Convey("Then it should return the expected size", func() {
				So(size, ShouldEqual, expectedSize)
			})
		})

		Convey("When the image does not contain 'reven-vmtest'", func() {
			image := "other-image"
			expectedSize := int64(13)

			size := getDiskSizeByBoard(image)

			Convey("Then it should return the expected size", func() {
				So(size, ShouldEqual, expectedSize)
			})
		})
	})
}

func getMockedVMProvisionInstall(mockClient *mocked_services.MockGenericProvisionServiceClient) *gomock.Call {
	return mockClient.EXPECT().Install(gomock.Any(),
		gomock.AssignableToTypeOf(&testapi.InstallRequest{}),
		gomock.Any())
}
