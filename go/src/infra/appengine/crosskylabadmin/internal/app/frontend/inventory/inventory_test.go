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
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	dataSV "infra/appengine/crosskylabadmin/internal/app/frontend/datastore/stableversion"
)

const (
	gpu = "fakeGPU"
	// dut should follow the following rules:
	// 1) entries should be in alphabetical order.
	// 2) indent is 2 spaces, no tabs.
	dut = `duts {
  common {
    environment: ENVIRONMENT_STAGING
    hostname: "dut_hostname"
    id: "dut_id_1"
    labels {
      capabilities {
        carrier: CARRIER_INVALID
        gpu_family: "%s"
        graphics: ""
        power: ""
        storage: ""
      }
      critical_pools: DUT_POOL_SUITES
      model: "link"
      peripherals {
      }
    }
  }
}
`

	emptyStableVersions = `{
	"cros": [],
	"faft": [],
	"firmware": []
}`

	stableVersions = `{
    "cros":[
        {
            "key":{
                "buildTarget":{
                    "name":"auron_paine"
                },
                "modelId":{
                    "value":""
                }
            },
            "version":"R78-12499.40.0"
        }
    ],
    "faft":[
        {
            "key": {
                "buildTarget": {
                    "name": "auron_paine"
                },
                "modelId": {
                    "value": "auron_paine"
                }
            },
            "version": "auron_paine-firmware/R39-6301.58.98"
        }
    ],
    "firmware":[
        {
            "key": {
                "buildTarget": {
                    "name": "auron_paine"
                },
                "modelId": {
                    "value": "auron_paine"
                }
            },
            "version": "Google_Auron_paine.6301.58.98"
        }
    ]
}`

	stableVersionWithEmptyVersions = `{
    "cros":[
        {
            "key":{
                "buildTarget":{
                    "name":"auron_paine"
                },
                "modelId":{
                    "value":""
                }
            },
            "version":""
        }
    ],
    "faft":[
        {
            "key": {
                "buildTarget": {
                    "name": "auron_paine"
                },
                "modelId": {
                    "value": "auron_paine"
                }
            },
            "version": ""
        }
    ],
    "firmware":[
        {
            "key": {
                "buildTarget": {
                    "name": "auron_paine"
                },
                "modelId": {
                    "value": "auron_paine"
                }
            },
            "version": ""
        }
    ]
}`
)

func TestDumpStableVersionToDatastore(t *testing.T) {
	Convey("Dump Stable version smoke test", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		tf.setStableVersionFactory("{}")
		is := tf.Inventory
		resp, err := is.DumpStableVersionToDatastore(ctx, nil)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
	})
	Convey("Update Datastore from empty stableversions file", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		tf.setStableVersionFactory(emptyStableVersions)
		_, err := tf.Inventory.DumpStableVersionToDatastore(ctx, nil)
		So(err, ShouldBeNil)
	})
	Convey("Update Datastore from non-empty stableversions file", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		tf.setStableVersionFactory(stableVersions)
		_, err := tf.Inventory.DumpStableVersionToDatastore(ctx, nil)
		So(err, ShouldBeNil)
		cros, err := dataSV.GetCrosStableVersion(ctx, "auron_paine", "auron_paine")
		So(err, ShouldBeNil)
		So(cros, ShouldEqual, "R78-12499.40.0")
		firmware, err := dataSV.GetFirmwareStableVersion(ctx, "auron_paine", "auron_paine")
		So(err, ShouldBeNil)
		So(firmware, ShouldEqual, "Google_Auron_paine.6301.58.98")
		faft, err := dataSV.GetFaftStableVersion(ctx, "auron_paine", "auron_paine")
		So(err, ShouldBeNil)
		So(faft, ShouldEqual, "auron_paine-firmware/R39-6301.58.98")
	})
	Convey("skip entries with empty version strings", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		tf.setStableVersionFactory(stableVersionWithEmptyVersions)
		defer validate()
		resp, err := tf.Inventory.DumpStableVersionToDatastore(ctx, nil)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		_, err = dataSV.GetCrosStableVersion(ctx, "auron_paine", "auron_paine")
		So(err, ShouldNotBeNil)
		_, err = dataSV.GetFirmwareStableVersion(ctx, "auron_paine", "auron_paine")
		So(err, ShouldNotBeNil)
		_, err = dataSV.GetFaftStableVersion(ctx, "auron_paine", "auron_paine")
		So(err, ShouldNotBeNil)
	})
}

func TestStableVersionFileParsing(t *testing.T) {
	Convey("Parse non-empty stableversions", t, func() {
		ctx := testingContext()
		parsed, err := parseStableVersions(stableVersions)
		So(err, ShouldBeNil)
		So(parsed, ShouldNotBeNil)
		So(len(parsed.GetCros()), ShouldEqual, 1)
		So(parsed.GetCros()[0].GetVersion(), ShouldEqual, "R78-12499.40.0")
		So(parsed.GetCros()[0].GetKey(), ShouldNotBeNil)
		So(parsed.GetCros()[0].GetKey().GetBuildTarget(), ShouldNotBeNil)
		So(parsed.GetCros()[0].GetKey().GetBuildTarget().GetName(), ShouldEqual, "auron_paine")
		records := getStableVersionRecords(ctx, parsed)
		So(len(records.cros), ShouldEqual, 1)
		So(len(records.firmware), ShouldEqual, 1)
		So(len(records.faft), ShouldEqual, 1)
	})
}
