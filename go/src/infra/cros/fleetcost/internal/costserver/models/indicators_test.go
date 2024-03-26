// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package models_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"go.chromium.org/luci/common/testing/assert/structuraldiff"
	"go.chromium.org/luci/common/testing/typed"
	"go.chromium.org/luci/gae/service/datastore"

	fleetcostpb "infra/cros/fleetcost/api/models"
	"infra/cros/fleetcost/internal/costserver/models"
	"infra/cros/fleetcost/internal/costserver/testsupport"
)

// TestCostIndicatorSimple tests putting a cost indicator into database and retrieving it.
func TestCostIndicatorSimple(t *testing.T) {
	t.Parallel()
	tf := testsupport.NewFixture(context.Background(), t)

	if err := datastore.Put(tf.Ctx, &models.CostIndicatorEntity{
		ID: "a",
		CostIndicator: &fleetcostpb.CostIndicator{
			Name:  "a",
			Board: "e",
		},
	}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if err := datastore.Get(tf.Ctx, &models.CostIndicatorEntity{
		ID: "a",
	}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

// TestCostIndicatorClone tests cloning a cost indicator
func TestCostIndicatorClone(t *testing.T) {
	t.Parallel()

	oldIndicator := &models.CostIndicatorEntity{
		ID: "a",
		CostIndicator: &fleetcostpb.CostIndicator{
			Name:  "a",
			Board: "e",
		},
	}

	newIndicator := oldIndicator.Clone()

	if diff := typed.Got(newIndicator).Want(oldIndicator).Options(cmp.AllowUnexported(*oldIndicator)).Diff(); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}

func TestPutCostIndicator(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	err := models.PutCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
		ID: "fake-cost-indicator",
		CostIndicator: &fleetcostpb.CostIndicator{
			Name:  "a",
			Board: "e",
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	result, err := models.GetCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
		ID: "fake-cost-indicator",
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if result.CostIndicator.GetName() != "a" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestGetCostIndicator(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixtureWithData(context.Background(), t)

	costIndicator, err := models.GetCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
		ID: "fake-cost-indicator",
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	want := &models.CostIndicatorEntity{
		ID: "fake-cost-indicator",
		CostIndicator: &fleetcostpb.CostIndicator{
			Name:  "a",
			Board: "e",
		},
	}

	if diff := structuraldiff.DebugCompare(costIndicator, want).String(); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}

func TestListCostIndicator(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixtureWithData(context.Background(), t)

	costIndicators, err := models.ListCostIndicators(tf.Ctx, 1)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	want := []*fleetcostpb.CostIndicator{
		{
			Name:  "a",
			Board: "e",
		},
	}

	if diff := structuraldiff.DebugCompare(costIndicators, want).String(); diff != "" {
		t.Errorf("unexpected diff: %s", diff)
	}
}

// TestUpdateCostIndicatorHappyPath tests updating a cost indicator that already exists.
//
// Note that when updating the record, we provide an argument that
func TestUpdateCostIndicatorHappyPath(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	if err := models.PutCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
		ID: "fake-cost-indicator",
		CostIndicator: &fleetcostpb.CostIndicator{
			Name:  "fake-cost-indicator",
			Board: "old-board",
		},
	}); err != nil {
		t.Fatalf("failed to insert cost indicator: %s", err)
	}

	got, err := models.UpdateCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
		ID: "fake-cost-indicator",
		CostIndicator: &fleetcostpb.CostIndicator{
			Name:  "fake-cost-indicator",
			Board: "new-board",
		},
	}, []string{"name", "board"})
	if err != nil {
		t.Errorf("unexpected error: %q", err)
	}

	if diff := typed.Got(got).Want(&models.CostIndicatorEntity{
		ID: "fake-cost-indicator",
		CostIndicator: &fleetcostpb.CostIndicator{
			Name:  "fake-cost-indicator",
			Board: "new-board",
		},
	}).Options(cmp.AllowUnexported(*got)).Diff(); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}
