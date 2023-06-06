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

type indexTableFn func(context.Context, string, *string) error

// IndexAssets updates the asset table thereby reindexing any new indexes that
// might be added to it. It is meant to be run during low-traffic/maintenance
// times as it attempts to index the entire table.
func IndexAssets(ctx context.Context) error {
	f := func(newCtx context.Context, namespace string, startToken *string) error {
		var err error
		var res []*ufspb.Asset
		res, *startToken, err = registration.ListAssets(newCtx, pageSize, *startToken, nil, false)
		if err != nil {
			return errors.Annotate(err, "indexAssets[%s] -- Failed to list", namespace).Err()
		}
		logging.Infof(ctx, "indexAssets -- Indexing %v assets in %s", len(res), namespace)
		// Update the assets back to datastore
		_, err = registration.BatchUpdateAssets(newCtx, res)
		if err != nil {
			return errors.Annotate(err, "indexAssets[%s] -- Failed to update", namespace).Err()
		}
		return nil
	}
	return indexTable(ctx, "assets", f)
}

// IndexMachines updates the machine table thereby reindexing any new indexes that
// might be added to it. It is meant to be run during low-traffic/maintenance
// times as it attempts to index the entire table.
func IndexMachines(ctx context.Context) error {
	f := func(newCtx context.Context, namespace string, startToken *string) error {
		var err error
		var res []*ufspb.Machine
		res, *startToken, err = registration.ListMachines(newCtx, pageSize, *startToken, nil, false)
		if err != nil {
			return errors.Annotate(err, "indexMachines[%s] -- Failed to list", namespace).Err()
		}
		logging.Infof(ctx, "indexMachines -- Indexing %v machines in %s", len(res), namespace)
		// Update the machines back to datastore
		_, err = registration.BatchUpdateMachines(newCtx, res)
		if err != nil {
			return errors.Annotate(err, "indexMachines[%s] -- Failed to update", namespace).Err()
		}
		return nil
	}
	return indexTable(ctx, "machines", f)
}

// indexRacks updates the rack table thereby reindexing any new indexes that
// might be added to it. It is meant to be run during low-traffic/maintenance
// times as it attempts to index the entire table.
func indexRacks(ctx context.Context) error {
	f := func(newCtx context.Context, namespace string, startToken *string) error {
		var err error
		var res []*ufspb.Rack
		res, *startToken, err = registration.ListRacks(newCtx, pageSize, *startToken, nil, false)
		if err != nil {
			return errors.Annotate(err, "indexRacks[%s] -- Failed to list", namespace).Err()
		}
		logging.Infof(ctx, "indexRacks -- Indexing %v racks in %s", len(res), namespace)
		// Update the rack back to datastore
		_, err = registration.BatchUpdateRacks(newCtx, res)
		if err != nil {
			return errors.Annotate(err, "indexRacks[%s] -- Failed to update", namespace).Err()
		}
		return nil
	}
	return indexTable(ctx, "racks", f)
}

func indexTable(ctx context.Context, tableName string, fn indexTableFn) error {
	logging.Infof(ctx, "indexTable -- Starting to index the %s table", tableName)
	for _, ns := range util.ClientToDatastoreNamespace {
		newCtx, err := util.SetupDatastoreNamespace(ctx, ns)
		if err != nil {
			logging.Errorf(ctx, "indexTable -- internal error, can't setup namespace %s. %v", ns, err)
			continue
		}
		for startToken := ""; ; {
			f := func(newCtx context.Context) error {
				return fn(newCtx, ns, &startToken)
			}
			if err := datastore.RunInTransaction(newCtx, f, nil); err != nil {
				// Log the error. No point in throwing it here as it will be ignored
				logging.Errorf(newCtx, "Cannot index %s in %s: %v", tableName, ns, err)
			}
			if startToken == "" {
				break
			}
		}
	}
	logging.Infof(ctx, "indexTable -- Done indexing the %s table", tableName)
	return nil
}
