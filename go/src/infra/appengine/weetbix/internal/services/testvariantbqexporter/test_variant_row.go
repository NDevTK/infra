// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testvariantbqexporter

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/spanner"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/googleapi"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.chromium.org/luci/common/bq"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/retry"
	"go.chromium.org/luci/common/retry/transient"
	"go.chromium.org/luci/server/span"
	"go.chromium.org/luci/server/tq"

	spanutil "infra/appengine/weetbix/internal/span"
	"infra/appengine/weetbix/pbutil"
	bqpb "infra/appengine/weetbix/proto/bq"
	pb "infra/appengine/weetbix/proto/v1"
)

func testVariantName(realm, testID, variantHash string) string {
	return fmt.Sprintf("realms/%s/tests/%s/variants/%s", realm, url.PathEscape(testID), variantHash)
}

// generateStatement generates a spanner statement from a text template.
func generateStatement(tmpl *template.Template, input interface{}) (spanner.Statement, error) {
	sql := &bytes.Buffer{}
	err := tmpl.Execute(sql, input)
	if err != nil {
		return spanner.Statement{}, err
	}
	return spanner.NewStatement(sql.String()), nil
}

func (b *BQExporter) populateQueryParameters() (inputs, params map[string]interface{}, err error) {
	inputs = map[string]interface{}{
		"WithFilter":   b.BqExport.Predicate != nil,
		"TestIdFilter": b.BqExport.GetPredicate().GetTestIdRegexp() != "",
		"StatusFilter": b.BqExport.GetPredicate().GetStatus() != pb.AnalyzedTestVariantStatus_STATUS_UNSPECIFIED,
	}

	params = map[string]interface{}{
		"realm":              b.BqExport.Realm,
		"flakyVerdictStatus": int(pb.VerdictStatus_VERDICT_FLAKY),
	}

	st, err := pbutil.AsTime(b.BqExport.TimeRange.GetEarliest())
	if err != nil {
		return nil, nil, err
	}
	params["startTime"] = st

	et, err := pbutil.AsTime(b.BqExport.TimeRange.GetLatest())
	if err != nil {
		return nil, nil, err
	}
	params["endTime"] = et

	if re := b.BqExport.GetPredicate().GetTestIdRegexp(); re != "" && re != ".*" {
		params["testIdRegexp"] = fmt.Sprintf("^%s$", re)
	}

	if status := b.BqExport.GetPredicate().GetStatus(); status != pb.AnalyzedTestVariantStatus_STATUS_UNSPECIFIED {
		params["status"] = int(status)
	}

	switch p := b.BqExport.GetPredicate().GetVariant().GetPredicate().(type) {
	case *pb.VariantPredicate_Equals:
		inputs["VariantHashEquals"] = true
		params["variantHashEquals"] = pbutil.VariantHash(p.Equals)
	case *pb.VariantPredicate_Contains:
		if len(p.Contains.Def) > 0 {
			inputs["VariantContains"] = true
			params["variantContains"] = pbutil.VariantToStrings(p.Contains)
		}
	case nil:
		// No filter.
	default:
		return nil, nil, errors.Reason("unexpected variant predicate %q", p).Err()
	}
	return
}

type result struct {
	UnexpectedResultCount spanner.NullInt64
	TotalResultCount      spanner.NullInt64
	FlakyVerdictCount     spanner.NullInt64
	TotalVerdictCount     spanner.NullInt64
	Invocations           []string
}

