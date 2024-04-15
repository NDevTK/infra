// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dumper

import (
	"context"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/encoding/protojson"

	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"

	ufsdevice "infra/unifiedfleet/api/v1/models/chromeos/device"
	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/external"
	"infra/unifiedfleet/app/model/configuration"
	"infra/unifiedfleet/app/util"
)

// grantRealmPerms grants `configurations.get` permissions on all realms to the
// existing context.
func grantRealmPerms(ctx context.Context, realms ...string) context.Context {
	perms := []authtest.RealmPermission{}

	for _, r := range realms {
		perms = append(perms, authtest.RealmPermission{
			Realm:      r,
			Permission: util.ConfigurationsGet,
		})
	}

	newCtx := auth.WithState(ctx, &authtest.FakeState{
		Identity:            "user:root@lab.com",
		IdentityPermissions: perms,
	})

	return newCtx
}

// ConstantRealmAssigner assigns a single realm to all device configs.
func ConstantRealmAssigner(d *ufsdevice.Config) string {
	return "chromeos:realm"
}

// TestSyncDeviceConfigs verifies the e2e sync works as expected
func TestSyncDeviceConfigs(t *testing.T) {
	t.Parallel()

	ctx := memory.UseWithAppID(context.Background(), ("dev~infra-unified-fleet-system"))
	ctx = external.WithTestingContext(ctx)
	ctx = grantRealmPerms(ctx, "chromeos:board1-model1", "chromeos:board2-model2", "chromeos:realm")

	Convey("When sync is run with a valid config", t, func() {
		namespaceToRealmAssignerMap = map[string]configuration.RealmAssignerFunc{
			"random-ns":             configuration.BoardModelRealmAssigner,
			util.OSPartnerNamespace: ConstantRealmAssigner,
		}
		ctx = config.Use(ctx, &config.Config{
			DeviceConfigsPushConfigs: &config.DeviceConfigPushConfigs{
				ConfigsPath: "test_device_config",
				Enabled:     true,
			},
		})

		err := syncDeviceConfigs(ctx)
		So(err, ShouldBeNil)

		Convey("DeviceConfigs should be fetchable in all namespaces specified", func() {
			for ns := range namespaceToRealmAssignerMap {
				ctx, err := util.SetupDatastoreNamespace(ctx, ns)
				So(err, ShouldBeNil)

				cfg, err := configuration.GetDeviceConfigACL(ctx, configuration.GetConfigID("board1", "model1", ""))
				So(cfg, ShouldResembleProto, expectedConfigs[0])
				So(err, ShouldBeNil)
				cfg2, err := configuration.GetDeviceConfigACL(ctx, configuration.GetConfigID("board2", "model2", ""))
				So(cfg2, ShouldResembleProto, expectedConfigs[1])
				So(err, ShouldBeNil)
			}
		})
		Convey("DeviceConfigs should only be fetchable in namespaces specified", func() {
			ctx, err := util.SetupDatastoreNamespace(ctx, "fake")
			So(err, ShouldBeNil)

			cfg, err := configuration.GetDeviceConfigACL(ctx, configuration.GetConfigID("board1", "model1", ""))
			So(cfg, ShouldBeNil)
			So(err, ShouldBeError)
			cfg2, err := configuration.GetDeviceConfigACL(ctx, configuration.GetConfigID("board2", "model2", ""))
			So(cfg2, ShouldBeNil)
			So(err, ShouldBeError)
		})
	})
}

// expectedConfigs contain the configs we read from the fake configs
// will be blank on any issue with the configs
var expectedConfigs = getExpectedConfigs()

func getExpectedConfigs() []*ufsdevice.Config {
	cfgs := &ufsdevice.AllConfigs{}
	content, err := os.ReadFile("../frontend/fake/device_config.cfg")
	if err != nil {
		return cfgs.Configs
	}
	if err := protojson.Unmarshal(content, cfgs); err != nil {
		return cfgs.Configs
	}
	return cfgs.Configs
}
