// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package rpc_services

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"
	swarmingapi "go.chromium.org/luci/swarming/proto/api_v2"
	moblabapipb "google.golang.org/genproto/googleapis/chromeos/moblab/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"infra/cros/satlab/common/dut"
	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/services"
	"infra/cros/satlab/common/services/build_service"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/executor"
	mk "infra/cros/satlab/satlabrpcserver/mocks"
	"infra/cros/satlab/satlabrpcserver/models"
	cpu "infra/cros/satlab/satlabrpcserver/platform/cpu_temperature"
	pb "infra/cros/satlab/satlabrpcserver/proto"
	"infra/cros/satlab/satlabrpcserver/services/dut_services"
	"infra/cros/satlab/satlabrpcserver/utils"
	"infra/cros/satlab/satlabrpcserver/utils/constants"
	mon "infra/cros/satlab/satlabrpcserver/utils/monitor"
	ufsModels "infra/unifiedfleet/api/v1/models"
	ufsApi "infra/unifiedfleet/api/v1/rpc"
	ufspb "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

type mockDeleteClient struct {
	getMachineLSECalls    []*ufspb.GetMachineLSERequest
	deleteMachineLSECalls []*ufspb.DeleteMachineLSERequest
	deleteAssetCalls      []*ufspb.DeleteAssetRequest
	deleteRackCalls       []*ufspb.DeleteRackRequest
}

func (c *mockDeleteClient) DeleteMachineLSE(ctx context.Context, req *ufsApi.DeleteMachineLSERequest, ops ...grpc.CallOption) (*emptypb.Empty, error) {
	c.deleteMachineLSECalls = append(c.deleteMachineLSECalls, req)
	return &emptypb.Empty{}, nil
}

func (c *mockDeleteClient) DeleteRack(ctx context.Context, req *ufsApi.DeleteRackRequest, ops ...grpc.CallOption) (*emptypb.Empty, error) {
	c.deleteRackCalls = append(c.deleteRackCalls, req)
	return &emptypb.Empty{}, nil
}

func (c *mockDeleteClient) DeleteAsset(ctx context.Context, req *ufsApi.DeleteAssetRequest, ops ...grpc.CallOption) (*emptypb.Empty, error) {
	c.deleteAssetCalls = append(c.deleteAssetCalls, req)
	return &emptypb.Empty{}, nil
}

func (c *mockDeleteClient) GetMachineLSE(ctx context.Context, req *ufsApi.GetMachineLSERequest, opts ...grpc.CallOption) (*ufsModels.MachineLSE, error) {
	c.getMachineLSECalls = append(c.getMachineLSECalls, req)
	return &ufsModels.MachineLSE{
		Name:     req.Name,
		Machines: []string{fmt.Sprintf("asset-%s", ufsUtil.RemovePrefix(req.Name))},
		Rack:     fmt.Sprintf("rack-%s", ufsUtil.RemovePrefix(req.Name)),
	}, nil
}

// checkShouldRaiseError it is a helper function to check the response should raise error.
func checkShouldRaiseError(t *testing.T, err error, expectedErr error) {
	if err == nil {
		t.Errorf("Should return error, but got no error")
	}

	if err.Error() != expectedErr.Error() {
		t.Errorf("Should return error, but get a different error. Expected %v, got %v", expectedErr, err)
	}
}

func createMockServer(t *testing.T) *SatlabRpcServiceServer {
	// Create a Mock `IBuildService`
	var mockBuildService = new(build_service.MockBuildService)

	// Create a Mock `IBucketService`
	var mockBucketService = new(mk.MockBucketServices)

	// Create a Mock `IDUTService`
	var mockDUTService = new(mk.MockDUTServices)

	// Create a Mock `ISwarmingService`
	var swarmingService = new(services.MockSwarmingService)

	// Create a SATLab Server
	return New(true, mockBuildService, mockBucketService, mockDUTService, nil, swarmingService)
}

