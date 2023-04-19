// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"fmt"

	empty "github.com/golang/protobuf/ptypes/empty"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/grpcutil"
	status "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"

	invV2Api "infra/appengine/cros/lab_inventory/api/v1"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/controller"
	"infra/unifiedfleet/app/external"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/util"
)

func verifyLSEPrototype(ctx context.Context, lse *ufspb.MachineLSE) error {
	if lse.GetChromeBrowserMachineLse() != nil {
		if !util.IsInBrowserZone(lse.GetMachineLsePrototype()) {
			return grpcStatus.Errorf(codes.InvalidArgument, "Prototype %s doesn't belong to browser lab", lse.GetMachineLsePrototype())
		}
		resp, err := controller.GetMachineLSEPrototype(ctx, lse.GetMachineLsePrototype())
		if err != nil {
			return grpcStatus.Errorf(codes.InvalidArgument, "Prototype %s doesn't exist", lse.GetMachineLsePrototype())
		}
		for _, v := range resp.GetVirtualRequirements() {
			if v.GetVirtualType() == ufspb.VirtualType_VIRTUAL_TYPE_VM {
				c := lse.GetChromeBrowserMachineLse().GetVmCapacity()
				if c < v.GetMin() || c > v.GetMax() {
					return grpcStatus.Errorf(codes.InvalidArgument, "Prototype %s is not matched to the vm capacity %d", lse.GetMachineLsePrototype(), c)
				}
			}
		}
	}
	return nil
}

// CreateMachineLSE creates machineLSE entry in database.
func (fs *FleetServerImpl) CreateMachineLSE(ctx context.Context, req *ufsAPI.CreateMachineLSERequest) (rsp *ufspb.MachineLSE, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := verifyLSEPrototype(ctx, req.GetMachineLSE()); err != nil {
		return nil, err
	}
	req.MachineLSE.Name = util.FormatDHCPHostname(req.MachineLSEId)
	req.MachineLSE.Hostname = util.FormatDHCPHostname(req.MachineLSE.Hostname)
	req.NetworkOption = updateNetworkOpt(req.MachineLSE.GetVlan(), req.MachineLSE.GetIp(), req.GetNetworkOption())

	machineLSE, err := controller.CreateMachineLSE(ctx, req.MachineLSE, req.GetNetworkOption())
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	machineLSE.Name = util.AddPrefix(util.MachineLSECollection, machineLSE.Name)
	return machineLSE, err
}

// UpdateMachineLSE updates the machineLSE information in database.
func (fs *FleetServerImpl) UpdateMachineLSE(ctx context.Context, req *ufsAPI.UpdateMachineLSERequest) (rsp *ufspb.MachineLSE, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	req.MachineLSE.Name = util.FormatDHCPHostname(util.RemovePrefix(req.MachineLSE.Name))
	req.MachineLSE.Hostname = util.FormatDHCPHostname(req.MachineLSE.Hostname)
	nwOpt := req.GetNetworkOptions()[req.MachineLSE.Name]
	nwOpt = updateNetworkOpt(req.MachineLSE.GetVlan(), req.MachineLSE.GetIp(), nwOpt)
	if nwOpt != nil {
		machinelse := req.MachineLSE
		var err error
		if req.UpdateMask != nil && len(req.UpdateMask.Paths) > 0 {
			machinelse, err = controller.UpdateMachineLSE(ctx, req.MachineLSE, req.UpdateMask)
			if err != nil {
				return nil, err
			}
		}

		// If network_option.delete is enabled, ignore network_option.vlan and return directly
		if nwOpt.GetDelete() {
			if err = controller.DeleteMachineLSEHost(ctx, req.MachineLSE.Name); err != nil {
				return nil, err
			}
			machinelse, err = controller.GetMachineLSE(ctx, req.MachineLSE.Name)
			if err != nil {
				return nil, err
			}
		} else if nwOpt.GetVlan() != "" || nwOpt.GetIp() != "" || nwOpt.GetNic() != "" {
			machinelse, err = controller.UpdateMachineLSEHost(ctx, req.MachineLSE.Name, nwOpt)
			if err != nil {
				return nil, err
			}
		}

		// https://aip.dev/122 - as per AIP guideline
		machinelse.Name = util.AddPrefix(util.MachineLSECollection, machinelse.Name)
		return machinelse, nil
	}

	machinelse, err := controller.UpdateMachineLSE(ctx, req.MachineLSE, req.UpdateMask)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	machinelse.Name = util.AddPrefix(util.MachineLSECollection, machinelse.Name)
	return machinelse, err
}

