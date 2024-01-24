// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	fleetcostpb "infra/cros/fleetcost/api"
)

// TestPing tests the ping API, which does nothing
func TestPingUFS(t *testing.T) {
	t.Parallel()

	tf := newFixture(context.Background(), t)

	tf.mockUFS.EXPECT().ListMachineLSEs(gomock.Any(), gomock.Any())

	_, err := tf.frontend.PingUFS(tf.ctx, &fleetcostpb.PingUFSRequest{})
	if err != nil {
		t.Error(err)
	}
}
