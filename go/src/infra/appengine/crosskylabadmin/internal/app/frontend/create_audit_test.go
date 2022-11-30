// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/app/frontend/routing"
)

// TestRouteAuditTaskImpl tests routing audit tasks.
func TestRouteAuditTaskImpl(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("no config", t, func() {
		tt, r := routeAuditTaskImpl(ctx, nil, "", 0.0)
		So(tt, ShouldEqual, routing.Paris)
		So(r, ShouldEqual, routing.ParisNotEnabled)
	})
	Convey("invalid random float", t, func() {
		tt, r := routeAuditTaskImpl(ctx, &config.RolloutConfig{}, "", 12.0)
		So(tt, ShouldEqual, routing.Paris)
		So(r, ShouldEqual, routing.InvalidRangeArgument)
	})
	Convey("bad permille info", t, func() {
		pat := &config.RolloutConfig{
			Pattern: []*config.RolloutConfig_Pattern{
				{Pattern: "^", ProdPermille: 4},
			},
		}
		res := pat.ComputePermilleData(ctx, "hostname")
		So(res, ShouldBeNil)
	})
	Convey("25-25 split", t, func() {
		pd := &config.RolloutConfig{Enable: true, ProdPermille: 250, LatestPermille: 250}
		Convey("0.24", func() {
			tt, r := routeAuditTaskImpl(ctx, pd, "", 0.24)
			So(tt, ShouldEqual, routing.ParisLatest)
			So(r, ShouldEqual, routing.ScoreBelowThreshold)
		})
		Convey("0.26", func() {
			tt, r := routeAuditTaskImpl(ctx, pd, "", 0.26)
			So(tt, ShouldEqual, routing.Paris)
			So(r, ShouldEqual, routing.ScoreBelowThreshold)
		})
	})
	Convey("Repair-only field", t, func() {
		pd := &config.RolloutConfig{Enable: true, OptinAllDuts: true}
		tt, r := routeAuditTaskImpl(ctx, pd, "", 0.24)
		So(tt, ShouldEqual, routing.Paris)
		So(r, ShouldEqual, routing.RepairOnlyField)
	})
}
