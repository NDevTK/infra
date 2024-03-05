// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package models_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"go.chromium.org/luci/common/testing/typed"
	"go.chromium.org/luci/gae/service/datastore"

	fleetcostpb "infra/cros/fleetcost/api"
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
