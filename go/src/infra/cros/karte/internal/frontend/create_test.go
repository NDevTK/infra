// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/gae/service/datastore"

	kartepb "infra/cros/karte/api"
	"infra/cros/karte/internal/identifiers"
	"infra/cros/karte/internal/scalars"
)

// TestCreateActionWithClock tests creating an action with the testing clock set to 10 seconds after
// the beginning of time (UTC midnight on 1970-01-01).
func TestCreateActionWithClock(t *testing.T) {
	Convey("test create action with clock", t, func() {
		ctx := gaetesting.TestingContext()
		ctx = identifiers.Use(ctx, identifiers.NewDefault())
		testClock := testclock.New(time.Unix(10, 0).UTC())
		ctx = clock.Set(ctx, testClock)
		datastore.GetTestable(ctx).Consistent(true)
		k := NewKarteFrontend()

		action, err := k.CreateAction(ctx, &kartepb.CreateActionRequest{
			Action: &kartepb.Action{},
		})
		So(err, ShouldBeNil)
		So(action.Name[0:10], ShouldEqual, "zzzzUzzzzz")
		action.Name = ""
		So(action, ShouldResemble, &kartepb.Action{
			CreateTime: scalars.ConvertTimeToTimestampPtr(time.Unix(10, 0)),
			SealTime:   scalars.ConvertTimeToTimestampPtr(time.Unix(43210, 0)),
		})
	})
}
