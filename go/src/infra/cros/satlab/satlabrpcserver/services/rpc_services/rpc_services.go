// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package rpc_services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"go.chromium.org/luci/common/logging"

	pb "go.chromium.org/chromiumos/infra/proto/go/satlabrpcserver"
	"infra/cmd/shivas/utils"
	"infra/cros/dutstate"
	"infra/cros/satlab/common/asset"
	"infra/cros/satlab/common/dns"
	"infra/cros/satlab/common/dut"
	"infra/cros/satlab/common/dut/shivas"
	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/run"
	"infra/cros/satlab/common/satlabcommands"
	"infra/cros/satlab/common/services"
	"infra/cros/satlab/common/services/build_service"
	"infra/cros/satlab/common/services/ufs"
	"infra/cros/satlab/common/setup"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/collection"
	e "infra/cros/satlab/common/utils/errors"
	"infra/cros/satlab/common/utils/executor"
	"infra/cros/satlab/common/utils/parser"
	"infra/cros/satlab/satlabrpcserver/platform/cpu_temperature"
	"infra/cros/satlab/satlabrpcserver/services/bucket_services"
	"infra/cros/satlab/satlabrpcserver/services/dut_services"
	u "infra/cros/satlab/satlabrpcserver/utils"
	"infra/cros/satlab/satlabrpcserver/utils/constants"
)

// SatlabRpcServiceServer is the gRPC service that provides every function.
type SatlabRpcServiceServer struct {
	pb.UnimplementedSatlabRpcServiceServer
	// dev is a flag indicate which environment we want to run
	dev bool
	// buildService the connector to `BuildClient`
	buildService build_service.IBuildService
	// bucketService the connector to partner bucket
	bucketService bucket_services.IBucketServices
	// dutService the service to connect to DUTs
	dutService dut_services.IDUTServices
	// cpuTemperatureOrchestrator the CPU temperature orchestrator
	cpuTemperatureOrchestrator *cpu_temperature.CPUTemperatureOrchestrator
	// commandExecutor provides an interface to run a command. It is good for testing
	commandExecutor executor.IExecCommander
	// swarmingService provides the swarming API services
	swarmingService services.ISwarmingService
}

func New(
	dev bool,
	buildService build_service.IBuildService,
	bucketService bucket_services.IBucketServices,
	dutService dut_services.IDUTServices,
	cpuTemperatureOrchestrator *cpu_temperature.CPUTemperatureOrchestrator,
	swarmingService services.ISwarmingService,
) *SatlabRpcServiceServer {
	return &SatlabRpcServiceServer{
		dev:                        dev,
		bucketService:              bucketService,
		buildService:               buildService,
		dutService:                 dutService,
		cpuTemperatureOrchestrator: cpuTemperatureOrchestrator,
		commandExecutor:            &executor.ExecCommander{},
		swarmingService:            swarmingService,
	}
}

func (s *SatlabRpcServiceServer) validateServices() error {
	if s.buildService == nil || s.bucketService == nil || s.swarmingService == nil {
		return errors.New("authentication failed, please configure the credentials.")
	}

	return nil
}

// ListBuildTargets the gRPC server entry point to list all the build targets.
//
// ListBuildTargetsRequest _ we don't need use any parameter from the request, but we need to
// define it as a parameter to satisfy the compiler.
// To see more, we can look at the `src/satlab_rpcserver/satlabrpc.proto`
func (s *SatlabRpcServiceServer) ListBuildTargets(ctx context.Context, _ *pb.ListBuildTargetsRequest) (*pb.ListBuildTargetsResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: list_build_targets")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	res, err := s.buildService.ListBuildTargets(ctx)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_build_targets: %w", err)
		return nil, err
	}

	return &pb.ListBuildTargetsResponse{
		BuildTargets: res,
	}, nil
}

// ListMilestones the gRPC server entry point to list all milestones from GCS bucket.
// TODO Add a cache for listing milestones
//
// pb.ListMilestonesRequest in the request from the client we use it as a filter to list the milestones.
func (s *SatlabRpcServiceServer) ListMilestones(ctx context.Context, in *pb.ListMilestonesRequest) (*pb.ListMilestonesResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: list_milestones")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	// Get the milestones from the partner bucket
	// If the milestones are in the partner bucket. they are staged.
	bucketMilestones, err := s.bucketService.GetMilestones(ctx, in.GetBoard())
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_milestones: %w", err)
		return nil, err
	}

	var remoteMilestones []string
	// Check the bucket is in asia, if it isn't in asia, we can fetch the milestones from `BuildClient`
	isBucketInAsia, err := s.bucketService.IsBucketInAsia(ctx)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_milestones: %w", err)
		return nil, err
	}

	if !isBucketInAsia {
		remoteMilestones, err = s.buildService.ListAvailableMilestones(ctx, in.GetBoard(), in.GetModel())
		if err != nil {
			logging.Errorf(ctx, "gRPC Service error: list_milestones: %w", err)
			return nil, err
		}
	}

	var res []*pb.BuildItem

	// Map bucketMilestones to response type `BuildItem`
	for _, item := range bucketMilestones {
		res = append(res, &pb.BuildItem{
			Value:    item,
			IsStaged: true,
			Status:   pb.BuildItem_BUILD_STATUS_PASS,
		})
	}

	// Filter the remoteMilestones not in the bucketMilestones,
	// and then mapping the milestones to response type `BuildItem`
	for _, item := range collection.Subtract(remoteMilestones, bucketMilestones, func(a string, b string) bool {
		return a == b
	}) {
		res = append(res, &pb.BuildItem{
			Value:    item,
			IsStaged: false,
			Status:   pb.BuildItem_BUILD_STATUS_PASS,
		})
	}

	// Sort the result
	sort.SliceStable(res, func(i, j int) bool {
		mA, errA := strconv.Atoi(res[i].Value)
		mB, errB := strconv.Atoi(res[j].Value)
		if errA != nil || errB != nil {
			return res[i].Value > res[j].Value
		}
		return mA > mB
	})

	return &pb.ListMilestonesResponse{
		Milestones: res,
	}, nil
}

