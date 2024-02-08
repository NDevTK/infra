// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller_test

import (
	"context"
	"testing"

	"go.chromium.org/luci/common/testing/assert/structuraldiff"

	fleetcostpb "infra/cros/fleetcost/api"
	"infra/cros/fleetcost/internal/costserver/controller"
	"infra/cros/fleetcost/internal/costserver/models"
	"infra/cros/fleetcost/internal/costserver/testsupport"
)

func TestPutCostIndicator(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	err := controller.PutCostIndicator(tf.Ctx, &models.CostIndicator{
		ID: "fake-cost-indicator",
		CostIndicator: &fleetcostpb.CostIndicator{
			Name:  "a",
			Board: "e",
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	result, err := controller.GetCostIndicator(tf.Ctx, &models.CostIndicator{
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

	costIndicator, err := controller.GetCostIndicator(tf.Ctx, &models.CostIndicator{
		ID: "fake-cost-indicator",
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	want := &models.CostIndicator{
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
