// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	testsupport "infra/cros/fleetcost/internal/costserver/testsupport"
)

// TestPing tests the ping API, which does nothing
func TestPingUFS(t *testing.T) {
	t.Parallel()

	tf := testsupport.NewFixture(context.Background(), t)

	tf.MockUFS.EXPECT().ListMachineLSEs(gomock.Any(), gomock.Any())

	_, err := tf.Frontend.PingUFS(tf.Ctx, &fleetcostAPI.PingUFSRequest{})
	if err != nil {
		t.Error(err)
	}
}