// ListAccessibleModels the gRPC server entry point to list all models for a given board
//
// pb.ListAccessibleModelsRequest in the request from the client we use it as a filter to list the models.
func (s *SatlabRpcServiceServer) ListAccessibleModels(ctx context.Context, in *pb.ListAccessibleModelsRequest) (*pb.ListAccessibleModelsResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: list_accessible_models")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	rawData, err := s.buildService.ListModels(ctx, in.GetBoard())
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_accessible_models: %w", err)
		return nil, err
	}

	data := make(map[string][]string)

	for _, item := range rawData {
		boardAndModelPair, err := parser.ExtractBoardAndModelFrom(item)
		if errors.Is(err, e.NotMatch) {
			logging.Warningf(ctx, "The model name (%s) doesn't match `buildTargets/{board}/models/{model}`", item)
		} else {
			data[boardAndModelPair.Model] = append(data[boardAndModelPair.Model], boardAndModelPair.Board)
		}
	}

	var res []*pb.Model

	for key, value := range data {
		res = append(res, &pb.Model{
			Name:   key,
			Boards: value,
		})
	}

	return &pb.ListAccessibleModelsResponse{
		Models: res,
	}, nil
}

// ListBuildVersions the gRPC server entry point to list all build versions for given board, model, and milestone.
// TODO Add a cache for listing build versions
//
// pb.ListBuildVersionsRequest in the request from the client we use to it as a filter to list the build versions.
func (s *SatlabRpcServiceServer) ListBuildVersions(ctx context.Context, in *pb.ListBuildVersionsRequest) (*pb.ListBuildVersionsResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: list_build_versions")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	// Get the builds from the partner bucket
	// If the builds are in the partner bucket. they are staged.
	bucketBuilds, err := s.bucketService.GetBuilds(ctx, in.GetBoard(), in.GetMilestone())
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_build_versions: %w", err)
		return nil, err
	}

	var remoteBuilds []*build_service.BuildVersion
	// Check the bucket is in asia, if it isn't in asia, we can fetch the builds from `BuildClient`
	isBucketInAsia, err := s.bucketService.IsBucketInAsia(ctx)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_build_versions: %w", err)
		return nil, err
	}

	if !isBucketInAsia {
		remoteBuilds, err = s.buildService.ListBuildsForMilestone(ctx, in.GetBoard(), in.GetModel(), in.GetMilestone())
		if err != nil {
			logging.Errorf(ctx, "gRPC Service error: list_build_versions: %w", err)
			return nil, err
		}
	}

	var res []*pb.BuildItem

	// Map the bucketBuilds to response type `BuildItem`
	for _, item := range bucketBuilds {
		res = append(res, &pb.BuildItem{
			Status:   pb.BuildItem_BUILD_STATUS_PASS,
			IsStaged: true,
			Value:    item,
		})
	}

	// Filter the remoteBuilds not in the bucketBuilds,
	// and then mapping the remoteBuilds to response type `BuildItem`
	for _, build := range collection.Subtract(remoteBuilds, bucketBuilds, func(a *build_service.BuildVersion, b string) bool {
		return a.Version == b
	}) {
		res = append(res, &pb.BuildItem{
			Value:    build.Version,
			IsStaged: false,
			Status:   constants.ToResponseBuildStatusMap[build.Status],
		})
	}

	// Sort the result
	sort.SliceStable(res, func(i, j int) bool {
		mA, errA := version.NewVersion(res[i].Value)
		mB, errB := version.NewVersion(res[j].Value)
		if errA != nil || errB != nil {
			return res[i].Value > res[j].Value
		}
		return mA.GreaterThanOrEqual(mB)
	})

	return &pb.ListBuildVersionsResponse{
		BuildVersions: res,
	}, nil
}

// StageBuild stage a build version in bucket.
//
// pb.StageBuildRequest in the request from client which we want to stage the artifact in the partner bucket.
func (s *SatlabRpcServiceServer) StageBuild(ctx context.Context, in *pb.StageBuildRequest) (*pb.StageBuildResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: stage_build")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	res, err := s.buildService.StageBuild(ctx, in.GetBoard(), in.GetModel(), in.GetBuildVersion(), site.GetGCSImageBucket())
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: stage_build: %w", err)
		return nil, err
	}

	return &pb.StageBuildResponse{
		BuildBucket: res.GetBucket(),
	}, nil

}