// GetMachineLSE gets the machineLSE information from database.
func (fs *FleetServerImpl) GetMachineLSE(ctx context.Context, req *ufsAPI.GetMachineLSERequest) (rsp *ufspb.MachineLSE, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.FormatDHCPHostname(util.RemovePrefix(req.Name))
	machineLSE, err := controller.GetMachineLSE(ctx, name)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	machineLSE.Name = util.AddPrefix(util.MachineLSECollection, machineLSE.Name)
	return machineLSE, err
}

// BatchGetMachineLSEs gets a batch of machineLSE information from database.
func (fs *FleetServerImpl) BatchGetMachineLSEs(ctx context.Context, req *ufsAPI.BatchGetMachineLSEsRequest) (rsp *ufsAPI.BatchGetMachineLSEsResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	lses, err := controller.BatchGetMachineLSEs(ctx, util.FormatDHCPHostnames(util.FormatInputNames(req.GetNames())))
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	for _, lse := range lses {
		lse.Name = util.AddPrefix(util.MachineLSECollection, lse.Name)
	}
	return &ufsAPI.BatchGetMachineLSEsResponse{
		MachineLses: lses,
	}, nil
}

// ListMachineLSEs list the machineLSEs information from database.
func (fs *FleetServerImpl) ListMachineLSEs(ctx context.Context, req *ufsAPI.ListMachineLSEsRequest) (rsp *ufsAPI.ListMachineLSEsResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	pageSize := util.GetPageSize(req.PageSize)
	result, nextPageToken, err := controller.ListMachineLSEs(ctx, pageSize, req.PageToken, req.Filter, req.KeysOnly, req.Full)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	for _, machineLSE := range result {
		machineLSE.Name = util.AddPrefix(util.MachineLSECollection, machineLSE.Name)
	}
	return &ufsAPI.ListMachineLSEsResponse{
		MachineLSEs:   result,
		NextPageToken: nextPageToken,
	}, nil
}

// DeleteMachineLSE deletes the machineLSE from database.
func (fs *FleetServerImpl) DeleteMachineLSE(ctx context.Context, req *ufsAPI.DeleteMachineLSERequest) (rsp *empty.Empty, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.FormatDHCPHostname(util.RemovePrefix(req.Name))
	err = controller.DeleteMachineLSE(ctx, name)
	return &empty.Empty{}, err
}

// RenameMachineLSE renames the machinelse in database.
func (fs *FleetServerImpl) RenameMachineLSE(ctx context.Context, req *ufsAPI.RenameMachineLSERequest) (lse *ufspb.MachineLSE, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	lse, err = controller.RenameMachineLSE(ctx, util.RemovePrefix(req.Name), util.RemovePrefix(req.NewName))
	if err != nil {
		return nil, err
	}
	lse.Name = util.AddPrefix(util.MachineLSECollection, lse.Name)
	return
}

func updateNetworkOpt(userVlan, ip string, nwOpt *ufsAPI.NetworkOption) *ufsAPI.NetworkOption {
	if userVlan == "" && ip == "" {
		return nwOpt
	}
	if nwOpt == nil {
		return &ufsAPI.NetworkOption{
			Vlan: userVlan,
			Ip:   ip,
		}
	}
	nwOpt.Vlan = userVlan
	nwOpt.Ip = ip
	return nwOpt
}

