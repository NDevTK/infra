// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ufspb "infra/unifiedfleet/api/v1/models"
	ufsds "infra/unifiedfleet/app/model/datastore"
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

func ListDefaultWifis(ctx context.Context, pageSize int32, pageToken, filter string, keysOnly bool) (res []*ufspb.DefaultWifi, nextPageToken string, err error) {
	// DefaultWifi has no filters.
	filterMap := map[string][]interface{}{}
	q, err := ufsds.ListQuery(ctx, registration.DefaultWifiKind, pageSize, pageToken, filterMap, keysOnly)
	if err != nil {
		return nil, "", err
	}
	var nextCur datastore.Cursor
	err = datastore.Run(ctx, q, func(ent *registration.DefaultWifiEntry, cb datastore.CursorCB) error {
		if keysOnly {
			wifi := &ufspb.DefaultWifi{
				Name: ent.ID,
			}
			res = append(res, wifi)
		} else {
			pm, err := ent.GetProto()
			if err != nil {
				logging.Errorf(ctx, "Failed to UnMarshal: %s", err)
				return nil
			}
			res = append(res, pm.(*ufspb.DefaultWifi))
		}
		if len(res) >= int(pageSize) {
			if nextCur, err = cb(); err != nil {
				return err
			}
			return datastore.Stop
		}
		return nil
	})
	if err != nil {
		logging.Errorf(ctx, "Failed to list DefaultWifi: %s", err)
		return nil, "", status.Errorf(codes.Internal, ufsds.InternalError)
	}
	if nextCur != nil {
		nextPageToken = nextCur.String()
	}
	return
}

func getDefaultWifiHistoryClient() *HistoryClient {
	return &HistoryClient{}
}
