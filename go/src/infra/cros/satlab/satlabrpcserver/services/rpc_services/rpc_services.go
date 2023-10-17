// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package rpc_services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/hashicorp/go-version"

	"infra/cmd/shivas/utils"
	"infra/cros/satlab/common/asset"
	"infra/cros/satlab/common/dns"
	"infra/cros/satlab/common/dut"
	"infra/cros/satlab/common/run"
	"infra/cros/satlab/common/satlabcommands"
	"infra/cros/satlab/common/services"
	"infra/cros/satlab/common/services/build_service"
	"infra/cros/satlab/common/services/ufs"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/collection"
	e "infra/cros/satlab/common/utils/errors"
	"infra/cros/satlab/common/utils/executor"
	"infra/cros/satlab/common/utils/parser"
	"infra/cros/satlab/satlabrpcserver/platform/cpu_temperature"
	pb "infra/cros/satlab/satlabrpcserver/proto"
	"infra/cros/satlab/satlabrpcserver/services/bucket_services"
	"infra/cros/satlab/satlabrpcserver/services/dut_services"
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
	rawData, err := s.buildService.ListModels(ctx, in.GetBoard())
	if err != nil {
		return nil, err
	}

	data := make(map[string][]string)

	for _, item := range rawData {
		boardAndModelPair, err := parser.ExtractBoardAndModelFrom(item)
		if errors.Is(err, e.NotMatch) {
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

	var remoteBuilds []*build_service.BuildVersion
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
	devices, err := s.dutService.GetConnectedIPs(ctx)
	if err != nil {
		return nil, err
	}

	IPs := []string{}
	for _, d := range devices {
		if d.IsConnected {
			IPs = append(IPs, d.IP)
		}
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
	r := &run.Run{
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

func (s *SatlabRpcServiceServer) RunTest(ctx context.Context, in *pb.RunTestRequest) (*pb.RunTestResponse, error) {
	r := &run.Run{
		Tests:     in.GetTests(),
		TestArgs:  in.GetTestArgs(),
		Board:     in.GetBoard(),
		Model:     in.GetModel(),
		Milestone: in.GetMilestone(),
		Build:     in.GetBuild(),
		Pool:      in.GetPool(),
	}
	buildLink, err := r.TriggerRun(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.RunTestResponse{BuildLink: buildLink}, nil
}

func (s *SatlabRpcServiceServer) GetVersionInfo(ctx context.Context, _ *pb.GetVersionInfoRequest) (*pb.GetVersionInfoResponse, error) {
	resp := pb.GetVersionInfoResponse{}
	hostId, err := satlabcommands.GetDockerHostBoxIdentifier(ctx, s.commandExecutor)
	if err != nil {
		return nil, err
	}
	resp.HostId = hostId
	osVersion, err := satlabcommands.GetOsVersion(ctx, s.commandExecutor)
	if err != nil {
		return nil, err
	}
	resp.Description = osVersion.Description
	resp.ChromeosVersion = osVersion.Version
	resp.Track = osVersion.Track
	version, err := satlabcommands.GetSatlabVersion(ctx, s.commandExecutor)
	if err != nil {
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
	IPToHostResult, err := dns.IPToHostname(ctx, s.commandExecutor, in.GetAddresses())
	if err != nil {
		return nil, err
	}

	for _, hostname := range IPToHostResult.Hostnames {
		if err = addPoolsToDUT(ctx, s.commandExecutor, hostname, []string{in.GetPool()}); err != nil {
			return nil, err
		}
	}

	return &pb.AddPoolResponse{}, nil
}

func removeAllPoolsFromDUT(ctx context.Context, executor executor.IExecCommander, hostname string) error {
	return addPoolsToDUT(ctx, executor, hostname, []string{"-"})
}

func (s *SatlabRpcServiceServer) UpdatePool(ctx context.Context, in *pb.UpdatePoolRequest) (*pb.UpdatePoolResponse, error) {
	IPHostMap, err := dns.ReadHostsToIPMap(ctx, s.commandExecutor)
	if err != nil {
		return nil, err
	}

	for _, item := range in.GetItems() {
		hostname, ok := IPHostMap[item.GetAddress()]
		if ok {
			// According to `shivas` CLI. If we add a pool ("-"). It will remove all pools from the
			// host.
			if err = removeAllPoolsFromDUT(ctx, s.commandExecutor, hostname); err != nil {
				return nil, err
			}

			// After removing the pools, we can add it the pools that we want to keep
			if err = addPoolsToDUT(ctx, s.commandExecutor, hostname, item.GetPools()); err != nil {
				return nil, err
			}
		}
	}

	return &pb.UpdatePoolResponse{}, nil
}

func (s *SatlabRpcServiceServer) GetDutDetail(ctx context.Context, in *pb.GetDutDetailRequest) (*pb.GetDutDetailResponse, error) {
	if s.swarmingService == nil {
		return nil, errors.New("need to login before using this")
	}

	IPToHostResult, err := dns.IPToHostname(ctx, s.commandExecutor, []string{in.GetAddress()})
	if err != nil {
		return nil, err
	}

	if len(IPToHostResult.InvalidAddresses) != 0 {
		return nil, errors.New(fmt.Sprintf("can't find the host by ip address {%v}", IPToHostResult.InvalidAddresses))
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
	if s.swarmingService == nil {
		return nil, errors.New("need to login before using this")
	}

	IPToHostResult, err := dns.IPToHostname(ctx, s.commandExecutor, []string{in.GetAddress()})
	if err != nil {
		return nil, err
	}

	if len(IPToHostResult.InvalidAddresses) != 0 {
		return nil, errors.New(fmt.Sprintf("can't find the host by ip address {%v}", IPToHostResult.InvalidAddresses))
	}

	r, err := s.swarmingService.ListBotTasks(ctx, IPToHostResult.Hostnames[0], in.GetCursor(), int(in.GetPageSize()))
	if err != nil {
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
		Cursor: r.Cursor,
		Tasks:  tasks,
	}, nil
}

func (s *SatlabRpcServiceServer) ListDutEvents(ctx context.Context, in *pb.ListDutEventsRequest) (*pb.ListDutEventsResponse, error) {
	if s.swarmingService == nil {
		return nil, errors.New("need to login before using this")
	}

	IPToHostResult, err := dns.IPToHostname(ctx, s.commandExecutor, []string{in.GetAddress()})
	if err != nil {
		return nil, err
	}

	if len(IPToHostResult.InvalidAddresses) != 0 {
		return nil, errors.New(fmt.Sprintf("can't find the host by ip address {%v}", IPToHostResult.InvalidAddresses))
	}

	r, err := s.swarmingService.ListBotEvents(ctx, IPToHostResult.Hostnames[0], in.GetCursor(), int(in.GetPageSize()))
	if err != nil {
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
		Cursor: r.Cursor,
		Events: events,
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
			Name:     dut.Name,
			Hostname: dut.Hostname,
			Pools:    dut.GetChromeosMachineLse().GetDeviceLse().GetDut().Pools,
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
	duts, err := getConnectedDuts(ctx, s.commandExecutor)
	if err != nil {
		return nil, err
	}

	return &pb.ListEnrolledDutsResponse{Duts: duts}, nil
}

func (s *SatlabRpcServiceServer) ListDuts(ctx context.Context, in *pb.ListDutsRequest) (*pb.ListDutsResponse, error) {
	connectedDevices, err := s.dutService.GetConnectedIPs(ctx)
	if err != nil {
		return nil, err
	}

	duts, err := getConnectedDuts(ctx, s.commandExecutor)
	if err != nil {
		return nil, err
	}

	enrolledIPs := []string{}

	for _, dut := range duts {
		for _, device := range connectedDevices {
			if dut.Address == device.IP {
				dut.IsConnected = device.IsConnected
				dut.MacAddress = device.MACAddress
				enrolledIPs = append(enrolledIPs, dut.Address)
			}
		}
	}

	unenrolledDevices := collection.Subtract(connectedDevices, enrolledIPs, func(a dut_services.Device, b string) bool {
		return a.IP == b
	})

	for _, device := range unenrolledDevices {
		duts = append(duts, &pb.Dut{
			Address:     device.IP,
			MacAddress:  device.MACAddress,
			IsConnected: device.IsConnected,
		})
	}

	return &pb.ListDutsResponse{Duts: duts}, nil
}

// DeleteDuts the RPC service for deleting DUTs
func (s *SatlabRpcServiceServer) DeleteDuts(ctx context.Context, in *pb.DeleteDutsRequest) (*pb.DeleteDutsResponse, error) {
	ctx = utils.SetupContext(ctx, site.GetNamespace(""))
	ufs, err := ufs.NewUFSClientWithDefaultOptions(ctx, site.GetUFSService(s.dev))
	if err != nil {
		return nil, err
	}

	res, invalidAddresses, err := innerDeleteDuts(ctx, s.commandExecutor, ufs, in.GetAddresses(), false)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteDutsResponse{
		Pass:             res.DutResults.Pass,
		Fail:             res.DutResults.Fail,
		InvalidAddresses: invalidAddresses,
	}, nil
}

// innerDeleteDuts the main logic of deleting the DUTs by given IP addresses.
// Create this function for testing easily
// This function returns a result of deleting DUTs result that contains pass and fail,
// and if we can not convert the IP address, we put the IP address to `invalidAddresses`
func innerDeleteDuts(ctx context.Context, executor executor.IExecCommander, ufs dut.DeleteClient, addresses []string, full bool) (*dut.DeleteDUTResult, []string, error) {
	IPToHostResult, err := dns.IPToHostname(ctx, executor, addresses)
	if err != nil {
		return nil, nil, err
	}

	d := dut.DeleteDUT{
		Names: IPToHostResult.Hostnames,
		Full:  full,
	}

	if err := d.Validate(); err != nil {
		return nil, IPToHostResult.InvalidAddresses, err
	}

	res, err := d.TriggerRun(ctx, executor, ufs)

	return res, IPToHostResult.InvalidAddresses, nil
}
