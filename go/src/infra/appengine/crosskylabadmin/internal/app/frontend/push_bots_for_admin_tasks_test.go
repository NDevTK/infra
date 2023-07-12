// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"testing"

	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/data/strpair"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/clients"
	"infra/appengine/crosskylabadmin/internal/app/config"
)

func TestPushBotsForAdminTasksImplSmokeTest(t *testing.T) {
	tf, validate := newTestFixture(t)
	defer validate()
	ctx := tf.C
	ctx = config.Use(ctx, &config.Config{
		Swarming: &config.Swarming{
			BotPool: "fake-bot-pool",
		},
	})
	swarmingClient := &fakeSwarmingClient{}
	req := &fleet.PushBotsForAdminTasksRequest{
		TargetDutState: fleet.DutState_NeedsRepair,
	}
	_, err := pushBotsForAdminTasksImpl(ctx, swarmingClient, req)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

type fakeSwarmingClient struct{}

// Guarantee that fakeSwarmingClient satisfies the SwarmingClient interface.
var a clients.SwarmingClient = &fakeSwarmingClient{};

func (sc *fakeSwarmingClient) ListAliveIdleBotsInPool(c context.Context, pool string, dims strpair.Map) ([]*swarming.SwarmingRpcsBotInfo, error) {
	return []*swarming.SwarmingRpcsBotInfo{
		{
			BotId: "fake-bot-1",
		},
		{
			BotId: "fake-bot-2",
		},
	}, nil
}

func (sc *fakeSwarmingClient) ListAliveBotsInPool(context.Context, string, strpair.Map) ([]*swarming.SwarmingRpcsBotInfo, error) {
	panic("ListAliveBotsInPool")
}

func (sc *fakeSwarmingClient) ListBotTasks(id string) clients.BotTasksCursor {
	panic("ListBotTasks")
}

func (sc *fakeSwarmingClient) ListRecentTasks(c context.Context, tags []string, state string, limit int) ([]*swarming.SwarmingRpcsTaskResult, error) {
	panic("ListRecentTasks")
}

func (sc *fakeSwarmingClient) ListSortedRecentTasksForBot(c context.Context, botID string, limit int) ([]*swarming.SwarmingRpcsTaskResult, error) {
	panic("ListSortedRecentTasksForBot")
}

func (sc *fakeSwarmingClient) CreateTask(c context.Context, name string, args *clients.SwarmingCreateTaskArgs) (string, error) {
	panic("CreateTask")
}

func (sc *fakeSwarmingClient) GetTaskResult(ctx context.Context, tid string) (*swarming.SwarmingRpcsTaskResult, error) {
	panic("GetTaskResult")
}