// CreateVM creates a vm entry in database.
func (fs *FleetServerImpl) CreateVM(ctx context.Context, req *ufsAPI.CreateVMRequest) (rsp *ufspb.VM, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	req.Vm.Name = util.FormatDHCPHostname(util.RemovePrefix(req.Vm.Name))
	req.Vm.Hostname = util.FormatDHCPHostname(util.RemovePrefix(req.Vm.Hostname))
	req.Vm.MachineLseId = util.FormatDHCPHostname(req.Vm.MachineLseId)
	req.NetworkOption = updateNetworkOpt(req.Vm.GetVlan(), req.Vm.GetIp(), req.GetNetworkOption())
	vm, err := controller.CreateVM(ctx, req.GetVm(), req.GetNetworkOption())
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	vm.Name = util.AddPrefix(util.VMCollection, vm.Name)
	return vm, err
}

// UpdateVM updates the vm information in database.
func (fs *FleetServerImpl) UpdateVM(ctx context.Context, req *ufsAPI.UpdateVMRequest) (rsp *ufspb.VM, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	req.Vm.Name = util.FormatDHCPHostname(util.RemovePrefix(req.Vm.Name))
	req.Vm.Hostname = util.FormatDHCPHostname(util.RemovePrefix(req.Vm.Hostname))
	req.Vm.MachineLseId = util.FormatDHCPHostname(req.Vm.MachineLseId)
	req.NetworkOption = updateNetworkOpt(req.Vm.GetVlan(), req.Vm.GetIp(), req.GetNetworkOption())
	if req.GetNetworkOption() != nil {
		vm := req.Vm
		var err error
		if req.UpdateMask != nil && len(req.UpdateMask.Paths) > 0 {
			vm, err = controller.UpdateVM(ctx, req.Vm, req.UpdateMask)
			if err != nil {
				return nil, err
			}
		}

		// If network_option.delete is enabled, ignore network_option.vlan and return directly
		if req.GetNetworkOption().GetDelete() {
			if err = controller.DeleteVMHost(ctx, req.Vm.Name); err != nil {
				return nil, err
			}
			vm, err = controller.GetVM(ctx, req.Vm.Name)
			if err != nil {
				return nil, err
			}
		} else if req.GetNetworkOption().GetVlan() != "" || req.GetNetworkOption().GetIp() != "" {
			vm, err = controller.UpdateVMHost(ctx, req.Vm.Name, req.GetNetworkOption())
			if err != nil {
				return nil, err
			}
		}

		// https://aip.dev/122 - as per AIP guideline
		vm.Name = util.AddPrefix(util.VMCollection, vm.Name)
		return vm, nil
	}

	vm, err := controller.UpdateVM(ctx, req.Vm, req.UpdateMask)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	vm.Name = util.AddPrefix(util.VMCollection, vm.Name)
	return vm, err
}

// DeleteVM deletes a VM from database.
func (fs *FleetServerImpl) DeleteVM(ctx context.Context, req *ufsAPI.DeleteVMRequest) (rsp *empty.Empty, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.FormatDHCPHostname(util.RemovePrefix(req.Name))
	err = controller.DeleteVM(ctx, name)
	return &empty.Empty{}, err
}

// GetVM gets the VM information from database.
func (fs *FleetServerImpl) GetVM(ctx context.Context, req *ufsAPI.GetVMRequest) (rsp *ufspb.VM, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.FormatDHCPHostname(util.RemovePrefix(req.Name))
	vm, err := controller.GetVM(ctx, name)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	vm.Name = util.AddPrefix(util.VMCollection, vm.Name)
	return vm, err
}

// BatchGetVMs gets a batch of vms from database.
func (fs *FleetServerImpl) BatchGetVMs(ctx context.Context, req *ufsAPI.BatchGetVMsRequest) (rsp *ufsAPI.BatchGetVMsResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	vms, err := controller.BatchGetVMs(ctx, util.FormatDHCPHostnames(util.FormatInputNames(req.GetNames())))
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	for _, v := range vms {
		v.Name = util.AddPrefix(util.VMCollection, v.Name)
	}
	return &ufsAPI.BatchGetVMsResponse{
		Vms: vms,
	}, nil
}

