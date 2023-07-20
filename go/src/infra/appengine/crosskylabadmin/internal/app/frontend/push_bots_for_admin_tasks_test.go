// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"testing"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/config"

	"github.com/golang/mock/gomock"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
)

func TestPushBotsForAdminTasksImplSmokeTest(t *testing.T) {
	tf, validate := newTestFixture(t)
	defer validate()
	ctx := tf.C
	tf.MockSwarming.EXPECT().ListAliveIdleBotsInPool(gomock.Any(), gomock.Any(), gomock.Any())
	ctx = config.Use(ctx, &config.Config{
		Swarming: &config.Swarming{
			BotPool: "fake-bot-pool",
		},
	})
	req := &fleet.PushBotsForAdminTasksRequest{
		TargetDutState: fleet.DutState_NeedsRepair,
	}
	_, err := pushBotsForAdminTasksImpl(ctx, tf.MockSwarming, nil, req)

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestPushBotsForAdminTasksWithUFSClient(t *testing.T) {
	tf, validate := newTestFixture(t)
	defer validate()
	ctx := tf.C
	tf.MockSwarming.EXPECT().ListAliveIdleBotsInPool(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*swarming.SwarmingRpcsBotInfo{
		{
			BotId: "fake-bot-a",
			Dimensions: []*swarming.SwarmingRpcsStringListPair{
				{
					Key:   "pool",
					Value: []string{"fake-bot-pool"},
				},
			},
		},
		{
			BotId: "fake-bot-b",
			Dimensions: []*swarming.SwarmingRpcsStringListPair{
				{
					Key:   "pool",
					Value: []string{"fake-bot-pool"},
				},
			},
		},
	}, nil)
	ctx = config.Use(ctx, &config.Config{
		Swarming: &config.Swarming{
			BotPool: "fake-bot-pool",
		},
	})
	req := &fleet.PushBotsForAdminTasksRequest{
		TargetDutState: fleet.DutState_NeedsRepair,
	}
	_, err := pushBotsForAdminTasksImpl(ctx, tf.MockSwarming, tf.MockUFS, req)

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
