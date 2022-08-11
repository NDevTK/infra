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

package inventory

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/gae/service/datastore"

	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/libs/git"
)

type testFixture struct {
	T *testing.T
	C context.Context

	Inventory *ServerImpl
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

func newTestFixtureWithContext(ctx context.Context, t *testing.T) (testFixture, func()) {
	tf := testFixture{T: t, C: ctx}
	mc := gomock.NewController(t)

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
		Tasker: &config.Tasker{
			BackgroundTaskExecutionTimeoutSecs: 3600,
			BackgroundTaskExpirationSecs:       300,
		},
		Swarming: &config.Swarming{
			Host:              "https://fake-host.appspot.com",
			BotPool:           "ChromeOSSkylab",
			FleetAdminTaskTag: "fake-tag",
			LuciProjectTag:    "fake-project",
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