// ListVMs list the vms information from database.
func (fs *FleetServerImpl) ListVMs(ctx context.Context, req *ufsAPI.ListVMsRequest) (rsp *ufsAPI.ListVMsResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	pageSize := util.GetPageSize(req.PageSize)
	result, nextPageToken, err := controller.ListVMs(ctx, pageSize, req.PageToken, req.Filter, req.KeysOnly)
	if err != nil {
		return nil, err
	}
	return &ufsAPI.ListVMsResponse{
		Vms:           result,
		NextPageToken: nextPageToken,
	}, nil
}

// CreateRackLSE creates rackLSE entry in database.
func (fs *FleetServerImpl) CreateRackLSE(ctx context.Context, req *ufsAPI.CreateRackLSERequest) (rsp *ufspb.RackLSE, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	req.RackLSE.Name = req.RackLSEId
	rackLSE, err := controller.CreateRackLSE(ctx, req.RackLSE)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	rackLSE.Name = util.AddPrefix(util.RackLSECollection, rackLSE.Name)
	return rackLSE, err
}

// UpdateRackLSE updates the rackLSE information in database.
func (fs *FleetServerImpl) UpdateRackLSE(ctx context.Context, req *ufsAPI.UpdateRackLSERequest) (rsp *ufspb.RackLSE, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	req.RackLSE.Name = util.RemovePrefix(req.RackLSE.Name)
	rackLSE, err := controller.UpdateRackLSE(ctx, req.RackLSE)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	rackLSE.Name = util.AddPrefix(util.RackLSECollection, rackLSE.Name)
	return rackLSE, err
}

// GetRackLSE gets the rackLSE information from database.
func (fs *FleetServerImpl) GetRackLSE(ctx context.Context, req *ufsAPI.GetRackLSERequest) (rsp *ufspb.RackLSE, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.RemovePrefix(req.Name)
	rackLSE, err := controller.GetRackLSE(ctx, name)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	rackLSE.Name = util.AddPrefix(util.RackLSECollection, rackLSE.Name)
	return rackLSE, err
}

// ListRackLSEs list the rackLSEs information from database.
func (fs *FleetServerImpl) ListRackLSEs(ctx context.Context, req *ufsAPI.ListRackLSEsRequest) (rsp *ufsAPI.ListRackLSEsResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	pageSize := util.GetPageSize(req.PageSize)
	result, nextPageToken, err := controller.ListRackLSEs(ctx, pageSize, req.PageToken, req.Filter, req.KeysOnly)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	for _, rackLSE := range result {
		rackLSE.Name = util.AddPrefix(util.RackLSECollection, rackLSE.Name)
	}
	return &ufsAPI.ListRackLSEsResponse{
		RackLSEs:      result,
		NextPageToken: nextPageToken,
	}, nil
}

// DeleteRackLSE deletes the rackLSE from database.
func (fs *FleetServerImpl) DeleteRackLSE(ctx context.Context, req *ufsAPI.DeleteRackLSERequest) (rsp *empty.Empty, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.RemovePrefix(req.Name)
	err = controller.DeleteRackLSE(ctx, name)
	return &empty.Empty{}, err
}

// ImportOSMachineLSEs imports chromeos devices machine lses
func (fs *FleetServerImpl) ImportOSMachineLSEs(ctx context.Context, req *ufsAPI.ImportOSMachineLSEsRequest) (response *status.Status, err error) {
	source := req.GetMachineDbSource()
	if err := ufsAPI.ValidateMachineDBSource(source); err != nil {
		return nil, err
	}
	es, err := external.GetServerInterface(ctx)
	if err != nil {
		return nil, err
	}
	client, err := es.NewCrosInventoryInterfaceFactory(ctx, source.GetHost())
	if err != nil {
		return nil, crosInventoryConnectionFailureStatus.Err()
	}
	resp, err := client.ListCrosDevicesLabConfig(ctx, &invV2Api.ListCrosDevicesLabConfigRequest{})
	if err != nil {
		return nil, crosInventoryServiceFailureStatus("ListCrosDevicesLabConfig").Err()
	}
	pageSize := fs.getImportPageSize()
	res, err := controller.ImportOSMachineLSEs(ctx, resp.GetLabConfigs(), pageSize)
	s := processImportDatastoreRes(res, err)
	if s.Err() != nil {
		return s.Proto(), s.Err()
	}
	return successStatus.Proto(), nil
}

