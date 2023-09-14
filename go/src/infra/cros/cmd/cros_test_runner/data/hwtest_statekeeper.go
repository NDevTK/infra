// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package data

import (
	"container/list"

	"go.chromium.org/chromiumos/config/go/build/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	artifactpb "go.chromium.org/chromiumos/config/go/test/artifact"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/protobuf/types/known/anypb"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/dutstate"
	vmlabapi "infra/libs/vmlab/api"
)

// HwTestStateKeeper represents all the data hw test execution flow requires.
type HwTestStateKeeper struct {
	interfaces.StateKeeper

	// Build related
	BuildState *build.State

	// Set from input
	CftTestRequest        *skylab_test_runner.CFTTestRequest
	CrosTestRunnerRequest *skylab_test_runner.CrosTestRunnerRequest

	// Request Queues
	ContainerQueue *list.List
	ProvisionQueue *list.List
	PreTestQueue   *list.List
	TestQueue      *list.List
	PostTestQueue  *list.List
	PublishQueue   *list.List

	// Dictionaries
	Injectables        *common.InjectableStorage
	ContainerInstances map[string]interfaces.ContainerInterface
	ContainerImages    map[string]*api.ContainerImageInfo

	// Dut related
	HostName                 string
	DeviceIdentifiers        []string
	Devices                  map[string]*testapi.CrosTestRequest_Device
	PrimaryDevice            *testapi.CrosTestRequest_Device
	PrimaryDeviceMetadata    *skylab_test_runner.CFTTestRequest_Device
	CompanionDevices         []*testapi.CrosTestRequest_Device
	CompanionDevicesMetadata []*skylab_test_runner.CFTTestRequest_Device
	DutTopology              *labapi.DutTopology
	DutServerAddress         *labapi.IpEndpoint
	CurrentDutState          dutstate.State
	// Only when DUT is a VM
	DutVmGceImage   *vmlabapi.GceImage
	DutVm           *vmlabapi.VmInstance
	LeaseVMResponse *testapi.LeaseVMResponse

	// Provision related
	InstallMetadata    *anypb.Any
	ProvisionResponses map[string][]*testapi.InstallResponse

	// Test related
	TestArgs      *testapi.AutotestExecutionMetadata
	TestResponses *testapi.CrosTestResponse

	// Publish related
	GcsUrl              string
	StainlessUrl        string
	TesthausUrl         string
	GcsPublishSrcDir    string
	CurrentInvocationId string
	TkoPublishSrcDir    string
	CpconPublishSrcDir  string
	TestResultForRdb    *artifactpb.TestResult

	// Build related
	SkylabResult *skylab_test_runner.Result

	// Tools and their related dependencies
	Ctr                   *crostoolrunner.CrosToolRunner
	DockerKeyFileLocation string
}

func NewHwTestStateKeeper() *HwTestStateKeeper {
	return &HwTestStateKeeper{
		ContainerQueue:           list.New(),
		ProvisionQueue:           list.New(),
		PreTestQueue:             list.New(),
		TestQueue:                list.New(),
		PostTestQueue:            list.New(),
		PublishQueue:             list.New(),
		Injectables:              common.NewInjectableStorage(),
		ContainerInstances:       make(map[string]interfaces.ContainerInterface),
		ContainerImages:          make(map[string]*api.ContainerImageInfo),
		DeviceIdentifiers:        []string{},
		Devices:                  make(map[string]*testapi.CrosTestRequest_Device),
		CompanionDevices:         []*testapi.CrosTestRequest_Device{},
		CompanionDevicesMetadata: []*skylab_test_runner.CFTTestRequest_Device{},
		ProvisionResponses:       make(map[string][]*testapi.InstallResponse),
	}
}
