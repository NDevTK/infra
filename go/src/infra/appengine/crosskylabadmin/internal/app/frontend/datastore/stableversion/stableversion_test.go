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
	"go.chromium.org/luci/appengine/gaetesting"
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

// TestImposeVersion tests updating and deleting a stable version entry in datastore through the ImposeVersion interface.
func TestImposeVersion(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContext()
	datastore.GetTestable(ctx).Consistent(true)
	Convey("test impose version", t, func() {
		Convey("cros", func() {
			e := &CrosStableVersionEntity{
				ID:   "eve;eve",
				Cros: "a",
			}
			So(datastore.Put(ctx, e), ShouldBeNil)
			So(e.ImposeVersion(ctx, "b"), ShouldBeNil)
			var ents []*CrosStableVersionEntity
			So(datastore.GetAll(ctx, datastore.NewQuery(CrosStableVersionKind), &ents), ShouldBeNil)
			So(len(ents), ShouldEqual, 1)
			So(ents[0].Cros, ShouldEqual, "b")
			So(e.ImposeVersion(ctx, ""), ShouldBeNil)
		})
		Convey("faft", func() {
			e := &FaftStableVersionEntity{
				ID:   "eve;eve",
				Faft: "a",
			}
			So(datastore.Put(ctx, e), ShouldBeNil)
			So(e.ImposeVersion(ctx, "b"), ShouldBeNil)
			var ents []*FaftStableVersionEntity
			So(datastore.GetAll(ctx, datastore.NewQuery(FaftStableVersionKind), &ents), ShouldBeNil)
			So(len(ents), ShouldEqual, 1)
			So(ents[0].Faft, ShouldEqual, "b")
			So(e.ImposeVersion(ctx, ""), ShouldBeNil)
		})
		Convey("firmware", func() {
			e := &FirmwareStableVersionEntity{
				ID:       "eve;eve",
				Firmware: "a",
			}
			So(datastore.Put(ctx, e), ShouldBeNil)
			So(e.ImposeVersion(ctx, "b"), ShouldBeNil)
			var ents []*FirmwareStableVersionEntity
			So(datastore.GetAll(ctx, datastore.NewQuery(FirmwareStableVersionKind), &ents), ShouldBeNil)
			So(len(ents), ShouldEqual, 1)
			So(ents[0].Firmware, ShouldEqual, "b")
			So(e.ImposeVersion(ctx, ""), ShouldBeNil)
		})
	})
}
