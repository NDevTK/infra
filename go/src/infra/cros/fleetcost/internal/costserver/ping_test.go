// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver_test

import (
	"context"
	"testing"

	fleetcostpb "infra/cros/fleetcost/api"
	testsupport "infra/cros/fleetcost/internal/costserver/testsupport"
)

// TestPing tests the ping API, which does nothing
func TestPing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tf := testsupport.NewFixture(ctx, t)

	_, err := tf.Frontend.Ping(tf.Ctx, &fleetcostpb.PingRequest{})
	if err != nil {
		t.Error(err)
	}
}
