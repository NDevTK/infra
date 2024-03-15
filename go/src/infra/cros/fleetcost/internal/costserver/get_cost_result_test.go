// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver_test

import (
	"context"
	"testing"

	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	testsupport "infra/cros/fleetcost/internal/costserver/testsupport"
)

func TestGetCostResult(t *testing.T) {
	t.Parallel()
	tf := testsupport.NewFixture(context.Background(), t)

	_, err := tf.Frontend.GetCostResult(tf.Ctx, &fleetcostAPI.GetCostResultRequest{})
	if err == nil {
		t.Error("err should not be nil")
	}
}
