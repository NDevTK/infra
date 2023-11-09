// Copyright 2018 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"go.chromium.org/luci/appengine/gaetesting"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/protobuf/types/known/durationpb"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/clients"
	"infra/appengine/crosskylabadmin/internal/app/clients/mock"
	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/tq"
	"infra/appengine/crosskylabadmin/internal/ufs/mockufs"
	"infra/cros/recovery/logger/metrics/mockmetrics"
	"infra/libs/git"
)

type testFixture struct {
	T *testing.T
	C context.Context

	Tracker   fleet.TrackerServer
	Inventory *ServerImpl

	MockSwarming       *mock.MockSwarmingClient
	MockBotTasksCursor *mock.MockBotTasksCursor
	MockUFS            *mockufs.MockClient
	MockKarte          *mockmetrics.MockMetrics
}

// newTextFixture creates a new testFixture to be used in unittests.
//
// The function returns the created testFixture and a validation function that
// must be deferred by the caller.
func newTestFixture(t *testing.T) (testFixture, func()) {
	return newTestFixtureWithContext(testingContext(), t)
}

func newTestFixtureWithContext(c context.Context, t *testing.T) (testFixture, func()) {
	// Configure the tq implementation: confirm that it's testable and set up queues used by CrOSSkylabAdmin.
	if tq.GetTestable(c) == nil {
		panic("internal error in app/frontend/test_common.go: in unit tests, taskqueue must be a testable implementation")
	}

	tf := testFixture{T: t, C: c}

	mc := gomock.NewController(t)

	tf.MockSwarming = mock.NewMockSwarmingClient(mc)
	tf.MockBotTasksCursor = mock.NewMockBotTasksCursor(mc)
	tf.Inventory = &ServerImpl{}
	tf.MockUFS = mockufs.NewMockClient(mc)
	tf.MockKarte = mockmetrics.NewMockMetrics(mc)
	tf.Tracker = &TrackerServerImpl{
		SwarmingFactory: func(context.Context, string) (clients.SwarmingClient, error) {
			return tf.MockSwarming, nil
		},
		MetricsClient: tf.MockKarte,
	}

	validate := func() {
		mc.Finish()
	}
	return tf, validate
}

// TestingContext returns a context suitable for unit tests.
func testingContext() context.Context {
	c := gaetesting.TestingContextWithAppID("dev~infra-crosskylabadmin")
	c = config.Use(c, &config.Config{
		AccessGroup: "fake-access-group",
		Swarming: &config.Swarming{
			Host:              "https://fake-host.appspot.com",
			BotPool:           "ChromeOSSkylab",
			FleetAdminTaskTag: "fake-tag",
			LuciProjectTag:    "fake-project",
			PoolCfgs: []*config.Swarming_PoolCfg{
				{
					PoolName:     "ChromeOSSkylab",
					AuditEnabled: true,
				},
			},
		},
		Tasker: &config.Tasker{
			BackgroundTaskExecutionTimeoutSecs: 3600,
			BackgroundTaskExpirationSecs:       300,
		},
		Cron: &config.Cron{
			FleetAdminTaskPriority:     33,
			EnsureTasksCount:           3,
			RepairIdleDuration:         durationpb.New(10),
			RepairAttemptDelayDuration: durationpb.New(10),
		},
		StableVersionConfig: &config.StableVersionConfig{
			GerritHost:            "xxx-fake-gerrit-review.googlesource.com",
			GitilesHost:           "xxx-gitiles.googlesource.com",
			Project:               "xxx-project",
			Branch:                "xxx-branch",
			StableVersionDataPath: "xxx-stable_version_data_path",
		},
	})
	datastore.GetTestable(c).Consistent(true)
	c = gologger.StdConfig.Use(c)
	c = logging.SetLevel(c, logging.Debug)
	return c
}

type fakeGitClient struct {
	getFile func(ctx context.Context, path string) (string, error)
}

func (f *fakeGitClient) GetFile(ctx context.Context, path string) (string, error) {
	return f.getFile(ctx, path)
}

func (f *fakeGitClient) SwitchProject(ctx context.Context, project string) error {
	return nil
}

func (tf *testFixture) setStableVersionFactory(stableVersionFileContent string) {
	is := tf.Inventory
	is.StableVersionGitClientFactory = func(c context.Context) (git.ClientInterface, error) {
		gc := &fakeGitClient{}
		gc.getFile = func(ctx context.Context, path string) (string, error) {
			return stableVersionFileContent, nil
		}
		return gc, nil
	}
}

// expectDefaultPerBotRefresh sets up the default expectations for refreshing
// each bot, once the list of bots is known.
//
// This is useful for tests that only target the initial Swarming bot listing
// logic.
func expectDefaultPerBotRefresh(tf testFixture) {
	tf.MockSwarming.EXPECT().ListSortedRecentTasksForBot(
		gomock.Any(), gomock.Any(), gomock.Any(),
	).AnyTimes().Return([]*swarming.SwarmingRpcsTaskResult{}, nil)
	tf.MockSwarming.EXPECT().ListBotTasks(gomock.Any()).AnyTimes().Return(
		tf.MockBotTasksCursor)
	tf.MockBotTasksCursor.EXPECT().Next(gomock.Any(), gomock.Any()).AnyTimes().Return(
		[]*swarming.SwarmingRpcsTaskResult{}, nil)
}

// BotForDUT returns BotInfos for DUTs with the given dut id.
//
// state is the bot's state dimension.
// dims is a convenient way to specify other bot dimensions.
// "a:x,y;b:z" will set the dimensions of the bot to ["a": ["x", "y"], "b":
//
//	["z"]]
func BotForDUT(id string, state string, dims string) *swarming.SwarmingRpcsBotInfo {
	sdims := make([]*swarming.SwarmingRpcsStringListPair, 0, 2)
	if dims != "" {
		ds := strings.Split(dims, ";")
		for _, d := range ds {
			d = strings.Trim(d, " ")
			kvs := strings.Split(d, ":")
			if len(kvs) != 2 {
				panic(fmt.Sprintf("dims string |%s|%s has a non-keyval dimension |%s|", dims, ds, d))
			}
			sdim := &swarming.SwarmingRpcsStringListPair{
				Key:   strings.Trim(kvs[0], " "),
				Value: []string{},
			}
			for _, v := range strings.Split(kvs[1], ",") {
				sdim.Value = append(sdim.Value, strings.Trim(v, " "))
			}
			sdims = append(sdims, sdim)
		}
	}
	sdims = append(sdims, &swarming.SwarmingRpcsStringListPair{
		Key:   "dut_state",
		Value: []string{state},
	})
	sdims = append(sdims, &swarming.SwarmingRpcsStringListPair{
		Key:   "dut_id",
		Value: []string{id},
	})
	sdims = append(sdims, &swarming.SwarmingRpcsStringListPair{
		Key:   "dut_name",
		Value: []string{id + "-host"},
	})
	return &swarming.SwarmingRpcsBotInfo{
		BotId:      fmt.Sprintf("bot_%s", id),
		Dimensions: sdims,
	}
}
