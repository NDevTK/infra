// Copyright 2018 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"net/http"
	"time"

	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/data/strpair"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/retry"
	"go.chromium.org/luci/grpc/grpcutil"
	"go.chromium.org/luci/server/auth"
	swarmingv2 "go.chromium.org/luci/swarming/proto/api_v2"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/clients"
	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/app/frontend/util"
	"infra/appengine/crosskylabadmin/internal/ufs"
	"infra/appengine/crosskylabadmin/site"
	"infra/cros/lab_inventory/utilization"
	"infra/cros/recovery/karte"
	"infra/cros/recovery/logger/metrics"
)

// SwarmingFactory is a constructor for a SwarmingClient.
type SwarmingFactory func(c context.Context, host string) (clients.SwarmingClient, error)

// TrackerServerImpl implements the fleet.TrackerServer interface.
type TrackerServerImpl struct {
	// SwarmingFactory is an optional factory function for creating clients.
	//
	// If SwarmingFactory is nil, clients.NewSwarmingClient is used.
	SwarmingFactory SwarmingFactory
	MetricsClient   metrics.Metrics
}

func (tsi *TrackerServerImpl) newSwarmingClient(c context.Context, host string) (clients.SwarmingClient, error) {
	if tsi.SwarmingFactory != nil {
		return tsi.SwarmingFactory(c, host)
	}
	return clients.NewSwarmingClient(c, host)
}

func (tsi *TrackerServerImpl) getKarteClient(ctx context.Context) (metrics.Metrics, error) {
	if tsi.MetricsClient != nil {
		return tsi.MetricsClient, nil
	}
	cfg := config.Get(ctx)
	// Create the Karte client
	transport, err := auth.GetRPCTransport(ctx, auth.AsSelf)
	if err != nil {
		return nil, errors.Annotate(err, "failed to get RPC transport").Err()
	}
	kClient, err := karte.NewMetricsWithHttp(ctx, &http.Client{
		Transport: transport,
	}, cfg.GetKarte().GetHost(), site.DefaultPRPCOptions)
	if err != nil {
		return nil, err
	}
	tsi.MetricsClient = kClient
	return kClient, nil
}

// PushBotsForAdminTasks implements the fleet.Tracker.pushBotsForAdminTasks() method.
func (tsi *TrackerServerImpl) PushBotsForAdminTasks(ctx context.Context, req *fleet.PushBotsForAdminTasksRequest) (res *fleet.PushBotsForAdminTasksResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()

	cfg := config.Get(ctx)
	sc, err := tsi.newSwarmingClient(ctx, cfg.Swarming.Host)
	if err != nil {
		return nil, errors.Annotate(err, "failed to obtain Swarming client").Err()
	}

	httpClient, err := ufs.NewHTTPClient(ctx)
	if err != nil {
		logging.Errorf(ctx, "error setting up UFS client: %s", err)
	}
	ufsClient, err := ufs.NewClient(ctx, httpClient, cfg.GetUFS().GetHost())
	if err != nil {
		logging.Errorf(ctx, "error setting up UFS client: %s", err)
	}
	metricsClient, err := tsi.getKarteClient(ctx)
	if err != nil {
		logging.Errorf(ctx, "error setting up Karte client: %s", err)
	}

	p := adminTaskBotPusher{
		ufsClient:      ufsClient,
		swarmingClient: sc,
		metricsClient:  metricsClient,
	}
	return p.pushBotsForAdminTasksImpl(ctx, req)
}

