// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package bqhelper is a tool for creating and updating BigQuery tables.
package bqhelper

import (
	"net/http"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
)

type tableDef struct {
	datasetID string
	tableID   string
	toUpdate  bigquery.TableMetadataToUpdate
}

type tableStore interface {
	getTableMetadata(ctx context.Context, datasetID, tableID string) (*bigquery.TableMetadata, error)
}

type bqTableStore struct {
	c *bigquery.Client
}

func errNotFound(e error) bool {
	err, ok := e.(*googleapi.Error)
	return ok && err.Code == http.StatusNotFound
}

func (bq *bqTableStore) getTableMetadata(ctx context.Context, datasetID, tableID string) (*bigquery.TableMetadata, error) {
	t := bq.c.Dataset(datasetID).Table(tableID)
	return t.Metadata(ctx)
}

func updateFromTableDef(ctx context.Context, td tableDef, ts tableStore) {
	md, err := ts.getTableMetadata(ctx, td.datasetID, td.tableID)
}