// GetChromeOSDeviceData gets the ChromeOSDeviceData(MachineLSE, Machine, Device config, Manufacturing config, Dutstate and Hwid data)
func (fs *FleetServerImpl) GetChromeOSDeviceData(ctx context.Context, req *ufsAPI.GetChromeOSDeviceDataRequest) (rsp *ufspb.ChromeOSDeviceData, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	osCtx, _ := util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	return controller.GetChromeOSDeviceData(osCtx, req.GetChromeosDeviceId(), req.GetHostname())
}

// UpdateMachineLSEDeployment updates the deployment record for a host
func (fs *FleetServerImpl) UpdateMachineLSEDeployment(ctx context.Context, req *ufsAPI.UpdateMachineLSEDeploymentRequest) (resp *ufspb.MachineLSEDeployment, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return controller.UpdateMachineLSEDeployment(ctx, req.GetMachineLseDeployment(), req.GetUpdateMask())
}

// BatchUpdateMachineLSEDeployment updates the deployment record for a batch of hosts
func (fs *FleetServerImpl) BatchUpdateMachineLSEDeployment(ctx context.Context, req *ufsAPI.BatchUpdateMachineLSEDeploymentRequest) (resp *ufsAPI.BatchUpdateMachineLSEDeploymentResponse, err error) {
	return nil, nil
}

// GetMachineLSEDeployment retrieves the deployment record for a host
func (fs *FleetServerImpl) GetMachineLSEDeployment(ctx context.Context, req *ufsAPI.GetMachineLSEDeploymentRequest) (resp *ufspb.MachineLSEDeployment, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.RemovePrefix(req.Name)
	dr, err := controller.GetMachineLSEDeployment(ctx, name)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	dr.SerialNumber = util.AddPrefix(util.MachineLSEDeploymentCollection, dr.SerialNumber)
	return dr, err
}

// BatchGetMachineLSEDeployments retrieves a batch of deployment records for hosts
func (fs *FleetServerImpl) BatchGetMachineLSEDeployments(ctx context.Context, req *ufsAPI.BatchGetMachineLSEDeploymentsRequest) (resp *ufsAPI.BatchGetMachineLSEDeploymentsResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	drs, err := controller.BatchGetMachineLSEDeployments(ctx, util.FormatInputNames(req.GetNames()))
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	for _, dr := range drs {
		dr.SerialNumber = util.AddPrefix(util.MachineLSEDeploymentCollection, dr.SerialNumber)
	}
	return &ufsAPI.BatchGetMachineLSEDeploymentsResponse{
		MachineLseDeployments: drs,
	}, nil
}

// ListMachineLSEDeployments retrieves a list of deployment records
func (fs *FleetServerImpl) ListMachineLSEDeployments(ctx context.Context, req *ufsAPI.ListMachineLSEDeploymentsRequest) (resp *ufsAPI.ListMachineLSEDeploymentsResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	pageSize := util.GetPageSize(req.PageSize)
	result, nextPageToken, err := controller.ListMachineLSEDeployments(ctx, pageSize, req.PageToken, req.Filter, req.KeysOnly)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline
	for _, res := range result {
		res.SerialNumber = util.AddPrefix(util.MachineLSEDeploymentCollection, res.GetSerialNumber())
	}
	return &ufsAPI.ListMachineLSEDeploymentsResponse{
		MachineLseDeployments: result,
		NextPageToken:         nextPageToken,
	}, nil
}

