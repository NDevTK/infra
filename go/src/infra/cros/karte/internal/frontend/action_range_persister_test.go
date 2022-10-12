// Copyright 2022 The ChromiumOS Authors. All rights reserved.
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
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/service/datastore"

	kartepb "infra/cros/karte/api"
	"infra/cros/karte/internal/identifiers"
)

// TestActionRangePersisterInsufficientInput tests that invoking actionRangePersister without
// specifying a time range fails with an error message telling you that the action that you took
// is perhaps technically valid, but is so likely to be an error that we're rejecting it for you
// as a favor.
func TestActionRangePersisterInsufficientInput(t *testing.T) {
	t.Parallel()
	Convey("insufficient input", t, func() {
		ctx := gaetesting.TestingContext()
		ctx = identifiers.Use(ctx, identifiers.NewNaive())
		testClock := testclock.New(time.Unix(10, 0).UTC())
		ctx = clock.Set(ctx, testClock)
		datastore.GetTestable(ctx).Consistent(true)

		_, err := makeQuery(ctx, &actionRangePersistOptions{})
		So(err, ShouldErrLike, "rejecting likely erroneous call")
	})
}

// TestActionRangePersisterSmokeTest tests that attempting to persist data in cases where no data
// actually exists, successfully does nothing.
func TestActionRangePersisterSmokeTest(t *testing.T) {
	t.Parallel()
	Convey("smoke test", t, func() {
		ctx := gaetesting.TestingContext()
		ctx = identifiers.Use(ctx, identifiers.NewNaive())
		testClock := testclock.New(time.Unix(10, 0).UTC())
		ctx = clock.Set(ctx, testClock)
		datastore.GetTestable(ctx).Consistent(true)
		a := &actionRangePersistOptions{
			startID: time.Unix(0, 0).UTC(),
			stopID:  time.Unix(20, 0).UTC(),
		}
		q, err := makeQuery(ctx, a)
		So(err, ShouldBeNil)
		_, _, err = persistActions(ctx, a, q.Query)
		So(err, ShouldBeNil)
		var actions []*ActionEntity
		So(datastore.GetAll(ctx, datastore.NewQuery(ActionKind), &actions), ShouldBeNil)
		So(len(actions), ShouldEqual, 0)
		So(persistObservations(ctx, a), ShouldBeNil)
		var observations []*ObservationEntity
		So(datastore.GetAll(ctx, datastore.NewQuery(ObservationKind), &observations), ShouldBeNil)
		So(len(observations), ShouldEqual, 0)
	})
}

// TestActionRangePersister tests persisting two actions and two observations.
func TestActionRangePersister(t *testing.T) {
	t.Parallel()
	Convey("test with several actions", t, func() {
		ctx := gaetesting.TestingContext()
		ctx = identifiers.Use(ctx, identifiers.NewNaive())
		testClock := testclock.New(time.Unix(10, 0).UTC())
		ctx = clock.Set(ctx, testClock)
		datastore.GetTestable(ctx).Consistent(true)
		fake := &fakeClient{}

		k := NewKarteFrontend()

		action1 := func() string {
			resp, err := k.CreateAction(ctx, &kartepb.CreateActionRequest{
				Action: &kartepb.Action{
					Name: "",
					Kind: "ssh-attempt",
				},
			})
			So(err, ShouldBeNil)
			return resp.GetName()
		}()
		So(action1, ShouldNotBeEmpty)

		action2 := func() string {
			resp, err := k.CreateAction(ctx, &kartepb.CreateActionRequest{
				Action: &kartepb.Action{
					Name: "",
					Kind: "ssh-attempt",
				},
			})
			So(err, ShouldBeNil)
			return resp.GetName()
		}()
		So(action2, ShouldNotBeEmpty)

		observation1 := func() string {
			resp, err := k.CreateObservation(ctx, &kartepb.CreateObservationRequest{
				Observation: &kartepb.Observation{
					ActionName: action1,
				},
			})
			So(err, ShouldBeNil)
			return resp.GetName()
		}()
		So(observation1, ShouldNotBeEmpty)

		observation2 := func() string {
			resp, err := k.CreateObservation(ctx, &kartepb.CreateObservationRequest{
				Observation: &kartepb.Observation{
					ActionName: action2,
				},
			})
			So(err, ShouldBeNil)
			return resp.GetName()
		}()
		So(observation2, ShouldNotBeEmpty)
		a := &actionRangePersistOptions{
			startID: time.Unix(0, 0).UTC(),
			stopID:  time.Unix(20, 0).UTC(),
			bq:      fake,
		}
		q, err := makeQuery(ctx, a)
		So(err, ShouldBeNil)
		_, _, err = persistActions(ctx, a, q.Query)
		So(err, ShouldBeNil)
		var actions []*ActionEntity
		So(datastore.GetAll(ctx, datastore.NewQuery(ActionKind), &actions), ShouldBeNil)
		So(len(actions), ShouldEqual, 2)
		So(persistObservations(ctx, a), ShouldBeNil)
		var observations []*ObservationEntity
		So(datastore.GetAll(ctx, datastore.NewQuery(ObservationKind), &observations), ShouldBeNil)
		So(len(observations), ShouldEqual, 2)
		// These two checks down here are the highest-value checks. They check the total number of
		// bigquery records produced and the number of those records that are observations, respectively.
		So(fake.size(), ShouldEqual, 4)
		So(fake.observationsSize(), ShouldEqual, 2)
	})
}
