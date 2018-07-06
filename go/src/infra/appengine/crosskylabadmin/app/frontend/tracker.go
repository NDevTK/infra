// Copyright 2018 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package frontend

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/duration"
	"go.chromium.org/gae/service/datastore"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/data/strpair"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/sync/parallel"
	"go.chromium.org/luci/grpc/grpcutil"
	"golang.org/x/net/context"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/app/clients"
)

// TrackerServerImpl implements the fleet.TrackerServer interface.
type TrackerServerImpl struct {
	clients.SwarmingFactory
}

// RefreshBots implements the fleet.Tracker.RefreshBots() method.
func (tsi *TrackerServerImpl) RefreshBots(c context.Context, req *fleet.RefreshBotsRequest) (res *fleet.RefreshBotsResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(c, err)
	}()

	sc, err := tsi.SwarmingClient(c, swarmingInstance)
	if err != nil {
		return nil, errors.Annotate(err, "failed to obtain Swarming client").Err()
	}

	bots, err := getBotsFromSwarming(c, sc, req.Selectors)
	if err != nil {
		return nil, errors.Annotate(err, "failed to get bots from Swarming").Err()
	}
	bsm, err := botInfoToSummary(bots)
	if err != nil {
		return nil, errors.Annotate(err, "failed to extract bot summary from bot info").Err()
	}
	if err := setIdleDuration(c, sc, bsm); err != nil {
		return nil, errors.Annotate(err, "failed to set idle time for bots").Err()
	}
	updated, err := insertBotSummary(c, bsm)
	if err != nil {
		return nil, errors.Annotate(err, "failed to insert bots").Err()
	}
	return &fleet.RefreshBotsResponse{
		DutIds: updated,
	}, nil
}

// SummarizeBots implements the fleet.Tracker.SummarizeBots() method.
func (tsi *TrackerServerImpl) SummarizeBots(c context.Context, req *fleet.SummarizeBotsRequest) (res *fleet.SummarizeBotsResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(c, err)
	}()

	bses, err := getBotSummariesFromDatastore(c, req.Selectors)
	if err != nil {
		return nil, err
	}
	bss := make([]*fleet.BotSummary, 0, len(bses))
	for _, bse := range bses {
		bs, err := bse.Decode()
		if err != nil {
			return nil, errors.Annotate(err, "failed to unmarshal bot summary for bot with dut_id %q", bse.DutID).Err()
		}
		bss = append(bss, bs)
	}
	return &fleet.SummarizeBotsResponse{
		Bots: bss,
	}, nil
}

// getBotsFromSwarming lists bots by calling the Swarming service.
func getBotsFromSwarming(c context.Context, sc clients.SwarmingClient, sels []*fleet.BotSelector) ([]*swarming.SwarmingRpcsBotInfo, error) {
	// No filters implies get all bots.
	if len(sels) == 0 {
		bots, err := sc.ListAliveBotsInPool(c, swarmingBotPool, strpair.Map{})
		if err != nil {
			return nil, errors.Annotate(err, "failed to get bots in pool %s", swarmingBotPool).Err()
		}
		return bots, nil
	}

	sels = dropDuplicateSelectors(sels)

	// For now, each selector can only yield 0 or 1 bot to update.
	bots := make([]*swarming.SwarmingRpcsBotInfo, 0, len(sels))
	// Protects access to bots
	m := &sync.Mutex{}
	err := parallel.WorkPool(clients.MaxConcurrentSwarmingCalls, func(workC chan<- func() error) {
		for i := range sels {
			// In-scope variable for goroutine closure.
			sel := sels[i]
			workC <- func() error {
				bs, ierr := getFilteredBotsFromSwarming(c, sc, sel)
				if ierr != nil {
					return ierr
				}
				m.Lock()
				defer m.Unlock()
				bots = append(bots, bs...)
				return nil
			}
		}
	})
	return bots, err
}

// getFilteredBotsFromSwarming lists bots for a single selector by calling the Swarming service.
// This function is intended to be used in a parallel.WorkPool().
func getFilteredBotsFromSwarming(c context.Context, sc clients.SwarmingClient, sel *fleet.BotSelector) ([]*swarming.SwarmingRpcsBotInfo, error) {
	dims := strpair.Map{
		clients.DutIDDimensionKey: []string{sel.DutId},
	}
	bs, err := sc.ListAliveBotsInPool(c, swarmingBotPool, dims)
	if err != nil {
		return nil, errors.Annotate(err, "failed to get bots in pool %s with dimensions %s", swarmingBotPool, dims).Err()
	}
	return bs, nil
}

var dutStateMap = map[string]fleet.DutState{
	"ready":         fleet.DutState_Ready,
	"needs_cleanup": fleet.DutState_NeedsCleanup,
	"needs_repair":  fleet.DutState_NeedsRepair,
	"needs_reset":   fleet.DutState_NeedsReset,
	"repair_failed": fleet.DutState_RepairFailed,
}

