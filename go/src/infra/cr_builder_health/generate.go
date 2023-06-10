// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"

	"infra/cr_builder_health/healthpb"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"

	"go.chromium.org/luci/auth"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/luciexe/build"

	"google.golang.org/api/iterator"
	"google.golang.org/grpc"
)

type Row struct {
	HealthScore      int        `bigquery:"health_score"`
	ScoreExplanation string     `bigquery:"score_explanation"`
	Date             civil.Date `bigquery:"date"`
	Project          string     `bigquery:"project"`
	Bucket           string     `bigquery:"bucket"`
	Builder          string     `bigquery:"builder"`
	Rotation         string     `bigquery:"rotation"`
	N                int        `bigquery:"n"`
	FailRate         float32    `bigquery:"fail_rate"`
	InfraFailRate    float32    `bigquery:"infra_fail_rate"`
	PendingMinsP50   float32    `bigquery:"pending_mins_p50"`
	PendingMinsP95   float32    `bigquery:"pending_mins_p95"`
	BuildMinsP50     float32    `bigquery:"build_mins_p50"`
	BuildMinsP95     float32    `bigquery:"build_mins_p95"`
}

type BBClient interface {
	SetBuilderHealth(ctx context.Context, in *buildbucketpb.SetBuilderHealthRequest, opts ...grpc.CallOption) (*buildbucketpb.SetBuilderHealthResponse, error)
}

func generate(ctx context.Context, input *healthpb.InputParams) error {
	bqClient, err := setup(ctx, input)
	if err != nil {
		return errors.Annotate(err, "Setup").Err()
	}
	defer bqClient.Close()

	thresholds, err := getThresholds(ctx)
	if err != nil {
		return errors.Annotate(err, "Get Thresholds").Err()
	}

	rows, err := getMetrics(ctx, bqClient, input)
	if err != nil {
		return errors.Annotate(err, "Get metrics").Err()
	}

	rowsWithIndicators, err := calculateIndicators(ctx, rows, *thresholds)
	if err != nil {
		return errors.Annotate(err, "Calculate indicators").Err()
	}

	client, err := bbClient(ctx)
	if err != nil {
		return errors.Annotate(err, "Make BB client").Err()
	}

	err = rpcBuildbucket(ctx, rows, client)
	if err != nil {
		return errors.Annotate(err, "RPC buildbucket").Err()
	}

	// Write out to BQ
	if err = writeIndicators(ctx, bqClient, rowsWithIndicators); err != nil {
		return errors.Annotate(err, "Write indicators").Err()
	}

	return nil
}

func setup(buildCtx context.Context, input *healthpb.InputParams) (*bigquery.Client, error) {
	var err error
	step, _ := build.StartStep(buildCtx, "Setup")
	defer func() { step.End(err) }()

	step.SetSummaryMarkdown(fmt.Sprintf("Date is %+v", input.Date))

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

	step.SetSummaryMarkdown(fmt.Sprintf("Queried %d builders", len(rows)))

	return rows, nil
}

func calculateIndicators(buildCtx context.Context, rows []Row, thresholds Thresholds) ([]Row, error) {
	var err error
	step, ctx := build.StartStep(buildCtx, "Calculate indicators")
	defer func() { step.End(err) }()

	failedBuilders := 0
	for i, row := range rows {
		if bucketThresholds, ok := thresholds.Thresholds[row.Bucket]; !ok {
			rows[i].HealthScore = 0
			continue
		} else if builderThresholds, ok := bucketThresholds[row.Builder]; !ok {
			rows[i].HealthScore = 0
			continue
		} else {
			if builderThresholds.Default == "_default" {
				if (builderThresholds.BuildTime != PercentileThresholds{} ||
					builderThresholds.TestPendingTime != PercentileThresholds{} ||
					builderThresholds.PendingTime != PercentileThresholds{} ||
					builderThresholds.FailRate != AverageThresholds{} ||
					builderThresholds.InfraFailRate != AverageThresholds{}) {
					rows[i].HealthScore = 0
					rows[i].ScoreExplanation = "Threshold config error: default sentinel and custom thresholds cannot both be set."
					logging.Errorf(ctx, "%s Bucket: %s. Builder: %s.", rows[i].ScoreExplanation, row.Bucket, row.Builder)
					failedBuilders += 1
					continue
				}
				builderThresholds = thresholds.Default
			} else if builderThresholds.Default != "" {
				rows[i].HealthScore = 0
				rows[i].ScoreExplanation = fmt.Sprintf("Threshold config error: Default set to unknown sentinel value: %s.", builderThresholds.Default)
				logging.Errorf(ctx, "%s Bucket: %s. Builder %s.", rows[i].ScoreExplanation, row.Bucket, row.Builder)
				failedBuilders += 1
				continue
			}
			if healthy, explanation := belowThresholds(row, builderThresholds); healthy {
				rows[i].HealthScore = 10
				rows[i].ScoreExplanation = explanation
			} else {
				rows[i].HealthScore = 1
				rows[i].ScoreExplanation = explanation
			}
		}
	}

	if failedBuilders > 0 {
		err = fmt.Errorf("Indicator calculation failed for %d builders", failedBuilders)
	}

	return rows, err
}

