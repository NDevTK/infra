// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/cros/recovery/logger/metrics"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
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
	_, err := pushBotsForAdminTasksImpl(ctx, tf.MockSwarming, nil, nil, req)

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
	_, err := pushBotsForAdminTasksImpl(ctx, tf.MockSwarming, tf.MockUFS, tf.MockKarte, req)

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

// TestGetDUTsForLabstations tests that getDUTsForLabstations returns the correct list of bot names given that the GetDUTsForLabstation RPC is functioning correctly.
func TestGetDUTsForLabstations(t *testing.T) {
	tf, validate := newTestFixture(t)
	defer validate()
	ctx := tf.C
	// Make the UFS call successfully return exactly one fake DUT.
	tf.MockUFS.EXPECT().GetDUTsForLabstation(gomock.Any(), gomock.Any()).Return(
		&ufsAPI.GetDUTsForLabstationResponse{
			Items: []*ufsAPI.GetDUTsForLabstationResponse_LabstationMapping{
				{
					Hostname: "fake-labstation-1",
					DutName:  []string{"fake-dut-1"},
				},
			},
		},
		nil,
	)
	duts, err := getDUTsForLabstations(ctx, tf.MockUFS, nil)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if diff := cmp.Diff([]string{"crossk-fake-dut-1"}, duts); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}

// TestGetLabstations tests that we return the hostname of every labstation in the records returned by
// metrics.Search. For this test, we assume that the output of the Karte API is correct.
// We have Search return two records describing the same labstation to test the deduplication logic.
func TestGetLabstations(t *testing.T) {
	var zero time.Time
	tf, validate := newTestFixture(t)
	defer validate()
	ctx := tf.C
	tf.MockKarte.EXPECT().Search(gomock.Any(), gomock.Any()).Return(
		&metrics.QueryResult{
			Actions: []*metrics.Action{
				{
					Hostname:   "fake-labstation-1",
					ActionKind: labstationRebootKind,
					Status:     metrics.ActionStatusSuccess,
				},
				{
					Hostname:   "fake-labstation-1",
					ActionKind: labstationRebootKind,
					Status:     metrics.ActionStatusSuccess,
				},
			},
		},
		nil,
	)
	labstations, err := getLabstations(ctx, tf.MockKarte, zero, zero)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if diff := cmp.Diff([]string{"fake-labstation-1"}, labstations); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}