// CreateSchedulingUnit creates SchedulingUnit entry in database.
func (fs *FleetServerImpl) CreateSchedulingUnit(ctx context.Context, req *ufsAPI.CreateSchedulingUnitRequest) (rsp *ufspb.SchedulingUnit, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	req.SchedulingUnit.Name = req.SchedulingUnitId
	cs, err := controller.CreateSchedulingUnit(ctx, req.SchedulingUnit)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline.
	cs.Name = util.AddPrefix(util.SchedulingUnitCollection, cs.Name)
	return cs, err
}

// UpdateSchedulingUnit updates the SchedulingUnit information in database.
func (fs *FleetServerImpl) UpdateSchedulingUnit(ctx context.Context, req *ufsAPI.UpdateSchedulingUnitRequest) (rsp *ufspb.SchedulingUnit, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	req.SchedulingUnit.Name = util.RemovePrefix(req.SchedulingUnit.Name)
	cs, err := controller.UpdateSchedulingUnit(ctx, req.SchedulingUnit, req.UpdateMask)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline.
	cs.Name = util.AddPrefix(util.SchedulingUnitCollection, cs.Name)
	return cs, err
}

// GetSchedulingUnit gets the SchedulingUnit information from database.
func (fs *FleetServerImpl) GetSchedulingUnit(ctx context.Context, req *ufsAPI.GetSchedulingUnitRequest) (rsp *ufspb.SchedulingUnit, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.RemovePrefix(req.Name)
	cs, err := controller.GetSchedulingUnit(ctx, name)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline.
	cs.Name = util.AddPrefix(util.SchedulingUnitCollection, cs.Name)
	return cs, err
}

// ListSchedulingUnits list the SchedulingUnits information from database.
func (fs *FleetServerImpl) ListSchedulingUnits(ctx context.Context, req *ufsAPI.ListSchedulingUnitsRequest) (rsp *ufsAPI.ListSchedulingUnitsResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	pageSize := util.GetPageSize(req.PageSize)
	result, nextPageToken, err := controller.ListSchedulingUnits(ctx, pageSize, req.PageToken, req.Filter, req.KeysOnly)
	if err != nil {
		return nil, err
	}
	// https://aip.dev/122 - as per AIP guideline.
	for _, cs := range result {
		cs.Name = util.AddPrefix(util.SchedulingUnitCollection, cs.Name)
	}
	return &ufsAPI.ListSchedulingUnitsResponse{
		SchedulingUnits: result,
		NextPageToken:   nextPageToken,
	}, nil
}

// DeleteSchedulingUnit deletes the SchedulingUnit from database.
func (fs *FleetServerImpl) DeleteSchedulingUnit(ctx context.Context, req *ufsAPI.DeleteSchedulingUnitRequest) (rsp *empty.Empty, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	name := util.RemovePrefix(req.Name)
	err = controller.DeleteSchedulingUnit(ctx, name)
	return &empty.Empty{}, err
}