// TestListBuildTargetsShouldSuccess test `ListBuildTargets` function.
//
// It should return some data without error.
func TestListBuildTargetsShouldSuccess(t *testing.T) {
	t.Parallel()
	// Create a SATLab Server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Setup some data to Mock
	expected := []string{"zork"}
	s.buildService.(*build_service.MockBuildService).On("ListBuildTargets", ctx).Return(
		expected, nil)

	req := &pb.ListBuildTargetsRequest{}

	res, err := s.ListBuildTargets(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	if !reflect.DeepEqual(expected, res.BuildTargets) {
		t.Errorf("Expected %v != got %v", expected, res.BuildTargets)
	}
}

// TestListBuildTargetsShouldSuccess test `ListBuildTargets` function.
//
// It should return error because it mocks some network error on calling
// `BuildClient` to fetch the data.
func TestListBuildTargetsShouldFailWhenMakeARequestToBuildClientFailed(t *testing.T) {
	t.Parallel()
	// Create a SATLab Server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Setup some data to Mock
	expectedErr := errors.New("network error")
	s.buildService.(*build_service.MockBuildService).On("ListBuildTargets", ctx).Return(
		[]string{}, expectedErr)

	req := &pb.ListBuildTargetsRequest{}

	_, err := s.ListBuildTargets(ctx, req)

	// Assert
	checkShouldRaiseError(t, err, expectedErr)
}

// TestListMilestonesShouldSuccess test `ListMilestones` function.
//
// It should return some data without error.
func TestListMilestonesShouldSuccess(t *testing.T) {
	t.Parallel()
	// Create a SATLab Server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Setup some data to Mock
	board := "zork"
	model := "dirinboz"
	expectedMilestones := []string{"114", "113"}
	s.buildService.(*build_service.MockBuildService).On("ListAvailableMilestones", ctx, board, model).Return(
		expectedMilestones, nil)

	localBucketMilestones := []string{"113"}
	s.bucketService.(*mk.MockBucketServices).On("GetMilestones", ctx, board).Return(
		localBucketMilestones, nil)
	s.bucketService.(*mk.MockBucketServices).On("IsBucketInAsia", ctx).Return(
		false, nil)

	req := &pb.ListMilestonesRequest{
		Board: board,
		Model: model,
	}

	res, err := s.ListMilestones(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	if len(res.Milestones) != 2 {
		t.Errorf("Expected %v items, but got %v", 2, len(res.Milestones))
	}

	// Assert
	expected := []*pb.BuildItem{
		{
			Value:    "114",
			IsStaged: false,
		},
		{
			Value:    "113",
			IsStaged: true,
		},
	}

	if !reflect.DeepEqual(expected, res.Milestones) {
		t.Errorf("Expected %v != got %v", expected, res.Milestones)
	}
}

// TestListMilestonesShouldSuccessWhenBucketInAsia test `ListMilestones` function.
func TestListMilestonesShouldSuccessWhenBucketInAsia(t *testing.T) {
	t.Parallel()
	// Create a SATLab Server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Setup some data to Mock
	board := "zork"
	model := "dirinboz"
	expectedMilestones := []string{"114", "113"}
	s.buildService.(*build_service.MockBuildService).On("ListAvailableMilestones", ctx, board, model).Return(
		expectedMilestones, nil)

	localBucketMilestones := []string{"113"}
	s.bucketService.(*mk.MockBucketServices).On("GetMilestones", ctx, board).Return(
		localBucketMilestones, nil)
	s.bucketService.(*mk.MockBucketServices).On("IsBucketInAsia", ctx).Return(
		true, nil)

	req := &pb.ListMilestonesRequest{
		Board: board,
		Model: model,
	}

	res, err := s.ListMilestones(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	if len(res.Milestones) != 1 {
		t.Errorf("Expected %v items, but got %v", 2, len(res.Milestones))
	}

	// Assert
	expected := []*pb.BuildItem{
		{
			Value:    "113",
			IsStaged: true,
		},
	}

	if !reflect.DeepEqual(expected, res.Milestones) {
		t.Errorf("Expected %v != got %v", expected, res.Milestones)
	}
}

// TestListMilestonesShouldSuccessWhenBucketInAsia test `ListMilestones` function.
func TestListMilestonesShouldFailWhenMakeARequestToBucketFailed(t *testing.T) {
	t.Parallel()
	// Create a SATLab Server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Setup some data to Mock
	board := "zork"
	model := "dirinboz"
	expectedMilestones := []string{"114", "113"}
	expectedErr := errors.New("can't make a request")
	s.buildService.(*build_service.MockBuildService).On("ListAvailableMilestones", ctx, board, model).Return(
		expectedMilestones, nil)

	localBucketMilestones := []string{"113"}
	s.bucketService.(*mk.MockBucketServices).On("GetMilestones", ctx, board).Return(
		localBucketMilestones, nil)
	s.bucketService.(*mk.MockBucketServices).On("IsBucketInAsia", ctx).Return(
		false, expectedErr)

	req := &pb.ListMilestonesRequest{
		Board: board,
		Model: model,
	}

	_, err := s.ListMilestones(ctx, req)

	// Assert
	checkShouldRaiseError(t, err, expectedErr)
}

// TestListAccessibleModelShouldSuccess test `ListAccessibleModel` function.
func TestListAccessibleModelShouldSuccess(t *testing.T) {
	t.Parallel()
	// Create a SATLab Server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Setup some data to Mock
	board := "zork"
	in := []string{"buildTargets/zork/models/model1", "buildTargets/zork/models/model2", "buildTargets/zork/models/dirinboz"}
	s.buildService.(*build_service.MockBuildService).On("ListModels", ctx, board).Return(
		in, nil)

	req := &pb.ListAccessibleModelsRequest{
		Board: board,
	}

	res, err := s.ListAccessibleModels(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	if len(res.Models) != 3 {
		t.Errorf("Should got %v difference models", 3)
	}

	expected := &pb.ListAccessibleModelsResponse{
		Models: []*pb.Model{
			{
				Name:   "model1",
				Boards: []string{"zork"},
			},
			{
				Name:   "dirinboz",
				Boards: []string{"zork"},
			},
			{
				Name:   "model2",
				Boards: []string{"zork"},
			},
		},
	}

	// Assert
	// ignore generated pb code
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(pb.ListAccessibleModelsResponse{}, pb.Model{})
	// Model ordering is not deterministic, need to sort before comparing
	sortModelsOpts := cmpopts.SortSlices(
		func(x, y *pb.Model) bool {
			return x.GetName() > y.GetName()
		})

	if diff := cmp.Diff(expected, res, ignorePBFieldOpts, sortModelsOpts); diff != "" {
		t.Errorf("Expected %v, got %v", expected, res.Models)
	}
}

// TestListAccessibleModelShouldSuccess test `ListAccessibleModel` function.
func TestListAccessibleModelShouldFailWhenMakeARequestToBucketFailed(t *testing.T) {
	t.Parallel()
	// Create a SATLab Server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Setup some data to Mock
	board := "zork"
	in := []string{"buildTargets/zork/models/model1", "buildTargets/zork/models/model2", "buildTargets/zork/models/dirinboz"}
	expectedErr := errors.New("can't make a request to bucket")
	s.buildService.(*build_service.MockBuildService).On("ListModels", ctx, board).Return(
		in, expectedErr)

	req := &pb.ListAccessibleModelsRequest{
		Board: board,
	}

	_, err := s.ListAccessibleModels(ctx, req)

	// Assert
	checkShouldRaiseError(t, err, expectedErr)
}

// TestListBuildVersionsShouldSuccess test `ListBuildVersions` function.
func TestListBuildVersionsShouldSuccess(t *testing.T) {
	t.Parallel()
	// Create a SATLab Server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Setup some data to Mock
	board := "zork1"
	model := "dirinboz1"
	var milestone int32 = 105
	s.bucketService.(*mk.MockBucketServices).
		On("GetBuilds", ctx, board, milestone).
		Return([]string{"14820.8.0"}, nil)

	s.buildService.(*build_service.MockBuildService).
		On("ListBuildsForMilestone", ctx, board, model, milestone).
		Return([]*build_service.BuildVersion{
			{
				Version: "14820.100.0",
				Status:  build_service.FAILED,
			},
			{
				Version: "14820.20.0",
				Status:  build_service.AVAILABLE,
			},
			{
				Version: "14820.8.0",
				Status:  build_service.AVAILABLE,
			},
		}, nil)

	s.bucketService.(*mk.MockBucketServices).On("IsBucketInAsia", ctx).Return(
		false, nil)

	req := &pb.ListBuildVersionsRequest{Board: board, Model: model, Milestone: milestone}

	res, err := s.ListBuildVersions(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	if len(res.BuildVersions) != 3 {
		t.Errorf("Should got %v difference models", 3)
	}

	expectedResult := []*pb.BuildItem{
		{
			Value:    "14820.100.0",
			Status:   pb.BuildItem_BUILD_STATUS_FAIL,
			IsStaged: false,
		},
		{
			Value:    "14820.20.0",
			Status:   pb.BuildItem_BUILD_STATUS_PASS,
			IsStaged: false,
		},
		{
			Value:    "14820.8.0",
			Status:   pb.BuildItem_BUILD_STATUS_PASS,
			IsStaged: true,
		},
	}

	// Assert
	if !reflect.DeepEqual(expectedResult, res.BuildVersions) {
		t.Errorf("Expected %v != got %v", expectedResult, res.BuildVersions)
	}
}

// TestListBuildVersionsShouldSuccess test `ListBuildVersions` function.
func TestListBuildVersionsShouldFailWhenMakeARequestToBuildClientFailed(t *testing.T) {
	t.Parallel()
	// Create a SATLab Server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Setup some data to Mock
	board := "zork"
	model := "dirinboz"
	var milestone int32 = 105
	expectedErr := errors.New("can't make a request to bucket")
	s.bucketService.(*mk.MockBucketServices).
		On("GetBuilds", ctx, board, milestone).
		Return([]string{"14826.0.0"}, nil)

	s.bucketService.(*mk.MockBucketServices).On("IsBucketInAsia", ctx).Return(
		false, nil)

	s.buildService.(*build_service.MockBuildService).
		On("ListBuildsForMilestone", ctx, board, model, milestone).
		Return([]*build_service.BuildVersion{}, expectedErr)

	req := &pb.ListBuildVersionsRequest{Board: board, Model: model, Milestone: milestone}

	_, err := s.ListBuildVersions(ctx, req)

	// Assert
	checkShouldRaiseError(t, err, expectedErr)
}

// TestStageBuildShouldSuccess test `StageBuild` function.
func TestStageBuildShouldSuccess(t *testing.T) {
	t.Parallel()
	// Create a SATLab Server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Setup some data to Mock
	board := "zork"
	model := "dirinboz"
	build := "1234.0.0"
	bucketName := site.GetGCSImageBucket()
	expectedArtifact := &moblabapipb.BuildArtifact{
		Build:  build,
		Name:   "artifacts",
		Bucket: bucketName,
		Path:   "buildTargets/zork/models/dirinboz/builds/1234.0.0/artifacts/chromeos-image-archive",
	}

	s.buildService.(*build_service.MockBuildService).
		On("StageBuild", ctx, board, model, build, bucketName).
		Return(expectedArtifact, nil)

	req := &pb.StageBuildRequest{
		Board:        board,
		Model:        model,
		BuildVersion: build,
	}

	res, err := s.StageBuild(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	if res.GetBuildBucket() != bucketName {
		t.Errorf("Expected %v, got: %v", bucketName, res.GetBuildBucket())
	}
}

// TestStageBuildShouldSuccess test `StageBuild` function.
func TestStageBuildShouldFailWhenMakeARequestToBuildClientFailed(t *testing.T) {
	t.Parallel()
	// Create a SATLab Server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Setup some data to Mock
	board := "zork"
	model := "dirinboz"
	build := "1234.0.0"
	bucketName := site.GetGCSImageBucket()
	expectedArtifact := &moblabapipb.BuildArtifact{
		Build:  build,
		Name:   "artifacts",
		Bucket: bucketName,
		Path:   "buildTargets/zork/models/dirinboz/builds/1234.0.0/artifacts/chromeos-image-archive",
	}
	expectedErr := errors.New("can't make a request")

	s.buildService.(*build_service.MockBuildService).
		On("StageBuild", ctx, board, model, build, bucketName).
		Return(expectedArtifact, expectedErr)

	req := &pb.StageBuildRequest{
		Board:        board,
		Model:        model,
		BuildVersion: build,
	}

	_, err := s.StageBuild(ctx, req)

	// Assert
	checkShouldRaiseError(t, err, expectedErr)
}

func TestListConnectedDUTsFirmwareShouldSuccess(t *testing.T) {
	t.Parallel()
	// Create a mock server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cmdOut := "{\n  \"fwid\": \"Google_Lindar.13672.291.0\",\n  \"model\": \"lillipup\",\n  \"fw_update\": {\n    \"lillipup\": {\n      \"host\": {\n        \"versions\": {\n          \"ro\": \"Google_Lindar.13672.207.0\",\n          \"rw\": \"Google_Lindar.13672.291.0\"\n        },\n        \"keys\": {\n          \"root\": \"b11d74edd286c144e1135b49e7f0bc20cf041f10\",\n          \"recovery\": \"c14bd720b70d97394257e3e826bd8f43de48d4ed\"\n        },\n        \"image\": \"images/bios-lindar.ro-13672-207-0.rw-13672-291-0.bin\"\n      },\n      \"ec\": {\n        \"versions\": {\n          \"ro\": \"lindar_v2.0.7573-4cf04a534f\",\n          \"rw\": \"lindar_v2.0.10133-063f551128\"\n        },\n        \"image\": \"images/ec-lindar.ro-2-0-7573.rw-2-0-10133.bin\"\n      },\n      \"signature_id\": \"lillipup\"\n    }\n  }\n}\n"

	// Mock some data
	IP := "192.168.100.1"
	s.dutService.(*mk.MockDUTServices).On("GetConnectedIPs", ctx).Return([]dut_services.Device{
		{IP: IP, IsConnected: true},
	}, nil)
	s.dutService.(*mk.MockDUTServices).
		On("RunCommandOnIPs", ctx, mock.Anything, constants.ListFirmwareCommand).
		Return([]*models.SSHResult{
			{IP: IP, Value: cmdOut},
		})

	req := &pb.ListConnectedDutsFirmwareRequest{}

	res, err := s.ListConnectedDutsFirmware(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	expected := []*pb.ConnectedDutFirmwareInfo{{
		Ip:              IP,
		CurrentFirmware: "Google_Lindar.13672.291.0",
		UpdateFirmware:  "Google_Lindar.13672.291.0",
	}}

	if !reflect.DeepEqual(expected, res.Duts) {
		t.Errorf("Expected: %v, got :%v", expected, res.Duts)
	}
}

func TestListConnectedDUTsFirmwareShouldGetEmptyListWhenCommandExecuteFailed(t *testing.T) {
	t.Parallel()
	// Create a mock server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	expectedError := errors.New("command execute failed")

	// Mock some data
	IP := "192.168.100.1"
	s.dutService.(*mk.MockDUTServices).On("GetConnectedIPs", ctx).Return([]dut_services.Device{
		{IP: IP, IsConnected: true},
	}, nil)
	s.dutService.(*mk.MockDUTServices).
		On("RunCommandOnIPs", ctx, mock.Anything, constants.ListFirmwareCommand).
		Return([]*models.SSHResult{
			{IP: IP, Error: expectedError},
		})

	req := &pb.ListConnectedDutsFirmwareRequest{}

	res, err := s.ListConnectedDutsFirmware(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	if len(res.Duts) != 0 {
		t.Errorf("Expected zero dut")
	}
}

func TestGetSystemInfoShouldWork(t *testing.T) {
	t.Parallel()
	// Create a mock server
	s := createMockServer(t)
	var mockCPUTemperature = new(mk.MockCPUTemperature)
	mockCPUTemperature.On("GetCurrentCPUTemperature").Return(float32(1.0), nil)
	var cpuOrchestrator = cpu.NewOrchestrator(mockCPUTemperature, 5)
	s.cpuTemperatureOrchestrator = cpuOrchestrator

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Make some data
	m := mon.New()
	m.Register(cpuOrchestrator, time.Second)
	time.Sleep(time.Second * 2)

	req := pb.GetSystemInfoRequest{}

	res, err := s.GetSystemInfo(ctx, &req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	expected := 1.0
	if !utils.NearlyEqual(float64(res.GetCpuTemperature()), expected) {
		t.Errorf("Expected %v, got %v", expected, res.GetCpuTemperature())
	}
}

func TestGetSystemInfoShouldWorkWithoutCPUOrchestrator(t *testing.T) {
	t.Parallel()
	// Create a mock server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	req := pb.GetSystemInfoRequest{}

	res, err := s.GetSystemInfo(ctx, &req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	expected := -1.0
	if !utils.NearlyEqual(float64(res.GetCpuTemperature()), expected) {
		t.Errorf("Expected %v, got %v", expected, res.GetCpuTemperature())
	}
}

func TestGetPeripheralInformationShouldSuccess(t *testing.T) {
	t.Parallel()
	// Create a mock server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	expectedResult := &models.SSHResult{
		Value: "<json data>",
		IP:    "192.168.231.100",
	}

	// Mock some data
	s.dutService.(*mk.MockDUTServices).
		On("RunCommandOnIP", ctx, mock.Anything, constants.GetPeripheralInfoCommand).
		Return(expectedResult, nil)

	req := &pb.GetPeripheralInformationRequest{}
	res, err := s.GetPeripheralInformation(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	if diff := cmp.Diff(expectedResult.Value, res.JsonInfo); diff != "" {
		t.Errorf("Return difference result. Expected %v, got %v", expectedResult.Value, res.JsonInfo)
	}
}

func TestGetPeripheralInformationShouldFailWhenExecuteCommandFailed(t *testing.T) {
	t.Parallel()
	// Create a mock server
	s := createMockServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	expectedResult := &models.SSHResult{
		Error: errors.New("execute cmd failed"),
		IP:    "192.168.231.100",
	}

	// Mock some data
	s.dutService.(*mk.MockDUTServices).
		On("RunCommandOnIP", ctx, mock.Anything, constants.GetPeripheralInfoCommand).
		Return(expectedResult, nil)

	req := &pb.GetPeripheralInformationRequest{}
	_, err := s.GetPeripheralInformation(ctx, req)

	// Assert
	if diff := cmp.Diff(expectedResult.Error.Error(), err.Error()); diff != "" {
		t.Errorf("Return difference result. Expected %v, got %v", expectedResult.Error, err)
	}
}

func TestUpdateDUTsFirmwareShouldSuccess(t *testing.T) {
	// Run this testcase parallel
	t.Parallel()
	// Create a mock server
	s := createMockServer(t)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// Create a mockData data
	mockData := []*models.SSHResult{
		{IP: "192.168.231.1", Value: "execute command success"},
		{IP: "192.168.231.2", Error: errors.New("failed to execute command")},
	}
	s.dutService.(*mk.MockDUTServices).
		On("RunCommandOnIPs", ctx, mock.Anything, constants.UpdateFirmwareCommand).
		Return(mockData, nil)

	// Act
	req := &pb.UpdateDutsFirmwareRequest{Ips: []string{"192.168.231.1", "192.168.231.2"}}
	resp, err := s.UpdateDutsFirmware(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: {%v}", err)
	}

	// Create a expected result
	expected := []*pb.FirmwareUpdateCommandOutput{
		{Ip: "192.168.231.1", CommandOutput: "execute command success"},
		{Ip: "192.168.231.2", CommandOutput: "failed to execute command"},
	}
	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(pb.FirmwareUpdateCommandOutput{})
	// sort the response and expected result when comparasion
	sortOpts := cmpopts.SortSlices(
		func(x, y *pb.FirmwareUpdateCommandOutput) bool {
			return x.GetIp() > y.GetIp()
		},
	)

	if diff := cmp.Diff(expected, resp.Outputs, ignorePBFieldOpts, sortOpts); diff != "" {
		t.Errorf("Expected: {%v}, got: {%v}", expected, resp.Outputs)
	}
}

func TestGetVersionInfoShouldSuccess(t *testing.T) {
	// Run this testcase parallel
	t.Parallel()
	// Create a mock server
	s := createMockServer(t)
	s.commandExecutor = &executor.FakeCommander{
		CmdOutput: "LABEL=output\ndescription:description\nversion:v\ntrack:track",
	}

	ctx := context.Background()
	req := &pb.GetVersionInfoRequest{}
	resp, err := s.GetVersionInfo(ctx, req)

	if err != nil {
		t.Errorf("Should success, but got an error: %v\n", err)
	}

	expected := &pb.GetVersionInfoResponse{
		HostId:          "label=output\ndescription:description\nversion:v\ntrack:track",
		Description:     "description",
		Track:           "track",
		ChromeosVersion: "v",
		Version:         "output",
	}

	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(pb.GetVersionInfoResponse{})

	if diff := cmp.Diff(expected, resp, ignorePBFieldOpts); diff != "" {
		t.Errorf("Expected: {%v}, got: {%v}, {%v}", expected, resp, diff)
	}
}

func TestGetVersionInfoShouldFail(t *testing.T) {
	// Run this testcase parallel
	t.Parallel()
	// Create a mock server
	s := createMockServer(t)
	s.commandExecutor = &executor.FakeCommander{
		Err: errors.New("exec command failed"),
	}

	ctx := context.Background()
	req := &pb.GetVersionInfoRequest{}
	res, err := s.GetVersionInfo(ctx, req)

	if err == nil {
		t.Errorf("Expected error")
	}

	if res != nil {
		t.Errorf("Expected the reuslt should be nil")
	}
}

func TestGetDUTDetailShouldSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// Create a mock data
	s := createMockServer(t)
	s.commandExecutor = &executor.FakeCommander{
		CmdOutput: `
192.168.231.137	satlab-0wgtfqin1846803b-one
192.168.231.137	satlab-0wgtfqin1846803b-host5
192.168.231.222	satlab-0wgtfqin1846803b-host11
192.168.231.222	satlab-0wgtfqin1846803b-host12
  `,
	}
	mockData := &swarmingapi.BotInfo{BotId: "test bot"}
	s.swarmingService.(*services.MockSwarmingService).
		On("GetBot", ctx, mock.Anything).
		Return(mockData, nil)

	req := &pb.GetDutDetailRequest{
		Address: "192.168.231.222",
	}
	resp, err := s.GetDutDetail(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: {%v}", err)
	}

	// Create a expected result
	expected := &pb.GetDutDetailResponse{
		BotId:      "test bot",
		Dimensions: []*pb.StringListPair{},
	}
	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(pb.GetDutDetailResponse{})

	if diff := cmp.Diff(expected, resp, ignorePBFieldOpts); diff != "" {
		t.Errorf("Expected: {%v}, got: {%v}, %v", expected, resp, diff)
	}
}

func TestListDutTasksShouldSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// Create a mock data
	s := createMockServer(t)
	s.commandExecutor = &executor.FakeCommander{
		CmdOutput: `
192.168.231.222	satlab-0wgtfqin1846803b-host12
  `,
	}
	mockData := &services.TasksIterator{
		Cursor: "next_cursor",
		Tasks: []services.Task{
			{
				Id: "task id",
			},
		},
	}
	s.swarmingService.(*services.MockSwarmingService).
		On("ListBotTasks", ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(mockData, nil)

	req := &pb.ListDutTasksRequest{
		Cursor:   "",
		PageSize: 1,
		Address:  "192.168.231.222",
	}
	resp, err := s.ListDutTasks(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: {%v}", err)
	}

	// Create a expected result
	expected := &pb.ListDutTasksResponse{
		Cursor: "next_cursor",
		Tasks: []*pb.Task{
			{
				Id: "task id",
			},
		},
	}
	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(pb.ListDutTasksResponse{}, pb.Task{})

	if diff := cmp.Diff(expected, resp, ignorePBFieldOpts); diff != "" {
		t.Errorf("Expected: {%v}, got: {%v}, %v", expected, resp, diff)
	}
}

func TestListDutEventsShouldSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// Create a mock data
	s := createMockServer(t)
	s.commandExecutor = &executor.FakeCommander{
		CmdOutput: `
192.168.231.222	satlab-0wgtfqin1846803b-host12
  `,
	}
	mockData := &services.BotEventsIterator{
		Cursor: "next_cursor",
		Events: []services.BotEvent{
			{
				TaskID: "task id",
			},
		},
	}
	s.swarmingService.(*services.MockSwarmingService).
		On("ListBotEvents", ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(mockData, nil)

	req := &pb.ListDutEventsRequest{
		Cursor:   "",
		PageSize: 1,
		Address:  "192.168.231.222",
	}
	resp, err := s.ListDutEvents(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: {%v}", err)
	}

	// Create a expected result
	expected := &pb.ListDutEventsResponse{
		Cursor: "next_cursor",
		Events: []*pb.BotEvent{
			{
				TaskId: "task id",
			},
		},
	}
	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(pb.ListDutEventsResponse{}, pb.BotEvent{})

	if diff := cmp.Diff(expected, resp, ignorePBFieldOpts); diff != "" {
		t.Errorf("Expected: {%v}, got: {%v}, %v", expected, resp, diff)
	}
}

func getDUTOutput() []byte {
	return []byte(`[{
"name": "satlab-0wgatfqi21498062-jeff137-c",
"machineLsePrototype": "",
"hostname": "satlab-0wgatfqi21498062-jeff137-c",
"chromeosMachineLse": {
	"deviceLse": {
        "dut": {
			"hostname": "satlab-0wgatfqi21498062-jeff137-c",
			"pools": [
				"jev-satlab"
			]
        }
	}
},
"machines": [
      "JEFF137-c"
]}]`)
}

func getAssetOutput() []byte {
	return []byte(`[{
"name": "JEFF137-c",
"type": "DUT",
"model": "atlas",
"info": {
	"model": "atlas",
	"buildTarget": "atlas"
}}]`)
}

func shivasTestHelper(hasData bool) executor.IExecCommander {
	return &executor.FakeCommander{
		FakeFn: func(in *exec.Cmd) ([]byte, error) {
			if in.Path == "/usr/local/bin/shivas" {
				for _, arg := range in.Args {
					if arg == "dut" {
						// execute a command to get dut
						if hasData {
							return getDUTOutput(), nil
						} else {
							return []byte("[]"), nil
						}
					}

					if arg == "asset" {
						// execute a command to get asset
						if hasData {
							return getAssetOutput(), nil
						} else {
							return []byte("[]"), nil
						}
					}
				}
			}

			if in.Path == "/usr/local/bin/get_host_identifier" {
				return []byte("satlab-id"), nil
			}
			if in.Path == "/usr/local/bin/docker" {
				return []byte("192.168.231.222	satlab-0wgatfqi21498062-jeff137-c"), nil
			}

			return nil, errors.New(fmt.Sprintf("handle command: %v", in.Path))
		},
	}

}

func TestListEnrolledDutsShouldSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// Create a mock data
	s := createMockServer(t)

	s.commandExecutor = shivasTestHelper(true)
	req := &pb.ListEnrolledDutsRequest{}
	resp, err := s.ListEnrolledDuts(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: {%v}", err)
	}

	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(pb.ListEnrolledDutsResponse{}, pb.Dut{})

	// Create a expected result
	expected := &pb.ListEnrolledDutsResponse{
		Duts: []*pb.Dut{
			{
				Name:     "satlab-0wgatfqi21498062-jeff137-c",
				Hostname: "satlab-0wgatfqi21498062-jeff137-c",
				Address:  "192.168.231.222",
				Pools:    []string{"jev-satlab"},
				Model:    "atlas",
				Board:    "atlas",
			},
		},
	}

	if diff := cmp.Diff(expected, resp, ignorePBFieldOpts); diff != "" {
		t.Errorf("diff: %v\n", diff)
	}
}

func TestListEnrolledDutsShouldFail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// Create a mock data
	s := createMockServer(t)

	s.commandExecutor = &executor.FakeCommander{
		Err: errors.New("execute command failed"),
	}
	req := &pb.ListEnrolledDutsRequest{}
	resp, err := s.ListEnrolledDuts(ctx, req)

	// Assert
	if err == nil {
		t.Errorf("Should be failed in this test case")
	}

	if resp != nil {
		t.Errorf("Should get the empty result.")
	}
}

func TestListConnectedAndEnrolledDutsShouldSuccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create a mock data
	s := createMockServer(t)
	s.dutService.(*mk.MockDUTServices).On("GetConnectedIPs", ctx).Return([]dut_services.Device{
		{IP: "192.168.231.222", MACAddress: "00:14:3d:14:c4:02", IsConnected: true},
		{IP: "192.168.231.2", MACAddress: "e8:9f:80:83:3d:c8", IsConnected: true},
	}, nil)
	s.commandExecutor = shivasTestHelper(true)

	req := &pb.ListDutsRequest{}
	resp, err := s.ListDuts(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: {%v}", err)
	}

	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(pb.ListDutsResponse{}, pb.Dut{})
	// Create a expected result
	expected := &pb.ListDutsResponse{
		Duts: []*pb.Dut{
			{
				Name:        "satlab-0wgatfqi21498062-jeff137-c",
				Hostname:    "satlab-0wgatfqi21498062-jeff137-c",
				Address:     "192.168.231.222",
				Pools:       []string{"jev-satlab"},
				Model:       "atlas",
				Board:       "atlas",
				IsConnected: true,
				MacAddress:  "00:14:3d:14:c4:02",
			},
			{
				Name:        "",
				Hostname:    "",
				Address:     "192.168.231.2",
				Pools:       nil,
				Model:       "",
				Board:       "",
				MacAddress:  "e8:9f:80:83:3d:c8",
				IsConnected: true,
			},
		},
	}

	sortModelsOpts := cmpopts.SortSlices(
		func(x, y *pb.Dut) bool {
			return x.GetAddress() > y.GetAddress()
		})

	if diff := cmp.Diff(expected, resp, ignorePBFieldOpts, sortModelsOpts); diff != "" {
		t.Errorf("diff: %v\n", diff)
	}
}

func TestListDisconnectedAndEnrolledDutsShouldSuccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// Create a mock data

	s := createMockServer(t)
	s.dutService.(*mk.MockDUTServices).On("GetConnectedIPs", ctx).Return([]dut_services.Device{
		{IP: "192.168.231.222", MACAddress: "00:14:3d:14:c4:02", IsConnected: false},
		{IP: "192.168.231.2", MACAddress: "e8:9f:80:83:3d:c8", IsConnected: false},
	}, nil)
	s.commandExecutor = shivasTestHelper(true)

	req := &pb.ListDutsRequest{}
	resp, err := s.ListDuts(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: {%v}", err)
	}

	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(pb.ListDutsResponse{}, pb.Dut{})

	// Create a expected result
	expected := &pb.ListDutsResponse{
		Duts: []*pb.Dut{
			{
				Name:        "satlab-0wgatfqi21498062-jeff137-c",
				Hostname:    "satlab-0wgatfqi21498062-jeff137-c",
				Address:     "192.168.231.222",
				Pools:       []string{"jev-satlab"},
				Model:       "atlas",
				Board:       "atlas",
				IsConnected: false,
				MacAddress:  "00:14:3d:14:c4:02",
			},
			{
				Name:        "",
				Hostname:    "",
				Address:     "192.168.231.2",
				Pools:       nil,
				Model:       "",
				Board:       "",
				MacAddress:  "e8:9f:80:83:3d:c8",
				IsConnected: false,
			},
		},
	}

	sortModelsOpts := cmpopts.SortSlices(
		func(x, y *pb.Dut) bool {
			return x.GetAddress() > y.GetAddress()
		})

	if diff := cmp.Diff(expected, resp, ignorePBFieldOpts, sortModelsOpts); diff != "" {
		t.Errorf("diff: %v\n", diff)
	}
}

func TestListConnectedAndUnenrolledDutsShouldSuccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create a mock data
	s := createMockServer(t)
	s.dutService.(*mk.MockDUTServices).On("GetConnectedIPs", ctx).Return([]dut_services.Device{
		{IP: "192.168.231.222", MACAddress: "00:14:3d:14:c4:02", IsConnected: true},
		{IP: "192.168.231.2", MACAddress: "e8:9f:80:83:3d:c8", IsConnected: true},
	}, nil)
	s.commandExecutor = shivasTestHelper(false)

	req := &pb.ListDutsRequest{}
	resp, err := s.ListDuts(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: {%v}", err)
	}

	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(pb.ListDutsResponse{}, pb.Dut{})

	// Create a expected result
	expected := &pb.ListDutsResponse{
		Duts: []*pb.Dut{
			{
				Name:        "",
				Hostname:    "",
				Address:     "192.168.231.222",
				Pools:       nil,
				Model:       "",
				Board:       "",
				IsConnected: true,
				MacAddress:  "00:14:3d:14:c4:02",
			},
			{
				Name:        "",
				Hostname:    "",
				Address:     "192.168.231.2",
				Pools:       nil,
				Model:       "",
				Board:       "",
				MacAddress:  "e8:9f:80:83:3d:c8",
				IsConnected: true,
			},
		},
	}

	sortModelsOpts := cmpopts.SortSlices(
		func(x, y *pb.Dut) bool {
			return x.GetAddress() > y.GetAddress()
		})

	if diff := cmp.Diff(expected, resp, ignorePBFieldOpts, sortModelsOpts); diff != "" {
		t.Errorf("diff: %v\n", diff)
	}
}

func TestListDisconnectedAndUnenrolledDutsShouldSuccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create a mock data
	s := createMockServer(t)
	s.dutService.(*mk.MockDUTServices).On("GetConnectedIPs", ctx).Return([]dut_services.Device{
		{IP: "192.168.231.222", MACAddress: "00:14:3d:14:c4:02", IsConnected: false},
		{IP: "192.168.231.2", MACAddress: "e8:9f:80:83:3d:c8", IsConnected: false},
	}, nil)
	s.commandExecutor = shivasTestHelper(false)

	req := &pb.ListDutsRequest{}
	resp, err := s.ListDuts(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: {%v}", err)
	}

	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(pb.ListDutsResponse{}, pb.Dut{})

	// Create a expected result
	expected := &pb.ListDutsResponse{
		Duts: []*pb.Dut{
			{
				Name:        "",
				Hostname:    "",
				Address:     "192.168.231.222",
				Pools:       nil,
				Model:       "",
				Board:       "",
				IsConnected: false,
				MacAddress:  "00:14:3d:14:c4:02",
			},
			{
				Name:        "",
				Hostname:    "",
				Address:     "192.168.231.2",
				Pools:       nil,
				Model:       "",
				Board:       "",
				MacAddress:  "e8:9f:80:83:3d:c8",
				IsConnected: false,
			},
		},
	}

	sortModelsOpts := cmpopts.SortSlices(
		func(x, y *pb.Dut) bool {
			return x.GetAddress() > y.GetAddress()
		})

	if diff := cmp.Diff(expected, resp, ignorePBFieldOpts, sortModelsOpts); diff != "" {
		t.Errorf("diff: %v\n", diff)
	}
}

func TestListConnectedDutsShouldFail(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create a mock data
	s := createMockServer(t)
	s.dutService.(*mk.MockDUTServices).On("GetConnectedIPs", ctx).Return([]dut_services.Device{
		{IP: "192.168.231.222", MACAddress: "00:14:3d:14:c4:02", IsConnected: false},
		{IP: "192.168.231.2", MACAddress: "e8:9f:80:83:3d:c8", IsConnected: false},
	}, nil)
	s.commandExecutor = &executor.FakeCommander{
		Err: errors.New("execute command failed"),
	}

	req := &pb.ListDutsRequest{}
	resp, err := s.ListDuts(ctx, req)

	// Assert
	if err == nil {
		t.Errorf("should fail")
	}
	if resp != nil {
		t.Errorf("response should be empty, but got %v", resp)
	}
}

func TestDeleteDutsShouldSuccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create a mock data
	e := &executor.FakeCommander{
		FakeFn: func(c *exec.Cmd) ([]byte, error) {
			// 192.168.231.222	satlab-0wgtfqin1846803b-host12
			if c.Path == paths.DockerPath {
				return []byte(`
192.168.231.222	satlab-0wgtfqin1846803b-host12
        `), nil
			} else if c.Path == paths.GetHostIdentifierScript {
				return []byte("0wgtfqin1846803b"), nil
			} else {
				return nil, errors.New(fmt.Sprintf("execute a command %v\n", c.Path))
			}
		},
	}

	addresses := []string{"192.168.231.222"}
	ufs := mockDeleteClient{}

	resp, invalidAddresses, err := innerDeleteDuts(ctx, e, &ufs, addresses, false)
	if err != nil {
		t.Errorf("unexpected error: %v\n", err)
		return
	}

	if len(invalidAddresses) != 0 {
		t.Errorf("invalid addresses should be empty")
	}

	expected := &dut.DeleteDUTResult{
		MachineLSEs: []*ufsModels.MachineLSE{
			{
				Name:     "machineLSEs/satlab-0wgtfqin1846803b-host12",
				Machines: []string{"asset-satlab-0wgtfqin1846803b-host12"},
				Rack:     "rack-satlab-0wgtfqin1846803b-host12",
			},
		},
		DutResults: &dut.Result{
			Pass: []string{"satlab-0wgtfqin1846803b-host12"},
			Fail: []string{},
		},
		AssetResults: &dut.Result{},
		RackResults:  &dut.Result{},
	}

	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(ufsModels.MachineLSE{})

	if diff := cmp.Diff(resp, expected, ignorePBFieldOpts); diff != "" {
		t.Errorf("unexpected diff: %v\n", diff)
	}
}

func TestDeleteDutsWithInvalidAddressesShouldSuccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create a mock data
	e := &executor.FakeCommander{
		FakeFn: func(c *exec.Cmd) ([]byte, error) {
			// 192.168.231.222	satlab-0wgtfqin1846803b-host12
			if c.Path == paths.DockerPath {
				return []byte(`
192.168.231.222	satlab-0wgtfqin1846803b-host12
        `), nil
			} else if c.Path == paths.GetHostIdentifierScript {
				return []byte("0wgtfqin1846803b"), nil
			} else {
				return nil, errors.New(fmt.Sprintf("execute a command %v\n", c.Path))
			}
		},
	}

	addresses := []string{"192.168.231.222", "192.168.231.221"}
	ufs := mockDeleteClient{}

	resp, invalidAddresses, err := innerDeleteDuts(ctx, e, &ufs, addresses, false)
	if err != nil {
		t.Errorf("unexpected error: %v\n", err)
		return
	}

	expectedInvalidAddresses := []string{"192.168.231.221"}

	if diff := cmp.Diff(invalidAddresses, expectedInvalidAddresses); diff != "" {
		t.Errorf("unexpected invalid addresses: %v", diff)
	}

	expected := &dut.DeleteDUTResult{
		MachineLSEs: []*ufsModels.MachineLSE{
			{
				Name:     "machineLSEs/satlab-0wgtfqin1846803b-host12",
				Machines: []string{"asset-satlab-0wgtfqin1846803b-host12"},
				Rack:     "rack-satlab-0wgtfqin1846803b-host12",
			},
		},
		DutResults: &dut.Result{
			Pass: []string{"satlab-0wgtfqin1846803b-host12"},
			Fail: []string{},
		},
		AssetResults: &dut.Result{},
		RackResults:  &dut.Result{},
	}

	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(ufsModels.MachineLSE{})

	if diff := cmp.Diff(resp, expected, ignorePBFieldOpts); diff != "" {
		t.Errorf("unexpected diff: %v\n", diff)
	}
}

func TestFullDeleteDutsShouldSuccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create a mock data
	e := &executor.FakeCommander{
		FakeFn: func(c *exec.Cmd) ([]byte, error) {
			// 192.168.231.222	satlab-0wgtfqin1846803b-host12
			if c.Path == paths.DockerPath {
				return []byte(`
192.168.231.222	satlab-0wgtfqin1846803b-host12
        `), nil
			} else if c.Path == paths.GetHostIdentifierScript {
				return []byte("0wgtfqin1846803b"), nil
			} else {
				return nil, errors.New(fmt.Sprintf("execute a command %v\n", c.Path))
			}
		},
	}

	addresses := []string{"192.168.231.222"}
	ufs := mockDeleteClient{}

	resp, invalidAddresses, err := innerDeleteDuts(ctx, e, &ufs, addresses, true)
	if err != nil {
		t.Errorf("unexpected error: %v\n", err)
		return
	}

	if len(invalidAddresses) != 0 {
		t.Errorf("invalid addresses should be empty")
	}

	expected := &dut.DeleteDUTResult{
		MachineLSEs: []*ufsModels.MachineLSE{
			{
				Name:     "machineLSEs/satlab-0wgtfqin1846803b-host12",
				Machines: []string{"asset-satlab-0wgtfqin1846803b-host12"},
				Rack:     "rack-satlab-0wgtfqin1846803b-host12",
			},
		},
		DutResults: &dut.Result{
			Pass: []string{"satlab-0wgtfqin1846803b-host12"},
			Fail: []string{},
		},
		AssetResults: &dut.Result{
			Pass: []string{"asset-satlab-0wgtfqin1846803b-host12"},
			Fail: []string{},
		},
		RackResults: &dut.Result{
			Pass: []string{"rack-satlab-0wgtfqin1846803b-host12"},
			Fail: []string{},
		},
	}

	// ignore pb fields in `FirmwareUpdateCommandOutput`
	ignorePBFieldOpts := cmpopts.IgnoreUnexported(ufsModels.MachineLSE{})

	if diff := cmp.Diff(resp, expected, ignorePBFieldOpts); diff != "" {
		t.Errorf("unexpected diff: %v\n", diff)
	}
}