// ListConnectedDutsFirmware get current and firmware update on each DUT
func (s *SatlabRpcServiceServer) ListConnectedDutsFirmware(ctx context.Context, _ *pb.ListConnectedDutsFirmwareRequest) (*pb.ListConnectedDutsFirmwareResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: list_connected_duts_firmware")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	devices, err := s.dutService.GetConnectedIPs(ctx)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_connected_duts_firmware: %w", err)
		return nil, err
	}

	IPs := []string{}
	for _, d := range devices {
		if d.IsPingable && d.HasTestImage {
			IPs = append(IPs, d.IP)
		}
	}

	res := s.dutService.RunCommandOnIPs(ctx, IPs, constants.ListFirmwareCommand)

	var DUTsResponse []*pb.ConnectedDutFirmwareInfo

	for _, cmdRes := range res {
		if cmdRes.Error != nil {
			// If we execute the command failed, we can just continue others. Don't block.
			logging.Errorf(ctx, "Got an error when execute command: %v", cmdRes.Error)
			continue
		}
		var cmdResponse dut_services.ListFirmwareCommandResponse
		err = json.Unmarshal([]byte(cmdRes.Value), &cmdResponse)
		if err != nil {
			// If something wrong, we can continue to decode another ip result.
			logging.Errorf(ctx, "Json decode error: %v", err)
			continue
		}

		model := cmdResponse.Model
		currentFirmware := cmdResponse.FwId
		updateFirmware := "null"
		if _, ok := cmdResponse.FwUpdate[model]; ok {
			updateFirmware = cmdResponse.FwUpdate[model].Host.Versions.RW
		}
		DUTsResponse = append(DUTsResponse, &pb.ConnectedDutFirmwareInfo{
			Ip: cmdRes.IP, CurrentFirmware: currentFirmware, UpdateFirmware: updateFirmware,
		})
	}

	return &pb.ListConnectedDutsFirmwareResponse{Duts: DUTsResponse}, nil
}

// GetSystemInfo get the system information
func (s *SatlabRpcServiceServer) GetSystemInfo(ctx context.Context, _ *pb.GetSystemInfoRequest) (*pb.GetSystemInfoResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: get_system_info")

	var averageTemperature float32 = -1.0
	if s.cpuTemperatureOrchestrator == nil {
		logging.Errorf(ctx, "This platform doesn't support getting the temperature")
	} else {
		averageTemperature = s.cpuTemperatureOrchestrator.GetAverageCPUTemperature()
	}

	startTime, err := satlabcommands.GetSatlabStartTime(ctx, s.commandExecutor)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: get_system_info: %w", err)
		return nil, err
	}

	return &pb.GetSystemInfoResponse{
		CpuTemperature: averageTemperature,
		StartTime:      startTime,
	}, nil
}

// GetPeripheralInformation get peripheral inforamtion by given DUT IP.
func (s *SatlabRpcServiceServer) GetPeripheralInformation(ctx context.Context, in *pb.GetPeripheralInformationRequest) (*pb.GetPeripheralInformationResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: get_peripheral_information")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	res, err := s.dutService.RunCommandOnIP(ctx, in.GetDutHostname(), constants.GetPeripheralInfoCommand)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: get_peripheral_information: %w", err)
		return nil, err
	}

	if res.Error != nil {
		logging.Errorf(ctx, "gRPC Service error: get_peripheral_information: %w", res.Error)
		return nil, res.Error
	}

	return &pb.GetPeripheralInformationResponse{
		JsonInfo: res.Value,
	}, nil
}

// UpdateDutsFirmware update Duts by given IPs
func (s *SatlabRpcServiceServer) UpdateDutsFirmware(ctx context.Context, in *pb.UpdateDutsFirmwareRequest) (*pb.UpdateDutsFirmwareResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: update_duts_firmware")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	// Run command on given IPs
	rawData := s.dutService.RunCommandOnIPs(ctx, in.GetIps(), constants.UpdateFirmwareCommand)

	// Create a response variable
	var resp = make([]*pb.FirmwareUpdateCommandOutput, len(rawData))

	// Loop over the raw data and then map to `FirmwareUpdateCommandOutput`
	for idx, cmdResp := range rawData {
		// Create a `FirmwareUpdateCommandOutput` object.
		out := &pb.FirmwareUpdateCommandOutput{
			Ip: cmdResp.IP,
		}
		// If the cmd response is an error,
		// we can show the error message to user.
		// Otherwise, we show the command output
		if cmdResp.Error != nil {
			out.CommandOutput = cmdResp.Error.Error()
		} else {
			out.CommandOutput = cmdResp.Value
		}

		resp[idx] = out
	}

	// Response the result to client
	return &pb.UpdateDutsFirmwareResponse{Outputs: resp}, nil
}

// Close clean up
func (s *SatlabRpcServiceServer) Close(ctx context.Context) {
	if err := s.buildService.Close(); err != nil {
		logging.Errorf(ctx, "Error while closing buildservice %v", err)
	}
}

// parseDims parse the dims input.
func parseDims(data []*pb.Dim) map[string]string {
	dims := map[string]string{}
	for _, d := range data {
		dims[d.GetKey()] = d.GetValue()
	}

	return dims
}

