// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"go.chromium.org/luci/common/testing/assert/structuraldiff"
	"go.chromium.org/luci/common/testing/typed"

	fleetcostpb "infra/cros/fleetcost/api/models"
	"infra/cros/fleetcost/internal/costserver/controller"
	"infra/cros/fleetcost/internal/costserver/models"
	"infra/cros/fleetcost/internal/costserver/testsupport"
)

func TestPutCostIndicator(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	err := controller.PutCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
		ID: "fake-cost-indicator",
		CostIndicator: &fleetcostpb.CostIndicator{
			Name:  "a",
			Board: "e",
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	result, err := controller.GetCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
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

	costIndicator, err := controller.GetCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
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

	costIndicators, err := controller.ListCostIndicators(tf.Ctx, 1)
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

	if err := controller.PutCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
		ID: "fake-cost-indicator",
		CostIndicator: &fleetcostpb.CostIndicator{
			Name:  "fake-cost-indicator",
			Board: "old-board",
		},
	}); err != nil {
		t.Fatalf("failed to insert cost indicator: %s", err)
	}

	got, err := controller.UpdateCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
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
