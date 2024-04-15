// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/gae/service/datastore"

	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/external"
)

// encTestingContext creates a testing context which mocks the logging and
// datastore services and loads a custom config,
// which will allow the loading of a dummy bot config file
func encTestingContext() context.Context {
	ctx := gaetesting.TestingContextWithAppID("dev~infra-unified-fleet-system")
	ctx = gologger.StdConfig.Use(ctx)
	ctx = logging.SetLevel(ctx, logging.Error)
	ctx = config.Use(ctx, &config.Config{
		OwnershipConfig: &config.OwnershipConfig{
			GitilesHost: "test_gitiles",
			Project:     "test_project",
			Branch:      "test_branch",
			EncConfig: []*config.OwnershipConfig_ConfigFile{
				{
					Name:       "test_name",
					RemotePath: "test_enc_git_path",
				},
			},
			SecurityConfig: []*config.OwnershipConfig_ConfigFile{
				{
					Name:       "test_name",
					RemotePath: "test_security_git_path",
				},
			},
		},
	})
	ctx = external.WithTestingContext(ctx)
	datastore.GetTestable(ctx).Consistent(true)
	return ctx
}

// Tests the functionality for loading and storing Ownership
// data from bot config files sepcified in the UFS config.
func TestGetEncBotConfigs(t *testing.T) {
	t.Parallel()

	Convey("Read Bot Configs", t, func() {
		Convey("happy path", func() {
			err := getBotConfigs(encTestingContext())
			So(err, ShouldBeNil)
		})
	})
}
