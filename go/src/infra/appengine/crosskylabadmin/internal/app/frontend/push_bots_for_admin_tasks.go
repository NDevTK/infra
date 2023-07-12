// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/data/strpair"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/clients"
	"infra/appengine/crosskylabadmin/internal/app/config"
)

func pushBotsForAdminTasksImpl(ctx context.Context, sc clients.SwarmingClient, req *fleet.PushBotsForAdminTasksRequest) (*fleet.PushBotsForAdminTasksResponse, error) {
	cfg := config.Get(ctx)
	dutState, ok := clients.DutStateRevMap[req.GetTargetDutState()]
	if !ok {
		return nil, fmt.Errorf("DutState=%#v does not map to swarming value", req.GetTargetDutState())
	}

	// Schedule admin tasks to idle DUTs.
	dims := make(strpair.Map)
	dims[clients.DutStateDimensionKey] = []string{dutState}
	bots, err := sc.ListAliveIdleBotsInPool(ctx, cfg.Swarming.BotPool, dims)
	if err != nil {
		reason := fmt.Sprintf("failed to list alive idle cros bots with dut_state %q", dutState)
		return nil, errors.Annotate(err, reason).Err()
	}
	logging.Infof(ctx, "successfully get %d alive idle cros bots with dut_state %q.", len(bots), dutState)

	// Parse BOT id to schedule tasks for readability.
	repairBOTs := identifyBotsForRepair(ctx, bots)
	err = clients.PushRepairDUTs(ctx, repairBOTs, dutState)
	if err != nil {
		logging.Infof(ctx, "push repair bots: %v", err)
		return nil, errors.New("failed to push repair duts")
	}
	return &fleet.PushBotsForAdminTasksResponse{}, nil
}
