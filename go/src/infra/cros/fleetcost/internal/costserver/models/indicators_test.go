// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package models_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/type/money"

	"go.chromium.org/luci/common/testing/assert/structuraldiff"
	"go.chromium.org/luci/common/testing/typed"
	"go.chromium.org/luci/gae/service/datastore"

	fleetcostpb "infra/cros/fleetcost/api/models"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver/models"
	"infra/cros/fleetcost/internal/costserver/testsupport"
)

// TestCostIndicatorSimple tests putting a cost indicator into database and retrieving it.
func TestCostIndicatorSimple(t *testing.T) {
	t.Parallel()
	tf := testsupport.NewFixture(context.Background(), t)

	if err := datastore.Put(tf.Ctx, &models.CostIndicatorEntity{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board:       "e",
			BurnoutRate: 12.0,
		},
	}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if err := datastore.Get(tf.Ctx, &models.CostIndicatorEntity{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board:       "e",
			BurnoutRate: 12.0,
		},
	}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

// TestCostIndicatorSimple tests that writing a cost indicator to the database populates the correct fields.
func TestCostIndicatorIndexedFields(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	oldIndicator := &models.CostIndicatorEntity{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board: "e",
			Model: "w",
		},
	}

	err := datastore.Put(tf.Ctx, oldIndicator)
	if err != nil {
		t.Fatal(err)
	}

	item := &models.CostIndicatorEntity{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board: "e",
			Model: "w",
		},
	}
	if err := datastore.Get(tf.Ctx, item); err != nil {
		t.Error(err)
	}

	if diff := typed.Got(item).Want(&models.CostIndicatorEntity{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board: "e",
			Model: "w",
		},
		Board: "e",
		Model: "w",
	}).Options(cmp.AllowUnexported(models.CostIndicatorEntity{})).Diff(); diff != "" {
		t.Errorf("unexpected error (-want +got): %s", diff)
	}
}

// TestCostIndicatorClone tests cloning a cost indicator
func TestCostIndicatorClone(t *testing.T) {
	t.Parallel()

	oldIndicator := &models.CostIndicatorEntity{
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
		CostIndicator: &fleetcostpb.CostIndicator{
			Board:       "e",
			BurnoutRate: 12.0,
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	result, err := models.GetCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board: "e",
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if result.CostIndicator.GetBurnoutRate() != 12.0 {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestGetCostIndicator(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixtureWithData(context.Background(), t)

	costIndicator, err := models.GetCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board: "e",
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	want := &models.CostIndicatorEntity{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board:       "e",
			BurnoutRate: 44.0,
		},
		Board: "e",
	}

	if diff := structuraldiff.DebugCompare(costIndicator, want).String(); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}

// TestListCostIndicator tests listing all cost indicators in a scenario where this is only
// one cost indicator.
func TestListCostIndicator(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixtureWithData(context.Background(), t)

	costIndicators, err := models.ListCostIndicators(tf.Ctx, 1, nil)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	want := []*fleetcostpb.CostIndicator{
		{
			Board:       "e",
			BurnoutRate: 44.0,
		},
	}

	if diff := structuraldiff.DebugCompare(costIndicators, want).String(); diff != "" {
		t.Errorf("unexpected diff: %s", diff)
	}
}

// TestListCostIndicatorWithModelFilter tests listing devices with a model filter.
func TestListCostIndicatorWithModelFilter(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)
	if _, err := tf.Frontend.CreateCostIndicator(tf.Ctx, &fleetcostAPI.CreateCostIndicatorRequest{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board:    "fake-board-1",
			Model:    "fake-model",
			Location: fleetcostpb.Location_LOCATION_ACS,
			Type:     fleetcostpb.IndicatorType_INDICATOR_TYPE_CLOUD,
			Cost: &money.Money{
				CurrencyCode: "USD",
				Units:        100,
			},
		},
	}); err != nil {
		panic(err)
	}
	if _, err := tf.Frontend.CreateCostIndicator(tf.Ctx, &fleetcostAPI.CreateCostIndicatorRequest{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board:    "fake-board-2",
			Model:    "fake-model",
			Location: fleetcostpb.Location_LOCATION_ACS,
			Type:     fleetcostpb.IndicatorType_INDICATOR_TYPE_CLOUD,
			Cost: &money.Money{
				CurrencyCode: "USD",
				Units:        200,
			},
		},
	}); err != nil {
		panic(err)
	}
	if _, err := tf.Frontend.CreateCostIndicator(tf.Ctx, &fleetcostAPI.CreateCostIndicatorRequest{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board:    "fake-board-2",
			Model:    "a-different-model",
			Location: fleetcostpb.Location_LOCATION_ACS,
			Type:     fleetcostpb.IndicatorType_INDICATOR_TYPE_CLOUD,
			Cost: &money.Money{
				CurrencyCode: "USD",
				Units:        200,
			},
		},
	}); err != nil {
		panic(err)
	}

	resp, err := tf.Frontend.ListCostIndicators(tf.Ctx, &fleetcostAPI.ListCostIndicatorsRequest{
		PageSize: 1000,
		Filter: &fleetcostAPI.ListCostIndicatorsFilter{
			Model: "fake-model",
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if diff := typed.Got(len(resp.GetCostIndicator())).Want(2).Diff(); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}

// TestUpdateCostIndicatorHappyPath tests updating a cost indicator that already exists.
//
// Note that when updating the record, we provide an argument that
func TestUpdateCostIndicatorHappyPath(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	if err := models.PutCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board:       "fake-board",
			BurnoutRate: 12.0,
		},
	}); err != nil {
		t.Fatalf("failed to insert cost indicator: %s", err)
	}

	got, err := models.UpdateCostIndicatorEntity(tf.Ctx, &models.CostIndicatorEntity{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board:       "fake-board",
			BurnoutRate: 14.0,
		},
		Board: "fake-board",
	}, []string{"burnout_rate"})
	if err != nil {
		t.Errorf("unexpected error: %q", err)
	}

	if diff := typed.Got(got).Want(&models.CostIndicatorEntity{
		CostIndicator: &fleetcostpb.CostIndicator{
			Board:       "fake-board",
			BurnoutRate: 14.0,
		},
		Board: "fake-board",
	}).Options(cmp.AllowUnexported(*got)).Diff(); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}