// Run suite triggers the test suite on the satlab. Right now, this is implemented using CTPBuildRequest
func (s *SatlabRpcServiceServer) RunSuite(ctx context.Context, in *pb.RunSuiteRequest) (*pb.RunSuiteResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: run_suite")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	r := &run.Run{
		Suite:      in.GetSuite(),
		Model:      in.GetModel(),
		Board:      in.GetBuildTarget(),
		Milestone:  in.GetMilestone(),
		Build:      in.GetBuildVersion(),
		Pool:       in.GetPool(),
		AddedDims:  parseDims(in.GetDims()),
		MaxTimeout: true,
	}
	buildLink, err := r.TriggerRun(ctx)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: run_suite: %w", err)
		return nil, err
	}
	return &pb.RunSuiteResponse{BuildLink: buildLink}, nil
}

func (s *SatlabRpcServiceServer) RunTest(ctx context.Context, in *pb.RunTestRequest) (*pb.RunTestResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: run_test")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	r := &run.Run{
		Tests:      in.GetTests(),
		TestArgs:   in.GetTestArgs(),
		Board:      in.GetBoard(),
		Model:      in.GetModel(),
		Milestone:  in.GetMilestone(),
		Build:      in.GetBuild(),
		Pool:       in.GetPool(),
		AddedDims:  parseDims(in.GetDims()),
		MaxTimeout: true,
	}
	buildLink, err := r.TriggerRun(ctx)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: run_test: %w", err)
		return nil, err
	}
	return &pb.RunTestResponse{BuildLink: buildLink}, nil
}

func (s *SatlabRpcServiceServer) GetVersionInfo(ctx context.Context, _ *pb.GetVersionInfoRequest) (*pb.GetVersionInfoResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: get_version_info")

	resp := pb.GetVersionInfoResponse{}
	hostId, err := satlabcommands.GetDockerHostBoxIdentifier(ctx, s.commandExecutor)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: get_version_info: %w", err)
		return nil, err
	}
	resp.HostId = hostId
	osVersion, err := satlabcommands.GetOsVersion(ctx, s.commandExecutor)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: get_version_info: %w", err)
		return nil, err
	}
	resp.Description = osVersion.Description
	resp.ChromeosVersion = osVersion.Version
	resp.Track = osVersion.Track
	version, err := satlabcommands.GetSatlabVersion(ctx, s.commandExecutor)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: get_version_info: %w", err)
		return nil, err
	}
	resp.Version = version
	return &resp, nil
}

func addPoolsToDUT(ctx context.Context, executor executor.IExecCommander, hostname string, pools []string) error {
	req := dut.UpdateDUT{
		Pools:    pools,
		Hostname: hostname,
	}
	return req.TriggerRun(ctx, executor)
}

func (s *SatlabRpcServiceServer) AddPool(ctx context.Context, in *pb.AddPoolRequest) (*pb.AddPoolResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: add_pool")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	IPToHostResult, err := dns.IPToHostname(ctx, s.commandExecutor, in.GetAddresses())
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: add_pool: %w", err)
		return nil, err
	}

	for _, hostname := range IPToHostResult.Hostnames {
		if err = addPoolsToDUT(ctx, s.commandExecutor, hostname, []string{in.GetPool()}); err != nil {
			logging.Errorf(ctx, "gRPC Service error: add_pool: %w", err)
			return nil, err
		}
	}

	return &pb.AddPoolResponse{}, nil
}

func removeAllPoolsFromDUT(ctx context.Context, executor executor.IExecCommander, hostname string) error {
	return addPoolsToDUT(ctx, executor, hostname, []string{"-"})
}

// validateUpdatePools validate remove pools from UI
// that shouldn't be remove all pools from a DUT
func validateUpdatePools(pools []string) bool {
	return len(pools) >= 1
}

func (s *SatlabRpcServiceServer) UpdatePool(ctx context.Context, in *pb.UpdatePoolRequest) (*pb.UpdatePoolResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: update_pool")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	IPHostMap, err := dns.ReadHostsToIPMap(ctx, s.commandExecutor)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: update_pool: %w", err)
		return nil, err
	}

	for _, item := range in.GetItems() {
		hostname, ok := IPHostMap[item.GetAddress()]
		if ok && validateUpdatePools(item.GetPools()) {
			// According to `shivas` CLI. If we add a pool ("-"). It will remove all pools from the
			// host.
			if err = removeAllPoolsFromDUT(ctx, s.commandExecutor, hostname); err != nil {
				logging.Errorf(ctx, "gRPC Service error: update_pool: %w", err)
				return nil, err
			}

			// After removing the pools, we can add it the pools that we want to keep
			if err = addPoolsToDUT(ctx, s.commandExecutor, hostname, item.GetPools()); err != nil {
				logging.Errorf(ctx, "gRPC Service error: update_pool: %w", err)
				return nil, err
			}
		}
	}

	return &pb.UpdatePoolResponse{}, nil
}

