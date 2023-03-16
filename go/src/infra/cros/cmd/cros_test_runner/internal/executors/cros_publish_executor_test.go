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
	testapi "go.chromium.org/chromiumos/config/go/test/api"

	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/containers"
	"infra/cros/cmd/cros_test_runner/internal/mocked_services"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"

	"go.chromium.org/chromiumos/config/go/longrunning"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestPublishServiceStart(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("Publish service with invalid type", t, func() {
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, NoExecutorType)
		So(exec, ShouldBeNil)
	})

	Convey("Publish service start with no template", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		resp, err := exec.Start(ctx, nil)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Publish service start process container fails", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		mocked_client := mocked_services.NewMockCrosToolRunnerContainerServiceClient(ctrl)
		ctr.CtrClient = mocked_client
		getMockedStartTemplatedContainer(mocked_client).Return(nil, fmt.Errorf("some error"))
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		gcsPublishTemplate := &testapi.CrosPublishTemplate{
			PublishType:   testapi.CrosPublishTemplate_PUBLISH_GCS,
			PublishSrcDir: "publish/src/dir"}
		contTemplate := &testapi.Template{
			Container: &testapi.Template_CrosPublish{
				CrosPublish: gcsPublishTemplate,
			},
		}
		resp, err := exec.Start(ctx, contTemplate)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})
}

func TestPublishServicePublishResults(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("Publish service publish results with no client", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		resp, err := exec.Publish(ctx, nil, nil)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Publish service publish results with empty request", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		mocked_client := mocked_services.NewMockGenericPublishServiceClient(ctrl)
		resp, err := exec.Publish(ctx, nil, mocked_client)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Publish service publish results with publish results error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		mocked_client := mocked_services.NewMockGenericPublishServiceClient(ctrl)
		getMockedPublishResults(mocked_client).Return(nil, fmt.Errorf("some_error"))
		resp, err := exec.Publish(ctx, &testapi.PublishRequest{}, mocked_client)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Publish service publish results with lro process failure", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		mocked_client := mocked_services.NewMockGenericPublishServiceClient(ctrl)
		getMockedPublishResults(mocked_client).Return(nil, nil)
		resp, err := exec.Publish(ctx, &testapi.PublishRequest{}, mocked_client)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("Publish service publish results success", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		mocked_client := mocked_services.NewMockGenericPublishServiceClient(ctrl)
		wantResp := &testapi.PublishResponse{}
		wantRespAnypb, _ := anypb.New(wantResp)
		getMockedPublishResults(mocked_client).Return(&longrunning.Operation{
			Done:   true,
			Result: &longrunning.Operation_Response{Response: wantRespAnypb}},
			nil)
		resp, err := exec.Publish(ctx, &testapi.PublishRequest{}, mocked_client)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		So(proto.Equal(resp, wantResp), ShouldBeTrue)
	})
}

func TestInvokePublishWithAsyncLogging(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("Publish service invoke publish with async logging error with empty request", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		err := exec.InvokePublishWithAsyncLogging(ctx, "testing-publish", nil, nil, nil)
		So(err, ShouldNotBeNil)
	})

	Convey("Publish service invoke publish with async logging error with empty client", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		err := exec.InvokePublishWithAsyncLogging(
			ctx,
			"testing-publish",
			&testapi.PublishRequest{},
			nil,
			nil)
		So(err, ShouldNotBeNil)
	})

	Convey("Publish service invoke publish with async logging error with empty container", t, func() {
		ctx := context.Background()
		exec := NewCrosPublishExecutor(nil, CrosGcsPublishExecutorType)
		err := exec.InvokePublishWithAsyncLogging(
			ctx,
			"testing-publish",
			&testapi.PublishRequest{},
			mocked_services.NewMockGenericPublishServiceClient(ctrl),
			nil)
		So(err, ShouldNotBeNil)
	})

	Convey("Publish service invoke publish with async logging error with no logs location", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		err := exec.InvokePublishWithAsyncLogging(
			ctx,
			"testing-publish",
			&testapi.PublishRequest{},
			mocked_services.NewMockGenericPublishServiceClient(ctrl),
			nil)
		So(err, ShouldNotBeNil)
	})

}
func TestPublishServiceExecuteCommand(t *testing.T) {
	t.Parallel()

	Convey("Publish service unsupported cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		err := exec.ExecuteCommand(ctx, NewUnsupportedCmd())
		So(err, ShouldNotBeNil)
	})

	Convey("Publish service gcs start cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		err := exec.ExecuteCommand(ctx, commands.NewGcsPublishServiceStartCmd(exec))
		So(err, ShouldNotBeNil)
	})

	Convey("Publish service gcs publish results cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosGcsPublishExecutorType)
		err := exec.ExecuteCommand(ctx, commands.NewGcsPublishUploadCmd(exec))
		So(err, ShouldNotBeNil)
	})

	Convey("Publish service tko start cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosTkoPublishExecutorType)
		err := exec.ExecuteCommand(ctx, commands.NewTkoPublishServiceStartCmd(exec))
		So(err, ShouldNotBeNil)
	})

	Convey("Publish service tko publish results cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosTkoPublishExecutorType)
		err := exec.ExecuteCommand(ctx, commands.NewTkoPublishUploadCmd(exec))
		So(err, ShouldNotBeNil)
	})

	Convey("Publish service cpcon start cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosCpconPublishExecutorType)
		err := exec.ExecuteCommand(ctx, commands.NewCpconPublishServiceStartCmd(exec))
		So(err, ShouldNotBeNil)
	})

	Convey("Publish service cpcon publish results cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosCpconPublishExecutorType)
		err := exec.ExecuteCommand(ctx, commands.NewCpconPublishUploadCmd(exec))
		So(err, ShouldNotBeNil)
	})

	Convey("Publish service rdb start cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosRdbPublishExecutorType)
		err := exec.ExecuteCommand(ctx, commands.NewRdbPublishServiceStartCmd(exec))
		So(err, ShouldNotBeNil)
	})

	Convey("Publish service rdb publish results cmd execution error", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosPublishTemplatedContainer(
			containers.CrosGcsPublishTemplatedContainerType,
			"container/image/path",
			ctr)
		exec := NewCrosPublishExecutor(cont, CrosRdbPublishExecutorType)
		err := exec.ExecuteCommand(ctx, commands.NewRdbPublishUploadCmd(exec))
		So(err, ShouldNotBeNil)
	})
}

func getMockedPublishResults(mockClient *mocked_services.MockGenericPublishServiceClient) *gomock.Call {
	return mockClient.EXPECT().Publish(gomock.Any(),
		gomock.AssignableToTypeOf(&testapi.PublishRequest{}),
		gomock.Any())
}
