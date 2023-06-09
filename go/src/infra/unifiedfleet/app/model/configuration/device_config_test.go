// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configuration

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"

	ufsdevice "infra/unifiedfleet/api/v1/models/chromeos/device"
	. "infra/unifiedfleet/app/model/datastore"
	"infra/unifiedfleet/app/util"
)

// makeDevCfgForTesting creates a basic DeviceConfig. These configs have no
// guarantee to make sense at a domain level, but can be used to verify code
// behavior.
func makeDevCfgForTesting(board, model, variant string, tams []string) *ufsdevice.Config {
	return &ufsdevice.Config{
		Id:  GetConfigID(board, model, variant),
		Tam: tams,
	}
}

// boardRealmAssigner just sets the realm to be equal to the board.
func boardRealmAssigner(c *ufsdevice.Config) string {
	return c.Id.PlatformId.Value
}

// grantRealmPerms grants `configuration.get` permissions in specified realms.
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

// constantRealmAssigner assigns all devices to a specific realm.
func constantRealmAssigner(d *ufsdevice.Config) string {
	return "chromeos:realm"
}

func TestBatchUpdateDeviceConfig(t *testing.T) {
	t.Parallel()
	baseCtx := memory.UseWithAppID(context.Background(), ("gae-test"))
	ctx := grantRealmPerms(baseCtx, "chromeos:realm")

	datastore.GetTestable(ctx).Consistent(true)

	Convey("When a valid config is added", t, func() {
		cfgs := make([]*ufsdevice.Config, 2)
		for i := 0; i < 2; i++ {
			cfgs[i] = makeDevCfgForTesting(fmt.Sprintf("board%d", i), fmt.Sprintf("model%d", i), fmt.Sprintf("variant%d", i), []string{fmt.Sprintf("test-%d", i)})
		}
		resp, err := BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{cfgs[0]}, constantRealmAssigner)
		So(err, ShouldBeNil)
		So(resp, ShouldResembleProto, []*ufsdevice.Config{cfgs[0]})
		Convey("That config is written to datastore", func() {
			cfg0, err := GetDeviceConfigACL(ctx, GetConfigID("board0", "model0", "variant0"))
			So(err, ShouldBeNil)
			So(cfg0, ShouldResembleProto, cfgs[0])
		})
		Convey("When both that config and another config is added", func() {
			resp, err := BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{cfgs[0], cfgs[1]}, constantRealmAssigner)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, []*ufsdevice.Config{cfgs[0], cfgs[1]})
			Convey("Both configs are accessible", func() {
				cfg0, err := GetDeviceConfigACL(ctx, GetConfigID("board0", "model0", "variant0"))
				So(err, ShouldBeNil)
				So(cfg0, ShouldResembleProto, cfgs[0])
				cfg1, err := GetDeviceConfigACL(ctx, GetConfigID("board1", "model1", "variant1"))
				So(err, ShouldBeNil)
				So(cfg1, ShouldResembleProto, cfgs[1])
			})
		})
	})

	Convey("When an invalid config is added in a batch request", t, func() {
		badCfg := &ufsdevice.Config{}
		goodCfg := makeDevCfgForTesting("board0", "model0", "variant", []string{"email"})

		resp, err := BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{badCfg, goodCfg}, constantRealmAssigner)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
		Convey("No configs from that request are added", func() {
			_, err := GetDeviceConfigACL(ctx, GetConfigID("board0", "model0", "variant"))
			So(err, ShouldNotBeNil)
		})
	})

	Convey("When inserting a config with a specific realm", t, func() {
		cfg := makeDevCfgForTesting("board", "model", "variant", []string{"email"})

		// note boardRealmAssigner
		resp, err := BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{cfg}, constantRealmAssigner)
		So(err, ShouldBeNil)
		So(resp, ShouldResembleProto, []*ufsdevice.Config{cfg})
		Convey("Entity in datastore has correct realm", func() {
			entity := &DeviceConfigEntity{
				ID: GetDeviceConfigIDStr(GetConfigID("board", "model", "variant")),
			}
			err := datastore.Get(ctx, entity)
			So(err, ShouldBeNil)
			So(entity.Realm, ShouldEqual, "chromeos:realm")
		})
	})
}

