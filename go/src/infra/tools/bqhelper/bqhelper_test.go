// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package bqhelper

import (
	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
)

type tableKey struct {
	datasetID, tableID string
}

type localTableStore map[tableKey]*bigquery.TableMetadata

func localTableStore() *localTableStore {
	return &localTableStore{
		tables:         map[tableKey]*bigquery.Table{},
		tableMetadatas: map[tableKey]*bigquery.TableMetadata{},
	}

}

func (ts localTableStore) getTableMetadata(ctx context.Context, datasetID, tableID string) (*bigquery.TableMetadata, error) {
	key := tableKey{datasetID: datasetID, tableID: tableID}
	md, ok := bq.tableMetadatas[tableKey]
	if ok {
		return md, nil
	}
	return nil, &googleapi.Error{Code: http.StatusNotFound}
}
