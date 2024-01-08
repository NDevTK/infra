// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/grpcutil"

	ufspb "infra/unifiedfleet/api/v1/models"
	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	api "infra/unifiedfleet/api/v1/rpc"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/controller"
	"infra/unifiedfleet/app/util"
)

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
	if err := controller.UpdateDutMeta(ctx, req.GetDutMeta()); err != nil {
		logging.Errorf(ctx, "fail to update dut meta: %s", err.Error())
		return nil, err
	}

	if err := controller.UpdateAssetMeta(ctx, req.GetDutMeta()); err != nil {
		logging.Errorf(ctx, "fail to update asset meta: %s", err.Error())
		return nil, err
	}

	if err := controller.UpdateLabMeta(ctx, req.GetLabMeta()); err != nil {
		logging.Errorf(ctx, "fail to update lab meta: %s", err.Error())
		return nil, err
	}

	res, err := controller.UpdateDutState(ctx, req.GetDutState())
	if err != nil {
		return nil, err
	}
	return res, nil
}

// UpdateDeviceRecoveryData update device configs for a DUT
func (fs *FleetServerImpl) UpdateDeviceRecoveryData(ctx context.Context, req *api.UpdateDeviceRecoveryDataRequest) (rsp *api.UpdateDeviceRecoveryDataResponse, err error) {
	if err := req.Validate(); err != nil {
		logging.Errorf(ctx, "UpdateDeviceRecoverData request validate fail - %s", err.Error())
		return nil, err
	}
	if err := controller.UpdateRecoveryData(ctx, req); err != nil {
		logging.Errorf(ctx, "fail to update device recovery data: %s", err.Error())
		return nil, err
	}
	return &api.UpdateDeviceRecoveryDataResponse{}, nil
}

// UpdateTestData updates the device date provide by Test runner.
func (fs *FleetServerImpl) UpdateTestData(ctx context.Context, req *api.UpdateTestDataRequest) (rsp *api.UpdateTestDataResponse, err error) {
	if err := req.Validate(); err != nil {
		logging.Errorf(ctx, "UpdateTestData request validate fail - %s", err.Error())
		return nil, err
	}
	if err := controller.UpdateTestData(ctx, req); err != nil {
		logging.Errorf(ctx, "fail to update test data: %s", err.Error())
		return nil, err
	}
	return &api.UpdateTestDataResponse{}, nil
}

// GetDutState gets the ChromeOS device DutState.
func (fs *FleetServerImpl) GetDutState(ctx context.Context, req *api.GetDutStateRequest) (rsp *chromeosLab.DutState, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	osCtx, _ := util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	return controller.GetDutState(osCtx, req.GetChromeosDeviceId(), req.GetHostname())
}

// ListDutStates list the DutStates information from database.
func (fs *FleetServerImpl) ListDutStates(ctx context.Context, req *api.ListDutStatesRequest) (rsp *api.ListDutStatesResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := ufsAPI.ValidateListRequest(req); err != nil {
		return nil, err
	}
	pageSize := util.GetPageSize(req.PageSize)
	result, nextPageToken, err := controller.ListDutStates(ctx, pageSize, req.PageToken, req.Filter, req.KeysOnly)
	if err != nil {
		return nil, err
	}
	return &api.ListDutStatesResponse{
		DutStates:     result,
		NextPageToken: nextPageToken,
	}, nil
}
