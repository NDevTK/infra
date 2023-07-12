// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package fakes

import (
	"context"

	"go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/data/strpair"

	"infra/appengine/crosskylabadmin/internal/app/clients"
)

type SwarmingClient struct{}

// Guarantee that SwarmingClient satisfies the SwarmingClient interface.
var _ clients.SwarmingClient = &SwarmingClient{}

func (sc *SwarmingClient) ListAliveIdleBotsInPool(c context.Context, pool string, dims strpair.Map) ([]*swarming.SwarmingRpcsBotInfo, error) {
	return []*swarming.SwarmingRpcsBotInfo{
		{
			BotId: "fake-bot-1",
		},
		{
			BotId: "fake-bot-2",
		},
	}, nil
}

func (sc *SwarmingClient) ListAliveBotsInPool(context.Context, string, strpair.Map) ([]*swarming.SwarmingRpcsBotInfo, error) {
	panic("ListAliveBotsInPool")
}

func (sc *SwarmingClient) ListBotTasks(id string) clients.BotTasksCursor {
	panic("ListBotTasks")
}

func (sc *SwarmingClient) ListRecentTasks(c context.Context, tags []string, state string, limit int) ([]*swarming.SwarmingRpcsTaskResult, error) {
	panic("ListRecentTasks")
}

func (sc *SwarmingClient) ListSortedRecentTasksForBot(c context.Context, botID string, limit int) ([]*swarming.SwarmingRpcsTaskResult, error) {
	panic("ListSortedRecentTasksForBot")
}

func (sc *SwarmingClient) CreateTask(c context.Context, name string, args *clients.SwarmingCreateTaskArgs) (string, error) {
	panic("CreateTask")
}

func (sc *SwarmingClient) GetTaskResult(ctx context.Context, tid string) (*swarming.SwarmingRpcsTaskResult, error) {
	panic("GetTaskResult")
}
