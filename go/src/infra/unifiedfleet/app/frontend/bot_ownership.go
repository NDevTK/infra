// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"

	ufspb "infra/unifiedfleet/api/v1/models"
	api "infra/unifiedfleet/api/v1/rpc"
)

// GetOwnershipData returns the ownership data for a given host.
func (fs *FleetServerImpl) GetOwnershipData(ctx context.Context, req *api.GetOwnershipDataRequest) (response *ufspb.OwnershipData, err error) {
	// TODO(b/248054750) - Implement this method
	return nil, nil
}
