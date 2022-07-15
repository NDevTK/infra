// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"infra/appengine/crosskylabadmin/internal/app/frontend/routing"
)

// TestRouteAuditTaskImpl tests routing audit tasks.
func TestRouteAuditTaskImpl(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("no config", t, func() {
		tt, r := routeAuditTaskImpl(ctx, nil)
		So(tt, ShouldEqual, routing.Legacy)
		So(r, ShouldEqual, routing.ParisNotEnabled)
	})
}
