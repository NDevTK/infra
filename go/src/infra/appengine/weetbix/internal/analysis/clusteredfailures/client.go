// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clusteredfailures

import (
	"context"
	"net/http"
	"time"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/bq"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/caching"
	"google.golang.org/api/option"

	"infra/appengine/weetbix/internal/config"
	bqp "infra/appengine/weetbix/proto/bq"
)

// schemaApplyer ensures BQ schema matches the row proto definitons.
var schemaApplyer = bq.NewSchemaApplyer(caching.RegisterLRUCache(50))

// NewClient creates a new client for exporting clustered failures.
func NewClient() *Client {
	return &Client{}
}

// Client provides methods to export clustered failures to BigQuery.
type Client struct {
}

func bqClient(ctx context.Context, luciProject string, gcpProject string) (*bigquery.Client, error) {
	tr, err := auth.GetRPCTransport(ctx, auth.AsProject, auth.WithProject(luciProject), auth.WithScopes(bigquery.Scope))
	if err != nil {
		return nil, err
	}

	return bigquery.NewClient(ctx, gcpProject, option.WithHTTPClient(&http.Client{
		Transport: tr,
	}))
}

// Insert inserts the given rows in BigQuery.
func (c *Client) Insert(ctx context.Context, luciProject string, tableCfg *config.BigQueryTable, rows []*bqp.ClusteredFailure) error {
	client, err := bqClient(ctx, luciProject, tableCfg.Project)
	if err != nil {
		return err
	}
	defer client.Close()

	tableMetadata := &bigquery.TableMetadata{
		TimePartitioning: &bigquery.TimePartitioning{
			Type:       bigquery.DayPartitioningType,
			Expiration: 540 * 24 * time.Hour,
			Field:      "partition_time",
		},
		Clustering: &bigquery.Clustering{
			Fields: []string{"cluster_algorithm", "cluster_id", "test_result_system", "test_result_id"},
		},
	}

	table := client.Dataset(tableCfg.Dataset).Table(tableCfg.Table)
	if err := schemaApplyer.EnsureTable(ctx, table, tableMetadata); err != nil {
		return errors.Annotate(err, "ensuring clustered failures schema").Err()
	}

	bqRows := make([]*bq.Row, 0, len(rows))
	for _, r := range rows {
		// bq.Row implements ValueSaver for arbitrary protos.
		bqRow := &bq.Row{
			Message:  r,
			InsertID: bigquery.NoDedupeID,
		}
		bqRows = append(bqRows, bqRow)
	}

	inserter := table.Inserter()
	if err := inserter.Put(ctx, bqRows); err != nil {
		errors.Annotate(err, "inserting clustered failures").Err()
	}
	return nil
}
