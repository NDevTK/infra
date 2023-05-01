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

	ufsdevice "infra/unifiedfleet/api/v1/models/chromeos/device"
	. "infra/unifiedfleet/app/model/datastore"
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

func TestBatchUpdateDeviceConfig(t *testing.T) {
	t.Parallel()
	ctx := memory.UseWithAppID(context.Background(), ("gae-test"))
	datastore.GetTestable(ctx).Consistent(true)

	Convey("When a valid config is added", t, func() {
		cfgs := make([]*ufsdevice.Config, 3)
		for i := 0; i < 2; i++ {
			cfgs[i] = makeDevCfgForTesting(fmt.Sprintf("board-%d", i), fmt.Sprintf("model-%d", i), fmt.Sprintf("variant-%d", i), []string{fmt.Sprintf("test-%d", i)})
		}
		resp, err := BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{cfgs[0]}, BlankRealmAssigner)
		So(err, ShouldBeNil)
		So(resp, ShouldResembleProto, []*ufsdevice.Config{cfgs[0]})
		Convey("That config is written to datastore", func() {
			cfg0, err := GetDeviceConfig(ctx, GetConfigID("board-0", "model-0", "variant-0"))
			So(err, ShouldBeNil)
			So(cfg0, ShouldResembleProto, cfgs[0])
		})
		Convey("When both that config and another config is added", func() {
			resp, err := BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{cfgs[0], cfgs[1]}, BlankRealmAssigner)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, []*ufsdevice.Config{cfgs[0], cfgs[1]})
			Convey("Both configs are accessible", func() {
				cfg0, err := GetDeviceConfig(ctx, GetConfigID("board-0", "model-0", "variant-0"))
				So(err, ShouldBeNil)
				So(cfg0, ShouldResembleProto, cfgs[0])
				cfg1, err := GetDeviceConfig(ctx, GetConfigID("board-1", "model-1", "variant-1"))
				So(err, ShouldBeNil)
				So(cfg1, ShouldResembleProto, cfgs[1])
			})
		})
	})

	Convey("When an invalid config is added in a batch request", t, func() {
		badCfg := &ufsdevice.Config{}
		goodCfg := makeDevCfgForTesting("board", "model", "variant", []string{"email"})

		resp, err := BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{badCfg, goodCfg}, BlankRealmAssigner)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
		Convey("No configs from that request are added", func() {
			_, err := GetDeviceConfig(ctx, GetConfigID("board", "model", "variant"))
			So(err, ShouldNotBeNil)
		})
	})

	Convey("When inserting a config with a specific realm", t, func() {
		cfg := makeDevCfgForTesting("board", "model", "variant", []string{"email"})

		// note boardRealmAssigner
		resp, err := BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{cfg}, boardRealmAssigner)
		So(err, ShouldBeNil)
		So(resp, ShouldResembleProto, []*ufsdevice.Config{cfg})
		Convey("Entity in datastore has correct realm", func() {
			entity := &DeviceConfigEntity{
				ID: GetDeviceConfigIDStr(GetConfigID("board", "model", "variant")),
			}
			err := datastore.Get(ctx, entity)
			So(err, ShouldBeNil)
			So(entity.Realm, ShouldEqual, "board")
		})
	})
}

func TestGetDeviceConfig(t *testing.T) {
	t.Parallel()
	ctx := memory.UseWithAppID(context.Background(), ("gae-test"))
	cfg := makeDevCfgForTesting("board", "model", "variant", []string{"email"})

	Convey("When a config is added", t, func() {
		resp, err := BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{cfg}, BlankRealmAssigner)
		So(err, ShouldBeNil)
		So(resp, ShouldResembleProto, []*ufsdevice.Config{cfg})
		Convey("That config can be accessed", func() {
			cfg_resp, err := GetDeviceConfig(ctx, GetConfigID("board", "model", "variant"))
			So(err, ShouldBeNil)
			So(cfg_resp, ShouldResembleProto, cfg)
		})
		Convey("Another config cannot be accessed", func() {
			cfg_resp, err := GetDeviceConfig(ctx, GetConfigID("board2", "model2", "variant2"))
			So(cfg_resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("A config with an invalid ID cannot be accessed", func() {
			cfg_resp, err := GetDeviceConfig(ctx, &ufsdevice.ConfigId{})
			So(cfg_resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestDeviceConfigsExist(t *testing.T) {
	t.Parallel()
	ctx := memory.UseWithAppID(context.Background(), ("gae-test"))
	cfg := makeDevCfgForTesting("board", "model", "variant", []string{"email"})

	Convey("When a config is added", t, func() {
		resp, err := BatchUpdateDeviceConfigs(ctx, []*ufsdevice.Config{cfg}, BlankRealmAssigner)
		So(err, ShouldBeNil)
		So(resp, ShouldResembleProto, []*ufsdevice.Config{cfg})
		Convey("DeviceConfigsExist should correctly report that config exists, and other config does not", func() {
			cfgIDs := []*ufsdevice.ConfigId{GetConfigID("board", "model", "variant"), GetConfigID("non", "existant", "config")}
			exists, err := DeviceConfigsExist(ctx, cfgIDs)

			So(err, ShouldBeNil)
			So(exists, ShouldResemble, []bool{true, false})
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
