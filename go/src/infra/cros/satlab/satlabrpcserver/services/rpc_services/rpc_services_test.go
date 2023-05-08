// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package rpc_services

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"
	moblabapipb "google.golang.org/genproto/googleapis/chromeos/moblab/v1beta1"
	mk "infra/cros/satlab/satlabrpcserver/mocks"
	pb "infra/cros/satlab/satlabrpcserver/proto"
	"infra/cros/satlab/satlabrpcserver/services/build_services"
	"infra/cros/satlab/satlabrpcserver/utils"
	"infra/cros/satlab/satlabrpcserver/utils/constants"
)

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
	var mockBuildService = new(mk.MockBuildServices)

	// Create a Mock `IBucketService`
	var mockBucketService = new(mk.MockBucketServices)

	// Create a Mock `IDUTService`
	var mockDUTService = new(mk.MockDUTServices)

	// Create a `LabelParser`
	var labelParser, err = utils.NewLabelParser()
	if err != nil {
		t.Fatalf("Failed to create a label parser %v", err)
	}
	// Create a SATLab Server
	return New(mockBuildService, mockBucketService, mockDUTService, labelParser)
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
	s.buildService.(*mk.MockBuildServices).On("ListBuildTargets", ctx).Return(
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
	s.buildService.(*mk.MockBuildServices).On("ListBuildTargets", ctx).Return(
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
	s.buildService.(*mk.MockBuildServices).On("ListAvailableMilestones", ctx, board, model).Return(
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
	s.buildService.(*mk.MockBuildServices).On("ListAvailableMilestones", ctx, board, model).Return(
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
	s.buildService.(*mk.MockBuildServices).On("ListAvailableMilestones", ctx, board, model).Return(
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
	s.buildService.(*mk.MockBuildServices).On("ListModels", ctx, board).Return(
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
	s.buildService.(*mk.MockBuildServices).On("ListModels", ctx, board).Return(
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

	s.buildService.(*mk.MockBuildServices).
		On("ListBuildsForMilestone", ctx, board, model, milestone).
		Return([]*build_services.BuildVersion{
			{
				Version: "14820.100.0",
				Status:  build_services.FAILED,
			},
			{
				Version: "14820.20.0",
				Status:  build_services.AVAILABLE,
			},
			{
				Version: "14820.8.0",
				Status:  build_services.AVAILABLE,
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

	s.buildService.(*mk.MockBuildServices).
		On("ListBuildsForMilestone", ctx, board, model, milestone).
		Return([]*build_services.BuildVersion{}, expectedErr)

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
	bucketName := "chromeos-moblab-cienet-dev"
	expectedArtifact := &moblabapipb.BuildArtifact{
		Build:  build,
		Name:   "artifacts",
		Bucket: bucketName,
		Path:   "buildTargets/zork/models/dirinboz/builds/1234.0.0/artifacts/chromeos-moblab-cienet-dev",
	}

	s.buildService.(*mk.MockBuildServices).
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
	bucketName := "chromeos-moblab-cienet-dev"
	expectedArtifact := &moblabapipb.BuildArtifact{
		Build:  build,
		Name:   "artifacts",
		Bucket: bucketName,
		Path:   "buildTargets/zork/models/dirinboz/builds/1234.0.0/artifacts/chromeos-moblab-cienet-dev",
	}
	expectedErr := errors.New("can't make a request")

	s.buildService.(*mk.MockBuildServices).
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
	s.dutService.(*mk.MockDUTServices).
		On("RunCommandOnIPs", ctx, mock.Anything, constants.ListFirmwareCommand).
		Return([]*utils.SSHResult{
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
	s.dutService.(*mk.MockDUTServices).
		On("RunCommandOnIPs", ctx, mock.Anything, constants.ListFirmwareCommand).
		Return([]*utils.SSHResult{
			{IP: "192.168.100.1", Error: expectedError},
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