func TestDeleteDutsShouldFail(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create a mock data
	e := &executor.FakeCommander{
		FakeFn: func(c *exec.Cmd) ([]byte, error) {
			return nil, errors.New("fail")
		},
	}

	addresses := []string{"192.168.231.222"}
	ufs := mockDeleteClient{}

	resp, invalidAddresses, err := innerDeleteDuts(ctx, e, &ufs, addresses, true)
	if err == nil {
		t.Errorf("should get an err")
	}

	if resp != nil {
		t.Errorf("result should be empty")
	}

	if len(invalidAddresses) != 0 {
		t.Errorf("invalid addresses should be empty: %v", invalidAddresses)
	}
}

func TestGetNetworkInfoShouldSuccess(t *testing.T) {
	t.Parallel()
	// Create a mock server
	s := createMockServer(t)

	expected := &pb.GetNetworkInfoResponse{
		Hostname:    "127.0.0.1",
		MacAddress:  "aa:bb:cc:dd:ee:ff",
		IsConnected: true,
	}

	s.commandExecutor = &executor.FakeCommander{
		FakeFn: func(in *exec.Cmd) ([]byte, error) {
			cmd := strings.Join(in.Args, " ")
			if cmd == "/usr/local/bin/get_host_ip" {
				return []byte(expected.Hostname), nil
			} else if cmd == "/usr/local/bin/docker exec dhcp cat /sys/class/net/eth0/address" {
				return []byte(expected.MacAddress), nil
			}
			return nil, errors.New(fmt.Sprintf("handle command: %v", in.Path))
		},
		CmdOutput: fmt.Sprintf("%v/24 dev eth0 scope link  src %v", expected.Hostname, expected.Hostname),
	}

	ctx := context.Background()

	req := &pb.GetNetworkInfoRequest{}

	res, err := s.GetNetworkInfo(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	if diff := cmp.Diff(expected, res, cmpopts.IgnoreUnexported(pb.GetNetworkInfoResponse{})); diff != "" {
		t.Errorf("Expected %v, got %v", expected, res)
	}
}
