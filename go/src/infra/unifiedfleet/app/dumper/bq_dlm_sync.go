// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/api/iterator"

	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/model/registration"
	"infra/unifiedfleet/app/util"
)

const (
	DlmBqCbxQuery = `
	SELECT googlePartNumber
	FROM ` + "`cros-device-lifecycle-manager.prod.devices`" + ` AS d
	  JOIN ` + "`cros-device-lifecycle-manager.prod.device_skus`" + ` AS s
	  ON d.deviceId=s.deviceId
	WHERE (d.hwXComplianceVersion != 0 OR s.hwXComplianceVersion !=0)
	  AND googlePartNumber IS NOT null
	`
	DlmBqFingerprintQuery = ""
	DlmBqTouchscreenQuery = ""
)

type deviceGpn struct {
	GooglePartNumber string
}

// SyncDutInfoFromDlmBq fetches certain asset data from DLM BQ table.
func SyncDutInfoFromDlmBq(ctx context.Context) error {
	if err := datastore.RunInTransaction(ctx, syncDutInfoFromDlmBqCbx, nil); err != nil {
		return errors.Annotate(err, "Failed to sync DLM BQ for Cbx").Err()
	}
	return nil
}

func syncDutInfoFromDlmBqCbx(ctx context.Context) error {
	ctx, err := util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	if err != nil {
		return errors.Annotate(err, "failed to set namespace").Err()
	}

	gpns, err := syncDutInfoFromDlmBqQueryBq(ctx, DlmBqCbxQuery)
	if err != nil {
		return err
	}

	lses, err := getMachineLSEsFromGpn(ctx, gpns)
	if err != nil {
		return err
	}
	for _, lse := range lses {
		if lse.GetChromeosMachineLse().GetDeviceLse().GetDut() == nil {
			logging.Infof(ctx, "Skipping machineLSE %s, DUT not found", lse.GetName())
		}
		lse.GetChromeosMachineLse().GetDeviceLse().GetDut().Cbx = true
	}

	if _, err := inventory.BatchUpdateMachineLSEs(ctx, lses); err != nil {
		return errors.Annotate(err, "Unable to update machinelses").Err()
	}
	logging.Debugf(ctx, "Successfully updated DUT cbx info to UFS datastore")
	return nil
}

func syncDutInfoFromDlmBqQueryBq(ctx context.Context, query string) ([]string, error) {
	bqClient := get(ctx)
	q := bqClient.Query(query)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	gpns := make([]string, 0)
	for {
		var gpn deviceGpn
		err = it.Next(&gpn)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		gpns = append(gpns, gpn.GooglePartNumber)
	}
	return gpns, nil
}

func getMachineLSEsFromGpn(ctx context.Context, gpns []string) ([]*ufspb.MachineLSE, error) {
	lses := make([]*ufspb.MachineLSE, 0)
	for _, g := range gpns {
		if g == "" {
			logging.Infof(ctx, "Skipping empty GPN")
			continue
		}
		machines, err := registration.QueryMachineByPropertyName(ctx, "gpn", g, true)
		if err != nil {
			return nil, errors.Annotate(err, "Failed to query machines for gpn %s", g).Err()
		}
		for _, m := range machines {
			machinelses, err := inventory.QueryMachineLSEByPropertyName(ctx, "machine_ids", m.GetName(), false)
			if err != nil {
				return nil, errors.Annotate(err, "Failed to query machinelse for machine %s", m.GetName()).Err()
			}
			lses = append(lses, machinelses...)
		}
	}
	return lses, nil
}