// PushBotsForAdminAuditTasks implements the fleet.Tracker.pushBotsForAdminTasks() method.
func (tsi *TrackerServerImpl) PushBotsForAdminAuditTasks(ctx context.Context, req *fleet.PushBotsForAdminAuditTasksRequest) (res *fleet.PushBotsForAdminAuditTasksResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()

	dutStates := map[fleet.DutState]bool{
		fleet.DutState_Ready:             true,
		fleet.DutState_NeedsRepair:       true,
		fleet.DutState_NeedsReset:        true,
		fleet.DutState_RepairFailed:      true,
		fleet.DutState_NeedsManualRepair: true,
		fleet.DutState_NeedsReplacement:  false,
		fleet.DutState_NeedsDeploy:       false,
	}

	var actions []string
	var taskname string
	switch req.Task {
	case fleet.AuditTask_ServoUSBKey:
		actions = []string{"verify-servo-usb-drive"}
		taskname = "USB-drive"
	case fleet.AuditTask_DUTStorage:
		actions = []string{"verify-dut-storage"}
		taskname = "Storage"
		dutStates[fleet.DutState_RepairFailed] = false
		dutStates[fleet.DutState_NeedsManualRepair] = false
	case fleet.AuditTask_RPMConfig:
		actions = []string{"verify-rpm-config"}
		taskname = "RPM Config"
		dutStates[fleet.DutState_RepairFailed] = false
		dutStates[fleet.DutState_NeedsManualRepair] = false
	}

	if len(actions) == 0 {
		logging.Infof(ctx, "No action specified", err)
		return nil, errors.New("failed to push audit bots")
	}

	scheduleTasks := func(swarmingHost, swarmingPool string) error {
		sc, err := tsi.newSwarmingClient(ctx, swarmingHost)
		if err != nil {
			return errors.Annotate(err, "failed to obtain Swarming client").Err()
		}
		// Schedule audit tasks to ready|needs_repair|needs_reset|repair_failed DUTs.
		var bots []*swarming.SwarmingRpcsBotInfo
		f := func() (err error) {
			dims := make(strpair.Map)
			bots, err = sc.ListAliveBotsInPool(ctx, swarmingPool, dims)
			return err
		}
		err = retry.Retry(ctx, simple3TimesRetry(), f, retry.LogCallback(ctx, "Try get list of the BOTs"))
		if err != nil {
			return errors.Annotate(err, "failed to list alive cros bots").Err()
		}
		logging.Infof(ctx, "successfully get %d alive cros bots", len(bots))
		botIDs := identifyBotsForAudit(ctx, bots, dutStates, req.Task)

		err = clients.PushAuditDUTs(ctx, botIDs, actions, taskname)
		if err != nil {
			logging.Infof(ctx, "failed push audit bots: %v", err)
			return errors.Reason("failed to push audit bots").Err()
		}
		return nil
	}
	cfg := config.Get(ctx)
	var errs []error
	for _, pool := range cfg.GetSwarming().GetPoolCfgs() {
		if !pool.GetAuditEnabled() {
			logging.Infof(ctx, "Audit is not enabled for %q.", pool.GetPoolName())
			continue
		}
		if err := scheduleTasks(cfg.GetSwarming().GetBotPool(), pool.GetPoolName()); err != nil {
			logging.Errorf(ctx, "Audit for %q failed: %s.", pool.GetPoolName(), err)
			errs = append(errs, errors.Annotate(err, "schedule tasks for %q", pool.GetPoolName()).Err())
		} else {
			logging.Infof(ctx, "Audit for %q succesful scheduled.", pool.GetPoolName())
		}
	}
	if len(errs) > 0 {
		return nil, errors.NewMultiError(errs...).AsError()
	}
	return &fleet.PushBotsForAdminAuditTasksResponse{}, nil
}

