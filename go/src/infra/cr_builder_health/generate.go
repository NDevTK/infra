// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"

	"infra/cr_builder_health/healthpb"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

	"google.golang.org/api/iterator"
)

type Row struct {
	Date           civil.Date `bigquery:"date"`
	Project        string     `bigquery:"project"`
	Bucket         string     `bigquery:"bucket"`
	Builder        string     `bigquery:"builder"`
	Rotation       string     `bigquery:"rotation"`
	N              int        `bigquery:"n"`
	FailRate       float32    `bigquery:"fail_rate"`
	InfraFailRate  float32    `bigquery:"infra_fail_rate"`
	PendingMinsP50 float32    `bigquery:"pending_mins_p50"`
	PendingMinsP95 float32    `bigquery:"pending_mins_p95"`
	BuildMinsP50   float32    `bigquery:"build_mins_p50"`
	BuildMinsP95   float32    `bigquery:"build_mins_p95"`
}

func generate(ctx context.Context, input *healthpb.InputParams) error {
	bqClient, err := setup(ctx)
	if err != nil {
		return errors.Annotate(err, "Setup").Err()
	}
	defer bqClient.Close()

	rows, err := getMetrics(ctx, bqClient, input)
	if err != nil {
		return errors.Annotate(err, "Get metrics").Err()
	}

	// TODO write to Buildbucket

	// Write out to BQ
	err = writeIndicators(ctx, bqClient, rows)
	if err != nil {
		return errors.Annotate(err, "Write indicators").Err()
	}

	return nil
}

func setup(buildCtx context.Context) (*bigquery.Client, error) {
	var err error
	step, _ := build.StartStep(buildCtx, "Setup")
	defer func() { step.End(err) }()

	bqClient, err := bigquery.NewClient(buildCtx, "cr-builder-health-indicators")
	if err != nil {
		return nil, errors.Annotate(err, "Initializing BigQuery client").Err()
	}

	return bqClient, nil
}

func getMetrics(buildCtx context.Context, bqClient *bigquery.Client, input *healthpb.InputParams) ([]Row, error) {
	var err error
	step, ctx := build.StartStep(buildCtx, "Get metrics")
	defer func() { step.End(err) }()

	q := bqClient.Query(`
	SELECT
	  DATE(@date) as date,
	  b.builder.project,
	  b.builder.bucket,
	  b.builder.builder,
	  JSON_VALUE_ARRAY(ANY_VALUE(b.input.properties), '$.sheriff_rotations')[OFFSET(0)] as rotation,
	  COUNT(*) as n,
	  COUNTIF(b.status = 'FAILURE') / COUNT(*) as fail_rate,
	  COUNTIF(b.status = 'INFRA_FAILURE') / COUNT(*) as infra_fail_rate,
	  IFNULL(APPROX_QUANTILES(TIMESTAMP_DIFF(start_time, create_time, SECOND), 100)[OFFSET(50)]/60, 0) as pending_mins_p50,
	  IFNULL(APPROX_QUANTILES(TIMESTAMP_DIFF(start_time, create_time, SECOND), 100)[OFFSET(95)]/60, 0) as pending_mins_p95,
	  IFNULL(APPROX_QUANTILES(TIMESTAMP_DIFF(end_time, start_time, SECOND), 100)[OFFSET(50)]/60, 0) as build_mins_p50,
	  IFNULL(APPROX_QUANTILES(TIMESTAMP_DIFF(end_time, start_time, SECOND), 100)[OFFSET(95)]/60, 0) as build_mins_p95
    ` +
		"FROM `cr-buildbucket.chrome.builds` as b" + `
	WHERE
		b.create_time < @date
		AND b.create_time >= TIMESTAMP_SUB(@date, INTERVAL 7 DAY)
		AND b.builder.bucket = 'ci'
		AND b.builder.project IN ('chromium', 'chrome')
		AND b.input.properties LIKE '%sheriff_rotations%'
	GROUP BY
	  b.builder.project, b.builder.bucket, b.builder.builder
	HAVING n >= 20
	ORDER BY
	  rotation,
	  LOWER(builder) ASC
	`)
	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "date",
			Value: input.Date.AsTime().UTC().Format(iso8601Format),
		},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "BQ query").Err()
	}

	var rows []Row

	for {
		var row Row
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			logging.Errorf(ctx, "Partially parsed bad row: %+v", row)
			return nil, errors.Annotate(err, "Row iterator").Err()
		}
		logging.Debugf(ctx, "%+v", row)
		rows = append(rows, row)
	}

	return rows, nil
}

func writeIndicators(buildCtx context.Context, bqClient *bigquery.Client, rows []Row) error {
	var err error
	step, ctx := build.StartStep(buildCtx, "Write indicators")
	defer func() { step.End(err) }()

	inserter := bqClient.Dataset("builder_health_indicators").Table("builder-health-indicators").Inserter()
	if err := inserter.Put(ctx, rows); err != nil {
		return err
	}

	return nil
}
