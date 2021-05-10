// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package schedulingunit implement loading scheduling unit from UFS
// for the worker.

package schedulingunit

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cmd/skylab_swarming_worker/internal/swmbot"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// Get a SchedulingUnit from UFS, unlike a DeviceUnderTest, a SchedulingUnit doesn't
// have ID field, so both dut_id and dut_name swarming dimensions are referred from
// name field of SchedulingUnit.
func GetSchedulingUnitFromUFS(ctx context.Context, b *swmbot.Info, name string) (*ufspb.SchedulingUnit, error) {
	req := &ufsAPI.GetSchedulingUnitRequest{
		Name: ufsUtil.AddPrefix(ufsUtil.SchedulingUnitCollection, name),
	}
	client, err := swmbot.UFSClient(ctx, b)
	if err != nil {
		return nil, errors.Annotate(err, "Get SchedulingUnit from UFS: initialize UFS client").Err()
	}
	return client.GetSchedulingUnit(swmbot.SetupContext(ctx, ufsUtil.OSNamespace), req)
}
