// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package regulator defines the service main flow.
package regulator

import (
	"context"
	"fmt"
	"strings"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	apipb "go.chromium.org/luci/swarming/proto/api_v2"

	"infra/cros/botsregulator/internal/clients"
	"infra/cros/botsregulator/internal/provider"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

const (
	// Common prefix for machineLSE keys.
	MachineLSEPrefix string = "machineLSEs/"
	// Common prefix for schedulingUnits keys.
	SchedulingUnitsPrefix string = "schedulingunits/"
)

type regulator struct {
	bpiClient      provider.BPI
	opts           *RegulatorOptions
	swarmingClient clients.SwarmingClient
	ufsClient      clients.UFSClient
}

func NewRegulator(ctx context.Context, opts *RegulatorOptions) (*regulator, error) {
	logging.Infof(ctx, "creating regulator with flags: %v\n", opts)
	bc, err := provider.NewProviderFromEnv(ctx, opts.BPI)
	if err != nil {
		return nil, err
	}
	sc, err := clients.NewSwarmingClient(ctx, opts.Swarming)
	if err != nil {
		return nil, err
	}
	uc, err := clients.NewUFSClient(ctx, opts.UFS)
	if err != nil {
		return nil, err
	}
	return &regulator{
		bpiClient:      bc,
		opts:           opts,
		swarmingClient: sc,
		ufsClient:      uc,
	}, nil
}

// FetchLSEsByHive fetches machineLSEs from UFS by hive.
func (r *regulator) FetchLSEsByHive(ctx context.Context) ([]*ufspb.MachineLSE, error) {
	ctx = clients.SetUFSNamespace(ctx, r.opts.Namespace)
	// TODO(b/328443703): Handle pagination. Current max value: 1000.
	res, err := r.ufsClient.ListMachineLSEs(ctx, &ufsAPI.ListMachineLSEsRequest{
		Filter: fmt.Sprintf("hive=%s", r.opts.Hive),
		// KeysOnly returns the entities' ID only. It is faster than a full query.
		KeysOnly: true,
		PageSize: 1000,
	})
	if err != nil {
		return nil, errors.Annotate(err, "could not list machinesLSEs").Err()
	}
	return res.GetMachineLSEs(), nil
}

// FetchAllSchedulingUnits fetches ALL Scheduling Units from UFS.
func (r *regulator) FetchAllSchedulingUnits(ctx context.Context) ([]*ufspb.SchedulingUnit, error) {
	ctx = clients.SetUFSNamespace(ctx, r.opts.Namespace)
	// TODO(b/328443703): Handle pagination. Current max value: 1000.
	res, err := r.ufsClient.ListSchedulingUnits(ctx, &ufsAPI.ListSchedulingUnitsRequest{
		PageSize: 1000,
	})
	if err != nil {
		return nil, err
	}
	return res.GetSchedulingUnits(), nil
}

// ListDroneBots returns list of running Drone Swarming bots.
func (r *regulator) ListDroneBots(ctx context.Context) ([]*apipb.BotInfo, error) {
	cursor := ""
	var bots []*apipb.BotInfo
	for {
		// TODO(b/328810743): The botsRequest dimensions should be derived from flag options to support Satlab flow.
		resp, err := r.swarmingClient.ListBots(ctx, &apipb.BotsRequest{
			Limit:  500,
			Cursor: cursor,
			Dimensions: []*apipb.StringPair{
				{
					Key:   "bot_config",
					Value: "skylab.py",
				},
				{
					Key:   "ufs_zone",
					Value: "ZONE_SFO36_OS",
				},
			},
			IsDead: apipb.NullableBool_FALSE,
		})
		if err != nil {
			return nil, err
		}
		bots = append(bots, resp.Items...)
		cursor = resp.Cursor
		if cursor == "" {
			break
		}
	}
	return bots, nil
}

// ConsolidateAvailableDUTs returns a list of available DUTs to create Swarming bots for.
// This list includes Scheduling Units and single DUTs, all sharing the same hive.
// The assumption is that all LSEs in a Scheduling Unit should share the same hive.
// This is enforced on UFS side.
func (r *regulator) ConsolidateAvailableDUTs(ctx context.Context, dbs []*apipb.BotInfo, lses []*ufspb.MachineLSE, sus []*ufspb.SchedulingUnit) ([]string, error) {
	// List of available DUTs requiring a Swarming bot.
	var ad []string
	// Map of Drone DUTs for easy search.
	droneDUTs := dutMapFromBots(dbs)
	// Map of all lses sharing the same hive (e.g. cloudbots).
	lsesInSU := make(map[string]bool, len(lses))
	for _, lse := range lses {
		l, ok := strings.CutPrefix(lse.GetName(), MachineLSEPrefix)
		if !ok {
			return nil, errors.Reason("could not parse LSE name: %v", lse).Err()
		}
		lsesInSU[l] = false
	}
	// Filtering SUs by hive.
	for _, su := range sus {
		seen := false
		for _, lse := range su.GetMachineLSEs() {
			if _, ok := lsesInSU[lse]; ok {
				lsesInSU[lse] = true
				seen = true
			}
		}
		// At least 1 DUT in the SU has the corresponding hive.
		if seen {
			s, ok := strings.CutPrefix(su.GetName(), SchedulingUnitsPrefix)
			if !ok {
				return nil, errors.Reason("could not parse SU name: %v", su).Err()
			}
			// The SU is still running on Drone.
			if _, ok := droneDUTs[s]; ok {
				logging.Infof(ctx, "Scheduling Unit %s is still running on Drone; skipping")
				continue
			}
			ad = append(ad, s)
		}
	}
	for lse, seen := range lsesInSU {
		// The DUT is part of a scheduling unit.
		if seen {
			continue
		}
		// The DUT is still running on Drone.
		if _, ok := droneDUTs[lse]; ok {
			logging.Infof(ctx, "DUT %s is still running on Drone; skipping")
			continue
		}
		ad = append(ad, lse)
	}
	return ad, nil
}

// UpdateConfig is a wrapper around the current provider UpdateConfig method.
func (r *regulator) UpdateConfig(ctx context.Context, hns []string) error {
	return r.bpiClient.UpdateConfig(ctx, hns, r.opts.CfID)
}

// dutMapFromBots return a map of DUT name from a list of Swarming bots.
func dutMapFromBots(dbs []*apipb.BotInfo) map[string]struct{} {
	duts := make(map[string]struct{}, len(dbs))
	for _, db := range dbs {
		for _, d := range db.GetDimensions() {
			if d.Key == "dut_name" {
				duts[d.Value[0]] = struct{}{}
				break
			}
		}
	}
	return duts
}