func (s *SatlabRpcServiceServer) GetDutDetail(ctx context.Context, in *pb.GetDutDetailRequest) (*pb.GetDutDetailResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: get_dut_detail")

	if err := s.validateServices(); err != nil {
		logging.Errorf(ctx, "gRPC Service error: get_dut_detail: need to login before using this")
		return nil, err
	}

	IPToHostResult, err := dns.IPToHostname(ctx, s.commandExecutor, []string{in.GetAddress()})
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: get_dut_detail: %w", err)
		return nil, err
	}

	if len(IPToHostResult.InvalidAddresses) != 0 {
		logging.Errorf(ctx, "gRPC Service error: get_dut_detail: can't find the host by ip address {%v}", IPToHostResult.InvalidAddresses)
		return nil, fmt.Errorf("can't find the host by ip address {%v}", IPToHostResult.InvalidAddresses)
	}

	r, err := s.swarmingService.GetBot(ctx, IPToHostResult.Hostnames[0])
	if err != nil {
		return nil, err
	}

	dimensions := []*pb.StringListPair{}

	for _, d := range r.GetDimensions() {
		dimensions = append(dimensions, &pb.StringListPair{
			Key:    d.GetKey(),
			Values: d.GetValue(),
		})
	}

	resp := pb.GetDutDetailResponse{
		BotId:           r.GetBotId(),
		TaskId:          r.GetTaskId(),
		ExternalIp:      r.GetExternalIp(),
		AuthenticatedAs: r.GetAuthenticatedAs(),
		FirstSeenTs:     r.GetFirstSeenTs(),
		IsDead:          r.GetIsDead(),
		LastSeenTs:      r.GetLastSeenTs(),
		Quarantined:     r.GetQuarantined(),
		MaintenanceMsg:  r.GetMaintenanceMsg(),
		TaskName:        r.GetTaskName(),
		Version:         r.GetVersion(),
		Dimensions:      dimensions,
	}

	return &resp, nil
}

func (s *SatlabRpcServiceServer) ListDutTasks(ctx context.Context, in *pb.ListDutTasksRequest) (*pb.ListDutTasksResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: list_dut_tasks")

	if err := s.validateServices(); err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_dut_tasks: need to login before using this")
		return nil, err
	}

	IPToHostResult, err := dns.IPToHostname(ctx, s.commandExecutor, []string{in.GetAddress()})
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_dut_tasks: %w", err)
		return nil, err
	}

	if len(IPToHostResult.InvalidAddresses) != 0 {
		logging.Errorf(ctx, "gRPC Service error: list_dut_tasks: can't find the host by ip address {%v}", IPToHostResult.InvalidAddresses)
		return nil, fmt.Errorf("can't find the host by ip address {%v}", IPToHostResult.InvalidAddresses)
	}

	r, err := s.swarmingService.ListBotTasks(ctx, IPToHostResult.Hostnames[0], in.GetPageToken(), int(in.GetPageSize()))
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_dut_tasks: %w", err)
		return nil, err
	}

	tasks := []*pb.Task{}

	for _, t := range r.Tasks {
		tasks = append(tasks, &pb.Task{
			Id:        t.Id,
			Name:      t.Name,
			StartAt:   t.StartAt,
			Duration:  t.Duration,
			Url:       t.Url,
			IsSuccess: t.IsSuccess,
		})
	}

	return &pb.ListDutTasksResponse{
		NextPageToken: r.Cursor,
		Tasks:         tasks,
	}, nil
}

func (s *SatlabRpcServiceServer) ListDutEvents(ctx context.Context, in *pb.ListDutEventsRequest) (*pb.ListDutEventsResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: list_dut_events")

	if err := s.validateServices(); err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_dut_events: need to login before using this")
		return nil, err
	}

	IPToHostResult, err := dns.IPToHostname(ctx, s.commandExecutor, []string{in.GetAddress()})
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_dut_events: %w", err)
		return nil, err
	}

	if len(IPToHostResult.InvalidAddresses) != 0 {
		logging.Errorf(ctx, "gRPC Service error: list_dut_events: can't find the host by ip address {%v}", IPToHostResult.InvalidAddresses)
		return nil, fmt.Errorf("can't find the host by ip address {%v}", IPToHostResult.InvalidAddresses)
	}

	r, err := s.swarmingService.ListBotEvents(ctx, IPToHostResult.Hostnames[0], in.GetPageToken(), int(in.GetPageSize()))
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_dut_events: %w", err)
		return nil, err
	}

	events := []*pb.BotEvent{}
	for _, e := range r.Events {
		events = append(events, &pb.BotEvent{
			Msg:       e.Message,
			EventType: e.Type,
			CreatedAt: e.Ts,
			TaskId:    e.TaskID,
			TaskLink:  e.TaskLink,
			Version:   e.Version,
		})
	}

	return &pb.ListDutEventsResponse{
		NextPageToken: r.Cursor,
		Events:        events,
	}, nil
}

func getConnectedDuts(ctx context.Context, executor executor.IExecCommander) ([]*pb.Dut, error) {
	satlabID, err := satlabcommands.GetDockerHostBoxIdentifier(ctx, executor)
	if err != nil {
		return nil, err
	}
	// Use rack and satlab id to filter
	satlabRackFilter := []string{site.MaybePrepend(site.Satlab, satlabID, "rack")}
	d := dut.GetDUT{
		Racks: satlabRackFilter,
	}
	a := asset.GetAsset{
		Racks: satlabRackFilter,
	}

	HostMap, err := dns.ReadHostsToHostMap(ctx, executor)
	if err != nil {
		return nil, err
	}

	duts, err := d.TriggerRun(ctx, executor, []string{})
	if err != nil {
		return nil, err
	}

	assets, err := a.TriggerRun(ctx, executor)
	if err != nil {
		return nil, err
	}

	res := []*pb.Dut{}

	for _, dut := range duts {
		e := &pb.Dut{
			Name:        dut.Name,
			Hostname:    dut.Hostname,
			Pools:       dut.GetChromeosMachineLse().GetDeviceLse().GetDut().Pools,
			ServoSerial: dut.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals().GetServo().GetServoSerial(),
			ServoType:   dut.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals().GetServo().GetServoType(),
			ServoPort:   dut.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals().GetServo().GetServoPort(),
			State:       dutstate.ConvertFromUFSState(dut.GetResourceState()).String(),
		}

		address := HostMap[dut.Hostname]
		e.Address = address

		for _, asset := range assets {
			if len(dut.Machines) > 0 {
				if asset.Name == dut.Machines[0] {
					e.Model = asset.Model
					e.Board = asset.Info.BuildTarget
				}
			}
		}

		res = append(res, e)
	}

	return res, nil
}

