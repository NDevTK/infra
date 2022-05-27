// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/auth"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/clients"
	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/app/frontend/routing"
	"infra/appengine/crosskylabadmin/internal/app/frontend/worker"
	"infra/appengine/crosskylabadmin/site"
	"infra/cros/recovery/tasknames"
	"infra/libs/skylab/buildbucket"
	"infra/libs/skylab/buildbucket/labpack"
	"infra/libs/skylab/common/heuristics"
)

// UFSErrorPolicy controls how UFS errors are handled.
type ufsErrorPolicy string

// UFS error policy constants.
// Error policy constants are defined in go/src/infra/appengine/crosskylabadmin/app/config/config.proto.
//
// Strict   -- fail on UFS error even if we don't need the result
// Fallback -- if we encounter a UFS error, fall back to the legacy path.
// Lax      -- if we do not need the UFS response to make a decision, do not fail the request.
const (
	// The strict policy causes all UFS error requests to be treated as fatal and causes the request to fail.
	ufsErrorPolicyStrict   ufsErrorPolicy = "strict"
	ufsErrorPolicyFallback                = "fallback"
	ufsErrorPolicyLax                     = "lax"
)

// NormalizeError policy normalizes a string into the canonical name for a policy.
func normalizeErrorPolicy(policy string) (ufsErrorPolicy, error) {
	policy = strings.ToLower(policy)
	switch policy {
	case "", "default", "fallback":
		return ufsErrorPolicyFallback, nil
	case "strict":
		return ufsErrorPolicyStrict, nil
	case "lax":
		return ufsErrorPolicyLax, nil
	}
	return "", fmt.Errorf("unrecognized policy: %q", policy)
}

// getRolloutConfig gets the applicable rolloutConfig.
func getRolloutConfig(ctx context.Context, taskType string, isLabstation bool, expectedState string) (*config.RolloutConfig, error) {
	if taskType == "" {
		return nil, errors.Reason("get rollout config: taskType cannot be empty").Err()
	}
	if taskType != "repair" {
		return nil, errors.Reason("getRolloutConfig: tasks other than repair are not supported, %q given", taskType).Err()
	}
	if isLabstation {
		return config.Get(ctx).GetParis().GetLabstationRepair(), nil
	}
	if expectedState == "" {
		return nil, errors.Reason("get rollout config: expectedState cannot be empty").Err()
	}
	switch expectedState {
	case "ready":
		return nil, errors.Reason("get rollout config: refusing to schedule repair task on ready dut").Err()
	case "needs_repair":
		return config.Get(ctx).GetParis().GetDutRepair(), nil
	case "repair_failed":
		return config.Get(ctx).GetParis().GetDutRepairOnRepairFailed(), nil
	case "needs_manual_repair":
		return config.Get(ctx).GetParis().GetDutRepairOnNeedsManualRepair(), nil
	}
	return nil, errors.Reason("get rollout config: expected state %q is not recognized", expectedState).Err()
}

// CreateRepairTask kicks off a repair job.
//
// This function will either schedule a legacy repair task or a PARIS repair task.
// Note that the ufs client can be nil.
func CreateRepairTask(ctx context.Context, botID string, expectedState string, pools []string, randFloat float64) (string, error) {
	logging.Infof(ctx, "Creating repair task for %q expected state %q with random input %f", botID, expectedState, randFloat)
	// If we encounter an error picking paris or legacy, do the safe thing and use legacy.
	taskType, err := RouteTask(ctx, "repair", botID, expectedState, pools, randFloat)
	if err != nil {
		logging.Infof(ctx, "Create repair task: falling back to legacy repair by default: %s", err)
	}
	switch taskType {
	case "paris", "latest":
		url, err := createBuildbucketRepairTask(ctx, createBuildBucketRepairTaskRequest{
			taskType:      labpack.TaskMap[taskType],
			botID:         botID,
			expectedState: expectedState,
		})
		if err != nil {
			logging.Errorf(ctx, "Attempted and failed to create buildbucket task: %s", err)
			logging.Errorf(ctx, "Falling back to legacy flow")
			url, err = createLegacyRepairTask(ctx, botID, expectedState)
			return url, errors.Annotate(err, "fallback legacy repair task somehow failed").Err()
		}
		return url, err
	default:
		return createLegacyRepairTask(ctx, botID, expectedState)
	}
}

// DUTRoutingInfo is all the deterministic information about a DUT that is necessary to decide
// whether to use a legacy task or a paris task.
//
// We need to know whether a DUT is a labstation or not.
// We also need to know its hostname so we can choose the pattern stanza that applies to it.
type dutRoutingInfo struct {
	hostname   string
	labstation bool
	pools      []string
}

