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
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kylelemons/godebug/pretty"
	. "github.com/smartystreets/goconvey/convey"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/app/frontend/test"
	"infra/appengine/crosskylabadmin/internal/app/gitstore/fakes"
	"infra/libs/skylab/inventory"
)

func TestResizePool(t *testing.T) {
	Convey("With 0 DUTs in target pool and 0 DUTs in spare pool", t, func(c C) {
		tf, validate := newTestFixture(t)
		defer validate()

		duts := []testInventoryDut{}
		err := tf.FakeGitiles.SetInventory(config.Get(tf.C).Inventory, fakes.InventoryData{
			Lab: inventoryBytesFromDUTs(duts),
			Infrastructure: inventoryBytesFromServers([]testInventoryServer{
				{
					hostname: "drone-queen-ENVIRONMENT_STAGING",
					dutIDs:   []string{},
				},
			}),
		})
		So(err, ShouldBeNil)

		Convey("ResizePool to 0 DUTs in target pool makes no changes", func() {
			resp, err := tf.Inventory.ResizePool(tf.C, &fleet.ResizePoolRequest{
				DutSelector: &fleet.DutSelector{
					Model: "link",
				},
				SparePool:      "DUT_POOL_SUITES",
				TargetPool:     "DUT_POOL_CQ",
				TargetPoolSize: 0,
			})
			So(err, ShouldBeNil)
			So(resp.Url, ShouldEqual, "")
			So(resp.Changes, ShouldHaveLength, 0)
		})

		Convey("ResizePool to 1 DUTs in target pool fails", func() {
			_, err := tf.Inventory.ResizePool(tf.C, &fleet.ResizePoolRequest{
				DutSelector: &fleet.DutSelector{
					Model: "link",
				},
				SparePool:      "DUT_POOL_SUITES",
				TargetPool:     "DUT_POOL_CQ",
				TargetPoolSize: 1,
			})
			So(err, ShouldNotBeNil)
		})
	})

	Convey("With 0 DUTs in target pool and 4 DUTs in spare pool", t, func(c C) {
		tf, validate := newTestFixture(t)
		defer validate()

		duts := []testInventoryDut{
			{"link_suites_0", "link_suites_0", "link", "DUT_POOL_SUITES"},
			{"link_suites_1", "link_suites_1", "link", "DUT_POOL_SUITES"},
			{"link_suites_2", "link_suites_2", "link", "DUT_POOL_SUITES"},
			{"link_suites_3", "link_suites_3", "link", "DUT_POOL_SUITES"},
		}
		err := tf.FakeGitiles.SetInventory(config.Get(tf.C).Inventory, fakes.InventoryData{
			Lab: inventoryBytesFromDUTs(duts),
			Infrastructure: inventoryBytesFromServers([]testInventoryServer{
				{
					hostname: "drone-queen-ENVIRONMENT_STAGING",
					dutIDs:   []string{"link_suites_0", "link_suites_1", "link_suites_2", "link_suites_3"},
				},
			}),
		})
		So(err, ShouldBeNil)

		Convey("ResizePool to 0 DUTs in target pool makes no changes", func() {
			resp, err := tf.Inventory.ResizePool(tf.C, &fleet.ResizePoolRequest{
				DutSelector: &fleet.DutSelector{
					Model: "link",
				},
				SparePool:      "DUT_POOL_SUITES",
				TargetPool:     "DUT_POOL_CQ",
				TargetPoolSize: 0,
			})
			So(err, ShouldBeNil)
			So(resp.Url, ShouldEqual, "")
			So(resp.Changes, ShouldHaveLength, 0)
		})

		Convey("ResizePool to 3 DUTs in target pool expands target pool", func() {
			resp, err := tf.Inventory.ResizePool(tf.C, &fleet.ResizePoolRequest{
				DutSelector: &fleet.DutSelector{
					Model: "link",
				},
				SparePool:      "DUT_POOL_SUITES",
				TargetPool:     "DUT_POOL_CQ",
				TargetPoolSize: 3,
			})
			So(err, ShouldBeNil)
			So(resp.Url, ShouldNotEqual, "")
			So(resp.Changes, ShouldHaveLength, 3)
			mc := poolChangeMap(resp.Changes)
			So(poolChangeCount(mc, "DUT_POOL_SUITES", "DUT_POOL_CQ"), ShouldEqual, 3)
		})

		Convey("ResizePool to 5 DUTs in target pool fails", func() {
			_, err := tf.Inventory.ResizePool(tf.C, &fleet.ResizePoolRequest{
				DutSelector: &fleet.DutSelector{
					Model: "link",
				},
				SparePool:      "DUT_POOL_SUITES",
				TargetPool:     "DUT_POOL_CQ",
				TargetPoolSize: 5,
			})
			So(err, ShouldNotBeNil)
		})
	})

	Convey("With 4 DUTs in target pool and 0 DUTs in spare pool", t, func(c C) {
		tf, validate := newTestFixture(t)
		defer validate()

		duts := []testInventoryDut{
			{"link_suites_0", "link_suites_0", "link", "DUT_POOL_CQ"},
			{"link_suites_1", "link_suites_1", "link", "DUT_POOL_CQ"},
			{"link_suites_2", "link_suites_2", "link", "DUT_POOL_CQ"},
			{"link_suites_3", "link_suites_3", "link", "DUT_POOL_CQ"},
		}
		err := tf.FakeGitiles.SetInventory(config.Get(tf.C).Inventory, fakes.InventoryData{
			Lab: inventoryBytesFromDUTs(duts),
			Infrastructure: inventoryBytesFromServers([]testInventoryServer{
				{
					hostname: "drone-queen-ENVIRONMENT_STAGING",
					dutIDs:   []string{"link_suites_0", "link_suites_1", "link_suites_2", "link_suites_3"},
				},
			}),
		})
		So(err, ShouldBeNil)

		Convey("ResizePool to 4 DUTs in target pool makes no changes", func() {
			resp, err := tf.Inventory.ResizePool(tf.C, &fleet.ResizePoolRequest{
				DutSelector: &fleet.DutSelector{
					Model: "link",
				},
				SparePool:      "DUT_POOL_SUITES",
				TargetPool:     "DUT_POOL_CQ",
				TargetPoolSize: 4,
			})
			So(err, ShouldBeNil)
			So(resp.Url, ShouldEqual, "")
			So(resp.Changes, ShouldHaveLength, 0)
		})

		Convey("ResizePool to 3 DUTs in target pool contracts target pool", func() {
			resp, err := tf.Inventory.ResizePool(tf.C, &fleet.ResizePoolRequest{
				DutSelector: &fleet.DutSelector{
					Model: "link",
				},
				SparePool:      "DUT_POOL_SUITES",
				TargetPool:     "DUT_POOL_CQ",
				TargetPoolSize: 3,
			})
			So(err, ShouldBeNil)
			So(resp.Url, ShouldNotEqual, "")
			So(resp.Changes, ShouldHaveLength, 1)
			mc := poolChangeMap(resp.Changes)
			So(poolChangeCount(mc, "DUT_POOL_CQ", "DUT_POOL_SUITES"), ShouldEqual, 1)
		})
	})

	Convey("With 4 DUTs in spare pool but 1 is in different env", t, func(c C) {
		tf, validate := newTestFixture(t)
		defer validate()

		duts := []testInventoryDut{
			{"link_suites_0", "link_suites_0", "link", "DUT_POOL_SUITES"},
			{"link_suites_1", "link_suites_1", "link", "DUT_POOL_SUITES"},
			{"link_suites_2", "link_suites_2", "link", "DUT_POOL_SUITES"},
			{"link_suites_3", "link_suites_3", "link", "DUT_POOL_SUITES"},
		}
		err := tf.FakeGitiles.SetInventory(config.Get(tf.C).Inventory, fakes.InventoryData{
			Lab: inventoryBytesFromDUTs(duts),
			Infrastructure: inventoryBytesFromServers([]testInventoryServer{
				{
					hostname: "drone-queen-ENVIRONMENT_STAGING",
					dutIDs:   []string{"link_suites_0", "link_suites_1", "link_suites_2"},
				},
			}),
		})
		So(err, ShouldBeNil)

		Convey("ResizePool to 4 DUTs in target pool raise error", func() {
			_, err := tf.Inventory.ResizePool(tf.C, &fleet.ResizePoolRequest{
				DutSelector: &fleet.DutSelector{
					Model: "link",
				},
				SparePool:      "DUT_POOL_SUITES",
				TargetPool:     "DUT_POOL_CQ",
				TargetPoolSize: 4,
			})
			So(err, ShouldNotBeNil)
		})

		Convey("ResizePool to 3 DUTs in target pool works", func() {
			resp, err := tf.Inventory.ResizePool(tf.C, &fleet.ResizePoolRequest{
				DutSelector: &fleet.DutSelector{
					Model: "link",
				},
				SparePool:      "DUT_POOL_SUITES",
				TargetPool:     "DUT_POOL_CQ",
				TargetPoolSize: 3,
			})
			So(err, ShouldBeNil)
			So(resp.Url, ShouldNotEqual, "")
			So(resp.Changes, ShouldHaveLength, 3)
			mc := poolChangeMap(resp.Changes)
			So(poolChangeCount(mc, "DUT_POOL_SUITES", "DUT_POOL_CQ"), ShouldEqual, 3)
		})
	})
}