func (s *SatlabRpcServiceServer) ListEnrolledDuts(ctx context.Context, in *pb.ListEnrolledDutsRequest) (*pb.ListEnrolledDutsResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: list_enrolled_duts")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	duts, err := getConnectedDuts(ctx, s.commandExecutor)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_enrolled_duts: %w", err)
		return nil, err
	}

	return &pb.ListEnrolledDutsResponse{Duts: duts}, nil
}

func (s *SatlabRpcServiceServer) ListDuts(ctx context.Context, in *pb.ListDutsRequest) (*pb.ListDutsResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: list_duts")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	connectedDevices, err := s.dutService.GetConnectedIPs(ctx)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_duts: %w", err)
		return nil, err
	}

	duts, err := getConnectedDuts(ctx, s.commandExecutor)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_duts: %w", err)
		return nil, err
	}

	// Get the USB device connected to extract Cr50/Ti50 and Servo serials serial numbers
	usbDevices, err := s.dutService.GetUSBDevicePaths(ctx)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_duts: %w", err)
		return nil, err
	}

	enrolledIPs := []string{}

	for _, dut := range duts {
		for _, device := range connectedDevices {
			if dut.Address == device.IP {
				dut.IsPingable = device.IsPingable
				dut.HasTestImage = device.HasTestImage
				dut.MacAddress = device.MACAddress
				enrolledIPs = append(enrolledIPs, dut.Address)
			}
		}
	}

	unenrolledDevices := collection.Subtract(connectedDevices, enrolledIPs, func(a dut_services.Device, b string) bool {
		return a.IP == b
	})

	for _, device := range unenrolledDevices {

		// TODO optimize we don't need to wait for
		// out dut executing command complete to fetch
		// the next dut board and model.
		var servoSerial = ""
		var board = ""
		var model = ""
		if device.IsPingable && device.HasTestImage {
			board, err = s.dutService.GetBoard(ctx, device.IP)
			if err != nil {
				// Skip when we can't get the board from the CLI.
				board = ""
			}
			model, err = s.dutService.GetModel(ctx, device.IP)
			if err != nil {
				// Skip when we can't get the model from the CLI.
				model = ""
			}
			var isServoConnected = false
			isServoConnected, servoSerial, err = s.dutService.GetServoSerial(ctx, device.IP, usbDevices)
			if err != nil {
				logging.Errorf(ctx, "gRPC Service error: list_duts: failed to find servo serial for %s: %v", device.IP, err)
			}
			// TODO Make UI handle this to display appropriate thing instead of setting it here.
			if isServoConnected && servoSerial == "" {
				servoSerial = "NOT DETECTED"
			}
		}
		duts = append(duts, &pb.Dut{
			Board:        board,
			Model:        model,
			Address:      device.IP,
			MacAddress:   device.MACAddress,
			IsPingable:   device.IsPingable,
			HasTestImage: device.HasTestImage,
			ServoSerial:  servoSerial,
		})
	}

	return &pb.ListDutsResponse{Duts: duts}, nil
}

// DeleteDuts the RPC service for deleting DUTs
func (s *SatlabRpcServiceServer) DeleteDuts(ctx context.Context, in *pb.DeleteDutsRequest) (*pb.DeleteDutsResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: delete_duts")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	ctx = utils.SetupContext(ctx, site.GetNamespace(""))
	ufs, err := ufs.NewUFSClientWithDefaultOptions(ctx, site.GetUFSService(s.dev))
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: delete_duts: %w", err)
		return nil, err
	}

	res, err := innerDeleteDuts(ctx, s.commandExecutor, ufs, in.GetHostnames(), false)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: delete_duts: %w", err)
		return nil, err
	}

	return &pb.DeleteDutsResponse{
		Pass: res.DutResults.Pass,
		Fail: res.DutResults.Fail,
	}, nil
}

// innerDeleteDuts the main logic of deleting the DUTs by given IP addresses.
// Create this function for testing easily
// This function returns a result of deleting DUTs result that contains pass and fail.
func innerDeleteDuts(ctx context.Context, executor executor.IExecCommander, ufs dut.DeleteClient, hostnames []string, full bool) (*dut.DeleteDUTResult, error) {
	d := dut.DeleteDUT{
		Names: hostnames,
		Full:  full,
	}

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return d.TriggerRun(ctx, executor, ufs)
}

// GetNetworkInfo gets newwork information of satlab.
func (s *SatlabRpcServiceServer) GetNetworkInfo(ctx context.Context, _ *pb.GetNetworkInfoRequest) (*pb.GetNetworkInfoResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: get_network_info")

	hostname, err := satlabcommands.GetHostIP(ctx, s.commandExecutor)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: get_network_info: %w", err)
		return nil, err
	}
	macAddress, err := satlabcommands.GetMacAddress(ctx, s.commandExecutor)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: get_network_info: %w", err)
		return nil, err
	}

	return &pb.GetNetworkInfoResponse{
		Hostname:    hostname,
		MacAddress:  macAddress,
		IsConnected: hostname != "" && hostname != "localhost",
	}, nil

}