// PushRepairJobsForLabstations implements the fleet.Tracker.pushLabstationsForRepair() method.
func (tsi *TrackerServerImpl) PushRepairJobsForLabstations(ctx context.Context, req *fleet.PushRepairJobsForLabstationsRequest) (res *fleet.PushRepairJobsForLabstationsResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()

	cfg := config.Get(ctx)
	sc, err := tsi.newSwarmingClient(ctx, cfg.Swarming.Host)
	if err != nil {
		return nil, errors.Annotate(err, "failed to obtain Swarming client").Err()
	}

	// Schedule repair jobs to idle labstations. It's for periodically checking
	// and rebooting labstations to ensure they're in good state.
	dims := make(strpair.Map)
	dims[clients.DutOSDimensionKey] = []string{"OS_TYPE_LABSTATION"}
	bots, err := sc.ListAliveIdleBotsInPool(ctx, cfg.GetSwarming().GetBotPool(), dims)
	if err != nil {
		return nil, errors.Annotate(err, "failed to list alive idle labstation bots").Err()
	}
	logging.Infof(ctx, "successfully get %d alive idle labstation bots.", len(bots))

	// Parse BOT id to schedule tasks for readability.
	botIDs := identifyLabstationsForRepair(ctx, bots)

	err = clients.PushRepairLabstations(ctx, botIDs)
	if err != nil {
		logging.Infof(ctx, "push repair labstations: %v", err)
		return nil, errors.New("failed to push repair labstations")
	}
	return &fleet.PushRepairJobsForLabstationsResponse{}, nil
}

// ReportBots reports metrics of swarming bots.
func (tsi *TrackerServerImpl) ReportBots(ctx context.Context, req *fleet.ReportBotsRequest) (res *fleet.ReportBotsResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()
	cfg := config.Get(ctx)
	sc, err := tsi.newSwarmingClient(ctx, cfg.Swarming.Host)
	if err != nil {
		return nil, errors.Annotate(err, "failed to obtain Swarming client").Err()
	}

	bots, err := sc.ListAliveBotsInPool(ctx, cfg.Swarming.BotPool, strpair.Map{})
	utilization.ReportMetrics(ctx, flattenAndDedpulicateBots([][]*swarming.SwarmingRpcsBotInfo{bots}))
	return &fleet.ReportBotsResponse{}, nil
}

func flattenAndDedpulicateBots(nb [][]*swarming.SwarmingRpcsBotInfo) []*swarming.SwarmingRpcsBotInfo {
	bm := make(map[string]*swarming.SwarmingRpcsBotInfo)
	for _, bs := range nb {
		for _, b := range bs {
			bm[b.BotId] = b
		}
	}
	bots := make([]*swarming.SwarmingRpcsBotInfo, 0, len(bm))
	for _, v := range bm {
		bots = append(bots, v)
	}
	return bots
}

var dutStatesForRepairTask = map[fleet.DutState]bool{
	fleet.DutState_NeedsRepair:       true,
	fleet.DutState_RepairFailed:      true,
	fleet.DutState_NeedsManualRepair: true,
}

// identifyBotsForRepair identifies duts that need run admin repair.
func identifyBotsForRepair(ctx context.Context, bots []*swarmingv2.BotInfo) (repairBOTs []string) {
	repairBOTs = make([]string, 0, len(bots))
	for _, b := range bots {
		dims := util.DimensionsMapV2(b.Dimensions)
		os, err := util.ExtractSingleValuedDimension(dims, clients.DutOSDimensionKey)
		// Some bot may not have os dimension(e.g. scheduling unit), so we ignore the error here.
		if err == nil && os == "OS_TYPE_LABSTATION" {
			continue
		}
		id, err := util.ExtractSingleValuedDimension(dims, clients.BotIDDimensionKey)
		if err != nil {
			logging.Warningf(ctx, "failed to obtain BOT id for bot %q", b.BotId)
			continue
		}

		s := clients.GetStateDimensionV2(b.GetDimensions())
		if dutStatesForRepairTask[s] {
			logging.Infof(ctx, "BOT: %s - Needs repair", id)
			repairBOTs = append(repairBOTs, id)
		}
	}
	return repairBOTs
}

