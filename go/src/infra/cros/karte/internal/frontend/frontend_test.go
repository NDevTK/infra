// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"fmt"
	"testing"
	"time"

	cloudBQ "cloud.google.com/go/bigquery"
	"github.com/google/go-cmp/cmp"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/protobuf/testing/protocmp"

	kartepb "infra/cros/karte/api"
	"infra/cros/karte/internal/identifiers"
	"infra/cros/karte/internal/scalars"
)

const invalidProjectID = "invalid project ID -- 5509d052-1fec-4ff6-bb2f-bb4e98951520"

// TestCreateAction makes sure that CreateAction returns the action it created and that the action is present in datastore.
func TestCreateAction(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContext()
	ctx = identifiers.Use(ctx, identifiers.NewNaive())
	datastore.GetTestable(ctx).Consistent(true)
	k := NewKarteFrontend()
	resp, err := k.CreateAction(ctx, &kartepb.CreateActionRequest{
		Action: &kartepb.Action{
			Name:       "",
			Kind:       "ssh-attempt",
			CreateTime: scalars.ConvertTimeToTimestampPtr(time.Unix(1, 2)),
		},
	})
	expected := &kartepb.Action{
		Name:       fmt.Sprintf("%sentity001000000000", identifiers.IDVersion),
		Kind:       "ssh-attempt",
		SealTime:   scalars.ConvertTimeToTimestampPtr(time.Unix(1+12*60*60, 2)),
		CreateTime: scalars.ConvertTimeToTimestampPtr(time.Unix(1, 2)),
	}
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(expected, resp, protocmp.Transform()); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	// Here we inspect the contents of datastore.
	q, err := newActionEntitiesQuery("", "")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	datastoreActionEntities, _, err := q.Next(ctx, 0)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if datastoreActionEntities == nil {
		t.Errorf("action entities should not be nil")
	}
	switch len(datastoreActionEntities) {
	case 0:
		t.Errorf("datastore should not be empty")
	case 1:
	default:
		t.Errorf("datastore should not have more than 1 item")
	}
}

// TestRejectActionWithUserDefinedName tests that an action with a user-defined name is rejected.
func TestRejectActionWithUserDefinedName(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContext()
	datastore.GetTestable(ctx).Consistent(true)
	k := NewKarteFrontend()
	resp, err := k.CreateAction(ctx, &kartepb.CreateActionRequest{
		Action: &kartepb.Action{
			Name: "aaaaa",
			Kind: "ssh-attempt",
		},
	})
	if resp != nil {
		t.Errorf("unexpected response: %s", resp.String())
	}
	if err == nil {
		t.Errorf("expected response to be rejected")
	}
}

// TestCreateActionWithNoTime tests that creating an action without a time succeeds and supplies the current time.
// See b/206651512 for details.
func TestCreateActionWithNoTime(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContext()
	datastore.GetTestable(ctx).Consistent(true)
	// Set a test clock to an arbitrary time to make sure that the correct time is supplied.
	testClock := testclock.New(time.Unix(3, 4))
	ctx = clock.Set(ctx, testClock)
	ctx = identifiers.Use(ctx, identifiers.NewDefault())

	k := NewKarteFrontend()

	resp, err := k.CreateAction(ctx, &kartepb.CreateActionRequest{
		Action: &kartepb.Action{
			Name: "",
			Kind: "ssh-attempt",
		},
	})

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if resp == nil {
		t.Errorf("resp should not be nil")
	}
	expected := time.Unix(3, 4)
	actual := scalars.ConvertTimestampPtrToTime(resp.GetCreateTime())
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected diff: %s", diff)
	}
}

