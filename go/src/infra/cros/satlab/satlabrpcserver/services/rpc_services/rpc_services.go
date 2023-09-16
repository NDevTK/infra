// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package rpc_services

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"strconv"

	"github.com/hashicorp/go-version"

	"infra/cros/satlab/common/dns"
	dut_pkg "infra/cros/satlab/common/dut"
	run_pkg "infra/cros/satlab/common/run"
	"infra/cros/satlab/common/satlabcommands"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/executor"
	"infra/cros/satlab/satlabrpcserver/platform/cpu_temperature"
	pb "infra/cros/satlab/satlabrpcserver/proto"
	"infra/cros/satlab/satlabrpcserver/services/bucket_services"
	"infra/cros/satlab/satlabrpcserver/services/build_services"
	"infra/cros/satlab/satlabrpcserver/services/dut_services"
	"infra/cros/satlab/satlabrpcserver/utils"
	"infra/cros/satlab/satlabrpcserver/utils/constants"
)

// SatlabRpcServiceServer is the gRPC service that provides every function.
type SatlabRpcServiceServer struct {
	pb.UnimplementedSatlabRpcServiceServer
	// buildService the connector to `BuildClient`
	buildService build_services.IBuildServices
	// bucketService the connector to partner bucket
	bucketService bucket_services.IBucketServices
	// dutService the service to connect to DUTs
	dutService dut_services.IDUTServices
	// labelParser the parser for parsing label
	labelParser *utils.LabelParser
	// cpuTemperatureOrchestrator the CPU temperature orchestrator
	cpuTemperatureOrchestrator *cpu_temperature.CPUTemperatureOrchestrator
	// commandExecutor provides an interface to run a command. It is good for testing
	commandExecutor executor.IExecCommander
}

func New(
	buildService build_services.IBuildServices,
	bucketService bucket_services.IBucketServices,
	dutService dut_services.IDUTServices,
	labelParser *utils.LabelParser,
	cpuTemperatureOrchestrator *cpu_temperature.CPUTemperatureOrchestrator,
) *SatlabRpcServiceServer {
	return &SatlabRpcServiceServer{
		bucketService:              bucketService,
		buildService:               buildService,
		dutService:                 dutService,
		labelParser:                labelParser,
		cpuTemperatureOrchestrator: cpuTemperatureOrchestrator,
		commandExecutor:            &executor.ExecCommander{},
	}
}

