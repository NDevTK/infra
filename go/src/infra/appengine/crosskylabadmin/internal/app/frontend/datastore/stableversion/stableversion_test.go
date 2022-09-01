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

package stableversion

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
)

func TestUpdateAndGet(t *testing.T) {
	// TODO(gregorynisbet): convert to test fixture
	ctx := context.Background()
	ctx = memory.Use(ctx)

	model := "xxx-model"
	buildTarget := "xxx-buildTarget"
	crosVersion := "xxx-cros-version"
	firmwareVersion := "xxx-firmware-version"
	faftVersion := "xxx-faft-version"

	Convey("StableVersion datastore", t, func() {
		Convey("Cros", func() {
			Convey("Cros not present initially", func() {
				item, err := GetCrosStableVersion(ctx, buildTarget, model)
				So(err, ShouldNotBeNil)
				So(item, ShouldEqual, "")
			})
			Convey("Cros write should succeed", func() {
				err := PutSingleCrosStableVersion(ctx, buildTarget, model, crosVersion)
				So(err, ShouldBeNil)
			})
			Convey("Cros present after write", func() {
				item, err := GetCrosStableVersion(ctx, buildTarget, model)
				So(err, ShouldBeNil)
				So(item, ShouldEqual, crosVersion)
			})
			Convey("Cros is not case-sensitive", func() {
				err := PutSingleCrosStableVersion(ctx, "AAA", "model", crosVersion)
				So(err, ShouldBeNil)
				item, err := GetCrosStableVersion(ctx, "Aaa", "model")
				So(err, ShouldBeNil)
				So(item, ShouldEqual, crosVersion)
			})
		})
		Convey("Faft", func() {
			Convey("Faft not present initially", func() {
				item, err := GetFaftStableVersion(ctx, buildTarget, model)
				So(err, ShouldNotBeNil)
				So(item, ShouldEqual, "")
			})
			Convey("Faft write should succeed", func() {
				err := PutSingleFaftStableVersion(ctx, buildTarget, model, faftVersion)
				So(err, ShouldBeNil)
			})
			Convey("Faft present after write", func() {
				item, err := GetFaftStableVersion(ctx, buildTarget, model)
				So(err, ShouldBeNil)
				So(item, ShouldEqual, faftVersion)
			})
			Convey("Faft is not case-sensitive", func() {
				err := PutSingleFaftStableVersion(ctx, "AAA", "BBB", faftVersion)
				So(err, ShouldBeNil)
				item, err := GetFaftStableVersion(ctx, "Aaa", "Bbb")
				So(err, ShouldBeNil)
				So(item, ShouldEqual, faftVersion)
			})
		})
		Convey("Firmware", func() {
			Convey("Firmware not present initially", func() {
				item, err := GetFirmwareStableVersion(ctx, buildTarget, model)
				So(err, ShouldNotBeNil)
				So(item, ShouldEqual, "")
			})
			Convey("Firmware write should succeed", func() {
				err := PutSingleFirmwareStableVersion(ctx, buildTarget, model, firmwareVersion)
				So(err, ShouldBeNil)
			})
			Convey("Firmware present after write", func() {
				item, err := GetFirmwareStableVersion(ctx, buildTarget, model)
				So(err, ShouldBeNil)
				So(item, ShouldEqual, firmwareVersion)
			})
			Convey("Firmware is not case-sensitive", func() {
				err := PutSingleFirmwareStableVersion(ctx, "AAA", "BBB", firmwareVersion)
				So(err, ShouldBeNil)
				item, err := GetFirmwareStableVersion(ctx, "Aaa", "Bbb")
				So(err, ShouldBeNil)
				So(item, ShouldEqual, firmwareVersion)
			})
		})
	})
}

func TestRemoveEmptyKeyOrValue(t *testing.T) {
	Convey("remove non-conforming keys and values", t, func() {
		ctx := context.Background()
		ctx = memory.Use(ctx)
		Convey("remove empty key good", func() {
			m := map[string]string{"": "a"}
			removeEmptyKeyOrValue(ctx, m)
			So(len(m), ShouldEqual, 0)
		})
		Convey("remove empty value good", func() {
			m := map[string]string{"a": ""}
			removeEmptyKeyOrValue(ctx, m)
			So(len(m), ShouldEqual, 0)
		})
		Convey("remove empty key and value good", func() {
			m := map[string]string{"": ""}
			removeEmptyKeyOrValue(ctx, m)
			So(len(m), ShouldEqual, 0)
		})
		Convey("remove conforming key and value bad", func() {
			m := map[string]string{"k": "v"}
			removeEmptyKeyOrValue(ctx, m)
			So(len(m), ShouldEqual, 1)
		})
	})
}

// TestDeleteAll tests the DeleteAll function by adding some things to datastore and then deleting them
func TestDeleteAll(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("test delete all", t, func() {
		const board = "01dcd1a3-ec21-4f92-9acf-f9fac3ec51a2"
		const model = "2da5161a-ab39-4f2f-9eab-0e44fd2f0c77"
		const cros = "942929ba-bab4-4267-b3a7-776a9aae26fe"

		ctx = memory.Use(ctx)
		datastore.GetTestable(ctx).Consistent(true)
		tally, err := datastore.Count(ctx, datastore.NewQuery(CrosStableVersionKind).KeysOnly(true))
		So(err, ShouldBeNil)
		So(tally, ShouldEqual, 0) // no records initially.
		err = PutSingleCrosStableVersion(ctx, board, model, cros)
		So(err, ShouldBeNil)
		tally, err = datastore.Count(ctx, datastore.NewQuery(CrosStableVersionKind).KeysOnly(true))
		So(err, ShouldBeNil)
		So(tally, ShouldEqual, 1) // one record now.
		err = DeleteAll(ctx, CrosStableVersionKind)
		So(err, ShouldBeNil)
		tally, err = datastore.Count(ctx, datastore.NewQuery(CrosStableVersionKind).KeysOnly(true))
		So(err, ShouldBeNil)
		So(tally, ShouldEqual, 0) // We delete the records, back to zero.
	})
}

// TestMakeKeyBatches tests reading out keys in batches.
func TestMakeKeyBatches(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("test make batches", t, func() {
		ctx = memory.Use(ctx)
		datastore.GetTestable(ctx).Consistent(true)

		So(PutSingleCrosStableVersion(ctx, "b", "1", "x"), ShouldBeNil)
		So(PutSingleCrosStableVersion(ctx, "b", "2", "x"), ShouldBeNil)
		So(PutSingleCrosStableVersion(ctx, "b", "3", "x"), ShouldBeNil)
		So(PutSingleCrosStableVersion(ctx, "b", "4", "x"), ShouldBeNil)
		So(PutSingleCrosStableVersion(ctx, "b", "5", "x"), ShouldBeNil)
		So(PutSingleCrosStableVersion(ctx, "b", "6", "x"), ShouldBeNil)

		tally, err := datastore.Count(ctx, datastore.NewQuery(CrosStableVersionKind).KeysOnly(true))
		So(err, ShouldBeNil)
		So(tally, ShouldEqual, 6)

		batches, err := makeKeyBatches(ctx, CrosStableVersionKind, 2)
		So(err, ShouldBeNil)
		So(len(batches), ShouldEqual, 3)
		for _, batch := range batches {
			So(len(batch), ShouldEqual, 2)
		}
	})
}