func (s *SatlabRpcServiceServer) ListTestPlans(ctx context.Context, _ *pb.ListTestPlansRequest) (*pb.ListTestPlansResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: list_test_plans")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	res, err := s.bucketService.ListTestplans(ctx)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: list_test_plans: %w", err)
		return nil, err
	}

	return &pb.ListTestPlansResponse{
		Names: res,
	}, nil
}

func (s *SatlabRpcServiceServer) AddDuts(ctx context.Context, in *pb.AddDutsRequest) (*pb.AddDutsResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: add_duts")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	var fail = make([]*pb.AddDutsResponse_FailedData, 0, len(in.GetDuts()))
	var pass = make([]*pb.AddDutsResponse_PassedData, 0, len(in.GetDuts()))

	for _, d := range in.GetDuts() {
		// The buffer we want to get the command output
		// we use this buffer to parse the deploy URL.
		var buf bytes.Buffer
		err := (&dut.AddDUT{
			Hostname:    d.GetHostname(),
			Address:     d.GetAddress(),
			Board:       d.GetBoard(),
			Model:       d.GetModel(),
			AssetType:   "dut",
			Asset:       uuid.NewString(),
			DeployTags:  []string{"satlab:true"},
			ServoSerial: d.GetServoSerial(),
		}).TriggerRun(ctx, s.commandExecutor, &buf)
		if err != nil {
			fail = append(fail, &pb.AddDutsResponse_FailedData{
				Hostname: d.GetHostname(),
				Reason:   err.Error(),
			})
		} else {
			// There are two results in the command output.
			// The first one is a success. Another one has failed.
			out := buf.String()
			rows := strings.Split(out, "\n")
			// If it deployed failed, The output
			// will contain the message `Failed to add DUT...`
			failed := ""
			for i := len(rows) - 1; i >= 0; i-- {
				if strings.HasPrefix(rows[i], "Failed to add DUT") {
					failed = rows[i]
					break
				}
			}

			if failed != "" {
				fail = append(fail, &pb.AddDutsResponse_FailedData{
					Hostname: d.GetHostname(),
					Reason:   failed,
				})
			} else {
				url, err := parser.ParseDeployURL(out)
				if err != nil {
					// Skip parsing error here, we don't want to
					// block user if any dut has been deployed successfully,
					// but we can't parse the url from the command output.
					url = ""
				}
				pass = append(pass, &pb.AddDutsResponse_PassedData{
					Hostname: d.GetHostname(),
					Url:      url,
				})
			}
		}
	}

	return &pb.AddDutsResponse{Pass: pass, Fail: fail}, nil
}

func (s *SatlabRpcServiceServer) RunTestPlan(ctx context.Context, in *pb.RunTestPlanRequest) (*pb.RunTestPlanResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: run_test_plan")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	r := &run.Run{
		Board:      in.GetBoard(),
		Model:      in.GetModel(),
		Milestone:  in.GetMilestone(),
		Build:      in.GetBuild(),
		Pool:       in.GetPool(),
		Testplan:   in.GetTestPlanName(),
		AddedDims:  parseDims(in.GetDims()),
		MaxTimeout: true,
	}

	buildLink, err := r.TriggerRun(ctx)
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: run_test_plan: %w", err)
		return nil, err
	}
	return &pb.RunTestPlanResponse{BuildLink: buildLink}, nil
}

func (s *SatlabRpcServiceServer) GetTestPlan(ctx context.Context, in *pb.GetTestPlanRequest) (*pb.GetTestPlanResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: get_test_plan")

	if err := s.validateServices(); err != nil {
		return nil, err
	}

	tp, err := s.bucketService.GetTestPlan(ctx, in.GetName())
	if err != nil {
		logging.Errorf(ctx, "gRPC Service error: get_test_plan: %w", err)
		return nil, err
	}

	return &pb.GetTestPlanResponse{Plan: tp}, nil
}

func (s *SatlabRpcServiceServer) SetCloudConfiguration(ctx context.Context, in *pb.SetCloudConfigurationRequest) (*pb.SetCloudConfigurationResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: set_cloud_configuration")

	if err := validateCloudConfiguration(in); err != nil {
		logging.Errorf(ctx, "gRPC Service error: set_cloud_configuration: %w", err)
		return nil, err
	}

	bucket := removeGCSBucketPrefixAndSuffix(in.GetGcsBucketUrl())

	r := setup.Setup{
		Bucket:            bucket,
		GSAccessKeyId:     in.GetBotoKeyId(),
		GSSecretAccessKey: in.GetBotoKeySecret(),
	}

	err := r.StartSetup(ctx)
	if err != nil {
		logging.Errorf(ctx, "StartSetup failed. got an error: %v\n", err)
		return nil, err
	}

	return &pb.SetCloudConfigurationResponse{}, nil
}

// removeGCSBucketPrefixAndSuffix remove the gcs bucket url
// e.g.
// gs://bucket/ -> bucket
// gs://bucket  -> bucket
// bucket/      -> bucket
func removeGCSBucketPrefixAndSuffix(bucket string) string {
	s := strings.TrimPrefix(bucket, "gs://")
	s = strings.TrimRight(s, "/")
	return s
}

