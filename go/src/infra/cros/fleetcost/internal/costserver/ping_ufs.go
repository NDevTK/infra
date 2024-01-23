// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package costserver

import (
	"context"
	"errors"

	fleetcostpb "infra/cros/fleetcost/api"
)

// PingUFS takes a PingUFSRequest which is empty and pings UFS, returning a descriptionof what it did.
func (f *FleetCostFrontend) PingUFS(context.Context, *fleetcostpb.PingUFSRequest) (*fleetcostpb.PingUFSResponse, error) {
	return nil, errors.New("not yet implemented")
}
