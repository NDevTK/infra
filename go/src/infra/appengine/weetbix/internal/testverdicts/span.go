// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testverdicts

import (
	"context"
	"text/template"
	"time"

	"cloud.google.com/go/spanner"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/server/span"
	"google.golang.org/protobuf/types/known/durationpb"

	"infra/appengine/weetbix/internal/pagination"
	spanutil "infra/appengine/weetbix/internal/span"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

const (
	testVerdictTTL      = 90 * 24 * time.Hour
	pageTokenTimeFormat = time.RFC3339Nano
)

// TestVerdict represents a row in the TestVerdicts table.
type TestVerdict struct {
	Project                      string
	TestID                       string
	PartitionTime                time.Time
	VariantHash                  string
	IngestedInvocationID         string
	SubRealm                     string
	ExpectedCount                int64
	UnexpectedCount              int64
	SkippedCount                 int64
	IsExonerated                 bool
	PassedAvgDuration            *time.Duration
	HasUnsubmittedChanges        bool
	HasContributedToClSubmission bool
}

// ReadTestVerdicts read test verdicts from the TestVerdicts table.
// Must be called in a spanner transactional context.
func ReadTestVerdicts(ctx context.Context, keys spanner.KeySet, fn func(tv *TestVerdict) error) error {
	var b spanutil.Buffer
	fields := []string{
		"Project", "TestId", "PartitionTime", "VariantHash", "IngestedInvocationId", "SubRealm", "ExpectedCount",
		"UnexpectedCount", "SkippedCount", "IsExonerated", "PassedAvgDurationUsec", "HasUnsubmittedChanges",
		"HasContributedToClSubmission",
	}
	return span.Read(ctx, "TestVerdicts", keys, fields).Do(
		func(row *spanner.Row) error {
			tv := &TestVerdict{}
			var passedAvgDurationUsec spanner.NullInt64
			err := b.FromSpanner(
				row,
				&tv.Project,
				&tv.TestID,
				&tv.PartitionTime,
				&tv.VariantHash,
				&tv.IngestedInvocationID,
				&tv.SubRealm,
				&tv.ExpectedCount,
				&tv.UnexpectedCount,
				&tv.SkippedCount,
				&tv.IsExonerated,
				&passedAvgDurationUsec,
				&tv.HasUnsubmittedChanges,
				&tv.HasContributedToClSubmission,
			)
			if err != nil {
				return err
			}
			if passedAvgDurationUsec.Valid {
				passedAvgDuration := time.Microsecond * time.Duration(passedAvgDurationUsec.Int64)
				tv.PassedAvgDuration = &passedAvgDuration
			}
			return fn(tv)
		})
}

// SaveUnverified saves the test verdict into the TestVerdicts table without
// verifying it.
// Must be called in spanner RW transactional context.
func (tvr *TestVerdict) SaveUnverified(ctx context.Context) {
	var passedAvgDuration spanner.NullInt64
	if tvr.PassedAvgDuration != nil {
		passedAvgDuration.Int64 = tvr.PassedAvgDuration.Microseconds()
		passedAvgDuration.Valid = true
	}

	row := map[string]interface{}{
		"Project":                      tvr.Project,
		"TestId":                       tvr.TestID,
		"PartitionTime":                tvr.PartitionTime,
		"VariantHash":                  tvr.VariantHash,
		"IngestedInvocationId":         tvr.IngestedInvocationID,
		"SubRealm":                     tvr.SubRealm,
		"ExpectedCount":                tvr.ExpectedCount,
		"UnexpectedCount":              tvr.UnexpectedCount,
		"SkippedCount":                 tvr.SkippedCount,
		"IsExonerated":                 tvr.IsExonerated,
		"PassedAvgDurationUsec":        passedAvgDuration,
		"HasUnsubmittedChanges":        tvr.HasUnsubmittedChanges,
		"HasContributedToClSubmission": tvr.HasContributedToClSubmission,
	}
	span.BufferWrite(ctx, spanner.InsertOrUpdateMap("TestVerdicts", spanutil.ToSpannerMap(row)))
}

// ReadTestHistoryOptions specifies options for ReadTestHistory().
type ReadTestHistoryOptions struct {
	Project          string
	TestID           string
	SubRealms        []string
	VariantPredicate *pb.VariantPredicate
	SubmittedFilter  pb.SubmittedFilter
	TimeRange        *pb.TimeRange
	PageSize         int
	PageToken        string
}

// statement generats a spanner statement for the specified query template.
func (opts ReadTestHistoryOptions) statement(ctx context.Context, tmpl string, paginationParams []string) (spanner.Statement, error) {
	now := clock.Now(ctx)

	params := map[string]interface{}{
		"project":    opts.Project,
		"testId":     opts.TestID,
		"subRealms":  opts.SubRealms,
		"afterTime":  now.Add(-testVerdictTTL),
		"beforeTime": now,
		"limit":      opts.PageSize,

		// If the filter is unspecified, this param will be ignored during the
		// statement generation step.
		"hasUnsubmittedChanges": opts.SubmittedFilter == pb.SubmittedFilter_ONLY_UNSUBMITTED,

		// Status enum variants.
		"unexpected":          int(pb.TestVerdictStatus_UNEXPECTED),
		"unexpectedlySkipped": int(pb.TestVerdictStatus_UNEXPECTEDLY_SKIPPED),
		"flaky":               int(pb.TestVerdictStatus_FLAKY),
		"exonerated":          int(pb.TestVerdictStatus_EXONERATED),
		"expected":            int(pb.TestVerdictStatus_EXPECTED),
	}

	if opts.TimeRange.GetEarliest() != nil {
		params["afterTime"] = opts.TimeRange.GetEarliest().AsTime()
	}
	if opts.TimeRange.GetLatest() != nil {
		params["beforeTime"] = opts.TimeRange.GetLatest().AsTime()
	}

	switch p := opts.VariantPredicate.GetPredicate().(type) {
	case *pb.VariantPredicate_Equals:
		params["variantHash"] = pbutil.VariantHash(p.Equals)
	case *pb.VariantPredicate_Contains:
		if len(p.Contains.Def) > 0 {
			params["variantKVs"] = pbutil.VariantToStrings(p.Contains)
		}
	case nil:
		// No filter.
	default:
		panic(errors.Reason("unexpected variant predicate %q", opts.VariantPredicate).Err())
	}

	if opts.PageToken != "" {
		tokens, err := pagination.ParseToken(opts.PageToken)
		if err != nil {
			return spanner.Statement{}, err
		}

		if len(tokens) != len(paginationParams) {
			return spanner.Statement{}, pagination.InvalidToken(errors.Reason("expected %d components, got %d", len(paginationParams), len(tokens)).Err())
		}

		// Keep all pagination params as strings and convert them to other data
		// types in the query as necessary. So we can have a unified way of handling
		// different page tokens.
		for i, param := range paginationParams {
			params[param] = tokens[i]
		}
	}

	stmt, err := spanutil.GenerateStatement(testHistoryQueryTmpl, tmpl, map[string]interface{}{
		"submittedFilterSpecified": opts.SubmittedFilter != pb.SubmittedFilter_SUBMITTED_FILTER_UNSPECIFIED,
		"pagination":               opts.PageToken != "",
		"params":                   params,
	})
	if err != nil {
		return spanner.Statement{}, err
	}
	stmt.Params = params

	return stmt, nil
}

// ReadTestHistory reads verdicts from the spanner database.
// Must be called in a spanner transactional context.
func ReadTestHistory(ctx context.Context, opts ReadTestHistoryOptions) (verdicts []*pb.TestVerdict, nextPageToken string, err error) {
	stmt, err := opts.statement(ctx, "testHistoryQuery", []string{"paginationTime", "paginationVariantHash", "paginationInvId"})
	if err != nil {
		return nil, "", err
	}

	var b spanutil.Buffer
	verdicts = make([]*pb.TestVerdict, 0, opts.PageSize)
	err = span.Query(ctx, stmt).Do(func(row *spanner.Row) error {
		tv := &pb.TestVerdict{
			TestId: opts.TestID,
		}
		var status int64
		var passedAvgDurationUsec spanner.NullInt64
		err := b.FromSpanner(
			row,
			&tv.InvocationId,
			&tv.VariantHash,
			&status,
			&tv.PartitionTime,
			&passedAvgDurationUsec,
		)
		if err != nil {
			return err
		}
		tv.Status = pb.TestVerdictStatus(status)
		if passedAvgDurationUsec.Valid {
			tv.PassedAvgDuration = durationpb.New(time.Microsecond * time.Duration(passedAvgDurationUsec.Int64))
		}
		verdicts = append(verdicts, tv)
		return nil
	})
	if err != nil {
		return nil, "", errors.Annotate(err, "query test history").Err()
	}

	if opts.PageSize != 0 && len(verdicts) == opts.PageSize {
		lastTV := verdicts[len(verdicts)-1]
		nextPageToken = pagination.Token(lastTV.PartitionTime.AsTime().Format(pageTokenTimeFormat), lastTV.VariantHash, lastTV.InvocationId)
	}
	return verdicts, nextPageToken, nil
}

// ReadTestHistoryStats reads stats of verdicts grouped by UTC dates from the
// spanner database.
// Must be called in a spanner transactional context.
func ReadTestHistoryStats(ctx context.Context, opts ReadTestHistoryOptions) (groups []*pb.QueryTestHistoryStatsResponse_Group, nextPageToken string, err error) {
	stmt, err := opts.statement(ctx, "testHistoryStatsQuery", []string{"paginationDate", "paginationVariantHash"})
	if err != nil {
		return nil, "", err
	}

	var b spanutil.Buffer
	groups = make([]*pb.QueryTestHistoryStatsResponse_Group, 0, opts.PageSize)
	err = span.Query(ctx, stmt).Do(func(row *spanner.Row) error {
		group := &pb.QueryTestHistoryStatsResponse_Group{}
		var (
			unexpectedCount, unexpectedlySkippedCount  int64
			flakyCount, exoneratedCount, expectedCount int64
			passedAvgDurationUsec                      spanner.NullInt64
		)
		err := b.FromSpanner(
			row,
			&group.PartitionTime,
			&group.VariantHash,
			&unexpectedCount, &unexpectedlySkippedCount,
			&flakyCount, &exoneratedCount, &expectedCount,
			&passedAvgDurationUsec,
		)
		if err != nil {
			return err
		}
		group.UnexpectedCount = int32(unexpectedCount)
		group.UnexpectedlySkippedCount = int32(unexpectedlySkippedCount)
		group.FlakyCount = int32(flakyCount)
		group.ExoneratedCount = int32(exoneratedCount)
		group.ExpectedCount = int32(expectedCount)
		if passedAvgDurationUsec.Valid {
			group.PassedAvgDuration = durationpb.New(time.Microsecond * time.Duration(passedAvgDurationUsec.Int64))
		}
		groups = append(groups, group)
		return nil
	})
	if err != nil {
		return nil, "", errors.Annotate(err, "query test history stats").Err()
	}

	if opts.PageSize != 0 && len(groups) == opts.PageSize {
		lastGroup := groups[len(groups)-1]
		nextPageToken = pagination.Token(lastGroup.PartitionTime.AsTime().Format(pageTokenTimeFormat), lastGroup.VariantHash)
	}
	return groups, nextPageToken, nil
}

var testHistoryQueryTmpl = template.Must(template.New("").Parse(`
	{{define "tvStatus"}}
		CASE
			WHEN IsExonerated THEN @exonerated
			WHEN UnexpectedCount = 0 THEN @expected
			WHEN SkippedCount = UnexpectedCount AND ExpectedCount = 0 THEN @unexpectedlySkipped
			WHEN ExpectedCount = 0 THEN @unexpected
			ELSE @flaky
		END TvStatus
	{{end}}

	{{define "testVerdictFilter"}}
		Project = @project
			AND TestId = @testId
			AND PartitionTime >= @afterTime
			AND PartitionTime < @beforeTime
			{{if .params.subRealms}}
				AND SubRealm IN UNNEST(@subRealms)
			{{end}}
			{{if .params.variantHash}}
				AND VariantHash = @variantHash
			{{end}}
			{{if .params.variantKVs}}
				AND VariantHash IN (
					SELECT DISTINCT VariantHash
					FROM TestVariantRealms
					WHERE
						Project = @project
						AND TestId = @testId
						{{if .params.subRealms}}
							AND SubRealm IN UNNEST(@subRealms)
						{{end}}
						AND (SELECT LOGICAL_AND(kv IN UNNEST(Variant)) FROM UNNEST(@variantKVs) kv)
				)
			{{end}}
			{{if .submittedFilterSpecified}}
				AND HasUnsubmittedChanges = @hasUnsubmittedChanges
			{{end}}
	{{end}}

	{{define "testHistoryQuery"}}
		SELECT
			IngestedInvocationId,
			VariantHash,
			{{template "tvStatus" .}},
			PartitionTime,
			PassedAvgDurationUsec
		FROM TestVerdicts
		WHERE
			{{template "testVerdictFilter" .}}
			{{if .pagination}}
				AND	(
					PartitionTime < TIMESTAMP(@paginationTime)
						OR (PartitionTime = TIMESTAMP(@paginationTime) AND VariantHash > @paginationVariantHash)
						OR (PartitionTime = TIMESTAMP(@paginationTime) AND VariantHash = @paginationVariantHash AND IngestedInvocationId > @paginationInvId)
				)
			{{end}}
		ORDER BY
			PartitionTime DESC,
			VariantHash ASC,
			IngestedInvocationId ASC
		{{if .params.limit}}
			LIMIT @limit
		{{end}}
	{{end}}

	{{define "testHistoryStatsQuery"}}
		WITH tv as (
			SELECT
				VariantHash,
				{{template "tvStatus" .}},
				PartitionTime,
				PassedAvgDurationUsec
			FROM TestVerdicts
			WHERE
				{{template "testVerdictFilter" .}}
				{{if .pagination}}
					AND	PartitionTime < TIMESTAMP_ADD(TIMESTAMP(@paginationDate), INTERVAL 1 DAY)
				{{end}}
		)
		SELECT
			TIMESTAMP_TRUNC(PartitionTime, DAY, "UTC") AS PartitionDate,
			VariantHash,
			COUNTIF(TvStatus = @unexpected) AS UnexpectedCount,
			COUNTIF(TvStatus = @unexpectedlySkipped) AS UnexpectedlySkippedCount,
			COUNTIF(TvStatus = @flaky) AS FlakyCount,
			COUNTIF(TvStatus = @exonerated) AS ExoneratedCount,
			COUNTIF(TvStatus = @expected) AS ExpectedCount,
			CAST(AVG(PassedAvgDurationUsec) AS INT64) AS AvgPassedAvgDurationUsec
		FROM tv
		GROUP BY PartitionDate, VariantHash
		{{if .pagination}}
			HAVING
				PartitionDate < TIMESTAMP(@paginationDate)
					OR (PartitionDate = TIMESTAMP(@paginationDate) AND VariantHash > @paginationVariantHash)
		{{end}}
		ORDER BY
			PartitionDate DESC,
			VariantHash ASC
		{{if .params.limit}}
			LIMIT @limit
		{{end}}
	{{end}}
`))