func bbClient(buildCtx context.Context) (buildbucketpb.BuildersClient, error) {
	var err error
	step, _ := build.StartStep(buildCtx, "Make BB client")
	defer func() { step.End(err) }()

	authenticator := auth.NewAuthenticator(buildCtx, auth.SilentLogin, auth.Options{})
	httpClient, err := authenticator.Client()
	if err != nil {
		return nil, errors.Annotate(err, "Initializing Auth").Err()
	}

	return buildbucketpb.NewBuildersPRPCClient(&prpc.Client{
		C:    httpClient,
		Host: "cr-buildbucket.appspot.com",
	}), nil
}

func rpcBuildbucket(buildCtx context.Context, rows []Row, client BBClient) error {
	var err error
	step, ctx := build.StartStep(buildCtx, "RPC Buildbucket")
	defer func() { step.End(err) }()

	healthProtos := make([]*buildbucketpb.SetBuilderHealthRequest_BuilderHealth, len(rows), len(rows))
	for i, row := range rows {
		constituentMetrics := map[string]float32{
			"FailRate":       row.FailRate,
			"InfraFailRate":  row.InfraFailRate,
			"PendingMinsP50": row.PendingMinsP50,
			"PendingMinsP95": row.PendingMinsP95,
			"BuildMinsP50":   row.BuildMinsP50,
			"BuildMinsP95":   row.BuildMinsP95,
		}

		healthProtos[i] = &buildbucketpb.SetBuilderHealthRequest_BuilderHealth{
			Id: &buildbucketpb.BuilderID{Project: row.Project, Bucket: row.Bucket, Builder: row.Builder},
			Health: &buildbucketpb.HealthStatus{
				HealthScore:   int64(row.HealthScore),
				HealthMetrics: constituentMetrics,
				Description:   row.ScoreExplanation,
				DocLinks:      map[string]string{"google.com": "go/builder-health-metrics-design"},
				DataLinks:     nil, // TODO add dashboard link for historical data
			},
		}
	}
	req := &buildbucketpb.SetBuilderHealthRequest{
		Health: healthProtos,
	}
	res, err := client.SetBuilderHealth(ctx, req)
	if err != nil {
		logging.Errorf(ctx, "Result: %+v. Error: %s", res, err)
		return errors.Annotate(err, "Set builder health").Err()
	}

	nErrors := 0
	for _, resp := range res.Responses {
		if resp.GetError() == nil {
			continue
		}

		nErrors += 1

		result := ""
		if resp.GetResult() != nil {
			result = resp.GetResult().String()
		}
		logging.Errorf(ctx, "Result: %s. Error: %s", result, resp.GetError().String())
	}

	if nErrors > 0 {
		return fmt.Errorf("%d set builder health requests failed", nErrors)
	}

	return nil
}

func writeIndicators(buildCtx context.Context, bqClient *bigquery.Client, rows []Row) error {
	var err error
	step, ctx := build.StartStep(buildCtx, "Write indicators")
	defer func() { step.End(err) }()

	step.SetSummaryMarkdown("Writing to BQ table cr-builder-health-indicators.builder_health_indicators.builder-health-indicators")

	inserter := bqClient.Dataset("builder_health_indicators").Table("builder-health-indicators").Inserter()
	if err := inserter.Put(ctx, rows); err != nil {
		return err
	}

	return nil
}
