// Copyright 2019 The LUCI Authors.
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
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/errors"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/app/gitstore/fakes"
	"infra/libs/skylab/inventory"
)

func mockValidateDeviceconfig(ctx context.Context, ic inventoryClient, nds []*inventory.CommonDeviceSpecs) error {
	return nil
}

func TestDeleteDutsWithSplitInventory(t *testing.T) {
	Convey("With 3 DUTs in the split inventory", t, func() {
		ctx := testingContext()
		ctx = withSplitInventory(ctx)
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		setSplitGitilesDuts(tf.C, tf.FakeGitiles, []testInventoryDut{
			{id: "dut1_id", hostname: "jetstream-host", model: "link", pool: "DUT_POOL_SUITES"},
			{id: "dut2_id", hostname: "chromeos6-rack1-row2-host3", model: "link", pool: "DUT_POOL_SUITES"},
			{id: "dut3_id", hostname: "chromeos15-rack1-row2-host3", model: "link", pool: "DUT_POOL_SUITES"},
		})

		Convey("DeleteDuts with no hostnames returns error", func() {
			_, err := tf.Inventory.DeleteDuts(tf.C, &fleet.DeleteDutsRequest{})
			So(err, ShouldNotBeNil)
		})

		Convey("DeleteDuts with unknown hostname deletes no duts", func() {
			resp, err := tf.Inventory.DeleteDuts(tf.C, &fleet.DeleteDutsRequest{Hostnames: []string{"unknown_hostname"}})
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.GetIds(), ShouldBeEmpty)
		})

		Convey("DeleteDuts with known hostnames deletes duts", func() {
			resp, err := tf.Inventory.DeleteDuts(tf.C, &fleet.DeleteDutsRequest{Hostnames: []string{"jetstream-host", "chromeos6-rack1-row2-host3"}})
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(stringset.NewFromSlice(resp.GetIds()...), ShouldResemble, stringset.NewFromSlice("dut1_id", "dut2_id"))

			So(tf.FakeGerrit.Changes, ShouldHaveLength, 1)
			change := tf.FakeGerrit.Changes[len(tf.FakeGerrit.Changes)-1]
			So(change.Files, ShouldHaveLength, 2)
			var paths []string
			for p := range change.Files {
				paths = append(paths, p)
			}
			So(stringset.NewFromSlice(paths...), ShouldResemble, stringset.NewFromSlice(
				"data/skylab/chromeos-misc/jetstream-host.textpb",
				"data/skylab/chromeos6/chromeos6-rack1-row2-host3.textpb",
			))
		})
	})
}

// TODO(xixuan): remove this after per-file inventory is landed and tested.
func TestDeleteDuts(t *testing.T) {
	Convey("With 3 DUTs in the inventory", t, func() {
		tf, validate := newTestFixture(t)
		defer validate()

		err := tf.FakeGitiles.SetInventory(config.Get(tf.C).Inventory, fakes.InventoryData{
			Lab: inventoryBytesFromDUTs([]testInventoryDut{
				{"dut_id_1", "dut_hostname_1", "link", "DUT_POOL_SUITES"},
				{"dut_id_2", "dut_hostname_2", "link", "DUT_POOL_SUITES"},
				{"dut_id_3", "dut_hostname_3", "link", "DUT_POOL_SUITES"},
			}),
		})
		So(err, ShouldBeNil)

		Convey("DeleteDuts with no hostnames returns error", func() {
			_, err := tf.Inventory.DeleteDuts(tf.C, &fleet.DeleteDutsRequest{})
			So(err, ShouldNotBeNil)
		})

		Convey("DeleteDuts with unknown hostname deletes no duts", func() {
			resp, err := tf.Inventory.DeleteDuts(tf.C, &fleet.DeleteDutsRequest{Hostnames: []string{"unknown_hostname"}})
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.GetIds(), ShouldBeEmpty)
		})

		Convey("DeleteDuts with known hostnames deletes duts", func() {
			resp, err := tf.Inventory.DeleteDuts(tf.C, &fleet.DeleteDutsRequest{Hostnames: []string{"dut_hostname_1", "dut_hostname_2"}})
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(stringset.NewFromSlice(resp.GetIds()...), ShouldResemble, stringset.NewFromSlice("dut_id_1", "dut_id_2"))
		})
	})

	Convey("With 2 DUTs with the same hostname", t, func() {
		tf, validate := newTestFixture(t)
		defer validate()

		err := tf.FakeGitiles.SetInventory(config.Get(tf.C).Inventory, fakes.InventoryData{
			Lab: inventoryBytesFromDUTs([]testInventoryDut{
				{"dut_id_1", "dut_hostname", "link", "DUT_POOL_SUITES"},
				{"dut_id_2", "dut_hostname", "link", "DUT_POOL_SUITES"},
			}),
		})
		So(err, ShouldBeNil)

		Convey("DeleteDuts with known hostname deletes both duts", func() {
			resp, err := tf.Inventory.DeleteDuts(tf.C, &fleet.DeleteDutsRequest{Hostnames: []string{"dut_hostname"}})
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(stringset.NewFromSlice(resp.GetIds()...), ShouldResemble, stringset.NewFromSlice("dut_id_1", "dut_id_2"))
		})
	})
}

// getLastChangeForHost gets the latest change for a given path of one host in gerrit for per-file inventory.
func getLastChangeForHost(fg *fakes.GerritClient, path string) (*inventory.Lab, error) {
	if len(fg.Changes) == 0 {
		return nil, errors.Reason("found no gerrit changes").Err()
	}

	change := fg.Changes[len(fg.Changes)-1]
	content, ok := change.Files[path]
	if !ok {
		return nil, errors.Reason(fmt.Sprintf("cannot find path %s in %v", path, change.Files)).Err()
	}
	var oneDutLab inventory.Lab
	err := inventory.LoadLabFromString(content, &oneDutLab)
	return &oneDutLab, err
}

// getLabFromLastChange gets the latest inventory.Lab committed to
// fakes.GerritClient
func getLabFromLastChange(fg *fakes.GerritClient) (*inventory.Lab, error) {
	if len(fg.Changes) == 0 {
		return nil, errors.Reason("found no gerrit changes").Err()
	}

	change := fg.Changes[len(fg.Changes)-1]
	f, ok := change.Files["data/skylab/lab.textpb"]
	if !ok {
		return nil, errors.Reason("No modification to Lab in gerrit change").Err()
	}
	var lab inventory.Lab
	err := inventory.LoadLabFromString(f, &lab)
	return &lab, err
}

// getInfrastructureFromLastChange gets the latest inventory.Infrastructure
// committed to fakes.GerritClient
func getInfrastructureFromLastChange(fg *fakes.GerritClient) (*inventory.Infrastructure, error) {
	if len(fg.Changes) == 0 {
		return nil, errors.Reason("found no gerrit changes").Err()
	}

	change := fg.Changes[len(fg.Changes)-1]
	f, ok := change.Files["data/skylab/server_db.textpb"]
	if !ok {
		return nil, errors.Reason("No modification to Infrastructure in gerrit change").Err()
	}
	var infra inventory.Infrastructure
	err := inventory.LoadInfrastructureFromString(f, &infra)
	return &infra, err
}
