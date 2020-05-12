// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	empty "github.com/golang/protobuf/ptypes/empty"
	"go.chromium.org/luci/common/logging"
	luciconfig "go.chromium.org/luci/config"
	"go.chromium.org/luci/config/server/cfgclient/textproto"
	"go.chromium.org/luci/grpc/grpcutil"
	"golang.org/x/net/context"
	status "google.golang.org/genproto/googleapis/rpc/status"

	proto "infra/unifiedfleet/api/v1/proto"
	api "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/model/registration"
	"infra/unifiedfleet/app/util"

	crimsonconfig "go.chromium.org/luci/machine-db/api/config/v1"
	crimson "go.chromium.org/luci/machine-db/api/crimson/v1"
)

// CreateMachine creates machine entry in database.
func (fs *FleetServerImpl) CreateMachine(ctx context.Context, req *api.CreateMachineRequest) (rsp *proto.Machine, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	req.Machine.Name = req.MachineId
	machine, err := registration.CreateMachine(ctx, req.Machine)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	machine.Name = util.AddPrefix(machineCollection, machine.Name)
	return machine, err
}

// UpdateMachine updates the machine information in database.
func (fs *FleetServerImpl) UpdateMachine(ctx context.Context, req *api.UpdateMachineRequest) (rsp *proto.Machine, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	req.Machine.Name = util.RemovePrefix(req.Machine.Name)
	machine, err := registration.UpdateMachine(ctx, req.Machine)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	machine.Name = util.AddPrefix(machineCollection, machine.Name)
	return machine, err
}

// GetMachine gets the machine information from database.
func (fs *FleetServerImpl) GetMachine(ctx context.Context, req *api.GetMachineRequest) (rsp *proto.Machine, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.RemovePrefix(req.Name)
	machine, err := registration.GetMachine(ctx, name)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	machine.Name = util.AddPrefix(machineCollection, machine.Name)
	return machine, err
}

// ListMachines list the machines information from database.
func (fs *FleetServerImpl) ListMachines(ctx context.Context, req *api.ListMachinesRequest) (rsp *api.ListMachinesResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	pageSize := util.GetPageSize(req.PageSize)
	result, nextPageToken, err := registration.ListMachines(ctx, pageSize, req.PageToken)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	for _, machine := range result {
		machine.Name = util.AddPrefix(machineCollection, machine.Name)
	}
	return &api.ListMachinesResponse{
		Machines:      result,
		NextPageToken: nextPageToken,
	}, nil
}

// DeleteMachine deletes the machine from database.
func (fs *FleetServerImpl) DeleteMachine(ctx context.Context, req *api.DeleteMachineRequest) (rsp *empty.Empty, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.RemovePrefix(req.Name)
	err = registration.DeleteMachine(ctx, name)
	return &empty.Empty{}, err
}

// ImportMachines imports the machines from parent sources.
func (fs *FleetServerImpl) ImportMachines(ctx context.Context, req *api.ImportMachinesRequest) (rsp *status.Status, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	source := req.GetMachineDbSource()
	if err := api.ValidateMachineDBSource(source); err != nil {
		return nil, err
	}
	mdbClient, err := fs.newMachineDBInterfaceFactory(ctx, source.GetHost())
	if err != nil {
		return nil, machineDBConnectionFailureStatus.Err()
	}
	logging.Debugf(ctx, "Querying machine-db to get the list of machines")
	resp, err := mdbClient.ListMachines(ctx, &crimson.ListMachinesRequest{})
	if err != nil {
		return nil, machineDBServiceFailureStatus("ListMachines").Err()
	}

	logging.Debugf(ctx, "Importing %d machines", len(resp.Machines))
	pageSize := fs.getImportPageSize()
	machines := util.ToChromeMachines(resp.GetMachines())
	for i := 0; ; i += pageSize {
		end := min(i+pageSize, len(machines))
		logging.Debugf(ctx, "importing %dth - %dth", i, end-1)
		res, err := registration.ImportMachines(ctx, machines[i:end])
		s := processImportDatastoreRes(res, err)
		if s.Err() != nil {
			return s.Proto(), s.Err()
		}
		if i+pageSize >= len(machines) {
			break
		}
	}
	return successStatus.Proto(), nil
}

// CreateRack creates rack entry in database.
func (fs *FleetServerImpl) CreateRack(ctx context.Context, req *api.CreateRackRequest) (rsp *proto.Rack, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	req.Rack.Name = req.RackId
	rack, err := registration.CreateRack(ctx, req.Rack)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	rack.Name = util.AddPrefix(rackCollection, rack.Name)
	return rack, err
}

// UpdateRack updates the rack information in database.
func (fs *FleetServerImpl) UpdateRack(ctx context.Context, req *api.UpdateRackRequest) (rsp *proto.Rack, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	req.Rack.Name = util.RemovePrefix(req.Rack.Name)
	rack, err := registration.UpdateRack(ctx, req.Rack)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	rack.Name = util.AddPrefix(rackCollection, rack.Name)
	return rack, err
}

// GetRack gets the rack information from database.
func (fs *FleetServerImpl) GetRack(ctx context.Context, req *api.GetRackRequest) (rsp *proto.Rack, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.RemovePrefix(req.Name)
	rack, err := registration.GetRack(ctx, name)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	rack.Name = util.AddPrefix(rackCollection, rack.Name)
	return rack, err
}