func TestResizePoolCommit(t *testing.T) {
	Convey("With 0 DUTs in target pool and 1 DUTs in spare pool", t, func(c C) {
		tf, validate := newTestFixture(t)
		defer validate()

		duts := []testInventoryDut{
			{"link_suites_0", "link_suites_0", "link", "DUT_POOL_SUITES"},
		}
		err := tf.FakeGitiles.SetInventory(config.Get(tf.C).Inventory, fakes.InventoryData{
			Lab: inventoryBytesFromDUTs(duts),
			Infrastructure: inventoryBytesFromServers([]testInventoryServer{
				{
					hostname: "drone-queen-ENVIRONMENT_STAGING",
					dutIDs:   []string{"link_suites_0"},
				},
			}),
		})
		So(err, ShouldBeNil)

		Convey("ResizePool to 1 DUTs in target pool commits changes", func() {
			_, err := tf.Inventory.ResizePool(tf.C, &fleet.ResizePoolRequest{
				DutSelector: &fleet.DutSelector{
					Model: "link",
				},
				SparePool:      "DUT_POOL_SUITES",
				TargetPool:     "DUT_POOL_CQ",
				TargetPoolSize: 1,
			})
			So(err, ShouldBeNil)
			assertLabInventoryChange(c, tf.FakeGerrit, []testInventoryDut{
				{"link_suites_0", "link_suites_0", "link", "DUT_POOL_CQ"},
			})
		})

		Convey("ResizePool does not commit changes on error", func() {
			_, err := tf.Inventory.ResizePool(tf.C, &fleet.ResizePoolRequest{
				DutSelector: &fleet.DutSelector{
					Model: "link",
				},
				SparePool:      "DUT_POOL_SUITES",
				TargetPool:     "DUT_POOL_CQ",
				TargetPoolSize: 4,
			})
			So(err, ShouldNotBeNil)
			So(tf.FakeGerrit.Changes, ShouldHaveLength, 0)
		})
	})
}

