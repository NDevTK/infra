// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inventory

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/logging"

	"infra/libs/skylab/inventory"
)

type invServiceClient struct {
}

func newInvServiceClient(ctx context.Context, host string) (inventoryClient, error) {
	return &invServiceClient{}, nil
}

func (client *invServiceClient) logInfo(ctx context.Context, t string, s ...interface{}) {
	logging.Infof(ctx, fmt.Sprintf("InventoryV2Clinet: %s", t), s...)
}

func (client *invServiceClient) addManyDUTsToFleet(ctx context.Context, nds []*inventory.CommonDeviceSpecs, pickServoPort bool) (string, []*inventory.CommonDeviceSpecs, error) {
	client.logInfo(ctx, "Adapter old data to inventory v2 proto")
	client.logInfo(ctx, "Call server RPC to add devices")
	client.logInfo(ctx, "Adapt the result back to old data format")
	return "No URL provided by inventory v2", nds, nil
}