func (b *BQExporter) timeRanges(previousStatus, currentStatus pb.AnalyzedTestVariantStatus, statusUpdateTime spanner.NullTime) map[pb.AnalyzedTestVariantStatus]*pb.TimeRange {
	// - if previousStatus satisfies the predicate, insert one row with range (earliest, statusUpdateTime)
	//   - verdicts also need to be bounded by narrowed time range
	// - if currentStatus satisfies the predicate, insert one row with range (statusUpdateTime, latest)
	// - if both satisfy, then 2 rows
	if !statusUpdateTime.Valid {
		panic("Empty Status Update time")
	}

	earliestP := b.BqExport.TimeRange.Earliest
	// The timestamps have been verified in populateQueryParameters.
	earliest, _ := pbutil.AsTime(earliestP)
	sut := statusUpdateTime.Time
	sutP := pbutil.MustTimestampProto(sut)

	if sut.Before(earliest) {
		// During the entire b.BqExport.TimeRange the test variant's status has
		// not changed. Return the original time range as it is.
		return map[pb.AnalyzedTestVariantStatus]*pb.TimeRange{currentStatus: b.BqExport.TimeRange}
	}

	trBefore := &pb.TimeRange{
		Earliest: earliestP,
		Latest:   sutP,
	}
	trAfter := &pb.TimeRange{
		Earliest: sutP,
		Latest:   b.BqExport.TimeRange.Latest,
	}

	exportStatus := b.BqExport.GetPredicate().GetStatus()
	switch {
	case previousStatus == pb.AnalyzedTestVariantStatus_STATUS_UNSPECIFIED:
		// The test variant is newly detected, reduce the time range to trAfter.
		return map[pb.AnalyzedTestVariantStatus]*pb.TimeRange{currentStatus: trAfter}
	case exportStatus == pb.AnalyzedTestVariantStatus_STATUS_UNSPECIFIED:
		// No requirement on test variant status.
		return map[pb.AnalyzedTestVariantStatus]*pb.TimeRange{previousStatus: trBefore, currentStatus: trAfter}
	case previousStatus == exportStatus:
		// Previously the test variant satisfied the predicate, need a row for before status change.
		return map[pb.AnalyzedTestVariantStatus]*pb.TimeRange{previousStatus: trBefore}
	case currentStatus == exportStatus:
		// Currently the test variant satisfies the predicate, need a row for after status change.
		return map[pb.AnalyzedTestVariantStatus]*pb.TimeRange{currentStatus: trAfter}
	}
	return nil
}

type verdictInfo struct {
	verdict               *bqpb.Verdict
	ingestionTime         time.Time
	unexpectedResultCount int
	totalResultCount      int
}

// convertVerdicts converts strings to verdictInfos.
// Ordered by IngestionTime.
func (b *BQExporter) convertVerdicts(vs []string) ([]verdictInfo, error) {
	verdicts := make([]verdictInfo, 0, len(vs))
	for _, v := range vs {
		parts := strings.Split(v, "/")
		if len(parts) != 6 {
			return nil, fmt.Errorf("verdict %s in wrong format", v)
		}
		verdict := &bqpb.Verdict{
			Invocation: parts[0],
		}
		s, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		verdict.Status = pb.VerdictStatus(s).String()

		ct, err := time.Parse(time.RFC3339Nano, parts[2])
		if err != nil {
			return nil, err
		}
		verdict.CreateTime = timestamppb.New(ct)

		it, err := time.Parse(time.RFC3339Nano, parts[3])
		if err != nil {
			return nil, err
		}

		unexpectedResultCount, err := strconv.Atoi(parts[4])
		if err != nil {
			return nil, err
		}

		totalResultCount, err := strconv.Atoi(parts[5])
		if err != nil {
			return nil, err
		}

		verdicts = append(verdicts, verdictInfo{
			verdict:               verdict,
			ingestionTime:         it,
			unexpectedResultCount: unexpectedResultCount,
			totalResultCount:      totalResultCount,
		})
	}

	sort.Slice(verdicts, func(i, j int) bool { return verdicts[i].ingestionTime.Before(verdicts[j].ingestionTime) })

	return verdicts, nil
}

func (b *BQExporter) populateVerdictsInRange(tv *bqpb.TestVariantRow, vs []verdictInfo, tr *pb.TimeRange) {
	earliest, _ := pbutil.AsTime(tr.Earliest)
	latest, _ := pbutil.AsTime(tr.Latest)
	var vsInRange []*bqpb.Verdict
	for _, v := range vs {
		if (v.ingestionTime.After(earliest) || v.ingestionTime.Equal(earliest)) && v.ingestionTime.Before(latest) {
			vsInRange = append(vsInRange, v.verdict)
		}
	}
	tv.Verdicts = vsInRange
}

