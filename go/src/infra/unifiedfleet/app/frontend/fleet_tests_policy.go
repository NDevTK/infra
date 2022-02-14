// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"

	api "infra/unifiedfleet/api/v1/rpc"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/controller"

	"go.chromium.org/luci/common/logging"
)

// CheckFleetTestsPolicy returns whether the given the test parameters are for a valid test.
func (fs *FleetServerImpl) CheckFleetTestsPolicy(ctx context.Context, req *api.CheckFleetTestsPolicyRequest) (response *ufsAPI.CheckFleetTestsPolicyResponse, err error) {
	isPublicUser, err := controller.IsPublicGroupMember(ctx)
	if err != nil {
		return nil, err
	}
	if !isPublicUser {
		return &ufsAPI.CheckFleetTestsPolicyResponse{
			IsTestValid: true,
			Status:      api.CheckFleetTestsPolicyResponse_PRIVATE_USER,
		}, nil
	}

	// Check test parameters
	status := api.CheckFleetTestsPolicyResponse_UNSPECIFIED
	err = controller.IsValidTest(ctx, req)
	if err == nil {
		return &ufsAPI.CheckFleetTestsPolicyResponse{
			IsTestValid: true,
			Status:      status,
		}, nil
	}

	switch err.(type) {
	case *controller.InvalidBoardError:
		logging.Infof(ctx, "Returning Invalid board error")
		status = api.CheckFleetTestsPolicyResponse_NOT_A_PUBLIC_BOARD
	case *controller.InvalidModelError:
		logging.Infof(ctx, "Returning Invalid model error")
		status = api.CheckFleetTestsPolicyResponse_NOT_A_PUBLIC_MODEL
	case *controller.InvalidTestError:
		logging.Infof(ctx, "Returning Invalid test error")
		status = api.CheckFleetTestsPolicyResponse_NOT_A_PUBLIC_TEST
	case *controller.InvalidImageError:
		logging.Infof(ctx, "Returning Invalid image error")
		status = api.CheckFleetTestsPolicyResponse_NOT_A_PUBLIC_IMAGE
	default:
		return nil, err
	}
	return &ufsAPI.CheckFleetTestsPolicyResponse{
		IsTestValid:   false,
		Status:        status,
		StatusMessage: err.Error(),
	}, nil
}
