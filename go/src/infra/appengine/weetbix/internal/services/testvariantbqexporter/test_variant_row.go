// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testvariantbqexporter

import (
	"bytes"
	"context"
	"fmt"
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
	"go.chromium.org/luci/server/span"
	"go.chromium.org/luci/server/tq"

	"infra/appengine/weetbix/internal/bqutil"
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
		"TestIdFilter": b.options.GetPredicate().GetTestIdRegexp() != "",
		"StatusFilter": b.options.GetPredicate().GetStatus() != pb.AnalyzedTestVariantStatus_STATUS_UNSPECIFIED,
	}

	params = map[string]interface{}{
		"realm":              b.options.Realm,
		"flakyVerdictStatus": int(pb.VerdictStatus_VERDICT_FLAKY),
	}

	st, err := pbutil.AsTime(b.options.TimeRange.GetEarliest())
	if err != nil {
		return nil, nil, err
	}
	params["startTime"] = st

	et, err := pbutil.AsTime(b.options.TimeRange.GetLatest())
	if err != nil {
		return nil, nil, err
	}
	params["endTime"] = et

	if re := b.options.GetPredicate().GetTestIdRegexp(); re != "" && re != ".*" {
		params["testIdRegexp"] = fmt.Sprintf("^%s$", re)
	}

	if status := b.options.GetPredicate().GetStatus(); status != pb.AnalyzedTestVariantStatus_STATUS_UNSPECIFIED {
		params["status"] = int(status)
	}

	switch p := b.options.GetPredicate().GetVariant().GetPredicate().(type) {
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

type statusAndTimeRange struct {
	status pb.AnalyzedTestVariantStatus
	tr     *pb.TimeRange
}

// timeRanges returns a list of updated time ranges.
//
// It checks if it's needed to narrow/split the original time range
// based on the current status and its update time, also the previous statuses and their update times.
// If b.options.Predicate.Status is specified:
// - if the test variant's status changes to the required status, time range should narrow to [statusUpdateTime, latest).
// - if the test variant's status changes from the required status to something else, time range should narrow to [earliest, statusUpdateTime).
//
// If b.options.Predicate.Status is not specified: time range should split to
// [earliest, statusUpdateTime) and [statusUpdateTime, latest).
func (b *BQExporter) timeRanges(currentStatus pb.AnalyzedTestVariantStatus, statusUpdateTime spanner.NullTime, ps []pb.AnalyzedTestVariantStatus, puts []time.Time) []statusAndTimeRange {
	if !statusUpdateTime.Valid {
		panic("Empty Status Update time")
	}

	earliestP := b.options.TimeRange.Earliest
	latestP := b.options.TimeRange.Latest
	// The timestamps have been verified in populateQueryParameters.
	earliest, _ := pbutil.AsTime(earliestP)
	latest, _ := pbutil.AsTime(latestP)

	exportStatus := b.options.GetPredicate().GetStatus()
	ps = append([]pb.AnalyzedTestVariantStatus{currentStatus}, ps...)
	puts = append([]time.Time{statusUpdateTime.Time}, puts...)

	var strs []statusAndTimeRange
	rightBound := latestP
	for i, t := range puts {
		if t.After(latest) {
			continue
		}

		s := ps[i]
		tp := pbutil.MustTimestampProto(t)

		shouldExport := s == exportStatus || exportStatus == pb.AnalyzedTestVariantStatus_STATUS_UNSPECIFIED
		if shouldExport {
			if t.After(earliest) {
				strs = append(strs, statusAndTimeRange{status: s, tr: &pb.TimeRange{
					Earliest: tp,
					Latest:   rightBound,
				}})
			} else {
				strs = append(strs, statusAndTimeRange{status: s, tr: &pb.TimeRange{
					Earliest: earliestP,
					Latest:   rightBound,
				}})
			}
		}
		rightBound = tp

		if !t.After(earliest) {
			break
		}

	}

	return strs
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
	return &pb.FlakeStatistics{
		FlakyVerdictCount:     0,
		TotalVerdictCount:     0,
		FlakyVerdictRate:      float32(0),
		UnexpectedResultCount: 0,
		TotalResultCount:      0,
		UnexpectedResultRate:  float32(0),
	}
}

func (b *BQExporter) populateFlakeStatistics(tv *bqpb.TestVariantRow, res *result, vs []verdictInfo, tr *pb.TimeRange) {
	if b.options.TimeRange.Earliest != tr.Earliest || b.options.TimeRange.Latest != tr.Latest {
		// The time range is different from the original one, so we cannot use the
		// statistics from query, instead we need to calculate using data from each verdicts.
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
	var status pb.AnalyzedTestVariantStatus
	var ps []pb.AnalyzedTestVariantStatus
	var puts []time.Time
	if err := bf.FromSpanner(
		row,
		&tv.Realm,
		&tv.TestId,
		&tv.VariantHash,
		&va,
		&tv.Tags,
		&tmd,
		&status,
		&statusUpdateTime,
		&ps,
		&puts,
		&vs,
	); err != nil {
		return nil, err
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

	timeRanges := b.timeRanges(status, statusUpdateTime, ps, puts)
	verdicts, err := b.convertVerdicts(vs[0].Invocations)
	if err != nil {
		return nil, err
	}

	var tvs []*bqpb.TestVariantRow
	for _, str := range timeRanges {
		newTV := deepCopy(tv)
		newTV.TimeRange = str.tr
		newTV.PartitionTime = str.tr.Latest
		newTV.Status = str.status.String()
		b.populateFlakeStatistics(newTV, vs[0], verdicts, str.tr)
		b.populateVerdictsInRange(newTV, verdicts, str.tr)
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
	rowCount := 0
	err := b.query(ctx, func(tvr *bqpb.TestVariantRow) error {
		tvrs = append(tvrs, tvr)
		rowCount++
		if len(tvrs) >= maxBatchRowCount {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case batchC <- tvrs:
			}
			tvrs = make([]*bqpb.TestVariantRow, 0, maxBatchRowCount)
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

	logging.Infof(ctx, "fetched %d rows for exporting %s test variants", rowCount, b.options.Realm)
	return nil
}

// inserter is implemented by bigquery.Inserter.
type inserter interface {
	// PutWithRetries uploads one or more rows to the BigQuery service.
	PutWithRetries(ctx context.Context, src []*bq.Row) error
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
			err := b.insertRows(ctx, ins, rows)
			if bqutil.FatalError(err) {
				err = tq.Fatal.Apply(err)
			}
			return err
		})
	}

	return eg.Wait()
}

// insertRows inserts rows into BigQuery.
// Retries on transient errors.
func (b *BQExporter) insertRows(ctx context.Context, ins inserter, rowProtos []*bqpb.TestVariantRow) error {
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

	return ins.PutWithRetries(ctx, rows)
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
		{{/* Filter by status */}}
		{{if .StatusFilter}}
			AND (
				(Status = @status AND StatusUpdateTime < @endTime)
				-- Status updated within the time range, we need to check if the previous
				-- status(es) satisfies the filter.
				OR StatusUpdateTime > @startTime
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
		StatusUpdateTime,
		PreviousStatuses,
		PreviousStatusUpdateTimes,
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
