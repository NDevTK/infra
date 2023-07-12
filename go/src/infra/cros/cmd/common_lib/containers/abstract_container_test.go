// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package containers

import (
	"context"
	"fmt"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"
)

func TestGetContainerType(t *testing.T) {
	t.Parallel()

	Convey("GetContainerType", t, func() {
		wantContType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(wantContType, "test-container", "container-image", ctr)
		gotContType := absContainer.GetContainerType()
		So(gotContType, ShouldNotBeNil)
		So(gotContType, ShouldEqual, wantContType)
	})
}

func TestGetLogsLocation(t *testing.T) {
	t.Parallel()

	Convey("GetLogsLocation_error", t, func() {
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		logsLoc, err := absContainer.GetLogsLocation()
		So(err, ShouldNotBeNil)
		So(logsLoc, ShouldEqual, "")
	})

	Convey("GetLogsLocation_success", t, func() {
		contType := CrosProvisionTemplatedContainerType
		wantLogsLoc := "temp/dir/loc"
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		absContainer.TempDirLoc = wantLogsLoc
		gotLogsLoc, err := absContainer.GetLogsLocation()
		So(err, ShouldBeNil)
		So(gotLogsLoc, ShouldEqual, wantLogsLoc)
	})
}

func TestInitializeBase(t *testing.T) {
	t.Parallel()

	Convey("InitializeBase_wrong_state", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		absContainer.state = ContainerStateStarted
		err := absContainer.InitializeBase(ctx)
		So(err, ShouldNotBeNil)
	})

	Convey("InitializeBase_missing_namePrefix", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "", "container-image", ctr)
		err := absContainer.InitializeBase(ctx)
		So(err, ShouldNotBeNil)
	})

	Convey("InitializeBase_missing_containerImage", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "", ctr)
		err := absContainer.InitializeBase(ctx)
		So(err, ShouldNotBeNil)
	})

	Convey("InitializeBase_success", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		err := absContainer.InitializeBase(ctx)
		So(err, ShouldBeNil)
		So(absContainer.TempDirLoc, ShouldNotEqual, "")
	})
}

func TestGetContainer(t *testing.T) {
	t.Parallel()

	Convey("GetContainer_wrong_state", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		resp, err := absContainer.GetContainer(ctx)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("GetContainer_empty_container_name", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		absContainer.state = ContainerStateStarted
		resp, err := absContainer.GetContainer(ctx)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

	Convey("GetContainer_without_starting_client", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		absContainer.state = ContainerStateStarted
		absContainer.Name = "container-1234"
		resp, err := absContainer.GetContainer(ctx)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})
}

func TestStopContainer(t *testing.T) {
	t.Parallel()

	Convey("StopContainer_wrong_state", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		err := absContainer.StopContainer(ctx)
		So(err, ShouldNotBeNil)
	})

	Convey("StopContainer_empty_container_name", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		absContainer.state = ContainerStateStarted
		err := absContainer.StopContainer(ctx)
		So(err, ShouldNotBeNil)
	})
	// TODO (azrahman): fix test for windows (sudo not in path)
	//
	//	Convey("GetContainer_with_no_client", t, func() {
	//		ctx := context.Background()
	//		contType := CrosProvisionTemplatedContainerType
	//		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
	//		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
	//		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
	//		absContainer.state = ContainerStateStarted
	//		absContainer.Name = "container-1234"
	//		err := absContainer.StopContainer(ctx)
	//		So(err, ShouldBeNil)
	//	})
}

func TestProcessContainer(t *testing.T) {
	t.Parallel()

	Convey("ProcessContainer_initialization_error", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		absContainer.ConcreteContainer = &MockContainer{InitializeErr: fmt.Errorf("some err")}
		address, err := absContainer.ProcessContainer(ctx, nil)
		So(err, ShouldNotBeNil)
		So(address, ShouldEqual, "")
	})

	Convey("ProcessContainer_startcontainer_error", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		absContainer.ConcreteContainer = &MockContainer{InitializeErr: nil, StartContainerResp: nil, StartContainerErr: fmt.Errorf("some err")}
		address, err := absContainer.ProcessContainer(ctx, nil)
		So(err, ShouldNotBeNil)
		So(address, ShouldEqual, "")
	})

	Convey("ProcessContainer_getcontainer_error", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		absContainer.ConcreteContainer = &MockContainer{InitializeErr: nil, StartContainerResp: &api.StartContainerResponse{}, StartContainerErr: nil, GetContainerResp: nil, GetContainerErr: fmt.Errorf("some err")}
		address, err := absContainer.ProcessContainer(ctx, nil)
		So(err, ShouldNotBeNil)
		So(address, ShouldEqual, "")
	})

	Convey("ProcessContainer_server_address_retrieval_error", t, func() {
		ctx := context.Background()
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		absContainer.ConcreteContainer = &MockContainer{InitializeErr: nil, StartContainerResp: &api.StartContainerResponse{}, StartContainerErr: nil, GetContainerResp: nil, GetContainerErr: nil}
		address, err := absContainer.ProcessContainer(ctx, nil)
		So(err, ShouldNotBeNil)
		So(address, ShouldEqual, "")
	})

	Convey("ProcessContainer_success", t, func() {
		ctx := context.Background()
		hostIp := "localhost"
		hostPort := int32(1234)
		contType := CrosProvisionTemplatedContainerType
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		absContainer := NewAbstractContainer(contType, "test-container", "container-image", ctr)
		getResp := &api.GetContainerResponse{Container: &api.Container{
			PortBindings: []*api.Container_PortBinding{
				{
					HostIp:   hostIp,
					HostPort: hostPort,
				},
			},
		},
		}
		absContainer.ConcreteContainer = &MockContainer{
			InitializeErr:      nil,
			StartContainerResp: &api.StartContainerResponse{},
			StartContainerErr:  nil,
			GetContainerResp:   getResp,
			GetContainerErr:    nil}
		address, err := absContainer.ProcessContainer(ctx, nil)
		So(err, ShouldBeNil)
		So(address, ShouldEqual, fmt.Sprintf("%s:%v", hostIp, hostPort))
	})
}

type MockContainer struct {
	ContainerType                 interfaces.ContainerType
	InitializeErr                 error
	ProcessContainerServerAddress string
	ProcessContainerErr           error
	StopContainerErr              error
	LogsLoc                       string
	LogsLocErr                    error
	StartContainerResp            *api.StartContainerResponse
	StartContainerErr             error
	GetContainerResp              *api.GetContainerResponse
	GetContainerErr               error
}

func (cont *MockContainer) GetContainerType() interfaces.ContainerType {
	return cont.ContainerType
}

func (cont *MockContainer) Initialize(context.Context, *api.Template) error {
	return cont.InitializeErr
}

func (cont *MockContainer) ProcessContainer(context.Context, *api.Template) (string, error) {
	return cont.ProcessContainerServerAddress, cont.ProcessContainerErr
}

func (cont *MockContainer) StopContainer(context.Context) error {
	return cont.StopContainerErr
}

func (cont *MockContainer) GetLogsLocation() (string, error) {
	return cont.LogsLoc, cont.LogsLocErr
}

func (cont *MockContainer) StartContainer(context.Context) (*api.StartContainerResponse, error) {
	return cont.StartContainerResp, cont.StartContainerErr
}

func (cont *MockContainer) GetContainer(context.Context) (*api.GetContainerResponse, error) {
	return cont.GetContainerResp, cont.GetContainerErr
}
