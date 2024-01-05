// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"google.golang.org/appengine/datastore"

	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/model/registration"
)

// CreateDefaultWifi creates a default wifi in the datastore.
func CreateDefaultWifi(ctx context.Context, wifi *ufspb.DefaultWifi) (*ufspb.DefaultWifi, error) {
	f := func(ctx context.Context) error {
		if _, err := registration.NonAtomicBatchCreateDefaultWifis(ctx, []*ufspb.DefaultWifi{wifi}); err != nil {
			return errors.Annotate(err, "Unable to create wifi %s", wifi).Err()
		}
		hc := getDefaultWifiHistoryClient()
		hc.logDefaultWifiChanges(nil, wifi)
		return hc.SaveChangeEvents(ctx)
	}
	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		return nil, errors.Annotate(err, "CreateDefaultWifi for %s", wifi).Err()
	}
	return wifi, nil
}

func GetDefaultWifi(ctx context.Context, name string) (*ufspb.DefaultWifi, error) {
	return registration.GetDefaultWifi(ctx, name)
}

func getDefaultWifiHistoryClient() *HistoryClient {
	return &HistoryClient{}
}