func zeroFlakyStatistics() *pb.FlakeStatistics {
	zero64 := int64(0)
	return &pb.FlakeStatistics{
		FlakyVerdictCount:     zero64,
		TotalVerdictCount:     zero64,
		FlakyVerdictRate:      float32(0),
		UnexpectedResultCount: zero64,
		TotalResultCount:      zero64,
		UnexpectedResultRate:  float32(0),
	}
}

func (b *BQExporter) populateFlakeStatistics(tv *bqpb.TestVariantRow, res *result, vs []verdictInfo, tr *pb.TimeRange) {
	if b.BqExport.TimeRange.Earliest != tr.Earliest || b.BqExport.TimeRange.Latest != tr.Latest {
		b.populateFlakeStatisticsByVerdicts(tv, vs, tr)
		return
	}
	zero64 := int64(0)
	if res.TotalVerdictCount.Valid && res.TotalVerdictCount.Int64 == zero64 {
		tv.FlakeStatistics = zeroFlakyStatistics()
		return
	}
	tv.FlakeStatistics = &pb.FlakeStatistics{
		FlakyVerdictCount:     res.FlakyVerdictCount.Int64,
		TotalVerdictCount:     res.TotalVerdictCount.Int64,
		FlakyVerdictRate:      float32(res.FlakyVerdictCount.Int64) / float32(res.TotalVerdictCount.Int64),
		UnexpectedResultCount: res.UnexpectedResultCount.Int64,
		TotalResultCount:      res.TotalResultCount.Int64,
		UnexpectedResultRate:  float32(res.UnexpectedResultCount.Int64) / float32(res.TotalResultCount.Int64),
	}
}

func (b *BQExporter) populateFlakeStatisticsByVerdicts(tv *bqpb.TestVariantRow, vs []verdictInfo, tr *pb.TimeRange) {
	if len(vs) == 0 {
		tv.FlakeStatistics = zeroFlakyStatistics()
		return
	}

	earliest, _ := pbutil.AsTime(tr.Earliest)
	latest, _ := pbutil.AsTime(tr.Latest)
	flakyVerdicts := 0
	totalVerdicts := 0
	unexpectedResults := 0
	totalResults := 0
	for _, v := range vs {
		if (v.ingestionTime.After(earliest) || v.ingestionTime.Equal(earliest)) && v.ingestionTime.Before(latest) {
			totalVerdicts++
			unexpectedResults += v.unexpectedResultCount
			totalResults += v.totalResultCount
			if v.verdict.Status == pb.VerdictStatus_VERDICT_FLAKY.String() {
				flakyVerdicts++
			}
		}
	}

	if totalVerdicts == 0 {
		tv.FlakeStatistics = zeroFlakyStatistics()
		return
	}

	tv.FlakeStatistics = &pb.FlakeStatistics{
		FlakyVerdictCount:     int64(flakyVerdicts),
		TotalVerdictCount:     int64(totalVerdicts),
		FlakyVerdictRate:      float32(flakyVerdicts) / float32(totalVerdicts),
		UnexpectedResultCount: int64(unexpectedResults),
		TotalResultCount:      int64(totalResults),
		UnexpectedResultRate:  float32(unexpectedResults) / float32(totalResults),
	}
}

func deepCopy(tv *bqpb.TestVariantRow) *bqpb.TestVariantRow {
	return &bqpb.TestVariantRow{
		Name:         tv.Name,
		Realm:        tv.Realm,
		TestId:       tv.TestId,
		VariantHash:  tv.VariantHash,
		Variant:      tv.Variant,
		TestMetadata: tv.TestMetadata,
		Tags:         tv.Tags,
	}
}