// GetDeviceData gets the requested device data (scheduling unit, chromeos data,
// attached device data) from UFS.
func (fs *FleetServerImpl) GetDeviceData(ctx context.Context, req *ufsAPI.GetDeviceDataRequest) (rsp *ufsAPI.GetDeviceDataResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Find LSE for hostname/asset tag
	req.Hostname = util.RemovePrefix(req.GetHostname())
	lse, err := getMachineLseIfExists(ctx, req.GetDeviceId(), req.GetHostname())
	ns := util.GetNamespaceFromCtx(ctx)
	logging.Infof(ctx, "querying namespace %q", ns)
	if err != nil {
		if ns == util.BrowserNamespace {
			// It may be VM bots
			rsp, err = getBrowserVMDataIfExists(ctx, req.GetHostname())
		} else {
			// Try to query and return SchedulingUnit if failed to fetch lse
			rsp, err = getSchedulingUnitDeviceDataIfExists(ctx, req.GetHostname())
		}
		if err != nil {
			logging.Errorf(ctx, err.Error())
		}
		if rsp != nil {
			return rsp, nil
		}
		return nil, grpcStatus.Error(codes.NotFound, "no valid device found")
	}

	// Get data based on device type
	if lse.GetChromeBrowserMachineLse() != nil {
		return &ufsAPI.GetDeviceDataResponse{
			Resource: &ufsAPI.GetDeviceDataResponse_BrowserDeviceData{
				BrowserDeviceData: &ufsAPI.BrowserDeviceData{
					Host: lse,
				},
			},
			ResourceType: ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_BROWSER_DEVICE,
		}, nil
	} else if lse.GetChromeosMachineLse() != nil {
		// TODO (justinsuen): refactor GetChromeOSDeviceData to take LSE as input.
		// Will remove machineId assignment after refactor.
		var machineId string
		if len(lse.GetMachines()) != 0 {
			machineId = lse.GetMachines()[0]
		} else if req.GetDeviceId() != "" {
			machineId = req.GetDeviceId()
		}
		device, err := controller.GetChromeOSDeviceData(ctx, machineId, req.GetHostname())
		if err != nil {
			return nil, errors.Annotate(err, "failed to get chromeos device data").Err()
		}
		return &ufsAPI.GetDeviceDataResponse{
			Resource: &ufsAPI.GetDeviceDataResponse_ChromeOsDeviceData{
				ChromeOsDeviceData: device,
			},
			ResourceType: ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_CHROMEOS_DEVICE,
		}, nil
	} else if lse.GetAttachedDeviceLse() != nil {
		device, err := controller.GetAttachedDeviceData(ctx, lse)
		if err != nil {
			return nil, errors.Annotate(err, "failed to get attached device data").Err()
		}
		return &ufsAPI.GetDeviceDataResponse{
			Resource: &ufsAPI.GetDeviceDataResponse_AttachedDeviceData{
				AttachedDeviceData: device,
			},
			ResourceType: ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_ATTACHED_DEVICE,
		}, nil
	}
	return nil, grpcStatus.Error(codes.NotFound, "no valid device found")
}

func getBrowserVMDataIfExists(ctx context.Context, hostname string) (*ufsAPI.GetDeviceDataResponse, error) {
	vm, err := controller.GetVM(ctx, hostname)
	if err != nil {
		return nil, err
	}
	if vm != nil {
		return &ufsAPI.GetDeviceDataResponse{
			Resource: &ufsAPI.GetDeviceDataResponse_BrowserDeviceData{
				BrowserDeviceData: &ufsAPI.BrowserDeviceData{
					Vm: vm,
				},
			},
			ResourceType: ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_BROWSER_DEVICE,
		}, nil
	}
	return nil, fmt.Errorf("failed to get vm for %s", hostname)
}

func getSchedulingUnitDeviceDataIfExists(ctx context.Context, hostname string) (*ufsAPI.GetDeviceDataResponse, error) {
	su, err := controller.GetSchedulingUnit(ctx, hostname)
	if err != nil {
		return nil, err
	}
	if su != nil {
		return &ufsAPI.GetDeviceDataResponse{
			Resource: &ufsAPI.GetDeviceDataResponse_SchedulingUnit{
				SchedulingUnit: su,
			},
			ResourceType: ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_SCHEDULING_UNIT,
		}, nil
	}
	return nil, fmt.Errorf("failed to get scheduling unit for %s", hostname)
}

func getMachineLseIfExists(ctx context.Context, id, hostname string) (*ufspb.MachineLSE, error) {
	// Find LSE for hostname/asset tag
	var lse *ufspb.MachineLSE
	var err error
	if hostname != "" {
		// Query MachineLSE by hostname
		lse, err = controller.GetMachineLSE(ctx, hostname)
		if err != nil {
			return nil, err
		}
	} else if id != "" {
		// Query MachineLSE by Machine id
		machinelses, err := inventory.QueryMachineLSEByPropertyName(ctx, "machine_ids", id, false)
		if err != nil {
			return nil, err
		}
		if len(machinelses) == 0 {
			return nil, grpcStatus.Error(codes.NotFound, fmt.Sprintf("device not found w/ asset id %s", id))
		}
		lse = machinelses[0]
	}
	return lse, nil
}
