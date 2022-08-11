// Copyright 2018 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package frontend

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

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
	"infra/libs/git"
)

type testFixture struct {
	T *testing.T
	C context.Context

	Tracker   fleet.TrackerServer
	Inventory *ServerImpl

	MockSwarming       *mock.MockSwarmingClient
	MockBotTasksCursor *mock.MockBotTasksCursor
}

// newTextFixture creates a new testFixture to be used in unittests.
//
// The function returns the created testFixture and a validation function that
// must be deferred by the caller.
func newTestFixture(t *testing.T) (testFixture, func()) {
	return newTestFixtureWithContext(testingContext(), t)
}

func newTestFixtureWithContext(c context.Context, t *testing.T) (testFixture, func()) {
	tf := testFixture{T: t, C: c}

	mc := gomock.NewController(t)

	tf.MockSwarming = mock.NewMockSwarmingClient(mc)
	tf.Tracker = &TrackerServerImpl{
		SwarmingFactory: func(context.Context, string) (clients.SwarmingClient, error) {
			return tf.MockSwarming, nil
		},
	}
	tf.MockBotTasksCursor = mock.NewMockBotTasksCursor(mc)
	tf.Inventory = &ServerImpl{}

	validate := func() {
		mc.Finish()
	}
	return tf, validate
}

func testingContext() context.Context {
	c := gaetesting.TestingContextWithAppID("dev~infra-crosskylabadmin")
	c = config.Use(c, &config.Config{
		AccessGroup: "fake-access-group",
		Swarming: &config.Swarming{
			Host:              "https://fake-host.appspot.com",
			BotPool:           "ChromeOSSkylab",
			FleetAdminTaskTag: "fake-tag",
			LuciProjectTag:    "fake-project",
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

// makeBotSelector returns a fleet.BotSelector selecting each of the duts
// recognized by the given Dut IDs.
func makeBotSelectorForDuts(duts []string) []*fleet.BotSelector {
	var bs []*fleet.BotSelector
	for _, d := range duts {
		bs = append(bs, &fleet.BotSelector{DutId: d})
	}
	return bs
}

// timeOffsetFromNowInSwarmingFormat returns a string representation of time offset
// from now as returned by Swarming.
func timeOffsetFromNowInSwarmingFormat(offset time.Duration) string {
	t := time.Now().UTC().Add(offset)
	return t.Format("2006-01-02T15:04:05.999999999")
}

// createTaskArgsMatcher is a gomock matcher to validate a subset of the fields
// of clients.SwarmingCreateTaskArgs argument.
type createTaskArgsMatcher struct {
	BotID                string
	DutID                string
	DutName              string
	DutState             string
	Priority             int64
	ExecutionTimeoutSecs int64
	ExpirationSecs       int64
	CmdSubString         string
}

func (m *createTaskArgsMatcher) Matches(x interface{}) bool {
	var args clients.SwarmingCreateTaskArgs
	switch a := x.(type) {
	case clients.SwarmingCreateTaskArgs:
		args = a
	case *clients.SwarmingCreateTaskArgs:
		args = *a
	default:
		return false
	}

	if (m.BotID != "" && args.BotID != m.BotID) ||
		(m.DutID != "" && args.DutID != m.DutID) ||
		(m.DutName != "" && args.DutName != m.DutName) ||
		(m.DutState != "" && args.DutState != m.DutState) ||
		(m.Priority != 0 && args.Priority != m.Priority) ||
		(m.ExecutionTimeoutSecs != 0 && args.ExecutionTimeoutSecs != m.ExecutionTimeoutSecs) ||
		(m.ExpirationSecs != 0 && args.ExpirationSecs != m.ExpirationSecs) {
		return false
	}
	if m.CmdSubString != "" {
		cmd := strings.Join(args.Cmd, " ")
		if !strings.Contains(cmd, m.CmdSubString) {
			return false
		}
	}
	return true
}

func (m *createTaskArgsMatcher) String() string {
	s := "is clients.SwarmingCreateTaskArgs with fields like"
	if m.BotID != "" {
		s = fmt.Sprintf("%s BotID: %s", s, m.BotID)
	}
	if m.DutID != "" {
		s = fmt.Sprintf("%s DutID: %s", s, m.DutID)
	}
	if m.DutName != "" {
		s = fmt.Sprintf("%s DutName: %s", s, m.DutName)
	}
	if m.DutState != "" {
		s = fmt.Sprintf("%s DutState: %s", s, m.DutState)
	}
	if m.Priority != 0 {
		s = fmt.Sprintf("%s DutState: %d", s, m.Priority)
	}
	if m.ExecutionTimeoutSecs != 0 {
		s = fmt.Sprintf("%s ExecutionTimeoutSecs: %d", s, m.ExecutionTimeoutSecs)
	}
	if m.ExpirationSecs != 0 {
		s = fmt.Sprintf("%s ExpirationSecs: %d", s, m.ExpirationSecs)
	}
	if m.CmdSubString != "" {
		s = fmt.Sprintf("%s CmdSubString: %s", s, m.CmdSubString)
	}
	return s
}

// BotForDUT returns BotInfos for DUTs with the given dut id.
//
// state is the bot's state dimension.
// dims is a convenient way to specify other bot dimensions.
// "a:x,y;b:z" will set the dimensions of the bot to ["a": ["x", "y"], "b":
//   ["z"]]
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
