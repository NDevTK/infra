// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dumper

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"

	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/model/registration"
	"infra/unifiedfleet/app/util"
)

// indexAssets updates the assset table thereby reindexing any new indexes that
// might be added to it. It is meant to be run during low-traffic/maintenance
// times as it attempts to index the entire table.
func indexAssets(ctx context.Context) error {
	logging.Infof(ctx, "indexAssets -- Starting to index the assets table")
	for _, ns := range util.ClientToDatastoreNamespace {
		newCtx, err := util.SetupDatastoreNamespace(ctx, ns)
		if err != nil {
			logging.Errorf(ctx, "indexAssets -- internal error, can't setup namespace %s. %v", ns, err)
			continue
		}
		for startToken := ""; ; {
			f := func(newCtx context.Context) error {
				var err error
				var res []*ufspb.Asset
				res, startToken, err = registration.ListAssets(newCtx, pageSize, startToken, nil, false)
				if err != nil {
					return errors.Annotate(err, "indexAssets[%s] -- Failed to list", ns).Err()
				}
				logging.Infof(ctx, "indexAssets -- Indexing %v assets in %s", len(res), ns)
				// Update the assets back to datastore
				_, err = registration.BatchUpdateAssets(newCtx, res)
				if err != nil {
					return errors.Annotate(err, "indexAssets[%s] -- Failed to update", ns).Err()
				}
				return nil
			}
			if err := datastore.RunInTransaction(newCtx, f, nil); err != nil {
				// Log the error. No point in throwing it here as it will be ignored
				logging.Errorf(newCtx, "Cannot index assets in %s: %v", ns, err)
			}
			if startToken == "" {
				break
			}
		}
	}
	logging.Infof(ctx, "indexAssets -- Done indexing the assets table")
	return nil
}
