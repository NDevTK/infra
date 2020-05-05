// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	proto "infra/unifiedfleet/api/v1/proto"
	api "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/model/configuration"
	"infra/unifiedfleet/app/model/datastore"

	crimsonconfig "go.chromium.org/luci/machine-db/api/config/v1"
)

var localPlatforms = []*crimsonconfig.Platform{
	{Name: "platform1"},
	{Name: "platform2"},
	{Name: "platform3"},
}

func mockParsePlatformsFunc(path string) (*crimsonconfig.Platforms, error) {
	return &crimsonconfig.Platforms{
		Platform: localPlatforms,
	}, nil
}

func TestImportChromePlatforms(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("Import chrome platforms", t, func() {
		Convey("happy path", func() {
			req := &api.ImportChromePlatformsRequest{
				Source: &api.ImportChromePlatformsRequest_ConfigSource{
					ConfigSource: &api.ConfigSource{
						ConfigServiceName: "",
						FileName:          "test.config",
					},
				},
			}
			parsePlatformsFunc = mockParsePlatformsFunc
			res, err := tf.Fleet.ImportChromePlatforms(ctx, req)
			So(err, ShouldBeNil)
			So(res.GetPassed(), ShouldHaveLength, len(localPlatforms))
			So(res.GetFailed(), ShouldHaveLength, 0)
			getRes, err := configuration.GetAllChromePlatforms(ctx)
			So(err, ShouldBeNil)
			So(getRes, ShouldHaveLength, len(localPlatforms))
			wants := getLocalPlatformNames()
			gets := getReturnedPlatformNames(*getRes)
			So(gets, ShouldResemble, wants)
		})
		Convey("import duplicated platforms", func() {
			req := &api.ImportChromePlatformsRequest{
				Source: &api.ImportChromePlatformsRequest_ConfigSource{
					ConfigSource: &api.ConfigSource{
						ConfigServiceName: "",
						FileName:          "test.config",
					},
				},
			}
			parsePlatformsFunc = func(_ string) (*crimsonconfig.Platforms, error) {
				return &crimsonconfig.Platforms{
					Platform: []*crimsonconfig.Platform{
						{Name: "platform1"},
						{Name: "platform4"},
					},
				}, nil
			}
			res, err := tf.Fleet.ImportChromePlatforms(ctx, req)
			So(err, ShouldBeNil)
			So(res.GetPassed(), ShouldHaveLength, 1)
			So(res.GetFailed(), ShouldHaveLength, 1)
			So(res.GetPassed()[0].GetPlatform().GetName(), ShouldEqual, "platform4")
			So(res.GetFailed()[0].GetPlatform().GetName(), ShouldEqual, "platform1")

			getRes, err := configuration.GetAllChromePlatforms(ctx)
			So(err, ShouldBeNil)
			So(getRes, ShouldHaveLength, len(localPlatforms)+1)
			wants := append(getLocalPlatformNames(), "platform4")
			gets := getReturnedPlatformNames(*getRes)
			So(gets, ShouldResemble, wants)
		})
	})
}

func getLocalPlatformNames() []string {
	wants := make([]string, len(localPlatforms))
	for i, p := range localPlatforms {
		wants[i] = p.GetName()
	}
	return wants
}

func getReturnedPlatformNames(res datastore.OpResults) []string {
	gets := make([]string, len(res))
	for i, r := range res {
		gets[i] = r.Data.(*proto.ChromePlatform).GetName()
	}
	return gets
}
