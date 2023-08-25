// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	. "github.com/smartystreets/goconvey/convey"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/errors"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/tq"
	"infra/cros/recovery/logger/metrics"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// TestPushBotsForAdminTasksImplSmokeTesttests that pushing bots for admin tasks
// calls the ListALiveIdleBotsInPool API.
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

// TestPushBotsForAdminTasksWithUFSClient tests that pushing bots for admin tasks with a UFS client succeeds.
func TestPushBotsForAdminTasksWithUFSClient(t *testing.T) {
	tf, validate := newTestFixture(t)
	defer validate()
	ctx := tf.C
	tq.GetTestable(ctx).CreateQueue("repair-bots")
	tf.MockSwarming.EXPECT().ListAliveIdleBotsInPool(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*swarming.SwarmingRpcsBotInfo{
		{
			BotId: "fake-bot-a",
			Dimensions: []*swarming.SwarmingRpcsStringListPair{
				{
					Key:   "id",
					Value: []string{"fake-bot-a"},
				},
				{
					Key:   "pool",
					Value: []string{"fake-bot-pool"},
				},
				{
					Key:   "dut_state",
					Value: []string{"needs_repair"},
				},
			},
		},
		{
			BotId: "fake-bot-b",
			Dimensions: []*swarming.SwarmingRpcsStringListPair{
				{
					Key:   "id",
					Value: []string{"fake-bot-b"},
				},
				{
					Key:   "pool",
					Value: []string{"fake-bot-pool"},
				},
				{
					Key:   "dut_state",
					Value: []string{"needs_repair"},
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
	numTasks := len(tq.GetTestable(ctx).GetScheduledTasks()["repair-bots"])
	if numTasks != 2 {
		t.Errorf("unexpected number of tasks %d", numTasks)
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

// TestPushBotsForAdminTasksWithPoolCfg tests that pushing bots for admin tasks with Pool
func TestPushBotsForAdminTasksWithPoolCfg(t *testing.T) {
	Convey("Handling PoolCfg bots", t, func() {
		tf, validate := newTestFixture(t)
		defer validate()
		ctx := tf.C
		tqt := tq.GetTestable(ctx)
		qn := "repair-bots"
		tqt.CreateQueue(qn)
		tf.MockSwarming.EXPECT().ListAliveIdleBotsInPool(gomock.Any(), "fake-bot-pool", gomock.Any()).Return([]*swarming.SwarmingRpcsBotInfo{
			{
				BotId: "fake-bot-a",
				Dimensions: []*swarming.SwarmingRpcsStringListPair{
					{
						Key:   "id",
						Value: []string{"fake-bot-a"},
					},
					{
						Key:   "pool",
						Value: []string{"fake-bot-pool"},
					},
					{
						Key:   "dut_state",
						Value: []string{"needs_repair"},
					},
				},
			},
			{
				BotId: "fake-bot-b",
				Dimensions: []*swarming.SwarmingRpcsStringListPair{
					{
						Key:   "id",
						Value: []string{"fake-bot-b"},
					},
					{
						Key:   "pool",
						Value: []string{"fake-bot-pool"},
					},
					{
						Key:   "dut_state",
						Value: []string{"needs_repair"},
					},
				},
			},
		}, nil)
		tf.MockSwarming.EXPECT().ListAliveIdleBotsInPool(gomock.Any(), "pool-cfg-a", gomock.Any()).Return([]*swarming.SwarmingRpcsBotInfo{
			{
				BotId: "pool-cfg-bot-a",
				Dimensions: []*swarming.SwarmingRpcsStringListPair{
					{
						Key:   "id",
						Value: []string{"pool-cfg-bot-a"},
					},
					{
						Key:   "pool",
						Value: []string{"pool-cfg-a"},
					},
					{
						Key:   "dut_state",
						Value: []string{"needs_repair"},
					},
				},
			},
		}, nil)

		tf.MockSwarming.EXPECT().ListAliveIdleBotsInPool(gomock.Any(), "pool-cfg-b", gomock.Any()).Return([]*swarming.SwarmingRpcsBotInfo{
			{
				BotId: "pool-cfg-bot-b",
				Dimensions: []*swarming.SwarmingRpcsStringListPair{
					{
						Key:   "id",
						Value: []string{"pool-cfg-bot-b"},
					},
					{
						Key:   "pool",
						Value: []string{"pool-cfg-b"},
					},
					{
						Key:   "dut_state",
						Value: []string{"needs_repair"},
					},
				},
			},
		}, nil)

		ctx = config.Use(ctx, &config.Config{
			Swarming: &config.Swarming{
				BotPool: "fake-bot-pool",
				PoolCfgs: []*config.Swarming_PoolCfg{
					{
						PoolName: "pool-cfg-a",
					},
					{
						PoolName:      "pool-cfg-b",
						BuilderBucket: "some_bucket",
					},
				},
			},
		})

		req := &fleet.PushBotsForAdminTasksRequest{
			TargetDutState: fleet.DutState_NeedsRepair,
		}
		_, err := pushBotsForAdminTasksImpl(ctx, tf.MockSwarming, tf.MockUFS, tf.MockKarte, req)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		tasks := tqt.GetScheduledTasks()[qn]
		fmt.Println(tasks)
		numTasks := len(tasks)
		So(numTasks, ShouldEqual, 4)
		var taskPaths, taskParams []string
		for _, v := range tasks {
			taskPaths = append(taskPaths, v.Path)
			taskParams = append(taskParams, string(v.Payload))
		}
		sort.Strings(taskPaths)
		sort.Strings(taskParams)
		expectedPaths := []string{"/internal/task/cros_repair/fake-bot-a", "/internal/task/cros_repair/fake-bot-b", "/internal/task/cros_repair/pool-cfg-bot-a", "/internal/task/cros_repair/pool-cfg-bot-b"}
		expectedParams := []string{"botID=fake-bot-a&builderBucket=&expectedState=needs_repair", "botID=fake-bot-b&builderBucket=&expectedState=needs_repair", "botID=pool-cfg-bot-a&builderBucket=&expectedState=needs_repair", "botID=pool-cfg-bot-b&builderBucket=some_bucket&expectedState=needs_repair"}
		So(taskPaths, ShouldResemble, expectedPaths)
		So(taskParams, ShouldResemble, expectedParams)
	})
}

// TestPushBotsForAdminTasksWithPoolCfgSkipError tests that error condition while pushing bots for admin tasks with Pool
func TestPushBotsForAdminTasksWithPoolCfgSkipError(t *testing.T) {
	Convey("Handling Errors for PoolCfg bots ", t, func() {
		tf, validate := newTestFixture(t)
		defer validate()
		ctx := tf.C
		tqt := tq.GetTestable(ctx)
		qn := "repair-bots"
		tqt.CreateQueue(qn)
		tf.MockSwarming.EXPECT().ListAliveIdleBotsInPool(gomock.Any(), "fake-bot-pool", gomock.Any()).Return([]*swarming.SwarmingRpcsBotInfo{
			{
				BotId: "fake-bot-a",
				Dimensions: []*swarming.SwarmingRpcsStringListPair{
					{
						Key:   "id",
						Value: []string{"fake-bot-a"},
					},
					{
						Key:   "pool",
						Value: []string{"fake-bot-pool"},
					},
					{
						Key:   "dut_state",
						Value: []string{"needs_repair"},
					},
				},
			},
			{
				BotId: "fake-bot-b",
				Dimensions: []*swarming.SwarmingRpcsStringListPair{
					{
						Key:   "id",
						Value: []string{"fake-bot-b"},
					},
					{
						Key:   "pool",
						Value: []string{"fake-bot-pool"},
					},
					{
						Key:   "dut_state",
						Value: []string{"needs_repair"},
					},
				},
			},
		}, nil)

		tf.MockSwarming.EXPECT().ListAliveIdleBotsInPool(gomock.Any(), "pool-cfg-a", gomock.Any()).Return(nil, errors.Reason("Fake Error").Err())

		tf.MockSwarming.EXPECT().ListAliveIdleBotsInPool(gomock.Any(), "pool-cfg-b", gomock.Any()).Return([]*swarming.SwarmingRpcsBotInfo{
			{
				BotId: "pool-cfg-bot-b",
				Dimensions: []*swarming.SwarmingRpcsStringListPair{
					{
						Key:   "id",
						Value: []string{"pool-cfg-bot-b"},
					},
					{
						Key:   "pool",
						Value: []string{"pool-cfg-b"},
					},
					{
						Key:   "dut_state",
						Value: []string{"needs_repair"},
					},
				},
			},
		}, nil)

		ctx = config.Use(ctx, &config.Config{
			Swarming: &config.Swarming{
				BotPool: "fake-bot-pool",
				PoolCfgs: []*config.Swarming_PoolCfg{
					{
						PoolName: "pool-cfg-a",
					},
					{
						PoolName:      "pool-cfg-b",
						BuilderBucket: "some_bucket",
					},
				},
			},
		})

		req := &fleet.PushBotsForAdminTasksRequest{
			TargetDutState: fleet.DutState_NeedsRepair,
		}
		_, err := pushBotsForAdminTasksImpl(ctx, tf.MockSwarming, tf.MockUFS, tf.MockKarte, req)
		if err != nil {
			if !strings.Contains(err.Error(), "Fake Error") {
				t.Errorf("unexpected error: %s", err)
			}
		}
		tasks := tqt.GetScheduledTasks()[qn]
		fmt.Println(tasks)
		numTasks := len(tasks)
		So(numTasks, ShouldEqual, 3)
		var taskPaths, taskParams []string
		for _, v := range tasks {
			taskPaths = append(taskPaths, v.Path)
			taskParams = append(taskParams, string(v.Payload))
		}
		sort.Strings(taskPaths)
		sort.Strings(taskParams)
		expectedPaths := []string{"/internal/task/cros_repair/fake-bot-a", "/internal/task/cros_repair/fake-bot-b", "/internal/task/cros_repair/pool-cfg-bot-b"}
		expectedParams := []string{"botID=fake-bot-a&builderBucket=&expectedState=needs_repair", "botID=fake-bot-b&builderBucket=&expectedState=needs_repair", "botID=pool-cfg-bot-b&builderBucket=some_bucket&expectedState=needs_repair"}
		So(taskPaths, ShouldResemble, expectedPaths)
		So(taskParams, ShouldResemble, expectedParams)
	})
}
