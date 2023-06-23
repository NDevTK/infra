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
