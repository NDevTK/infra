// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package zone_selector

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"

	"infra/vm_leaser/internal/constants"
)

func TestSelectZone(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	allZones := []string{}
	for _, a := range constants.ChromeOSZones {
		for _, b := range a {
			allZones = append(allZones, b)
		}
	}

	Convey("Test SelectZone", t, func() {
		Convey("SelectZone - zone provided; return zone", func() {
			req := &api.LeaseVMRequest{
				HostReqs: &api.VMRequirements{
					GceImage:  "test-image",
					GceRegion: "test-region",
				},
			}
			z := SelectZone(ctx, req, 1)
			So(z, ShouldEqual, "test-region")
		})
		Convey("SelectZone - select single random zone", func() {
			req := &api.LeaseVMRequest{
				HostReqs: &api.VMRequirements{
					GceImage: "test-image",
				},
			}
			z := SelectZone(ctx, req, 1)
			So(allZones, ShouldContain, z)
		})
		Convey("SelectZone - select single random zone for ChromeOS testing client", func() {
			req := &api.LeaseVMRequest{
				HostReqs: &api.VMRequirements{
					GceImage: "test-image",
				},
				TestingClient: api.VMTestingClient_VM_TESTING_CLIENT_CHROMEOS,
			}
			z := SelectZone(ctx, req, 1)
			So(allZones, ShouldContain, z)
		})
		Convey("SelectZone - check distribution of zones", func() {
			req := &api.LeaseVMRequest{
				HostReqs: &api.VMRequirements{
					GceImage: "test-image",
				},
			}

			zonesDict := make(map[string]int)

			// run select zone for a reasonably large number
			var numZones int64 = 10000
			var c int64 = 0
			for c < numZones {
				z := SelectZone(ctx, req, c)
				zonesDict[z] = zonesDict[z] + 1
				c = c + 1
			}

			// all keys should be valid zones
			for k := range zonesDict {
				So(allZones, ShouldContain, k)
			}

			// Since we have 4 main zones and 13 total zones, the distribution for
			// each zone must be around 9% per zone. To give the randomness some
			// leeway, the zone distribution should be 4-14% (9±5%) per subzone.
			dist := 1.0 / float64(len(allZones))
			for _, a := range allZones {
				So(zonesDict[a], ShouldBeLessThan, float64(numZones)*(dist+0.05))
				So(zonesDict[a], ShouldBeGreaterThan, float64(numZones)*(dist-0.05))
			}
		})
	})
}

func TestGetZoneSubnet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Test GetZoneSubnet", t, func() {
		Convey("GetZoneSubnet - happy path", func() {
			z, err := GetZoneSubnet(ctx, "test-region-1")
			So(err, ShouldBeNil)
			So(z, ShouldEqual, "regions/test-region/subnetworks/test-region")
		})
		Convey("GetZoneSubnet - bad zone", func() {
			z, err := GetZoneSubnet(ctx, "test-region")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "zone is malformed; needs to be xxx-yyy-zzz")
			So(z, ShouldEqual, "")
		})
	})
}

func TestExtractGoogleApiZone(t *testing.T) {
	t.Parallel()

	Convey("Test ExtractGoogleApiZone", t, func() {
		Convey("ExtractGoogleApiZone - happy path", func() {
			z, err := ExtractGoogleApiZone("https://www.googleapis.com/compute/v1/projects/chrome-fleet-vm-leaser-dev/zones/us-central1-b")
			So(err, ShouldBeNil)
			So(z, ShouldEqual, "us-central1-b")
		})
		Convey("ExtractGoogleApiZone - bad zone", func() {
			z, err := ExtractGoogleApiZone("https://www.googleapis.com/compute/v1/projects/chrome-fleet-vm-leaser-dev/zones/us-central1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "google api zone uri is malformed")
			So(z, ShouldEqual, "")
		})
		Convey("ExtractGoogleApiZone - no zone", func() {
			z, err := ExtractGoogleApiZone("https://www.googleapis.com/compute/v1/projects/chrome-fleet-vm-leaser-dev/zones")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "google api zone uri is malformed")
			So(z, ShouldEqual, "")
		})
	})
}

func TestValidateZone(t *testing.T) {
	t.Parallel()

	Convey("Test validateZone", t, func() {
		Convey("validateZone - happy path", func() {
			err := validateZone("us-central1-b")
			So(err, ShouldBeNil)
		})
		Convey("validateZone - bad zone", func() {
			err := validateZone("us-central1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "zone is malformed; needs to be xxx-yyy-zzz")
		})
		Convey("validateZone - no zone", func() {
			err := validateZone("")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "zone is malformed; needs to be xxx-yyy-zzz")
		})
	})
}
