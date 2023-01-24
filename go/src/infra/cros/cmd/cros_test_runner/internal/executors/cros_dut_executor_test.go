// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executors

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	test_api "go.chromium.org/chromiumos/config/go/test/api"
	lab_api "go.chromium.org/chromiumos/config/go/test/lab/api"

	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/containers"
	"infra/cros/cmd/cros_test_runner/internal/mocked_services"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
)

func TestDutServiceStart(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("Dut service start with no cache server address", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosDutTemplatedContainer("container/image/path", ctr)
		exec := NewCrosDutExecutor(cont)
		err := exec.Start(ctx, nil, nil)
		So(err, ShouldNotBeNil)
	})

	Convey("Dut service start with no dut ssh address", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosDutTemplatedContainer("container/image/path", ctr)
		exec := NewCrosDutExecutor(cont)
		err := exec.Start(ctx, &lab_api.IpEndpoint{}, nil)
		So(err, ShouldNotBeNil)
	})

	Convey("Dut service start without starting ctr", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosDutTemplatedContainer("container/image/path", ctr)
		exec := NewCrosDutExecutor(cont)
		err := exec.Start(ctx, &lab_api.IpEndpoint{}, &lab_api.IpEndpoint{})
		So(err, ShouldNotBeNil)
	})

	Convey("Dut service start process container fails", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		mocked_client := mocked_services.NewMockCrosToolRunnerContainerServiceClient(ctrl)
		ctr.CtrClient = mocked_client
		getMockedStartTemplatedContainer(mocked_client).Return(nil, fmt.Errorf("some error"))
		cont := containers.NewCrosDutTemplatedContainer("container/image/path", ctr)
		exec := NewCrosDutExecutor(cont)
		err := exec.Start(ctx, &lab_api.IpEndpoint{}, &lab_api.IpEndpoint{})
		So(err, ShouldNotBeNil)
	})

	Convey("Dut service start process address fails", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		mocked_client := mocked_services.NewMockCrosToolRunnerContainerServiceClient(ctrl)
		ctr.CtrClient = mocked_client
		getMockedStartTemplatedContainer(mocked_client).Return(&test_api.StartContainerResponse{}, nil)
		getMockedGetContainer(mocked_client).Return(&test_api.GetContainerResponse{Container: &test_api.Container{PortBindings: []*test_api.Container_PortBinding{{}}}}, nil)
		cont := containers.NewCrosDutTemplatedContainer("container/image/path", ctr)
		exec := NewCrosDutExecutor(cont)
		err := exec.Start(ctx, &lab_api.IpEndpoint{}, &lab_api.IpEndpoint{})
		So(err, ShouldNotBeNil)
	})
}

func TestDutServiceExecuteCommand(t *testing.T) {
	t.Parallel()

	Convey("Dut service unsupported cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosDutTemplatedContainer("container/image/path", ctr)
		exec := NewCrosDutExecutor(cont)
		err := exec.ExecuteCommand(ctx, NewUnsupportedCmd())
		So(err, ShouldNotBeNil)
	})

	Convey("Dut service start cmd process container execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosDutTemplatedContainer("container/image/path", ctr)
		exec := NewCrosDutExecutor(cont)
		err := exec.ExecuteCommand(ctx, commands.NewDutServiceStartCmd(exec))
		So(err, ShouldNotBeNil)
	})
}

func getMockedStartTemplatedContainer(mctrclient *mocked_services.MockCrosToolRunnerContainerServiceClient) *gomock.Call {
	return mctrclient.EXPECT().StartTemplatedContainer(gomock.Any(), gomock.AssignableToTypeOf(&test_api.StartTemplatedContainerRequest{}), gomock.Any())
}

func getMockedGetContainer(mctrclient *mocked_services.MockCrosToolRunnerContainerServiceClient) *gomock.Call {
	return mctrclient.EXPECT().GetContainer(gomock.Any(), gomock.AssignableToTypeOf(&test_api.GetContainerRequest{}), gomock.Any())
}