func TestGetDeviceConfig(t *testing.T) {
	t.Parallel()
	baseCtx := memory.UseWithAppID(context.Background(), ("gae-test"))
	ctx := grantRealmPerms(baseCtx, "chromeos:board-model")

	cfg := makeDevCfgForTesting("board", "model", "variant", []string{"email"})

	Convey("When a config is added", t, func() {
		resp, err := BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{cfg}, BoardModelRealmAssigner)
		So(err, ShouldBeNil)
		So(resp, ShouldResembleProto, []*ufsdevice.Config{cfg})
		Convey("That config can be accessed", func() {
			cfg_resp, err := GetDeviceConfigACL(ctx, GetConfigID("board", "model", "variant"))
			So(err, ShouldBeNil)
			So(cfg_resp, ShouldResembleProto, cfg)
		})
		Convey("Another config cannot be accessed", func() {
			cfg_resp, err := GetDeviceConfigACL(ctx, GetConfigID("board2", "model2", "variant2"))
			So(cfg_resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("A config with an invalid ID cannot be accessed", func() {
			cfg_resp, err := GetDeviceConfigACL(ctx, &ufsdevice.ConfigId{})
			So(cfg_resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
		Convey("A config can't be accessed without permissions", func() {
			otherPermsCtx := grantRealmPerms(baseCtx, "chromeos:other-board")
			cfg_resp, err := GetDeviceConfigACL(otherPermsCtx, GetConfigID("board", "model", "variant"))
			So(cfg_resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, PermissionDenied)
		})
	})
}

func TestDeviceConfigsExist(t *testing.T) {
	t.Parallel()
	baseCtx := memory.UseWithAppID(context.Background(), ("gae-test"))
	ctx := grantRealmPerms(baseCtx, "chromeos:board-model")

	cfg := makeDevCfgForTesting("board", "model", "variant", []string{"email"})
	cfg1 := makeDevCfgForTesting("board-hidden", "model-hidden", "variant", []string{"email"})

	Convey("When a config is added", t, func() {
		resp, err := BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{cfg, cfg1}, BoardModelRealmAssigner)
		So(err, ShouldBeNil)
		So(resp, ShouldResembleProto, []*ufsdevice.Config{cfg, cfg1})
		Convey("DeviceConfigsExist should correctly report that config exists, and other config does not", func() {
			cfgIDs := []*ufsdevice.ConfigId{GetConfigID("board", "model", "variant"), GetConfigID("non", "existant", "config")}
			exists, err := DeviceConfigsExistACL(ctx, cfgIDs)

			So(err, ShouldBeNil)
			So(exists, ShouldResemble, []bool{true, false})
		})
		Convey("DeviceConfigsExist should only report that configs the user can see are returned", func() {
			cfgIDs := []*ufsdevice.ConfigId{GetConfigID("board", "model", "variant"), GetConfigID("board-hidden", "model-hidden", "variant")}
			exists, err := DeviceConfigsExistACL(ctx, cfgIDs)

			So(err, ShouldBeNil)
			So(exists, ShouldResemble, []bool{true, false})

			fullPermsCtx := grantRealmPerms(baseCtx, "chromeos:board-model", "chromeos:board-hidden-model-hidden")
			exists, err = DeviceConfigsExistACL(fullPermsCtx, cfgIDs)

			So(err, ShouldBeNil)
			So(exists, ShouldResemble, []bool{true, true})
		})
	})
}

func TestGetDeviceConfigIDStr(t *testing.T) {
	t.Parallel()

	Convey("test full config", t, func() {
		cfgID := GetConfigID("board", "model", "variant")
		id := GetDeviceConfigIDStr(cfgID)
		So(id, ShouldEqual, "board.model.variant")
	})
	Convey("test board/model", t, func() {
		cfgID := GetConfigID("board", "model", "")
		id := GetDeviceConfigIDStr(cfgID)
		So(id, ShouldEqual, "board.model.")
	})
	Convey("test empty config", t, func() {
		cfgID := &ufsdevice.ConfigId{}
		id := GetDeviceConfigIDStr(cfgID)
		So(id, ShouldEqual, "..")
	})
}
