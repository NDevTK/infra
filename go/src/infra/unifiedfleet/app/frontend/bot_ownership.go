// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/grpcutil"

	ufspb "infra/unifiedfleet/api/v1/models"
	api "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/controller"
)

// GetOwnershipData returns the ownership data for a given host.
func (fs *FleetServerImpl) GetOwnershipData(ctx context.Context, req *api.GetOwnershipDataRequest) (response *ufspb.OwnershipData, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	ownership, err := controller.GetOwnershipData(ctx, req.GetHostname())
	if err == nil {
		return ownership, nil
	}
	logging.Errorf(ctx, "Error while querying ownership data : %v", err)
	return nil, err
}
