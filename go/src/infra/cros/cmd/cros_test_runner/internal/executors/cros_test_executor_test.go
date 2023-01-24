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

	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/containers"
	"infra/cros/cmd/cros_test_runner/internal/mocked_services"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"

	"go.chromium.org/chromiumos/config/go/longrunning"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestTestServiceStart(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("Test service start without starting ctr", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestTemplatedContainer("container/image/path", ctr)
		exec := NewCrosTestExecutor(cont)
		err := exec.Start(ctx)
		So(err, ShouldNotBeNil)
	})

	Convey("Test service start process container fails", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		mocked_client := mocked_services.NewMockCrosToolRunnerContainerServiceClient(ctrl)
		ctr.CtrClient = mocked_client
		getMockedStartTemplatedContainer(mocked_client).Return(nil, fmt.Errorf("some error"))
		cont := containers.NewCrosTestTemplatedContainer("container/image/path", ctr)
		exec := NewCrosTestExecutor(cont)
		err := exec.Start(ctx)
		So(err, ShouldNotBeNil)
	})
}

func TestTestServiceExecuteTests(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("Test service test execution with nil request", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestTemplatedContainer("container/image/path", ctr)
		exec := NewCrosTestExecutor(cont)
		resp, err := exec.ExecuteTests(ctx, nil)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Test service test execution with no established client", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestTemplatedContainer("container/image/path", ctr)
		exec := NewCrosTestExecutor(cont)
		resp, err := exec.ExecuteTests(ctx, &test_api.CrosTestRequest{})
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Test service test execution with run tests error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestTemplatedContainer("container/image/path", ctr)
		exec := NewCrosTestExecutor(cont)
		mocked_client := mocked_services.NewMockExecutionServiceClient(ctrl)
		exec.CrosTestServiceClient = mocked_client
		getMockedExecuteTests(mocked_client).Return(nil, fmt.Errorf("some_error"))
		resp, err := exec.ExecuteTests(ctx, &test_api.CrosTestRequest{})
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Test service test execution with lro process failure", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestTemplatedContainer("container/image/path", ctr)
		exec := NewCrosTestExecutor(cont)
		mocked_client := mocked_services.NewMockExecutionServiceClient(ctrl)
		exec.CrosTestServiceClient = mocked_client
		getMockedExecuteTests(mocked_client).Return(nil, fmt.Errorf("some_error"))
		resp, err := exec.ExecuteTests(ctx, &test_api.CrosTestRequest{})
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Test service test execution success", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestTemplatedContainer("container/image/path", ctr)
		exec := NewCrosTestExecutor(cont)
		mocked_client := mocked_services.NewMockExecutionServiceClient(ctrl)
		exec.CrosTestServiceClient = mocked_client
		wantResp := &test_api.CrosTestResponse{}
		wantRespAnypb, _ := anypb.New(wantResp)
		getMockedExecuteTests(mocked_client).Return(&longrunning.Operation{Done: true, Result: &longrunning.Operation_Response{Response: wantRespAnypb}}, nil)
		resp, err := exec.ExecuteTests(ctx, &test_api.CrosTestRequest{})
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		So(proto.Equal(resp, wantResp), ShouldBeTrue)
	})
}

func TestTestServiceExecuteCommand(t *testing.T) {
	t.Parallel()

	Convey("Test service unsupported cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestTemplatedContainer("container/image/path", ctr)
		exec := NewCrosTestExecutor(cont)
		err := exec.ExecuteCommand(ctx, NewUnsupportedCmd())
		So(err, ShouldNotBeNil)
	})

	Convey("Test service start cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestTemplatedContainer("container/image/path", ctr)
		exec := NewCrosTestExecutor(cont)
		err := exec.ExecuteCommand(ctx, commands.NewTestServiceStartCmd(exec))
		So(err, ShouldNotBeNil)
	})

	Convey("Test service test execution cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestTemplatedContainer("container/image/path", ctr)
		exec := NewCrosTestExecutor(cont)
		err := exec.ExecuteCommand(ctx, commands.NewTestsExecutionCmd(exec))
		So(err, ShouldNotBeNil)
	})
}

func getMockedExecuteTests(mockClient *mocked_services.MockExecutionServiceClient) *gomock.Call {
	return mockClient.EXPECT().RunTests(gomock.Any(), gomock.AssignableToTypeOf(&test_api.CrosTestRequest{}), gomock.Any())
}
