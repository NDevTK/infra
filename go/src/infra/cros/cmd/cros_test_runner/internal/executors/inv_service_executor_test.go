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
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"

	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/mocked_services"
)

type UnsupportedCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor
}

func NewUnsupportedCmd() interfaces.CommandInterface {
	absCmd := interfaces.NewAbstractCmd(commands.UnSupportedCmdType)
	absSingleCmdByNoExec := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: absCmd}
	return &UnsupportedCmd{AbstractSingleCmdByNoExecutor: absSingleCmdByNoExec}
}

func TestInvServiceStart(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("Inventory service start with an already established connection", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		exec.InventoryServiceClient = mocked_services.NewMockInventoryServiceClient(ctrl)
		err := exec.Start(ctx, exec.InventoryServiceAddress)
		So(err, ShouldBeNil)
	})

	Convey("Inventory service start with empty server address", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		err := exec.Start(ctx, "")
		So(err, ShouldNotBeNil)
	})
}

func TestInvServiceStop(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("Inventory service stop with no established server", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		err := exec.Stop(ctx)
		So(err, ShouldBeNil)
	})

	Convey("Inventory service stop with no grpc connection", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		exec.InventoryServiceClient = mocked_services.NewMockInventoryServiceClient(ctrl)
		err := exec.Stop(ctx)
		So(err, ShouldNotBeNil)
	})
}

func TestInvServiceGetDutTopology(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	hostName := "DUT-1234"

	Convey("GetDutTopology with empty host name", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		dutTopology, err := exec.GetDUTTopology(ctx, "")
		So(dutTopology, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("GetDutTopology with no service client", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		dutTopology, err := exec.GetDUTTopology(ctx, hostName)
		So(dutTopology, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("GetDutTopology_Success", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		mockInvServiceClient := mocked_services.NewMockInventoryServiceClient(ctrl)
		mockInvServiceGDTClient := mocked_services.NewMockInventoryService_GetDutTopologyClient(ctrl)
		exec.InventoryServiceClient = mockInvServiceClient

		getMockedGetDutTopologyRcvMsgSuccess(mockInvServiceGDTClient, hostName)
		getMockedGetDutTopology(mockInvServiceClient, hostName).Return(mockInvServiceGDTClient, nil)

		dutTopology, err := exec.GetDUTTopology(ctx, hostName)
		So(dutTopology, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("GetDutTopology_grpc_failure", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		mockInvServiceClient := mocked_services.NewMockInventoryServiceClient(ctrl)
		exec.InventoryServiceClient = mockInvServiceClient

		getMockedGetDutTopology(mockInvServiceClient, hostName).Return(nil, fmt.Errorf("some error"))

		dutTopology, err := exec.GetDUTTopology(ctx, hostName)
		So(dutTopology, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("GetDutTopology_grpc_response_failure", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		mockInvServiceClient := mocked_services.NewMockInventoryServiceClient(ctrl)
		mockInvServiceGDTClient := mocked_services.NewMockInventoryService_GetDutTopologyClient(ctrl)
		exec.InventoryServiceClient = mockInvServiceClient

		getMockedGetDutTopologyRcvMsgFailure(mockInvServiceGDTClient)
		getMockedGetDutTopology(mockInvServiceClient, hostName).Return(mockInvServiceGDTClient, nil)

		dutTopology, err := exec.GetDUTTopology(ctx, hostName)
		So(dutTopology, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestInvServiceExecuteCommand(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	Convey("Inventory service unsupported cmd execution error", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		err := exec.ExecuteCommand(ctx, NewUnsupportedCmd())
		So(err, ShouldNotBeNil)
	})

	Convey("Inventory service start cmd execution error", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		exec.InventoryServiceAddress = ""
		startCmd := commands.NewInvServiceStartCmd(exec)
		err := exec.ExecuteCommand(ctx, startCmd)
		So(err, ShouldNotBeNil)
	})

	Convey("Inventory service load dut topology cmd execution error", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		loadDuttopologyCmd := commands.NewLoadDutTopologyCmd(exec)
		err := exec.ExecuteCommand(ctx, loadDuttopologyCmd)
		So(err, ShouldNotBeNil)
	})

	Convey("Inventory service stop cmd execution error", t, func() {
		ctx := context.Background()
		exec := NewInvServiceExecutor("")
		exec.InventoryServiceClient = mocked_services.NewMockInventoryServiceClient(ctrl)
		stopCmd := commands.NewInvServiceStopCmd(exec)
		err := exec.ExecuteCommand(ctx, stopCmd)
		So(err, ShouldNotBeNil)
	})
}

func getMockedGetDutTopology(
	mis *mocked_services.MockInventoryServiceClient,
	hostName string) *gomock.Call {
	return mis.EXPECT().GetDutTopology(
		gomock.Any(),
		gomock.Eq(&labapi.GetDutTopologyRequest{
			Id: &labapi.DutTopology_Id{
				Value: hostName,
			},
		},
		),
	)
}

func getMockedGetDutTopologyRcvMsgSuccess(
	misgtc *mocked_services.MockInventoryService_GetDutTopologyClient,
	hostName string) *gomock.Call {

	return misgtc.EXPECT().RecvMsg(gomock.Eq(&labapi.GetDutTopologyResponse{})).DoAndReturn(
		func(resp *labapi.GetDutTopologyResponse) error {
			resp.Result = &labapi.GetDutTopologyResponse_Success_{
				Success: &labapi.GetDutTopologyResponse_Success{
					DutTopology: &labapi.DutTopology{
						Id: &labapi.DutTopology_Id{
							Value: hostName,
						},
					},
				},
			}
			return nil
		})
}

func getMockedGetDutTopologyRcvMsgFailure(
	misgtc *mocked_services.MockInventoryService_GetDutTopologyClient) *gomock.Call {
	return misgtc.EXPECT().RecvMsg(gomock.Eq(&labapi.GetDutTopologyResponse{})).DoAndReturn(
		func(resp *labapi.GetDutTopologyResponse) error {
			return fmt.Errorf("some error")
		})
}