// validateCloudConfiguration validate the config form the GRPC call.
func validateCloudConfiguration(in *pb.SetCloudConfigurationRequest) error {
	if strings.TrimSpace(in.GetGcsBucketUrl()) == "" {
		return errors.New("bucket is empty")
	}

	if strings.TrimSpace(in.GetBotoKeyId()) == "" {
		return errors.New("boto key is empty")
	}

	if strings.TrimSpace(in.GetBotoKeySecret()) == "" {
		return errors.New("secret key is empty")
	}

	return nil
}

// GetCloudConfiguration get the cloud configuration from env and boto file.
func (s *SatlabRpcServiceServer) GetCloudConfiguration(ctx context.Context, in *pb.GetCloudConfigurationRequest) (*pb.GetCloudConfigurationResponse, error) {
	logging.Infof(ctx, "gRPC Service triggered: get_cloud_configuration")

	bucket := site.GetGCSImageBucket()
	p := site.GetBotoPath()

	f, err := os.Open(p)
	if err != nil {
		// If `boto` file doesn't exist, it means the user
		// doesn't login. we return empty information
		return &pb.GetCloudConfigurationResponse{}, nil
	}
	key := setup.ReadBotoKey(f)

	return &pb.GetCloudConfigurationResponse{
		GcsBucketUrl: bucket,
		BotoKeyId:    key,
	}, nil
}

func innerReboot(executor executor.IExecCommander) {
	// Start the proccess immediately
	// because `reboot` won't return anything.
	// we don't want to wait for the stdout
	cmd := exec.Command(paths.Reboot, "-h", "now")
	executor.CombinedOutput(cmd)
}

// Reboot call a reboot command on RPC container
func (s *SatlabRpcServiceServer) Reboot(context.Context, *pb.RebootRequest) (*pb.RebootResponse, error) {
	// use defer and go rountine to make sure
	// UI receive the response first, and reboot after.
	defer func() {
		go innerReboot(s.commandExecutor)
	}()

	return &pb.RebootResponse{}, nil
}

func (s *SatlabRpcServiceServer) UploadLog(ctx context.Context, _ *pb.UploadLogRequest) (*pb.UploadLogResponse, error) {
	if err := s.validateServices(); err != nil {
		return nil, err
	}

	hostId, err := satlabcommands.GetDockerHostBoxIdentifier(ctx, s.commandExecutor)
	if err != nil {
		return nil, err
	}

	// prepare the tar filename and bucket path
	filename := fmt.Sprintf("%d.tar.gz", time.Now().Unix())
	now := time.Now().Format("2006-01-02_15:04:05")
	out := fmt.Sprintf("/tmp/%s", filename)
	if err := u.TarGz(constants.LogDirectory, out); err != nil {
		return nil, err
	}
	// clean up the tar file
	defer os.Remove(out)

	bPath := fmt.Sprintf("%s/%s/%s", hostId, now, filename)
	// upload the tar file to the bucket
	gsPath, err := s.bucketService.UploadLog(ctx, bPath, out)
	if err != nil {
		return nil, err
	}

	return &pb.UploadLogResponse{
		BucketLink: fmt.Sprintf(constants.GCSObjectURLTemplate, gsPath),
	}, nil
}

func (s *SatlabRpcServiceServer) DownloadLog(_ *pb.DownloadLogRequest, server pb.SatlabRpcService_DownloadLogServer) error {
	filename := fmt.Sprintf("%d.tar.gz", time.Now().Unix())
	out := fmt.Sprintf("/tmp/%s", filename)
	if err := u.TarGz(constants.LogDirectory, out); err != nil {
		return err
	}

	bufSize := 64 * 1024
	f, err := os.Open(out)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := make([]byte, bufSize)
	for {
		c, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		resp := &pb.DownloadLogResponse{
			FileChunk: buf[:c],
		}
		err = server.Send(resp)
		if err != nil {
			return err
		}
	}

	return nil
}

// RepairDuts a GRPC call for repairing the DUTs.
func (s *SatlabRpcServiceServer) RepairDuts(ctx context.Context, in *pb.RepairDutsRequest) (*pb.RepairDutsResponse, error) {
	if err := s.validateServices(); err != nil {
		return nil, err
	}
	satlabID, err := satlabcommands.GetDockerHostBoxIdentifier(ctx, s.commandExecutor)
	if err != nil {
		return nil, err
	}
	action := shivas.Normal
	if in.GetDeep() {
		action = shivas.DeepRepair
	}

	res := []*pb.RepairDutsResponse_RepairResult{}
	for _, h := range in.GetHostnames() {
		r := shivas.DUTRepairer{
			Name:      site.MaybePrepend(site.Satlab, satlabID, h),
			Namespace: site.GetNamespace(""),
			Executor:  s.commandExecutor,
		}
		resp, err := r.Repair(ctx, action)
		buildLink := ""
		taskLink := ""
		if err == nil {
			buildLink = resp.BuildLink
			taskLink = resp.TaskLink
		}
		res = append(res, &pb.RepairDutsResponse_RepairResult{
			Hostname:  h,
			BuildLink: buildLink,
			TaskLink:  taskLink,
			IsSuccess: err == nil,
		})

	}

	return &pb.RepairDutsResponse{Result: res}, nil
}
