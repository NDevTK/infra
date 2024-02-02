// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"
	"testing"

	fleetcostpb "infra/cros/fleetcost/api"
)

func TestCreateCostIndicator(t *testing.T) {
	t.Parallel()
	tf := newFixture(context.Background(), t)

	_, err := tf.frontend.CreateCostIndicator(tf.ctx, &fleetcostpb.CreateCostIndicatorRequest{})
	if err == nil {
		t.Error("err is unexpectedly nil")
	}
}
