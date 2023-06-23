// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package zone_selector

import (
	"context"
	"math/rand"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"

	"infra/vm_leaser/internal/constants"
)

// SelectZone selects a random zone based on the specified testing client.
func SelectZone(ctx context.Context, r *api.LeaseVMRequest, seed int64) string {
	// Call Seed once to seed any subsequent rand calls.
	rand.Seed(seed)

	if r.GetHostReqs().GetGceRegion() != "" {
		return r.GetHostReqs().GetGceRegion()
	}
	switch r.GetTestingClient() {
	case api.VMTestingClient_VM_TESTING_CLIENT_CHROMEOS:
		logging.Infof(ctx, "selecting random zone for ChromeOS testing client")
		return getRandomZone(ctx, constants.ChromeOSZones)
	default:
		logging.Infof(ctx, "selecting random zone for unspecified testing client")
		return getRandomZone(ctx, constants.ChromeOSZones)
	}
}

// getRandomZone takes an array of arrays of zones and returns a random one.
func getRandomZone(ctx context.Context, zones [][]string) string {
	mainIdx := rand.Intn(len(zones))
	subIdx := rand.Intn(len(zones[mainIdx]))
	logging.Infof(ctx, "selected zone for VM creation: %v", zones[mainIdx][subIdx])
	return zones[mainIdx][subIdx]
}