// generateTestVariantRows converts a bq.Row to *bqpb.TestVariantRows.
//
// For the most cases it should return one row. But if the test variant
// changes status during the default time range, it may need to export 2 rows
// for the previous and current statuses with smaller time ranges.
func (b *BQExporter) generateTestVariantRows(row *spanner.Row, bf spanutil.Buffer) ([]*bqpb.TestVariantRow, error) {
	tv := &bqpb.TestVariantRow{}
	va := &pb.Variant{}
	var vs []*result
	var statusUpdateTime spanner.NullTime
	var tmd spanutil.Compressed
	var status, previousStatus pb.AnalyzedTestVariantStatus
	var psInt spanner.NullInt64
	if err := bf.FromSpanner(
		row,
		&tv.Realm,
		&tv.TestId,
		&tv.VariantHash,
		&va,
		&tv.Tags,
		&tmd,
		&status,
		&psInt,
		&statusUpdateTime,
		&vs,
	); err != nil {
		return nil, err
	}

	if psInt.Valid {
		previousStatus = pb.AnalyzedTestVariantStatus(psInt.Int64)
	} else {
		previousStatus = pb.AnalyzedTestVariantStatus_STATUS_UNSPECIFIED
	}

	tv.Name = testVariantName(tv.Realm, tv.TestId, tv.VariantHash)
	if len(vs) != 1 {
		return nil, fmt.Errorf("fail to get verdicts for test variant %s", tv.Name)
	}

	tv.Variant = pbutil.VariantToStringPairs(va)
	tv.Status = status.String()

	if len(tmd) > 0 {
		tv.TestMetadata = &pb.TestMetadata{}
		if err := proto.Unmarshal(tmd, tv.TestMetadata); err != nil {
			return nil, errors.Annotate(err, "error unmarshalling test_metadata for %s", tv.Name).Err()
		}
	}

	timeRanges := b.timeRanges(previousStatus, status, statusUpdateTime)
	verdicts, err := b.convertVerdicts(vs[0].Invocations)
	if err != nil {
		return nil, err
	}

	var tvs []*bqpb.TestVariantRow
	for s, r := range timeRanges {
		newTV := deepCopy(tv)
		newTV.TimeRange = r
		newTV.PartitionTime = r.Latest
		newTV.Status = s.String()
		b.populateFlakeStatistics(newTV, vs[0], verdicts, r)
		b.populateVerdictsInRange(newTV, verdicts, r)
		tvs = append(tvs, newTV)
	}

	return tvs, nil
}

func (b *BQExporter) query(ctx context.Context, f func(*bqpb.TestVariantRow) error) error {
	inputs, params, err := b.populateQueryParameters()
	if err != nil {
		return err
	}
	st, err := generateStatement(testVariantRowsTmpl, inputs)
	if err != nil {
		return err
	}
	st.Params = params

	var bf spanutil.Buffer
	return span.Query(ctx, st).Do(
		func(row *spanner.Row) error {
			tvrs, err := b.generateTestVariantRows(row, bf)
			if err != nil {
				return err
			}
			for _, tvr := range tvrs {
				if err := f(tvr); err != nil {
					return err
				}
			}
			return nil
		},
	)
}

func (b *BQExporter) queryTestVariantsToExport(ctx context.Context, batchC chan []*bqpb.TestVariantRow) error {
	ctx, cancel := span.ReadOnlyTransaction(ctx)
	defer cancel()

	tvrs := make([]*bqpb.TestVariantRow, 0, maxBatchRowCount)
	batchSize := 0
	rowCount := 0
	err := b.query(ctx, func(tvr *bqpb.TestVariantRow) error {
		tvrs = append(tvrs, tvr)
		batchSize += proto.Size(tvr)
		rowCount++
		if len(tvrs) >= maxBatchRowCount || batchSize >= maxBatchTotalSize {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case batchC <- tvrs:
			}
			tvrs = make([]*bqpb.TestVariantRow, 0, maxBatchRowCount)
			batchSize = 0
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(tvrs) > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case batchC <- tvrs:
		}
	}

	logging.Infof(ctx, "fetched %d rows for exporting %s test variants", rowCount, b.BqExport.Realm)
	return nil
}

// inserter is implemented by bigquery.Inserter.
type inserter interface {
	// Put uploads one or more rows to the BigQuery service.
	Put(ctx context.Context, src []*bq.Row) error
}

func hasReason(apiErr *googleapi.Error, reason string) bool {
	for _, e := range apiErr.Errors {
		if e.Reason == reason {
			return true
		}
	}
	return false
}

func (b *BQExporter) batchExportRows(ctx context.Context, ins inserter, batchC chan []*bqpb.TestVariantRow) error {
	eg, ctx := errgroup.WithContext(ctx)
	defer eg.Wait()

	for rows := range batchC {
		rows := rows
		if err := b.batchSem.Acquire(ctx, 1); err != nil {
			return err
		}

		eg.Go(func() error {
			defer b.batchSem.Release(1)
			err := b.insertRowsWithRetries(ctx, ins, rows)
			if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == http.StatusForbidden && hasReason(apiErr, "accessDenied") {
				err = tq.Fatal.Apply(err)
			}
			return err
		})
	}

	return eg.Wait()
}

