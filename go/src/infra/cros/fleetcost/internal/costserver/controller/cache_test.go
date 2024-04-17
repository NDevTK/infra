// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.chromium.org/luci/gae/service/datastore"

	models "infra/cros/fleetcost/api/models"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/costserver/controller"
	"infra/cros/fleetcost/internal/costserver/entities"
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

	_, readErr := controller.ReadCachedCostResult(tf.Ctx, &fleetcostAPI.GetCostResultRequest{
		Hostname: "fake-hostname",
	}, tf.Time)
	if !datastore.IsErrNoSuchEntity(readErr) {
		t.Errorf("unexpected error in empty db: %s", readErr)
	}

	controller.StoreCachedCostResultDefer(tf.Ctx, "fake-hostname", tf.Time, &fleetcostAPI.GetCostResultResponse{
		Result: &models.CostResult{
			DedicatedCost: 30,
		},
	}, nil, nil)

	result, readErr := controller.ReadCachedCostResult(tf.Ctx, &fleetcostAPI.GetCostResultRequest{
		Hostname: "fake-hostname",
	}, tf.Time.Add(controller.CacheTTL-time.Second))
	if readErr != nil {
		t.Errorf("error writing cache record: %s", readErr)
	}
	if cost := result.GetResult().GetDedicatedCost(); cost != 30 {
		t.Errorf("unexpected dedicated cost %f != 30", cost)
	}

	_, readErr = controller.ReadCachedCostResult(tf.Ctx, &fleetcostAPI.GetCostResultRequest{
		Hostname: "fake-hostname",
	}, tf.Time.Add(controller.CacheTTL+time.Second))
	switch {
	case readErr == nil:
		t.Error("error is unexpectedly nil")
	case !strings.Contains(readErr.Error(), "too early"):
		t.Errorf("unexpected error: %s", readErr)
	}
}

// TestStoreCacheResultSkipWhenParentFunctionErrors tests that we do NOT write to the cache when the parent function has errored out.
//
// Note that the structure of this test is deliberately somewhat complicated.
// It has the same structure as the real call site of StoreCachedCostResult.
func TestStoreCacheResultDeferSkipWhenParentFunctionErrors(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	didErr := false

	f := func(ctx context.Context) (parentErr error) {
		defer controller.StoreCachedCostResultDefer(ctx, "a", tf.Time, &fleetcostAPI.GetCostResultResponse{
			Result: &models.CostResult{
				DedicatedCost: 100,
			},
		}, parentErr, func(e error) { didErr = true })
		return errors.New("aaaa an error")
	}

	_ = f(tf.Ctx)

	if didErr {
		t.Error("onErr callback should not have fired.")
	}
	n, err := datastore.Count(tf.Ctx, datastore.NewQuery(entities.CostIndicatorKind))
	if err != nil {
		t.Errorf("unexpected error when counting cost indicators: %s", err)
	}
	if n != 0 {
		t.Errorf("unexpected number of cost indicators in db after test: %d", n)
	}
}
