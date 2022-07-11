// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package queries

import (
	"context"
	"sync"

	"go.chromium.org/luci/common/tsmon"
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"

	"infra/appengine/drone-queen/internal/entities"
)

var (
	retryUniqueUUID = metric.NewCounter(
		"chromeos/drone-queen/irregular-event/retry-unique-uuid",
		"retry when attempting to generate a unique drone UUID",
		nil,
		field.String("instance"),
	)
	agentCount = metric.NewInt(
		"chromeos/drone-queen/agent/count",
		"count of drone agents",
		nil,
		field.String("hive"),
	)
	agentCapacity = metric.NewInt(
		"chromeos/drone-queen/agent/capacity",
		"capacity of drone agents",
		nil,
		field.String("hive"),
		field.String("type"),
	)
	// agentLoadTracker tracks the load of all agents.
	agentLoadTracker     = make(map[entities.DroneID]agentLoad)
	agentLoadTrackerLock = sync.Mutex{}
)

// agentLoad is a struct to record drone agent load info.
type agentLoad struct {
	hive          string
	totalCapacity int
	usedCapacity  int
}

func init() {
	tsmon.RegisterCallback(func(ctx context.Context) {
		freeCapacityByHive := make(map[string]int)
		usedCapacityByHive := make(map[string]int)
		agentCountByHive := make(map[string]int)

		agentLoadTrackerLock.Lock()
		for _, al := range agentLoadTracker {
			freeCapacityByHive[al.hive] += al.totalCapacity - al.usedCapacity
			usedCapacityByHive[al.hive] += al.usedCapacity
			agentCountByHive[al.hive]++
		}
		agentLoadTrackerLock.Unlock()

		for k, v := range freeCapacityByHive {
			agentCapacity.Set(ctx, int64(v), k, "free")
		}
		for k, v := range usedCapacityByHive {
			agentCapacity.Set(ctx, int64(v), k, "used")
		}
		for k, v := range agentCountByHive {
			agentCount.Set(ctx, int64(v), k)
		}
	})
}

func updateAgentLoad(d entities.DroneID, l agentLoad) {
	agentLoadTrackerLock.Lock()
	defer agentLoadTrackerLock.Unlock()
	agentLoadTracker[d] = l
}

func deleteAgentLoad(d entities.DroneID) {
	agentLoadTrackerLock.Lock()
	defer agentLoadTrackerLock.Unlock()
	delete(agentLoadTracker, d)
}