// botInfoToSummary initializes fleet.BotSummary for each bot.
//
// This function returns a map from the bot ID to fleet.BotSummary object for it.
func botInfoToSummary(bots []*swarming.SwarmingRpcsBotInfo) (map[string]*fleet.BotSummary, error) {
	bsm := make(map[string]*fleet.BotSummary, len(bots))
	for _, bi := range bots {
		dutID, err := extractSingleValuedDimension(bi, clients.DutIDDimensionKey)
		if err != nil {
			return bsm, errors.Annotate(err, "failed to obtain dutID for bot %q", bi.BotId).Err()
		}
		dutStateStr, err := extractSingleValuedDimension(bi, clients.DutStateDimensionKey)
		if err != nil {
			return bsm, errors.Annotate(err, "failed to obtain DutState for bot %q", bi.BotId).Err()
		}
		dutState, ok := dutStateMap[dutStateStr]
		if !ok {
			dutState = fleet.DutState_DutStateInvalid
		}
		bsm[bi.BotId] = &fleet.BotSummary{
			DutId:    dutID,
			DutState: dutState,
		}
	}
	return bsm, nil
}

// setIdleDuration updates the bot summaries with the duration each bot has been idle.
func setIdleDuration(c context.Context, sc clients.SwarmingClient, bsm map[string]*fleet.BotSummary) error {
	return parallel.WorkPool(clients.MaxConcurrentSwarmingCalls, func(workC chan<- func() error) {
		for bid := range bsm {
			// In-scope variable for goroutine closure.
			bid := bid
			bs := bsm[bid]
			workC <- func() error {
				idle, err := getIdleDuration(c, sc, bid)
				if err != nil {
					return err
				}
				bs.IdleDuration = idle
				return nil
			}
		}
	})
}

// getIdleDuration queries swarming for the duration since last task on the bot.
func getIdleDuration(c context.Context, sc clients.SwarmingClient, botID string) (*duration.Duration, error) {
	trs, err := sc.ListSortedRecentTasksForBot(c, botID, 1)
	if err != nil {
		return nil, errors.Annotate(err, "failed to list recent tasks for bot %s", botID).Err()
	}
	if len(trs) == 0 {
		return nil, nil
	}
	return clients.TimeSinceBotTask(trs[0])
}

// insertBotSummary returns the dut_ids of bots inserted.
func insertBotSummary(c context.Context, bsm map[string]*fleet.BotSummary) ([]string, error) {
	updated := make([]string, 0, len(bsm))
	bses := make([]*fleetBotSummaryEntity, 0, len(bsm))
	for bid, bs := range bsm {
		data, err := proto.Marshal(bs)
		if err != nil {
			return nil, errors.Annotate(err, "failed to marshal BotSummary for dut %q", bs.DutId).Err()
		}
		bses = append(bses, &fleetBotSummaryEntity{
			DutID: bs.DutId,
			BotID: bid,
			Data:  data,
		})
		updated = append(updated, bs.DutId)
	}
	if err := datastore.Put(c, bses); err != nil {
		return nil, errors.Annotate(err, "failed to put BotSummaries").Err()
	}
	return updated, nil
}

func extractSingleValuedDimension(bi *swarming.SwarmingRpcsBotInfo, key string) (string, error) {
	for _, dim := range bi.Dimensions {
		if dim.Key == key {
			switch len(dim.Value) {
			case 1:
				return dim.Value[0], nil
			case 0:
				return "", fmt.Errorf("no value for dimension %s", key)
			default:
				return "", fmt.Errorf("multiple values for dimension %s", key)
			}
		}
	}
	return "", fmt.Errorf("failed to find dimension %s", key)
}

func getBotSummariesFromDatastore(c context.Context, sels []*fleet.BotSelector) ([]*fleetBotSummaryEntity, error) {
	// No selectors implies summarize all bots.
	if len(sels) == 0 {
		bses := []*fleetBotSummaryEntity{}
		q := datastore.NewQuery(botSummaryKind)
		err := datastore.GetAll(c, q, &bses)
		if err != nil {
			return nil, errors.Annotate(err, "failed to get all bots from datastore").Err()
		}
		return bses, nil
	}

	// For now, each selector can only yield 0 or 1 BotSummary.
	bses := make([]*fleetBotSummaryEntity, 0, len(sels))
	for _, s := range sels {
		// datastore rejects search for empty key with InvalidKey.
		// For us, this is simply an impossible filter.
		if s.DutId == "" {
			continue
		}

		bses = append(bses, &fleetBotSummaryEntity{
			DutID: s.DutId,
		})
	}

	if err := datastore.Get(c, bses); err != nil {
		switch err := err.(type) {
		case errors.MultiError:
			return filterNotFoundEntities(bses, err)
		default:
			return nil, err
		}
	}
	return bses, nil
}

func filterNotFoundEntities(bses []*fleetBotSummaryEntity, merr errors.MultiError) ([]*fleetBotSummaryEntity, error) {
	if len(bses) != len(merr) {
		panic(fmt.Sprintf("Length of bot summary (%d) does not match length of multierror (%d)", len(bses), len(merr)))
	}
	filtered := make([]*fleetBotSummaryEntity, 0, len(bses))
	errs := make(errors.MultiError, 0, len(merr))
	for i, bse := range bses {
		err := merr[i]
		if err != nil {
			if !datastore.IsErrNoSuchEntity(err) {
				errs = append(errs, err)
			}
			continue
		}
		filtered = append(filtered, bse)
	}
	if errs.First() != nil {
		return nil, errs
	}
	return filtered, nil
}

func dropDuplicateSelectors(sels []*fleet.BotSelector) []*fleet.BotSelector {
	msels := make(map[string]*fleet.BotSelector, len(sels))
	for _, s := range sels {
		msels[s.DutId] = s
	}
	usels := make([]*fleet.BotSelector, 0, len(msels))
	for _, s := range msels {
		usels = append(usels, s)
	}
	return usels
}
