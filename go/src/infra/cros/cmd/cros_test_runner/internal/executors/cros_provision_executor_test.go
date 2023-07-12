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
	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"

	"infra/cros/cmd/common_lib/containers"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/mocked_services"

	"go.chromium.org/chromiumos/config/go/longrunning"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestProvisionServiceStart(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("Provision service start with nil provision request", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosProvisionExecutor(cont)
		err := exec.Start(ctx, nil)
		So(err, ShouldNotBeNil)
	})

	Convey("Provision service start without starting ctr", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosProvisionExecutor(cont)
		err := exec.Start(ctx, &api.CrosProvisionRequest{})
		So(err, ShouldNotBeNil)
	})

	Convey("Provision service start process container fails", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		mocked_client := mocked_services.NewMockCrosToolRunnerContainerServiceClient(ctrl)
		ctr.CtrClient = mocked_client
		getMockedStartTemplatedContainer(mocked_client).Return(nil, fmt.Errorf("some error"))
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosProvisionExecutor(cont)
		err := exec.Start(ctx, &api.CrosProvisionRequest{})
		So(err, ShouldNotBeNil)
	})
}

func TestProvisionServiceInstall(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("Provision service install with nil install request", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosProvisionExecutor(cont)
		resp, err := exec.Install(ctx, nil)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Provision service install with no established client", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosProvisionExecutor(cont)
		resp, err := exec.Install(ctx, &testapi.InstallRequest{})
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Provision service install with install error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosProvisionExecutor(cont)
		mocked_client := mocked_services.NewMockGenericProvisionServiceClient(ctrl)
		exec.CrosProvisionServiceClient = mocked_client
		getMockedProvisionInstall(mocked_client).Return(nil, fmt.Errorf("some_error"))
		resp, err := exec.Install(ctx, &testapi.InstallRequest{})
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Provision service install with lro process failure", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosProvisionExecutor(cont)
		mocked_client := mocked_services.NewMockGenericProvisionServiceClient(ctrl)
		exec.CrosProvisionServiceClient = mocked_client
		getMockedProvisionInstall(mocked_client).Return(nil, nil)
		resp, err := exec.Install(ctx, &testapi.InstallRequest{})
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Provision service install success", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosProvisionExecutor(cont)
		mocked_client := mocked_services.NewMockGenericProvisionServiceClient(ctrl)
		exec.CrosProvisionServiceClient = mocked_client
		wantResp := &testapi.InstallResponse{}
		wantRespAnypb, _ := anypb.New(wantResp)
		getMockedProvisionInstall(mocked_client).Return(&longrunning.Operation{
			Done: true,
			Result: &longrunning.Operation_Response{
				Response: wantRespAnypb,
			},
		},
			nil)
		resp, err := exec.Install(ctx, &testapi.InstallRequest{})
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		//So(resp, ShouldEqual, wantResp)
		So(proto.Equal(resp, wantResp), ShouldBeTrue)
	})
}

func TestProvisionServiceExecuteCommand(t *testing.T) {
	t.Parallel()

	Convey("Provision service unsupported cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosProvisionExecutor(cont)
		err := exec.ExecuteCommand(ctx, NewUnsupportedCmd())
		So(err, ShouldNotBeNil)
	})

	Convey("Provision service start cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosProvisionExecutor(cont)
		err := exec.ExecuteCommand(ctx, commands.NewProvisionServiceStartCmd(exec))
		So(err, ShouldNotBeNil)
	})

	Convey("Provision service install cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosProvisionTemplatedContainer("container/image/path", ctr)
		exec := NewCrosProvisionExecutor(cont)
		err := exec.ExecuteCommand(ctx, commands.NewProvisionInstallCmd(exec))
		So(err, ShouldNotBeNil)
	})
}

func getMockedProvisionInstall(mockClient *mocked_services.MockGenericProvisionServiceClient) *gomock.Call {
	return mockClient.EXPECT().Install(gomock.Any(),
		gomock.AssignableToTypeOf(&testapi.InstallRequest{}),
		gomock.Any())
}
