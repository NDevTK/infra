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
	// Check test parameters
	status := api.CheckFleetTestsPolicyResponse_UNSPECIFIED
	err = controller.IsValidTest(ctx, req)
	if err == nil {
		return &ufsAPI.CheckFleetTestsPolicyResponse{
			Status: api.CheckFleetTestsPolicyResponse_OK,
		}, nil
	}

	logging.Errorf(ctx, "Returning error %s", err.Error())
	switch err.(type) {
	case *controller.InvalidBoardError:
		status = api.CheckFleetTestsPolicyResponse_NOT_A_PUBLIC_BOARD
	case *controller.InvalidModelError:
		status = api.CheckFleetTestsPolicyResponse_NOT_A_PUBLIC_MODEL
	case *controller.InvalidTestError:
		status = api.CheckFleetTestsPolicyResponse_NOT_A_PUBLIC_TEST
	case *controller.InvalidImageError:
		status = api.CheckFleetTestsPolicyResponse_NOT_A_PUBLIC_IMAGE
	default:
		return nil, err
	}
	return &ufsAPI.CheckFleetTestsPolicyResponse{
		Status:        status,
		StatusMessage: err.Error(),
	}, nil
}