// insertRowsWithRetries inserts rows into BigQuery.
// Retries on transient errors.
func (b *BQExporter) insertRowsWithRetries(ctx context.Context, ins inserter, rowProtos []*bqpb.TestVariantRow) error {
	if err := b.putLimiter.Wait(ctx); err != nil {
		return err
	}

	rows := make([]*bq.Row, 0, len(rowProtos))
	for _, ri := range rowProtos {
		row := &bq.Row{
			Message:  ri,
			InsertID: bigquery.NoDedupeID,
		}
		rows = append(rows, row)
	}

	return retry.Retry(ctx, transient.Only(retry.Default), func() error {
		err := ins.Put(ctx, rows)

		switch e := err.(type) {
		case *googleapi.Error:
			if e.Code == http.StatusForbidden && hasReason(e, "quotaExceeded") {
				err = transient.Tag.Apply(err)
			}
		}

		return err
	}, retry.LogCallback(ctx, "bigquery_put"))
}

func (b *BQExporter) exportTestVariantRows(ctx context.Context, ins inserter) error {
	batchC := make(chan []*bqpb.TestVariantRow)
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return b.batchExportRows(ctx, ins, batchC)
	})

	eg.Go(func() error {
		defer close(batchC)
		return b.queryTestVariantsToExport(ctx, batchC)
	})

	return eg.Wait()
}

var testVariantRowsTmpl = template.Must(template.New("testVariantRowsTmpl").Parse(`
	@{USE_ADDITIONAL_PARALLELISM=TRUE}
	WITH test_variants AS (
		SELECT
			Realm,
			TestId,
			VariantHash,
		FROM AnalyzedTestVariants
		WHERE Realm = @realm
		{{/* Filter by TestId */}}
		{{if .TestIdFilter}}
			AND REGEXP_CONTAINS(TestId, @testIdRegexp)
		{{end}}
		{{/* Filter by Variant */}}
		{{if .VariantHashEquals}}
			AND VariantHash = @variantHashEquals
		{{end}}
		{{if .VariantContains }}
			AND (SELECT LOGICAL_AND(kv IN UNNEST(Variant)) FROM UNNEST(@variantContains) kv)
		{{end}}
		AND StatusUpdateTime < @endTime
		{{/* Filter by status */}}
		{{if .StatusFilter}}
			AND (
				Status = @status
				OR (StatusUpdateTime > @startTime AND PreviousStatus = @status)
			)
    {{end}}
	)

	SELECT
		Realm,
		TestId,
		VariantHash,
		Variant,
		Tags,
		TestMetadata,
		Status,
		PreviousStatus,
		StatusUpdateTime,
		ARRAY(
		SELECT
			AS STRUCT SUM(UnexpectedResultCount) UnexpectedResultCount,
			SUM(TotalResultCount) TotalResultCount,
			COUNTIF(Status=30) FlakyVerdictCount,
			COUNT(*) TotalVerdictCount,
			-- Using struct here will trigger the "null-valued array of struct" query shape
			-- which is not supported by Spanner.
			-- Use a string to work around it.
			ARRAY_AGG(FORMAT('%s/%d/%s/%s/%d/%d', InvocationId, Status, FORMAT_TIMESTAMP("%FT%H:%M:%E*S%Ez", InvocationCreationTime), FORMAT_TIMESTAMP("%FT%H:%M:%E*S%Ez", IngestionTime), UnexpectedResultCount, TotalResultCount)) Invocations
		FROM
			Verdicts
		WHERE
			Verdicts.Realm = test_variants.Realm
			AND Verdicts.TestId=test_variants.TestId
			AND Verdicts.VariantHash=test_variants.VariantHash
			AND IngestionTime >= @startTime
			AND IngestionTime < @endTime ) Results
	FROM
		test_variants
	JOIN
		AnalyzedTestVariants
	USING
		(Realm,
			TestId,
			VariantHash)
`))
