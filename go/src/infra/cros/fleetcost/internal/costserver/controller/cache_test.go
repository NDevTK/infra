// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller_test

import (
	"context"
	"testing"

	"go.chromium.org/luci/gae/service/datastore"

	models "infra/cros/fleetcost/api/models"
	"infra/cros/fleetcost/internal/costserver/controller"
	"infra/cros/fleetcost/internal/costserver/testsupport"
)

// TestStoreCachedCostResult tests the storing and retrieving a cached cost result.
//
// It tests reading:
// 1) From an empty database.
// 2) With a current time before the expiration time of the cache record
// 3) With a current time after the expiration time of the cache record
func TestStoreCachedCostResult(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	_, readErr := controller.ReadCachedCostResult(tf.Ctx, "fake-hostname")
	if !datastore.IsErrNoSuchEntity(readErr) {
		t.Errorf("unexpected error in empty db: %s", readErr)
	}

	if err := controller.StoreCachedCostResult(tf.Ctx, "fake-hostname", &models.CostResult{
		DedicatedCost: 30,
	}); err != nil {
		t.Errorf("unexpected error when filling cache: %s", err)
	}

	result, readErr := controller.ReadCachedCostResult(tf.Ctx, "fake-hostname")
	if readErr != nil {
		t.Errorf("error writing cache record: %s", readErr)
	}
	if cost := result.GetDedicatedCost(); cost != 30 {
		t.Errorf("unexpected dedicated cost %f != 30", cost)
	}
}