// identifyBotsForAudit identifies duts to run admin audit.
func identifyBotsForAudit(ctx context.Context, bots []*swarming.SwarmingRpcsBotInfo, dutStateMap map[fleet.DutState]bool, auditTask fleet.AuditTask) []string {
	logging.Infof(ctx, "Filtering bots for task: %s", auditTask)
	botIDs := make([]string, 0, len(bots))
	for _, b := range bots {
		dims := util.DimensionsMap(b.Dimensions)
		os, err := util.ExtractSingleValuedDimension(dims, clients.DutOSDimensionKey)
		// Some bot may not have os dimension(e.g. scheduling unit), so we ignore the error here.
		if err == nil && os == "OS_TYPE_LABSTATION" {
			continue
		}

		// TODO(xixuan): b/243448732, remove this check after VM prototype
		model, err := util.ExtractSingleValuedDimension(dims, clients.DutModelDimensionKey)
		// Exclude betty bots for audit
		if err == nil && model == "betty" {
			continue
		}

		id, err := util.ExtractSingleValuedDimension(dims, clients.BotIDDimensionKey)
		if err != nil {
			logging.Warningf(ctx, "failed to obtain BOT id for bot %q", b.BotId)
			continue
		}
		switch auditTask {
		case fleet.AuditTask_ServoUSBKey:
			// Disable skip to verify flakiness. (b/229656121)
			// state := swarming_utils.ExtractBotState(b).ServoUSBState
			// if len(state) > 0 && state[0] == "NEED_REPLACEMENT" {
			// 	logging.Infof(ctx, "Skipping BOT with id: %q as USB-key marked for replacement", b.BotId)
			// 	continue
			// }
		case fleet.AuditTask_DUTStorage:
			state := util.ExtractBotState(b).StorageState
			if len(state) > 0 && state[0] == "NEED_REPLACEMENT" {
				logging.Infof(ctx, "Skipping BOT with id: %q as storage marked for replacement", b.BotId)
				continue
			}
		case fleet.AuditTask_RPMConfig:
			state := util.ExtractBotState(b).RpmState
			if len(state) > 0 && state[0] != "UNKNOWN" {
				// expecting that RPM is going through check everytime when we do any update on setup.
				logging.Infof(ctx, "Skipping BOT with id: %q as RPM was already audited", b.BotId)
				continue
			}
		}

		s := clients.GetStateDimension(b.Dimensions)
		if v, ok := dutStateMap[s]; ok && v {
			botIDs = append(botIDs, id)
		} else {
			logging.Infof(ctx, "Skipping BOT with id: %q", b.BotId)
		}
	}
	return botIDs
}

// identifyLabstationsForRepair identifies labstations that need repair.
func identifyLabstationsForRepair(ctx context.Context, bots []*swarmingv2.BotInfo) []string {
	botIDs := make([]string, 0, len(bots))
	for _, b := range bots {
		dims := util.DimensionsMapV2(b.GetDimensions())
		os, err := util.ExtractSingleValuedDimension(dims, clients.DutOSDimensionKey)
		if err != nil {
			logging.Warningf(ctx, "failed to obtain os type for bot %q", b.BotId)
			continue
		} else if os != "OS_TYPE_LABSTATION" {
			continue
		}

		id, err := util.ExtractSingleValuedDimension(dims, clients.BotIDDimensionKey)
		if err != nil {
			logging.Warningf(ctx, "failed to obtain BOT id for bot %q", b.BotId)
			continue
		}

		botIDs = append(botIDs, id)
	}
	return botIDs
}

// simple3TimesRetryIterator simple retry iterator to try 3 times.
var simple3TimesRetryIterator = retry.ExponentialBackoff{
	Limited: retry.Limited{
		Delay:   200 * time.Millisecond,
		Retries: 3,
	},
}

// simple3TimesRetry returns a retry.Factory based on simple3TimesRetryIterator.
func simple3TimesRetry() retry.Factory {
	return func() retry.Iterator {
		return &simple3TimesRetryIterator
	}
}
