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
	"infra/libs/skylab/common/heuristics"
)

// TestRouteAuditTaskImpl tests routing audit tasks.
func TestRouteAuditTaskImpl(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("no config", t, func() {
		tt, r := routeAuditTaskImpl(ctx, nil, "", 0.0)
		So(tt, ShouldEqual, routing.Legacy)
		So(r, ShouldEqual, routing.ParisNotEnabled)
	})
	Convey("invalid random float", t, func() {
		tt, r := routeAuditTaskImpl(ctx, &config.RolloutConfig{}, "", 12.0)
		So(tt, ShouldEqual, routing.Legacy)
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
	Convey("routing not enabled", t, func() {
		ctx := context.Background()
		tt, r := routeAuditTaskImpl(ctx, &config.RolloutConfig{Enable: false}, "", 0.0)
		So(tt, ShouldEqual, heuristics.LegacyTaskType)
		So(r, ShouldEqual, routing.ParisNotEnabled)
	})
}
