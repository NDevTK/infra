// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inventory

import (
	"context"
	"time"

	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	dsinventory "infra/appengine/crosskylabadmin/internal/app/frontend/datastore/inventory"
	"infra/appengine/crosskylabadmin/internal/app/gitstore"
)

type inventoryClient interface {
	getDutInfo(context.Context, *fleet.GetDutInfoRequest) ([]byte, time.Time, error)
}

type gitStoreClient struct {
	store *gitstore.InventoryStore
}

func newGitStoreClient(ctx context.Context, gs *gitstore.InventoryStore) (inventoryClient, error) {
	return &gitStoreClient{
		store: gs,
	}, nil
}

func (client *gitStoreClient) getDutInfo(ctx context.Context, req *fleet.GetDutInfoRequest) ([]byte, time.Time, error) {
	var dut *dsinventory.DeviceUnderTest
	var now time.Time
	var err error
	if req.Id != "" {
		dut, err = dsinventory.GetSerializedDUTByID(ctx, req.Id)
	} else {
		dut, err = dsinventory.GetSerializedDUTByHostname(ctx, req.Hostname)
	}
	if err != nil {
		if datastore.IsErrNoSuchEntity(err) {
			return nil, now, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, now, err
	}
	return dut.Data, dut.Updated, nil
}