// RouteLabstationRepairTask takes a repair task for a labstation and routes it.
func routeRepairTaskImpl(ctx context.Context, r *config.RolloutConfig, info *dutRoutingInfo, randFloat float64) (string, routing.Reason) {
	if info == nil {
		logging.Errorf(ctx, "info cannot be nil, falling back to legacy")
		return routing.Legacy, routing.NilArgument
	}
	// TODO(gregorynisbet): Log instead of silently falling back to the default error handling policy.
	ufsErrorPolicy, err := normalizeErrorPolicy(r.GetUfsErrorPolicy())
	if err != nil {
		logging.Infof(ctx, "error while routing labstation repair task: %s", err)
	}
	// Check that the feature is enabled at all.
	if !r.GetEnable() {
		return routing.Legacy, routing.ParisNotEnabled
	}
	// Check for malformed input data that would cause us to be unable to make a decision.
	if len(info.pools) == 0 {
		switch ufsErrorPolicy {
		case ufsErrorPolicyLax:
			// Do nothing.
		case ufsErrorPolicyStrict:
			// TODO(gregorynisbet): Make strict error handling propagate the failure back up.
			return routing.Legacy, routing.NoPools
		case ufsErrorPolicyFallback:
			return routing.Legacy, routing.NoPools
		default:
			return routing.Legacy, routing.MalformedPolicy
		}
	}

	d, err := r.ComputePermilleData(ctx, info.hostname)
	if err != nil {
		return routing.Legacy, routing.ErrorExtractingPermilleInfo
	}

	// threshold is the chance of using Paris at all, which is equal to prod + latest.
	threshold := d.Prod + d.Latest
	// latestThreshold is a smaller threshold for using latest specifically.
	latestThreshold := d.Latest
	myValue := math.Round(1000.0 * randFloat)
	// If the threshold is zero, let's reject all possible values of myValue.
	// This way a threshold of zero actually means 0.0% instead of 0.1%.
	valueBelowThreshold := threshold != 0 && myValue <= threshold
	valueBelowLatestThreshold := latestThreshold != 0 && myValue <= latestThreshold
	if r.GetOptinAllDuts() {
		switch {
		case valueBelowLatestThreshold:
			return routing.ParisLatest, routing.ScoreBelowThreshold
		case valueBelowThreshold:
			return routing.Paris, routing.ScoreBelowThreshold
		default:
			return routing.Legacy, routing.ScoreTooHigh
		}
	}
	if threshold == 0 {
		return routing.Legacy, routing.ThresholdZero
	}
	if !r.GetOptinAllDuts() && len(r.GetOptinDutPool()) > 0 && isDisjoint(info.pools, r.GetOptinDutPool()) {
		return routing.Legacy, routing.WrongPool
	}
	switch {
	case valueBelowLatestThreshold:
		return routing.ParisLatest, routing.ScoreBelowThreshold
	case valueBelowThreshold:
		return routing.Paris, routing.ScoreBelowThreshold
	default:
		return routing.Legacy, routing.ScoreTooHigh
	}
}

// createBuildBucketRepairTaskRequest consists of the parameters needed to schedule a buildbucket repair task.
type createBuildBucketRepairTaskRequest struct {
	taskType labpack.CIPDVersion
	// botID is the ID of the bot, for example, "crossk-chromeos...".
	botID         string
	expectedState string
}

// CreateBuildbucketRepairTask creates a new repair task for a labstation.
// Err should be non-nil if and only if a task was created.
// We rely on this signal to decide whether to fall back to the legacy flow.
func createBuildbucketRepairTask(ctx context.Context, params createBuildBucketRepairTaskRequest) (string, error) {
	if err := params.taskType.Validate(); err != nil {
		return "", errors.Annotate(err, "create buildbucket repair task: invalid task type %v", params.taskType).Err()
	}
	logging.Infof(ctx, "Using new repair flow for bot %q with expected state %q", params.botID, params.expectedState)
	transport, err := auth.GetRPCTransport(ctx, auth.AsSelf)
	if err != nil {
		return "", errors.Annotate(err, "failed to get RPC transport").Err()
	}
	hc := &http.Client{
		Transport: transport,
	}
	bc, err := buildbucket.NewClient2(ctx, hc, site.DefaultPRPCOptions, "chromeos", "labpack", "labpack")
	if err != nil {
		logging.Errorf(ctx, "error creating buildbucket client: %q", err)
		return "", errors.Annotate(err, "create buildbucket repair task").Err()
	}
	p := &labpack.Params{
		UnitName:       heuristics.NormalizeBotNameToDeviceName(params.botID),
		TaskName:       string(tasknames.Recovery),
		EnableRecovery: true,
		// TODO(gregorynisbet): This is our own name, move it to the config.
		AdminService: "chromeos-skylab-bot-fleet.appspot.com",
		// NOTE: We use the UFS service, not the Inventory service here.
		InventoryService: config.Get(ctx).GetUFS().GetHost(),
		NoStepper:        false,
		NoMetrics:        false,
		UpdateInventory:  true,
		// TODO(gregorynisbet): Pass config file to labpack task.
		Configuration: "",
	}
	taskID, err := labpack.ScheduleTask(ctx, bc, params.taskType, p)
	if err != nil {
		logging.Errorf(ctx, "error scheduling task: %q", err)
		return "", errors.Annotate(err, "create buildbucket repair task").Err()
	}
	return bc.BuildURL(taskID), nil
}

// CreateLegacyRepairTask creates a legacy repair task for a labstation.
func createLegacyRepairTask(ctx context.Context, botID string, expectedState string) (string, error) {
	logging.Infof(ctx, "Using legacy repair flow for bot %q", botID)
	at := worker.AdminTaskForType(ctx, fleet.TaskType_Repair)
	sc, err := clients.NewSwarmingClient(ctx, config.Get(ctx).Swarming.Host)
	if err != nil {
		return "", errors.Annotate(err, "failed to obtain swarming client").Err()
	}
	cfg := config.Get(ctx)
	taskURL, err := runTaskByBotID(ctx, at, sc, botID, expectedState, cfg.Tasker.BackgroundTaskExpirationSecs, cfg.Tasker.BackgroundTaskExecutionTimeoutSecs)
	if err != nil {
		return "", errors.Annotate(err, "fail to create repair task for %s", botID).Err()
	}
	return taskURL, nil
}

// IsDisjoint returns true if and only if two sequences have no elements in common.
func isDisjoint(a []string, b []string) bool {
	bMap := make(map[string]bool, len(b))
	for _, item := range b {
		bMap[item] = true
	}
	for _, item := range a {
		if bMap[item] {
			return false
		}
	}
	return true
}
