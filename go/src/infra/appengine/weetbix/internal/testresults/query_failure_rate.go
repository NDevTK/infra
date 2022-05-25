// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testresults

import (
	"context"
	"text/template"
	"time"

	"cloud.google.com/go/spanner"
	"go.chromium.org/luci/server/span"
	"google.golang.org/protobuf/types/known/timestamppb"

	spanutil "infra/appengine/weetbix/internal/span"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

// QueryFailureRateOptions specifies options for QueryFailureRate().
type QueryFailureRateOptions struct {
	Project            string
	TestVariants       []*pb.TestVariantIdentifier
	AfterPartitionTime time.Time
}

// ReadVariants reads all the variants of the specified test from the
// spanner database.
// Must be called in a spanner transactional context.
func QueryFailureRate(ctx context.Context, opts QueryFailureRateOptions) ([]*pb.TestVariantFailureRateAnalysis, error) {
	type testVariant struct {
		TestID      string
		VariantHash string
	}
	tvs := make([]testVariant, 0, len(opts.TestVariants))
	for _, ptv := range opts.TestVariants {
		tvs = append(tvs, testVariant{
			TestID:      ptv.TestId,
			VariantHash: pbutil.VariantHash(ptv.Variant),
		})
	}

	stmt, err := spanutil.GenerateStatement(failureRateQueryTmpl, failureRateQueryTmpl.Name(), nil)
	if err != nil {
		return nil, err
	}
	stmt.Params = map[string]interface{}{
		"project":            opts.Project,
		"testVariants":       tvs,
		"afterPartitionTime": opts.AfterPartitionTime,
	}

	results := make([]*pb.TestVariantFailureRateAnalysis, 0, len(tvs))

	index := 0
	var b spanutil.Buffer
	err = span.Query(ctx, stmt).Do(func(row *spanner.Row) error {
		var testID, variantHash string
		var originalVerdictCount, finalVerdictCount int64
		var recentFailures, recentTotal int64
		var recentFailExamples []*verdictExample
		var failPasses, passes int64
		var failPassExamples []*verdictExample

		err := b.FromSpanner(
			row,
			&testID,
			&variantHash,
			&originalVerdictCount,
			&finalVerdictCount,
			&recentFailures,
			&recentTotal,
			&recentFailExamples,
			&failPasses,
			&passes,
			&failPassExamples,
		)
		if err != nil {
			return err
		}

		analysis := &pb.TestVariantFailureRateAnalysis{}
		if testID != tvs[index].TestID || variantHash != tvs[index].VariantHash {
			// This should never happen, as the SQL statement is designed
			// to return results in the same order as test variants requested.
			panic("results in incorrect order")
		}

		analysis.TestId = testID
		analysis.Variant = opts.TestVariants[index].Variant

		analysis.Sample = &pb.TestVariantFailureRateAnalysis_Sample{
			Verdicts:                 finalVerdictCount,
			VerdictsPreDeduplication: originalVerdictCount,
		}
		analysis.FailingRunRatio = &pb.FailingRunRatio{
			Numerator:   recentFailures,
			Denominator: recentTotal,
		}

		analysis.FailingRunExamples = toPBVerdictExamples(recentFailExamples)
		analysis.FlakyVerdictRatio = &pb.FlakyVerdictRatio{
			Numerator:   failPasses,
			Denominator: failPasses + passes,
		}
		analysis.FlakyVerdictExamples = toPBVerdictExamples(failPassExamples)
		results = append(results, analysis)
		index++
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// verdictExample is used to store an example verdict returned by
// a Spanner query.
type verdictExample struct {
	PartitionTime        time.Time
	IngestedInvocationId string
	ChangelistHosts      []string
	ChangelistChanges    []int64
	ChangelistPatchsets  []int64
}

func toPBVerdictExamples(ves []*verdictExample) []*pb.VerdictExample {
	results := make([]*pb.VerdictExample, 0, len(ves))
	for _, ve := range ves {
		cls := make([]*pb.Changelist, 0, len(ve.ChangelistHosts))
		if len(ve.ChangelistHosts) != len(ve.ChangelistChanges) ||
			len(ve.ChangelistChanges) != len(ve.ChangelistPatchsets) {
			panic("data consistency issue: length of changelist arrays do not match")
		}
		for i := range ve.ChangelistHosts {
			cls = append(cls, &pb.Changelist{
				Host:     ve.ChangelistHosts[i],
				Change:   ve.ChangelistChanges[i],
				Patchset: int32(ve.ChangelistPatchsets[i]),
			})
		}
		results = append(results, &pb.VerdictExample{
			PartitionTime:        timestamppb.New(ve.PartitionTime),
			IngestedInvocationId: ve.IngestedInvocationId,
			Changelists:          cls,
		})
	}
	return results
}

var failureRateQueryTmpl = template.Must(template.New("").Parse(`
WITH test_variant_verdicts AS (
	SELECT
		Index,
		TestId,
		VariantHash,
		ARRAY(
			-- Select unique test verdicts (at most once per changelist).
			SELECT AS STRUCT
				ANY_VALUE(STRUCT(
				PartitionTime,
				IngestedInvocationId,
				FirstRunFailed,
				HasSecondRun,
				SecondRunFailed,
				ChangelistHosts,
				ChangelistChanges,
				ChangelistPatchsets)
				-- Keep the verdict with the highest partition time (or if there are multiple
				-- verdicts with equal partition times for the same changelist, pick
				-- any verdict).
				HAVING MAX PartitionTime) AS Verdict,
				COUNT(*) AS OriginalVerdictCount,
			FROM (
				-- Flatten test runs to test verdicts and limit.
				SELECT
					PartitionTime,
					IngestedInvocationId,
					LOGICAL_OR(RunFailed AND RunIndex = 0) AS FirstRunFailed,
					LOGICAL_OR(RunIndex = 1) AS HasSecondRun,
					LOGICAL_OR(RunFailed AND RunIndex = 1) AS SecondRunFailed,
					ANY_VALUE(ChangelistHosts) AS ChangelistHosts,
					ANY_VALUE(ChangelistChanges) AS ChangelistChanges,
					ANY_VALUE(ChangelistPatchsets) AS ChangelistPatchsets
				FROM (
					-- Flatten test results to test runs.
					SELECT
						PartitionTime,
						IngestedInvocationId,
						RunIndex,
						LOGICAL_AND(COALESCE(IsUnexpected, FALSE)) AS RunFailed,
						ANY_VALUE(ChangelistHosts) AS ChangelistHosts,
						ANY_VALUE(ChangelistChanges) AS ChangelistChanges,
						ANY_VALUE(ChangelistPatchsets) AS ChangelistPatchsets
					FROM TestResults
					WHERE Project = @project
						AND PartitionTime > @afterPartitionTime
						AND TestId = tv.TestId And VariantHash = tv.VariantHash
						-- Exclude test results testing multiple CLs, as
						-- we cannot ensure at most one verdict per CL for
						-- them.
						AND ARRAY_LENGTH(ChangelistHosts) <= 1
					GROUP BY PartitionTime, IngestedInvocationId, RunIndex
				)
				GROUP BY PartitionTime, IngestedInvocationId
				ORDER BY PartitionTime DESC, IngestedInvocationId
				-- Apply an early limit to avoid pulling back excessive data
				-- for tests with lots of results. The limit must be applied
				-- before cutting out duplicate verdicts for the same
				-- changelist, as otherwise it will not be useful in improving
				-- performance.
				-- This is because the unique verdict per changelist aggregation
				-- cannot be expressed as a stream aggregation on the underlying
				-- data, so limits on the output do not flow to the input.
				-- If this limit is too small, we will not pull back enough data
				-- to get enough unique verdicts. This can be monitored via the
				-- OriginalVerdictCount and UniqueVerdictCount columns in the
				-- ultimate output.
				LIMIT 2000
			)
			GROUP BY IF(ARRAY_LENGTH(ChangelistHosts) > 0,
						CONCAT(ChangelistHosts[OFFSET(0)], '/', CAST(ChangelistChanges[OFFSET(0)] AS STRING)),
						IngestedInvocationId)
			ORDER BY Verdict.PartitionTime DESC, Verdict.IngestedInvocationId
			LIMIT 1000
			) AS RecentVerdicts,
	FROM UNNEST(@testVariants) tv WITH OFFSET Index
	ORDER BY Index
)

SELECT
	TestId,
	VariantHash,
	COALESCE((SELECT SUM(OriginalVerdictCount) FROM UNNEST(RecentVerdicts)),0) AS OriginalVerdictCount,
	ARRAY_LENGTH(RecentVerdicts) AS FinalVerdictCount,
	(SELECT COUNTIF(Verdict.FirstRunFailed) FROM UNNEST(RecentVerdicts) WITH OFFSET o WHERE o < 10) As RecentFailures,
	LEAST(ARRAY_LENGTH(RecentVerdicts), 10) As RecentTotal,
	ARRAY(
		SELECT AS STRUCT
			Verdict.PartitionTime,
			Verdict.IngestedInvocationId,
			Verdict.ChangelistHosts,
			Verdict.ChangelistChanges,
			Verdict.ChangelistPatchsets
		FROM UNNEST(RecentVerdicts) WITH OFFSET o
		WHERE o < 10 AND Verdict.FirstRunFailed
		ORDER BY PartitionTime DESC, IngestedInvocationId
		LIMIT 10
	) AS RecentFailExamples,
	((SELECT COUNTIF(Verdict.FirstRunFailed AND Verdict.HasSecondRun AND NOT Verdict.SecondRunFailed) FROM UNNEST(RecentVerdicts))) As FailPass,
	((SELECT COUNTIF(NOT Verdict.FirstRunFailed) FROM UNNEST(RecentVerdicts))) AS Passes,
	ARRAY(
		SELECT AS STRUCT
			Verdict.PartitionTime,
			Verdict.IngestedInvocationId,
			Verdict.ChangelistHosts,
			Verdict.ChangelistChanges,
			Verdict.ChangelistPatchsets
		FROM UNNEST(RecentVerdicts)
   	  	WHERE Verdict.FirstRunFailed AND Verdict.HasSecondRun AND NOT Verdict.SecondRunFailed
		ORDER BY PartitionTime DESC, IngestedInvocationId
		LIMIT 10
	) AS FailPassExamples,
FROM test_variant_verdicts
ORDER BY Index
`))