// ListRacks list the racks information from database.
func (fs *FleetServerImpl) ListRacks(ctx context.Context, req *api.ListRacksRequest) (rsp *api.ListRacksResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	pageSize := util.GetPageSize(req.PageSize)
	result, nextPageToken, err := registration.ListRacks(ctx, pageSize, req.PageToken)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	for _, rack := range result {
		rack.Name = util.AddPrefix(rackCollection, rack.Name)
	}
	return &api.ListRacksResponse{
		Racks:         result,
		NextPageToken: nextPageToken,
	}, nil
}

// DeleteRack deletes the rack from database.
func (fs *FleetServerImpl) DeleteRack(ctx context.Context, req *api.DeleteRackRequest) (rsp *empty.Empty, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.RemovePrefix(req.Name)
	err = registration.DeleteRack(ctx, name)
	return &empty.Empty{}, err
}

// ImportNics imports the nics info in batch.
func (fs *FleetServerImpl) ImportNics(ctx context.Context, req *api.ImportNicsRequest) (response *status.Status, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	source := req.GetMachineDbSource()
	if err := api.ValidateMachineDBSource(source); err != nil {
		return nil, err
	}
	mdbClient, err := fs.newMachineDBInterfaceFactory(ctx, source.GetHost())
	if err != nil {
		return nil, machineDBConnectionFailureStatus.Err()
	}
	logging.Debugf(ctx, "Querying machine-db to get the list of nics")
	resp, err := mdbClient.ListNICs(ctx, &crimson.ListNICsRequest{})
	if err != nil {
		return nil, machineDBServiceFailureStatus("ListMachines").Err()
	}
	logging.Debugf(ctx, "Importing %d nics", len(resp.Nics))
	return successStatus.Proto(), nil
}

// ImportDatacenters imports the datacenter and its related info in batch.
func (fs *FleetServerImpl) ImportDatacenters(ctx context.Context, req *api.ImportDatacentersRequest) (response *status.Status, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	configSource := req.GetConfigSource()
	if err := api.ValidateConfigSource(configSource); err != nil {
		return nil, err
	}
	if configSource.ConfigServiceName == "" {
		return nil, invalidConfigServiceName.Err()
	}

	logging.Debugf(ctx, "Importing datacenters from luci-config: %s", configSource.FileName)
	cfgInterface := fs.newCfgInterface(ctx)
	fetchedConfigs, err := cfgInterface.GetConfig(ctx, luciconfig.ServiceSet(configSource.ConfigServiceName), configSource.FileName, false)
	if err != nil {
		return nil, configServiceFailureStatus.Err()
	}
	dc := &crimsonconfig.Datacenter{}
	logging.Debugf(ctx, "%#v", fetchedConfigs)
	resolver := textproto.Message(dc)
	resolver.Resolve(fetchedConfigs)
	logging.Debugf(ctx, "processing datacenter: %s", dc.GetName())
	racks, kvms, switches, dhcps := processDatacenters(dc)
	logging.Debugf(ctx, "Got %d racks, %d kvms, %d switches, %d dhcp configs", len(racks), len(kvms), len(switches), len(dhcps))
	return successStatus.Proto(), nil
}

func processDatacenters(dc *crimsonconfig.Datacenter) ([]*proto.Rack, []*proto.KVM, []*proto.Switch, []*proto.DHCPConfig) {
	dcName := dc.GetName()
	switches := make([]*proto.Switch, 0)
	racks := make([]*proto.Rack, 0)
	rackToKvms := make(map[string][]string, 0)
	kvms := make([]*proto.KVM, 0)
	dhcps := make([]*proto.DHCPConfig, 0)
	for _, oldKVM := range dc.GetKvm() {
		name := oldKVM.GetName()
		k := &proto.KVM{
			Name:           name,
			MacAddress:     oldKVM.GetMacAddress(),
			ChromePlatform: oldKVM.GetPlatform(),
		}
		kvms = append(kvms, k)
		rackName := oldKVM.GetRack()
		rackToKvms[rackName] = append(rackToKvms[rackName], name)
		dhcps = append(dhcps, &proto.DHCPConfig{
			MacAddress: oldKVM.GetMacAddress(),
			Hostname:   name,
			Ip:         oldKVM.GetIpv4(),
		})
	}
	for _, old := range dc.GetRack() {
		rackName := old.GetName()
		switchNames := make([]string, 0)
		for _, crimsonSwitch := range old.GetSwitch() {
			s := &proto.Switch{
				Name:         crimsonSwitch.GetName(),
				CapacityPort: crimsonSwitch.GetPorts(),
			}
			switches = append(switches, s)
			switchNames = append(switchNames, s.GetName())
		}
		r := &proto.Rack{
			Name:     rackName,
			Location: util.ToLocation(rackName, dcName),
			Rack: &proto.Rack_ChromeBrowserRack{
				ChromeBrowserRack: &proto.ChromeBrowserRack{
					Switches: switchNames,
					Kvms:     rackToKvms[rackName],
				},
			},
		}
		racks = append(racks, r)
	}
	return racks, kvms, switches, dhcps
}
