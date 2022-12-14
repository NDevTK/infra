// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package frontend

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/gae/service/datastore"

	ufspb "infra/unifiedfleet/api/v1/models"
	api "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/controller"
	"infra/unifiedfleet/app/external"
	"infra/unifiedfleet/app/model/registration"
)

// encTestingContext creates a testing context which mocks the logging and datastore services.
// Also loads a custom config, which will allow the loading of a dummy bot config file
func encTestingContext() context.Context {
	c := gaetesting.TestingContextWithAppID("dev~infra-unified-fleet-system")
	c = gologger.StdConfig.Use(c)
	c = logging.SetLevel(c, logging.Error)
	c = config.Use(c, &config.Config{
		OwnershipConfig: &config.OwnershipConfig{
			GitilesHost: "test_gitiles",
			Project:     "test_project",
			Branch:      "test_branch",
			EncConfig: []*config.OwnershipConfig_ENCConfigFile{
				{
					Name:       "test_name",
					RemotePath: "test_enc_git_path",
				},
			},
		},
	})
	c = external.WithTestingContext(c)
	datastore.GetTestable(c).Consistent(true)
	return c
}

// Tests the RPC for getting ownership data
func TestGetOwnershipData(t *testing.T) {
	t.Parallel()
	ctx := encTestingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("Get Ownership Data for Bots", t, func() {
		Convey("happy path", func() {
			resp, err := registration.CreateMachine(ctx, &ufspb.Machine{
				Name: "testing-1"})
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			err = controller.ImportENCBotConfig(ctx)
			So(err, ShouldBeNil)
			req := &api.GetOwnershipDataRequest{
				Hostname: "testing-1",
			}

			res, err := tf.Fleet.GetOwnershipData(ctx, req)

			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.PoolName, ShouldEqual, "test")
			So(res.SwarmingInstance, ShouldEqual, "test_name")
		})
		Convey("Missing host - returns error", func() {
			req := &api.GetOwnershipDataRequest{
				Hostname: "blah-1",
			}
			res, err := tf.Fleet.GetOwnershipData(ctx, req)
			So(err, ShouldNotBeNil)
			So(res, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "not found")
		})
	})
}