// partialPoolChange contains a subset of the fleet.PoolChange fields.
//
// This struct is used for easy validation of relevant fields of
// fleet.PoolChange values returned from API responses.
type partialPoolChange struct {
	NewPool string
	OldPool string
}

// poolChangeMap converts a list of fleet.PoolChanges to a map from DutId to
// partialPoolChange.
//
// The returned map is more convenient for comparison with ShouldResemble
// assertions than the original list.
func poolChangeMap(pcs []*fleet.PoolChange) map[string]*partialPoolChange {
	mc := make(map[string]*partialPoolChange)
	for _, c := range pcs {
		mc[c.DutId] = &partialPoolChange{
			NewPool: c.NewPool,
			OldPool: c.OldPool,
		}
	}
	return mc
}

// poolChangeCount counts the number of partialPoolChanges in the map that move
// a DUT from oldPool to newPool.
func poolChangeCount(pcs map[string]*partialPoolChange, oldPool, newPool string) int {
	c := 0
	for _, pc := range pcs {
		if pc.OldPool == oldPool && pc.NewPool == newPool {
			c++
		}
	}
	return c
}

// assertLabInventoryChange verifies that the CL uploaded to gerrit contains the
// inventory of duts provided.
func assertLabInventoryChange(c C, fg *fakes.GerritClient, duts []testInventoryDut) {
	p := "data/skylab/lab.textpb"
	changes := fg.Changes
	So(changes, ShouldHaveLength, 1)
	change := changes[0]
	So(change.Files, ShouldContainKey, p)
	var actualLab inventory.Lab
	err := inventory.LoadLabFromString(change.Files[p], &actualLab)
	So(err, ShouldBeNil)
	var expectedLab inventory.Lab
	err = inventory.LoadLabFromString(string(inventoryBytesFromDUTs(duts)), &expectedLab)
	So(err, ShouldBeNil)
	// Sort before comparison
	want, _ := inventory.WriteLabToString(&expectedLab)
	got, _ := inventory.WriteLabToString(&actualLab)
	c.Printf("submitted incorrect lab -want +got: %s", pretty.Compare(strings.Split(want, "\n"), strings.Split(got, "\n")))
	So(want, ShouldEqual, got)
}

func expectDutsHealthFromSwarming(tf testFixture, bots []*swarming.SwarmingRpcsBotInfo) {
	tf.MockSwarming.EXPECT().ListAliveBotsInPool(
		gomock.Any(), gomock.Eq(config.Get(tf.C).Swarming.BotPool), gomock.Any(),
	).AnyTimes().DoAndReturn(test.FakeListAliveBotsInPool(bots))
}

func collectFailures(mrs map[string]*fleet.EnsurePoolHealthyResponse) []fleet.EnsurePoolHealthyResponse_Failure {
	ret := make([]fleet.EnsurePoolHealthyResponse_Failure, 0)
	for _, res := range mrs {
		ret = append(ret, res.Failures...)
	}
	return ret
}