// ListBuildTargets the gRPC server entry point to list all the build targets.
//
// ListBuildTargetsRequest _ we don't need use any parameter from the request, but we need to
// define it as a parameter to satisfy the compiler.
// To see more, we can look at the `src/satlab_rpcserver/satlabrpc.proto`
func (s *SatlabRpcServiceServer) ListBuildTargets(ctx context.Context, _ *pb.ListBuildTargetsRequest) (*pb.ListBuildTargetsResponse, error) {
	res, err := s.buildService.ListBuildTargets(ctx)
	if err != nil {
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
	// Get the milestones from the partner bucket
	// If the milestones are in the partner bucket. they are staged.
	bucketMilestones, err := s.bucketService.GetMilestones(ctx, in.GetBoard())
	if err != nil {
		return nil, err
	}

	var remoteMilestones []string
	// Check the bucket is in asia, if it isn't in asia, we can fetch the milestones from `BuildClient`
	isBucketInAsia, err := s.bucketService.IsBucketInAsia(ctx)
	if err != nil {
		return nil, err
	}

	if !isBucketInAsia {
		remoteMilestones, err = s.buildService.ListAvailableMilestones(ctx, in.GetBoard(), in.GetModel())
		if err != nil {
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
	for _, item := range utils.Subtract(remoteMilestones, bucketMilestones, func(a string, b string) bool {
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
	rawData, err := s.buildService.ListModels(ctx, in.GetBoard())
	if err != nil {
		return nil, err
	}

	data := make(map[string][]string)

	for _, item := range rawData {
		boardAndModelPair, err := s.labelParser.ExtractBoardAndModelFrom(item)
		if err == utils.NotMatch {
			log.Printf("The model name (%s) doesn't match `buildTargets/{board}/models/{model}`", item)
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
	// Get the builds from the partner bucket
	// If the builds are in the partner bucket. they are staged.
	bucketBuilds, err := s.bucketService.GetBuilds(ctx, in.GetBoard(), in.GetMilestone())
	if err != nil {
		return nil, err
	}

	var remoteBuilds []*build_services.BuildVersion
	// Check the bucket is in asia, if it isn't in asia, we can fetch the builds from `BuildClient`
	isBucketInAsia, err := s.bucketService.IsBucketInAsia(ctx)
	if err != nil {
		return nil, err
	}

	if !isBucketInAsia {
		remoteBuilds, err = s.buildService.ListBuildsForMilestone(ctx, in.GetBoard(), in.GetModel(), in.GetMilestone())
		if err != nil {
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
	for _, build := range utils.Subtract(remoteBuilds, bucketBuilds, func(a *build_services.BuildVersion, b string) bool {
		return a.Version == b
	}) {
		res = append(res, &pb.BuildItem{
			Value:    build.Version,
			IsStaged: false,
			Status:   build_services.ToResponseBuildStatusMap[build.Status],
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
	res, err := s.buildService.StageBuild(ctx, in.GetBoard(), in.GetModel(), in.GetBuildVersion(), site.GetGCSImageBucket())
	if err != nil {
		return nil, err
	}

	return &pb.StageBuildResponse{
		BuildBucket: res.GetBucket(),
	}, nil

}

// ListConnectedDutsFirmware get current and firmware update on each DUT
func (s *SatlabRpcServiceServer) ListConnectedDutsFirmware(ctx context.Context, _ *pb.ListConnectedDutsFirmwareRequest) (*pb.ListConnectedDutsFirmwareResponse, error) {
	IPs, err := utils.GetConnectedDUTIPs()
	if err != nil {
		return nil, err
	}
	res := s.dutService.RunCommandOnIPs(ctx, IPs, constants.ListFirmwareCommand)

	var DUTsResponse []*pb.ConnectedDutFirmwareInfo

	for _, cmdRes := range res {
		if cmdRes.Error != nil {
			// If we execute the command failed, we can just continue others. Don't block.
			log.Printf("Got an error when execute command: %v", cmdRes.Error)
			continue
		}
		var cmdResponse dut_services.ListFirmwareCommandResponse
		err = json.Unmarshal([]byte(cmdRes.Value), &cmdResponse)
		if err != nil {
			// If something wrong, we can continue to decode another ip result.
			log.Printf("Json decode error: %v", err)
			continue
		}

		model := cmdResponse.Model
		currentFirmware := cmdResponse.FwId
		updateFirmware := cmdResponse.FwUpdate[model].Host.Versions.RW
		DUTsResponse = append(DUTsResponse, &pb.ConnectedDutFirmwareInfo{
			Ip: cmdRes.IP, CurrentFirmware: currentFirmware, UpdateFirmware: updateFirmware,
		})
	}

	return &pb.ListConnectedDutsFirmwareResponse{Duts: DUTsResponse}, nil
}

// GetSystemInfo get the system information
func (s *SatlabRpcServiceServer) GetSystemInfo(_ context.Context, _ *pb.GetSystemInfoRequest) (*pb.GetSystemInfoResponse, error) {
	var averageTemperature float32 = -1.0
	if s.cpuTemperatureOrchestrator == nil {
		log.Println("This platform doesn't support getting the temperature")
	} else {
		averageTemperature = s.cpuTemperatureOrchestrator.GetAverageCPUTemperature()
	}

	return &pb.GetSystemInfoResponse{
		CpuTemperature: averageTemperature,
	}, nil
}

// GetPeripheralInformation get peripheral inforamtion by given DUT IP.
func (s *SatlabRpcServiceServer) GetPeripheralInformation(ctx context.Context, in *pb.GetPeripheralInformationRequest) (*pb.GetPeripheralInformationResponse, error) {
	res, err := s.dutService.RunCommandOnIP(ctx, in.GetDutHostname(), constants.GetPeripheralInfoCommand)
	if err != nil {
		return nil, err
	}

	if res.Error != nil {
		return nil, res.Error
	}

	return &pb.GetPeripheralInformationResponse{
		JsonInfo: res.Value,
	}, nil
}

// UpdateDutsFirmware update Duts by given IPs
func (s *SatlabRpcServiceServer) UpdateDutsFirmware(ctx context.Context, in *pb.UpdateDutsFirmwareRequest) (*pb.UpdateDutsFirmwareResponse, error) {
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
func (s *SatlabRpcServiceServer) Close() {
	var err error
	err = s.bucketService.Close()
	if err != nil {
		log.Println(err)
	}
	err = s.buildService.Close()
	if err != nil {
		log.Println(err)
	}
}

// Run suite triggers the test suite on the satlab. Right now, this is implemented using CTPBuildRequest
func (s *SatlabRpcServiceServer) RunSuite(ctx context.Context, in *pb.RunSuiteRequest) (*pb.RunSuiteResponse, error) {
	r := &run_pkg.Run{
		Suite:     in.Suite,
		Model:     in.Model,
		Board:     in.BuildTarget,
		Milestone: in.Milestone,
		Build:     in.BuildVersion,
		Pool:      in.Pool,
	}
	buildLink, err := r.TriggerRun(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.RunSuiteResponse{BuildLink: buildLink}, nil
}

func (s *SatlabRpcServiceServer) GetVersionInfo(ctx context.Context, _ *pb.GetVersionInfoRequest) (*pb.GetVersionInfoResponse, error) {
	resp := pb.GetVersionInfoResponse{}
	hostId, err := satlabcommands.GetDockerHostBoxIdentifier(s.commandExecutor)
	if err != nil {
		return nil, err
	}
	resp.HostId = hostId
	//TODO get the install id
	resp.InstallId = ""
	osVersion, err := satlabcommands.GetOsVersion(s.commandExecutor)
	if err != nil {
		return nil, err
	}
	resp.Description = osVersion.Description
	resp.ChromeosVersion = osVersion.Version
	resp.Track = osVersion.Track
	version, err := satlabcommands.GetSatlabVersion(s.commandExecutor)
	if err != nil {
		return nil, err
	}
	resp.Version = version
	return &resp, nil
}

func (s *SatlabRpcServiceServer) AddPool(ctx context.Context, in *pb.AddPoolRequest) (*pb.AddPoolResponse, error) {
	IPHostMap, err := dns.ReadHostsToIPMap(s.commandExecutor)
	if err != nil {
		return nil, err
	}

	for _, address := range in.GetAddresses() {
		hostname, ok := IPHostMap[address]
		if ok {
			req := dut_pkg.UpdateDUT{
				Pools:    []string{in.GetPool()},
				Hostname: hostname,
			}

			err = req.TriggerRun(ctx, s.commandExecutor)
			if err != nil {
				return nil, err
			}

		}
	}

	return &pb.AddPoolResponse{}, nil
}
