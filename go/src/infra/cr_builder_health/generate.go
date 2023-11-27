// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"net/url"

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
	ContactTeamEmail string     `bigquery:"contact_team_email"`
	N                int        `bigquery:"n"`
	Metrics          []*Metric  `bigquery:"metrics"`
}

type Metric struct {
	Type        string  `bigquery:"type"`
	Value       float32 `bigquery:"value"`
	Threshold   float32 `bigquery:"threshold"`
	HealthScore int     `bigquery:"health_score"`
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

	srcConfig, err := getSrcConfig(ctx)
	if err != nil {
		return errors.Annotate(err, "Get Src Config").Err()
	}

	rows, err := getMetrics(ctx, bqClient, input)
	if err != nil {
		return errors.Annotate(err, "Get metrics").Err()
	}

	rowsWithIndicators, err := calculateIndicators(ctx, rows, *srcConfig)
	if err != nil {
		return errors.Annotate(err, "Calculate indicators").Err()
	}

	err = logIndicators(ctx, rowsWithIndicators)
	if err != nil {
		return errors.Annotate(err, "Log indicators").Err()
	}

	if input.DryRun {
		return nil
	}

	client, err := bbClient(ctx)
	if err != nil {
		return errors.Annotate(err, "Make BB client").Err()
	}

	err = rpcBuildbucket(ctx, rowsWithIndicators, client)
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

	step.SetSummaryMarkdown(fmt.Sprintf("Date is %s", input.Date.AsTime().String()))

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
	  [
		STRUCT('fail_rate' as type, COUNTIF(b.status = 'FAILURE') / COUNT(*) as value),
		STRUCT('infra_fail_rate' as type, COUNTIF(b.status = 'INFRA_FAILURE') / COUNT(*) as value),
		STRUCT('pending_mins_p50' as type, IFNULL(APPROX_QUANTILES(TIMESTAMP_DIFF(start_time, create_time, SECOND), 100)[OFFSET(50)]/60, 0) as value),
		STRUCT('pending_mins_p95' as type, IFNULL(APPROX_QUANTILES(TIMESTAMP_DIFF(start_time, create_time, SECOND), 100)[OFFSET(95)]/60, 0) as value),
		STRUCT('build_mins_p50' as type, IFNULL(APPROX_QUANTILES(TIMESTAMP_DIFF(end_time, start_time, SECOND), 100)[OFFSET(50)]/60, 0) as value),
		STRUCT('build_mins_p95' as type, IFNULL(APPROX_QUANTILES(TIMESTAMP_DIFF(end_time, start_time, SECOND), 100)[OFFSET(95)]/60, 0) as value)
	  ] as metrics
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

func calculateIndicators(buildCtx context.Context, rows []Row, srcConfig SrcConfig) ([]Row, error) {
	var stepErr error
	step, ctx := build.StartStep(buildCtx, "Calculate indicators")
	defer func() { step.End(stepErr) }()

	failedBuilders := 0
	for i, row := range rows {
		if bucketSpec, ok := srcConfig.BucketSpecs[row.Bucket]; !ok {
			rows[i].HealthScore = 0
			continue
		} else if builderSpec, ok := bucketSpec[row.Builder]; !ok {
			rows[i].HealthScore = UNSET_SCORE
			continue
		} else {
			if len(builderSpec.ProblemSpecs) == 0 {
				rows[i].HealthScore = UNSET_SCORE
				rows[i].ScoreExplanation = fmt.Sprintf("Src Config error: Bucket: %s, Builder: %s has no ProblemSpecs", row.Bucket, row.Builder)
				logging.Errorf(ctx, "Src Config error: Bucket: %s, Builder: %s has no ProblemSpecs", row.Bucket, row.Builder)
				continue
			}

			sortProblemSpecs(builderSpec.ProblemSpecs) // to give lower scores precedence
			for problemName, problemSpec := range builderSpec.ProblemSpecs {
				if defaultSpec, ok := srcConfig.DefaultSpecs[problemSpec.Thresholds.Default]; ok {
					if (problemSpec.Thresholds.BuildTime != PercentileThresholds{} ||
						problemSpec.Thresholds.TestPendingTime != PercentileThresholds{} ||
						problemSpec.Thresholds.PendingTime != PercentileThresholds{} ||
						problemSpec.Thresholds.FailRate != AverageThresholds{} ||
						problemSpec.Thresholds.InfraFailRate != AverageThresholds{}) {
						rows[i].HealthScore = UNSET_SCORE
						rows[i].ScoreExplanation = "Threshold config error: default sentinel and custom thresholds cannot both be set."
						logging.Errorf(ctx, "%s Bucket: %s. Builder: %s. ProblemName: %s", rows[i].ScoreExplanation, row.Bucket, row.Builder, problemName)
						failedBuilders += 1
						continue
					}
					problemSpec.Thresholds = defaultSpec.Thresholds
				} else if problemSpec.Thresholds.Default != "" {
					rows[i].HealthScore = UNSET_SCORE
					rows[i].ScoreExplanation = fmt.Sprintf("Threshold config error: Default set to unknown sentinel value: %s.", problemSpec.Thresholds.Default)
					logging.Errorf(ctx, "%s Bucket: %s. Builder %s.", rows[i].ScoreExplanation, row.Bucket, row.Builder)
					failedBuilders += 1
					continue
				}
				stepErr = errors.Join(stepErr, compareThresholds(ctx, &rows[i], &problemSpec))
			}
			rows[i].ContactTeamEmail = builderSpec.ContactTeamEmail
		}
	}

	if failedBuilders > 0 {
		stepErr = errors.Join(stepErr, fmt.Errorf("Indicator calculation failed for %d builders", failedBuilders))
	}

	return rows, stepErr
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
		simplifiedMetrics := map[string]float32{}
		for _, metric := range row.Metrics {
			simplifiedMetrics[metric.Type] = metric.Value
		}

		dashboardLink := fmt.Sprintf("http://go/builder-health-indicators?f=builder:in:%s&f=bucket:in:%s&f=project:in:%s",
			url.QueryEscape(row.Builder), url.QueryEscape(row.Bucket), url.QueryEscape(row.Project))
		const designDocLink = "http://go/builder-health-metrics-design"
		healthProtos[i] = &buildbucketpb.SetBuilderHealthRequest_BuilderHealth{
			Id: &buildbucketpb.BuilderID{Project: row.Project, Bucket: row.Bucket, Builder: row.Builder},
			Health: &buildbucketpb.HealthStatus{
				HealthScore:   int64(row.HealthScore),
				HealthMetrics: simplifiedMetrics,
				Description:   row.ScoreExplanation,
				DocLinks: map[string]string{
					"":             "https://chromium.googlesource.com/chromium/src/+/refs/heads/main/infra/config/generated/health-specs/health-specs.json",
					"google.com":   designDocLink,
					"chromium.org": designDocLink,
				},
				DataLinks: map[string]string{
					"":             "https://chromium.googlesource.com/chromium/src/+/refs/heads/main/infra/config/generated/health-specs/health-specs.json",
					"google.com":   dashboardLink,
					"chromium.org": dashboardLink,
				},
			},
		}
	}
	req := &buildbucketpb.SetBuilderHealthRequest{
		Health: healthProtos,
	}
	res, err := client.SetBuilderHealth(ctx, req)
	if err != nil {
		logging.Errorf(ctx, "Set builder health error result: %+v. Error: %s", res, err)
		return errors.Annotate(err, "Set builder health").Err()
	}

	nErrors := 0
	for _, resp := range res.Responses {
		if resp.GetError() == nil {
			continue
		}

		nErrors += 1
		logging.Errorf(ctx, "Set builder health error: %s.", resp.GetError().String())
	}

	if nErrors > 0 {
		step.SetSummaryMarkdown(fmt.Sprintf("%d set builder health requests failed", nErrors))
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

func logIndicators(buildCtx context.Context, rowsWithIndicators []Row) error {
	var stepErr error
	step, ctx := build.StartStep(buildCtx, "Print indicators")
	defer func() { step.End(stepErr) }()

	for _, row := range rowsWithIndicators {
		logging.Errorf(ctx, "%s/%s/%s: HealthScore: %d.", row.Project, row.Bucket, row.Builder, row.HealthScore)
		for _, metric := range row.Metrics {
			logging.Errorf(ctx, "%+v", *metric)
		}
		logging.Errorf(ctx, "")
	}

	return nil
}
