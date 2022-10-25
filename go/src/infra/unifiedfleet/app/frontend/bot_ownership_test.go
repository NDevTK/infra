// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package frontend

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	configpb "go.chromium.org/luci/swarming/proto/config"

	ufspb "infra/unifiedfleet/api/v1/models"
	api "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/controller"
	"infra/unifiedfleet/app/model/registration"
)

// Tests the RPC for getting ownership data
func TestGetOwnershipData(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("Get Ownership Data for Bots", t, func() {
		Convey("happy path", func() {
			resp, err := registration.CreateMachine(ctx, &ufspb.Machine{
				Name: "test1-1"})
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			controller.ParseBotConfig(ctx, &configpb.BotsCfg{
				BotGroup: []*configpb.BotGroup{
					{
						BotId:      []string{"test{1,2}-1"},
						Dimensions: []string{"pool:abc"},
					},
				},
			}, "testSwarming")
			req := &api.GetOwnershipDataRequest{
				Hostname: "test1-1",
			}

			res, err := tf.Fleet.GetOwnershipData(ctx, req)

			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.PoolName, ShouldEqual, "abc")
			So(res.SwarmingInstance, ShouldEqual, "testSwarming")
		})
		Convey("Missing host - returns error", func() {
			req := &api.GetOwnershipDataRequest{
				Hostname: "test2-1",
			}
			res, err := tf.Fleet.GetOwnershipData(ctx, req)
			So(err, ShouldNotBeNil)
			So(res, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "not found")
		})
	})
}
