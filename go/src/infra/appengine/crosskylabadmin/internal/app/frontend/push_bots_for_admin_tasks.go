// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.chromium.org/luci/common/data/strpair"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	swarmingv2 "go.chromium.org/luci/swarming/proto/api_v2"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/clients"
	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/ufs"
	"infra/cros/recovery/logger/metrics"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

const labstationRebootKind = `action:Labstation reboot`

type adminTaskBotPusher struct {
	metricsClient  metrics.Metrics
	ufsClient      ufs.Client
	swarmingClient clients.SwarmingClient
}

// getLabstations takes a start time and a stop time and returns the labstations with reboot events in that time range.
func (p *adminTaskBotPusher) getLabstations(ctx context.Context, startTime time.Time, stopTime time.Time) ([]string, error) {
	// TODO(gregorynisbet): look at "action:Power cycle by RPM" as well.
	results, err := p.metricsClient.Search(ctx, &metrics.Query{
		StartTime:  startTime,
		StopTime:   stopTime,
		ActionKind: labstationRebootKind,
		Limit:      2000,
	})
	if err != nil {
		return nil, err
	}
	labstationMap := map[string]struct{}{}
	for _, action := range results.Actions {
		if action.Status == metrics.ActionStatusSuccess {
			labstationMap[action.Hostname] = struct{}{}
		}
	}
	var labstations []string
	for k := range labstationMap {
		labstations = append(labstations, k)
	}
	sort.Strings(labstations)
	return labstations, err
}

// getDUTsForLabstations gets all the DUTs associated with a labstation.
func (p *adminTaskBotPusher) getDUTsForLabstations(ctx context.Context, labstations []string) ([]string, error) {
	var duts []string
	resp, err := p.ufsClient.GetDUTsForLabstation(ctx, &ufsAPI.GetDUTsForLabstationRequest{
		Hostname: labstations,
	})
	if err != nil {
		return nil, err
	}
	for _, item := range resp.GetItems() {
		for _, hostname := range item.GetDutName() {
			duts = append(duts, fmt.Sprintf("crossk-%s", hostname))
		}
	}
	return duts, nil
}

// pushRepairDUTsForGivenPool pushes repair jobs for duts in a given pool.
// sc           -- the swarming client
// swarmingPool -- the swarming-level pool (NOT "label-pool") to push duts in
// dutState     -- the DUT state (e.g. "ready", "needs-repair")
// dims         -- a list of additional dimensions to map
// holdouts     -- a list of bot names to exclude (NOT dut names). Holdouts is read-only, so this parameter may be nil.
func (p *adminTaskBotPusher) pushRepairDUTsForGivenPool(ctx context.Context, swarmingPool string, dutState string, dims strpair.Map, holdouts map[string]bool) error {
	var bots []*swarmingv2.BotInfo
	rawBots, err := p.swarmingClient.ListAliveIdleBotsInPool(ctx, swarmingPool, dims)
	for _, bot := range rawBots {
		if !holdouts[bot.BotId] {
			bots = append(bots, bot)
		}
	}
	if err != nil {
		return errors.Annotate(err, "failed to list alive idle bots with dut_state %q", dutState).Err()
	} else {
		logging.Infof(ctx, "successfully get %d alive idle cros bots with dut_state %q in pool %q.", len(bots), dutState, swarmingPool)
		//Parse BOT id to schedule tasks for readability.
		repairBOTs := identifyBotsForRepair(ctx, bots)
		err = clients.PushRepairDUTs(ctx, repairBOTs, dutState, swarmingPool)
		if err != nil {
			logging.Infof(ctx, "Push repair bots in pool %q: %v", swarmingPool, err)
			return errors.Annotate(err, "Failed to push repair duts in pool %q", swarmingPool).Err()
		}
	}
	return nil
}

func (p *adminTaskBotPusher) pushBotsForAdminTasksImpl(ctx context.Context, req *fleet.PushBotsForAdminTasksRequest) (*fleet.PushBotsForAdminTasksResponse, error) {
	if p.swarmingClient == nil {
		return nil, errors.Reason("swarming client cannot be nil").Err()
	}
	cfg := config.Get(ctx)
	dutState, ok := clients.DutStateRevMap[req.GetTargetDutState()]
	if !ok {
		return nil, fmt.Errorf("DutState=%#v does not map to swarming value", req.GetTargetDutState())
	}

	var merr errors.MultiError
	// Schedule admin tasks to idle DUTs.
	dims := make(strpair.Map)
	dims[clients.DutStateDimensionKey] = []string{dutState}

	//TODO (prasadv): Create PoolCfg for ChromeOSSkylab and push admin tasks similar to other pool configs.
	// Once the Config is updated, remove the below code to push repair DUTs for admin task for Swarming.BotPool
	if cfg.Swarming.BotPool != "" {
		if err := p.pushRepairDUTsForGivenPool(ctx, cfg.Swarming.BotPool, dutState, dims, nil); err != nil {
			merr = append(merr, errors.Annotate(err, "Failed to push repair duts in pool %q", cfg.Swarming.BotPool).Err())
		} else {
			logging.Infof(ctx, "Successfully pushed repair duts with dut_state %q in pool %q.", dutState, cfg.Swarming.BotPool)
		}
	}

	// Loop through all the Swarming Pool configs and push duts for repair.
	for _, c := range cfg.Swarming.PoolCfgs {
		//TODO (prasadv): Remove this condition once BotPool is added to PoolCfg.
		if cfg.Swarming.BotPool != c.PoolName {
			if err := p.pushRepairDUTsForGivenPool(ctx, c.PoolName, dutState, dims, nil); err != nil {
				merr = append(merr, errors.Annotate(err, "Failed to push repair duts in pool %q", c.PoolName).Err())
			} else {
				logging.Infof(ctx, "Successfully pushed repair duts with dut_state %q in pool %q.", dutState, c.PoolName)
			}
		}
	}
	if len(merr) > 0 {
		return nil, merr
	}
	logging.Infof(ctx, "Successfully pushed repair duts in all pools")
	return &fleet.PushBotsForAdminTasksResponse{}, nil
}