// TestCreateActionWithSwarmingAndBuildbucketID tests creating a new action with an swarming ID and a buildbucket ID and reading it back.
func TestCreateActionWithSwarmingAndBuildbucketID(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContext()
	datastore.GetTestable(ctx).Consistent(true)
	testClock := testclock.New(time.Unix(3, 4))
	ctx = clock.Set(ctx, testClock)
	ctx = identifiers.Use(ctx, identifiers.NewNaive())

	k := NewKarteFrontend()

	expected := []*kartepb.Action{
		{
			Name:           fmt.Sprintf(identifiers.NaiveIDFmt, identifiers.IDVersion, identifiers.NaiveFirstID),
			Kind:           "ssh-attempt",
			SwarmingTaskId: "a",
			BuildbucketId:  "b",
			CreateTime:     scalars.ConvertTimeToTimestampPtr(time.Unix(3, 0)),
			SealTime:       scalars.ConvertTimeToTimestampPtr(time.Unix(3+12*60*60, 0)),
		},
	}

	_, err := k.CreateAction(ctx, &kartepb.CreateActionRequest{
		Action: &kartepb.Action{
			Name:           "",
			Kind:           "ssh-attempt",
			SwarmingTaskId: "a",
			BuildbucketId:  "b",
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	resp, err := k.ListActions(ctx, &kartepb.ListActionsRequest{
		Filter: `kind == "ssh-attempt"`,
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	actual := resp.GetActions()

	if diff := cmp.Diff(expected, actual, protocmp.Transform()); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}

// TestCreateObservation makes sure that that CreateObservation fails because
// it isn't implemented.
func TestCreateObservation(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContext()
	datastore.GetTestable(ctx).Consistent(true)
	k := NewKarteFrontend()
	_, err := k.CreateObservation(ctx, &kartepb.CreateObservationRequest{})
	if err == nil {
		t.Error("expected Create Observation to fail")
	}
}

// TestListActionsSmokeTest tests that ListActions does not error.
func TestListActionsSmokeTest(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContext()
	datastore.GetTestable(ctx).Consistent(true)
	k := NewKarteFrontend()
	resp, err := k.ListActions(ctx, &kartepb.ListActionsRequest{})
	if resp == nil {
		t.Errorf("expected resp to not be nil")
	}
	if len(resp.GetActions()) != 0 {
		t.Errorf("expected actions to be trivial")
	}
	if err != nil {
		t.Errorf("expected error to be nil not %s", err)
	}
}

// TestListActions tests that ListActions errors.
func TestListActions(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContext()
	datastore.GetTestable(ctx).Consistent(true)
	if err := PutActionEntities(
		ctx,
		&ActionEntity{
			ID: "aaaa",
		},
	); err != nil {
		t.Error(err)
	}
	k := NewKarteFrontend()
	resp, err := k.ListActions(ctx, &kartepb.ListActionsRequest{})
	if err != nil {
		t.Errorf("expected error to be nil not %s", err)
	}
	if resp == nil {
		t.Errorf("expected resp to not be nil")
	}
	if resp.GetActions() == nil {
		t.Errorf("expected actions to not be nil")
	}
	if len(resp.GetActions()) != 1 {
		t.Errorf("expected len(actions) to be 1 not %d", len(resp.GetActions()))
	}
}

// TestListObservations tests that ListObservations errors.
func TestListObservations(t *testing.T) {
	t.Parallel()
	k := NewKarteFrontend()
	ctx := gaetesting.TestingContext()
	datastore.GetTestable(ctx).Consistent(true)
	resp, err := k.ListObservations(ctx, &kartepb.ListObservationsRequest{})
	if resp == nil {
		t.Errorf("expected resp to not be nil")
	}
	if err != nil {
		t.Errorf("expected error to be nil not %s", err)
	}
}

type fakeClient struct {
	items        [][]cloudBQ.ValueSaver
	observations [][]cloudBQ.ValueSaver
}

func (c *fakeClient) getInserter(dataset string, table string) bqInserter {
	return func(ctx context.Context, item []cloudBQ.ValueSaver) error {
		c.items = append(c.items, item)
		if table == "observations" {
			c.observations = append(c.observations, item)
		}
		return nil
	}
}

// size returns the total number of items.
func (c *fakeClient) size() int {
	out := 0
	for _, row := range c.items {
		out += len(row)
	}
	return out
}

// observationsSize returns the total number of observations.
func (c *fakeClient) observationsSize() int {
	out := 0
	for _, row := range c.observations {
		out += len(row)
	}
	return out
}

// TestPersistActionRangeImpl_SmokeTest tests that persisting a range of actions
// returns a non-error response given an empty dataset
func TestPersistActionRangeImpl_SmokeTest(t *testing.T) {
	t.Parallel()
	k := NewKarteFrontend().(*karteFrontend)
	ctx := gaetesting.TestingContext()
	datastore.GetTestable(ctx).Consistent(true)
	fake := &fakeClient{}

	resp, err := k.persistActionRangeImpl(ctx, fake, &kartepb.PersistActionRangeRequest{
		StartTime: scalars.ConvertTimeToTimestampPtr(time.Unix(1, 0)),
		StopTime:  scalars.ConvertTimeToTimestampPtr(time.Unix(2, 0)),
	})
	if resp == nil {
		t.Errorf("expected resp not to be nil")
	}
	if err != nil {
		t.Errorf("expected resp to be nil not %s", err)
	}
}
