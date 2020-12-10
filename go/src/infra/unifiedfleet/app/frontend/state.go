// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/grpcutil"
	crimson "go.chromium.org/luci/machine-db/api/crimson/v1"
	status "google.golang.org/genproto/googleapis/rpc/status"

	ufspb "infra/unifiedfleet/api/v1/models"
	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	api "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/controller"
	"infra/unifiedfleet/app/external"
)

// ImportStates imports states of crimson objects.
func (fs *FleetServerImpl) ImportStates(ctx context.Context, req *api.ImportStatesRequest) (response *status.Status, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	source := req.GetMachineDbSource()
	if err := api.ValidateMachineDBSource(source); err != nil {
		return nil, err
	}
	es, err := external.GetServerInterface(ctx)
	if err != nil {
		return nil, err
	}
	mdbClient, err := es.NewMachineDBInterfaceFactory(ctx, source.GetHost())
	if err != nil {
		return nil, machineDBConnectionFailureStatus.Err()
	}
	// Skip importing the states of racks, kvms, switches, and vlans, as their states should
	// be referenced by the states of their related racks & machines.
	logging.Debugf(ctx, "Querying machine-db to list the machines")
	machines, err := mdbClient.ListMachines(ctx, &crimson.ListMachinesRequest{})
	if err != nil {
		return nil, machineDBServiceFailureStatus("ListMachines").Err()
	}
	if err := api.ValidateResourceKey(machines.GetMachines(), "Name"); err != nil {
		return nil, errors.Annotate(err, "machines has invalid chars").Err()
	}
	logging.Debugf(ctx, "Querying machine-db to list the vms")
	vms, err := mdbClient.ListVMs(ctx, &crimson.ListVMsRequest{})
	if err != nil {
		return nil, machineDBServiceFailureStatus("ListVMs").Err()
	}
	if err := api.ValidateResourceKey(vms.GetVms(), "Name"); err != nil {
		return nil, errors.Annotate(err, "vms has invalid chars").Err()
	}
	logging.Debugf(ctx, "Querying machine-db to list the vlans")
	vlans, err := mdbClient.ListVLANs(ctx, &crimson.ListVLANsRequest{})
	if err != nil {
		return nil, machineDBServiceFailureStatus("ListVLANs").Err()
	}
	logging.Debugf(ctx, "Querying machine-db to list the kvms")
	kvms, err := mdbClient.ListKVMs(ctx, &crimson.ListKVMsRequest{})
	if err != nil {
		return nil, machineDBServiceFailureStatus("ListKVMs").Err()
	}
	if err := api.ValidateResourceKey(kvms.GetKvms(), "Name"); err != nil {
		return nil, errors.Annotate(err, "kvms has invalid chars").Err()
	}
	logging.Debugf(ctx, "Querying machine-db to list the switches")
	switches, err := mdbClient.ListSwitches(ctx, &crimson.ListSwitchesRequest{})
	if err != nil {
		return nil, machineDBServiceFailureStatus("ListSwitches").Err()
	}
	if err := api.ValidateResourceKey(switches.GetSwitches(), "Name"); err != nil {
		return nil, errors.Annotate(err, "switches has invalid chars").Err()
	}
	logging.Debugf(ctx, "Querying machine-db to list the racks")
	racks, err := mdbClient.ListRacks(ctx, &crimson.ListRacksRequest{})
	if err != nil {
		return nil, machineDBServiceFailureStatus("ListRacks").Err()
	}
	if err := api.ValidateResourceKey(racks.GetRacks(), "Name"); err != nil {
		return nil, errors.Annotate(err, "racks has invalid chars").Err()
	}
	logging.Debugf(ctx, "Querying machine-db to list the hosts")
	hosts, err := mdbClient.ListPhysicalHosts(ctx, &crimson.ListPhysicalHostsRequest{})
	if err != nil {
		return nil, machineDBServiceFailureStatus("ListPhysicalHosts").Err()
	}
	if err := api.ValidateResourceKey(hosts.GetHosts(), "Name"); err != nil {
		return nil, errors.Annotate(err, "hosts has invalid chars").Err()
	}

	pageSize := fs.getImportPageSize()
	res, err := controller.ImportStates(ctx, machines.GetMachines(), racks.GetRacks(), hosts.GetHosts(), vms.GetVms(), vlans.GetVlans(), kvms.GetKvms(), switches.GetSwitches(), pageSize)
	s := processImportDatastoreRes(res, err)
	if s.Err() != nil {
		return s.Proto(), s.Err()
	}
	return successStatus.Proto(), nil
}

// UpdateState updates the state for a resource.
func (fs *FleetServerImpl) UpdateState(ctx context.Context, req *api.UpdateStateRequest) (response *ufspb.StateRecord, err error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	stateRecord, err := controller.UpdateState(ctx, req.State)
	if err != nil {
		return nil, err
	}
	return stateRecord, err
}

// GetState returns the state for a resource.
func (fs *FleetServerImpl) GetState(ctx context.Context, req *api.GetStateRequest) (response *ufspb.StateRecord, err error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return controller.GetState(ctx, req.ResourceName)
}

// UpdateDutState updates DUT state for a DUT.
func (fs *FleetServerImpl) UpdateDutState(ctx context.Context, req *api.UpdateDutStateRequest) (response *chromeosLab.DutState, err error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return nil, nil
}
