// Copyright 2018 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package inventory

import (
	"context"

	"go.chromium.org/luci/common/retry"
	"go.chromium.org/luci/grpc/grpcutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/frontend/dutpool"
)

// ResizePool implements the method from fleet.InventoryServer interface.
func (is *ServerImpl) ResizePool(ctx context.Context, req *fleet.ResizePoolRequest) (resp *fleet.ResizePoolResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(ctx, err)
	}()

	if err = req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	err = retry.Retry(
		ctx,
		transientErrorRetries(),
		func() error {
			var ierr error
			resp, ierr = is.resizePoolNoRetry(ctx, req)
			return ierr
		},
		retry.LogCallback(ctx, "resizePoolNoRetry"),
	)
	return resp, err
}

func (is *ServerImpl) resizePoolNoRetry(ctx context.Context, req *fleet.ResizePoolRequest) (*fleet.ResizePoolResponse, error) {
	ic, err := is.newInventoryClient(ctx)

	duts, err := ic.selectDutsFromInventory(ctx, req.DutSelector)
	if err != nil {
		return nil, err
	}
	changes, err := dutpool.Resize(duts, req.TargetPool, int(req.TargetPoolSize), req.SparePool)
	if err != nil {
		return nil, err
	}
	u, err := ic.commitBalancePoolChanges(ctx, changes)
	if err != nil {
		return nil, err
	}
	return &fleet.ResizePoolResponse{
		Url:     u,
		Changes: changes,
	}, nil
}
