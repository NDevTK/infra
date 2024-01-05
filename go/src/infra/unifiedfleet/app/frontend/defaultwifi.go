// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"

	"go.chromium.org/luci/grpc/grpcutil"

	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/controller"
	"infra/unifiedfleet/app/util"
)

// CreateDefaultWifi creates DefaultWifi entry in database.
func (fs *FleetServerImpl) CreateDefaultWifi(ctx context.Context, req *ufsAPI.CreateDefaultWifiRequest) (rsp *ufspb.DefaultWifi, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	if err = req.Validate(); err != nil {
		return
	}
	req.DefaultWifi.Name = req.DefaultWifiId
	rsp, err = controller.CreateDefaultWifi(ctx, req.DefaultWifi)
	if err != nil {
		return nil, err
	}
	rsp.Name = util.AddPrefix(util.DefaultWifiCollection, rsp.Name)
	return
}
