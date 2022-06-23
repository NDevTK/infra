// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testresults

import (
	"context"
	"sort"
	"text/template"
	"time"

	"cloud.google.com/go/spanner"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/server/span"
	"google.golang.org/protobuf/types/known/durationpb"

	"infra/appengine/weetbix/internal/pagination"
	spanutil "infra/appengine/weetbix/internal/span"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

const pageTokenTimeFormat = time.RFC3339Nano

// The suffix used for all gerrit hostnames.
const GerritHostnameSuffix = "-review.googlesource.com"

var (
	// minTimestamp is the minimum Timestamp value in Spanner.
	// https://cloud.google.com/spanner/docs/reference/standard-sql/data-types#timestamp_type
	minSpannerTimestamp = time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	// maxSpannerTimestamp is the max Timestamp value in Spanner.
	// https://cloud.google.com/spanner/docs/reference/standard-sql/data-types#timestamp_type
	maxSpannerTimestamp = time.Date(9999, time.December, 31, 23, 59, 59, 999999999, time.UTC)
)

// Changelist represents a gerrit changelist.
type Changelist struct {
	// Host is the gerrit hostname, excluding "-review.googlesource.com".
	Host     string
	Change   int64
	Patchset int64
}

// PresubmitRun represents information about the presubmit run a test result
// was part of.
type PresubmitRun struct {
	Owner string
	Mode  pb.PresubmitRunMode
}

// SortChangelists sorts a slice of changelists to be in ascending
// lexicographical order by (host, change, patchset).
func SortChangelists(cls []Changelist) {
	sort.Slice(cls, func(i, j int) bool {
		// Returns true iff cls[i] is less than cls[j].
		if cls[i].Host < cls[j].Host {
			return true
		}
		if cls[i].Host == cls[j].Host && cls[i].Change < cls[j].Change {
			return true
		}
		if cls[i].Host == cls[j].Host && cls[i].Change == cls[j].Change && cls[i].Patchset < cls[j].Patchset {
			return true
		}
		return false
	})
}

// IngestedInvocation represents a row in the IngestedInvocations table.
type IngestedInvocation struct {
	Project string
	// IngestedInvocationID is the ID of the (root) ResultDB invocation
	// being ingested, excluding "invocations/".
	IngestedInvocationID string
	SubRealm             string
	PartitionTime        time.Time
	BuildStatus          pb.BuildStatus
	PresubmitRun         *PresubmitRun

	// The following fields describe the commit tested, excluding any
	// unsubmitted changelists. If information is not available,
	// CommitPosition is zero and the other fields are their default
	// values.
	GitReferenceHash []byte
	CommitPosition   int64
	CommitHash       string

	// The unsubmitted changelists tested (if any). Limited to
	// at most 10 changelists.
	Changelists []Changelist
}

// ReadIngestedInvocations read ingested invocations from the
// IngestedInvocations table.
// Must be called in a spanner transactional context.
func ReadIngestedInvocations(ctx context.Context, keys spanner.KeySet, fn func(inv *IngestedInvocation) error) error {
	var b spanutil.Buffer
	fields := []string{
		"Project", "IngestedInvocationId", "SubRealm", "PartitionTime",
		"BuildStatus",
		"PresubmitRunMode", "PresubmitRunOwner",
		"GitReferenceHash", "CommitPosition", "CommitHash",
		"ChangelistHosts", "ChangelistChanges", "ChangelistPatchsets",
	}
	return span.Read(ctx, "IngestedInvocations", keys, fields).Do(
		func(row *spanner.Row) error {
			inv := &IngestedInvocation{}
			var presubmitRunMode spanner.NullInt64
			var presubmitRunOwner spanner.NullString
			var gitReferenceHash []byte
			var commitPosition spanner.NullInt64
			var commitHash spanner.NullString
			var changelistHosts []string
			var changelistChanges []int64
			var changelistPatchsets []int64

			err := b.FromSpanner(
				row,
				&inv.Project, &inv.IngestedInvocationID, &inv.SubRealm, &inv.PartitionTime,
				&inv.BuildStatus,
				&presubmitRunMode, &presubmitRunOwner,
				&gitReferenceHash, &commitPosition, &commitHash,
				&changelistHosts, &changelistChanges, &changelistPatchsets,
			)
			if err != nil {
				return err
			}

			// Data in Spanner should be consistent, so presubmitRunMode.Valid ==
			//   presubmitRunOwner.Valid.
			if presubmitRunMode.Valid {
				inv.PresubmitRun = &PresubmitRun{
					Mode:  pb.PresubmitRunMode(presubmitRunMode.Int64),
					Owner: presubmitRunOwner.StringVal,
				}
			}

			// Data in Spanner should be consistent, so commitPosition.Valid ==
			// commitHash.Valid == (gitReferenceHash != nil).
			if commitPosition.Valid {
				inv.GitReferenceHash = gitReferenceHash
				inv.CommitPosition = commitPosition.Int64
				inv.CommitHash = commitHash.StringVal
			}

			// Data in Spanner should be consistent, so
			// len(changelistHosts) == len(changelistChanges)
			//    == len(changelistPatchsets).
			if len(changelistHosts) != len(changelistChanges) ||
				len(changelistChanges) != len(changelistPatchsets) {
				panic("Changelist arrays have mismatched length in Spanner")
			}
			changelists := make([]Changelist, 0, len(changelistHosts))
			for i := range changelistHosts {
				changelists = append(changelists, Changelist{
					Host:     changelistHosts[i],
					Change:   changelistChanges[i],
					Patchset: changelistPatchsets[i],
				})
			}
			inv.Changelists = changelists
			return fn(inv)
		})
}

// SaveUnverified returns a mutation to insert the ingested invocation into
// the IngestedInvocations table. The ingested invocation is not validated.
func (inv *IngestedInvocation) SaveUnverified() *spanner.Mutation {
	var presubmitRunMode spanner.NullInt64
	var presubmitRunOwner spanner.NullString
	if inv.PresubmitRun != nil {
		presubmitRunMode = spanner.NullInt64{Valid: true, Int64: int64(inv.PresubmitRun.Mode)}
		presubmitRunOwner = spanner.NullString{Valid: true, StringVal: inv.PresubmitRun.Owner}
	}

	var gitReferenceHash []byte
	var commitPosition spanner.NullInt64
	var commitHash spanner.NullString
	if inv.CommitPosition > 0 {
		gitReferenceHash = inv.GitReferenceHash
		commitPosition = spanner.NullInt64{Valid: true, Int64: inv.CommitPosition}
		commitHash = spanner.NullString{Valid: true, StringVal: inv.CommitHash}
	}

	changelistHosts := make([]string, 0, len(inv.Changelists))
	changelistChanges := make([]int64, 0, len(inv.Changelists))
	changelistPatchsets := make([]int64, 0, len(inv.Changelists))
	for _, cl := range inv.Changelists {
		changelistHosts = append(changelistHosts, cl.Host)
		changelistChanges = append(changelistChanges, cl.Change)
		changelistPatchsets = append(changelistPatchsets, cl.Patchset)
	}

	row := map[string]interface{}{
		"Project":              inv.Project,
		"IngestedInvocationId": inv.IngestedInvocationID,
		"SubRealm":             inv.SubRealm,
		"PartitionTime":        inv.PartitionTime,
		"BuildStatus":          inv.BuildStatus,
		"GitReferenceHash":     gitReferenceHash,
		"CommitPosition":       commitPosition,
		"CommitHash":           commitHash,
		"PresubmitRunMode":     presubmitRunMode,
		"PresubmitRunOwner":    presubmitRunOwner,
		"ChangelistHosts":      changelistHosts,
		"ChangelistChanges":    changelistChanges,
		"ChangelistPatchsets":  changelistPatchsets,
	}
	return spanner.InsertOrUpdateMap("IngestedInvocations", spanutil.ToSpannerMap(row))
}

// TestResult represents a row in the TestResults table.
type TestResult struct {
	Project              string
	TestID               string
	PartitionTime        time.Time
	VariantHash          string
	IngestedInvocationID string
	RunIndex             int64
	ResultIndex          int64
	IsUnexpected         bool
	RunDuration          *time.Duration
	Status               pb.TestResultStatus
	// Properties of the test variant in the invocation (stored denormalised) follow.
	ExonerationReasons []pb.ExonerationReason
	// Properties of the invocation (stored denormalised) follow.
	SubRealm         string
	BuildStatus      pb.BuildStatus
	PresubmitRun     *PresubmitRun
	GitReferenceHash []byte
	CommitPosition   int64
	Changelists      []Changelist
}

// ReadTestResults reads test results from the TestResults table.
// Must be called in a spanner transactional context.
func ReadTestResults(ctx context.Context, keys spanner.KeySet, fn func(tr *TestResult) error) error {
	var b spanutil.Buffer
	fields := []string{
		"Project", "TestId", "PartitionTime", "VariantHash", "IngestedInvocationId",
		"RunIndex", "ResultIndex",
		"IsUnexpected", "RunDurationUsec", "Status",
		"ExonerationReasons",
		"SubRealm", "BuildStatus",
		"PresubmitRunMode", "PresubmitRunOwner",
		"GitReferenceHash", "CommitPosition",
		"ChangelistHosts", "ChangelistChanges", "ChangelistPatchsets",
	}
	return span.Read(ctx, "TestResults", keys, fields).Do(
		func(row *spanner.Row) error {
			tr := &TestResult{}
			var runDurationUsec spanner.NullInt64
			var isUnexpected spanner.NullBool
			var presubmitRunMode spanner.NullInt64
			var presubmitRunOwner spanner.NullString
			var gitReferenceHash []byte
			var commitPosition spanner.NullInt64
			var changelistHosts []string
			var changelistChanges []int64
			var changelistPatchsets []int64
			err := b.FromSpanner(
				row,
				&tr.Project, &tr.TestID, &tr.PartitionTime, &tr.VariantHash, &tr.IngestedInvocationID,
				&tr.RunIndex, &tr.ResultIndex,
				&isUnexpected, &runDurationUsec, &tr.Status,
				&tr.ExonerationReasons,
				&tr.SubRealm, &tr.BuildStatus,
				&presubmitRunMode, &presubmitRunOwner,
				&gitReferenceHash, &commitPosition,
				&changelistHosts, &changelistChanges, &changelistPatchsets,
			)
			if err != nil {
				return err
			}
			if runDurationUsec.Valid {
				runDuration := time.Microsecond * time.Duration(runDurationUsec.Int64)
				tr.RunDuration = &runDuration
			}
			tr.IsUnexpected = isUnexpected.Valid && isUnexpected.Bool

			// Data in Spanner should be consistent, so presubmitRunMode.Valid ==
			//   presubmitRunOwner.Valid.
			if presubmitRunMode.Valid {
				tr.PresubmitRun = &PresubmitRun{
					Mode:  pb.PresubmitRunMode(presubmitRunMode.Int64),
					Owner: presubmitRunOwner.StringVal,
				}
			}

			// Data in Spanner should be consistent, so commitPosition.Valid ==
			// (gitReferenceHash != nil).
			if commitPosition.Valid {
				tr.GitReferenceHash = gitReferenceHash
				tr.CommitPosition = commitPosition.Int64
			}

			// Data in spanner should be consistent, so
			// len(changelistHosts) == len(changelistChanges)
			//    == len(changelistPatchsets).
			if len(changelistHosts) != len(changelistChanges) ||
				len(changelistChanges) != len(changelistPatchsets) {
				panic("Changelist arrays have mismatched length in Spanner")
			}
			changelists := make([]Changelist, 0, len(changelistHosts))
			for i := range changelistHosts {
				changelists = append(changelists, Changelist{
					Host:     changelistHosts[i],
					Change:   changelistChanges[i],
					Patchset: changelistPatchsets[i],
				})
			}
			tr.Changelists = changelists
			return fn(tr)
		})
}

// TestResultSaveCols is the set of columns written to in a test result save.
// Allocated here once to avoid reallocating on every test result save.
var TestResultSaveCols = []string{
	"Project", "TestId", "PartitionTime", "VariantHash",
	"IngestedInvocationId", "RunIndex", "ResultIndex",
	"IsUnexpected", "RunDurationUsec", "Status",
	"ExonerationReasons", "SubRealm", "BuildStatus",
	"PresubmitRunMode", "PresubmitRunOwner",
	"GitReferenceHash", "CommitPosition",
	"ChangelistHosts", "ChangelistChanges", "ChangelistPatchsets",
}

// SaveUnverified prepare a mutation to insert the test result into the
// TestResults table. The test result is not validated.
func (tr *TestResult) SaveUnverified() *spanner.Mutation {
	var runDurationUsec spanner.NullInt64
	if tr.RunDuration != nil {
		runDurationUsec.Int64 = tr.RunDuration.Microseconds()
		runDurationUsec.Valid = true
	}

	var presubmitRunMode spanner.NullInt64
	var presubmitRunOwner spanner.NullString
	if tr.PresubmitRun != nil {
		presubmitRunMode = spanner.NullInt64{Valid: true, Int64: int64(tr.PresubmitRun.Mode)}
		presubmitRunOwner = spanner.NullString{Valid: true, StringVal: tr.PresubmitRun.Owner}
	}

	var gitReferenceHash []byte
	var commitPosition spanner.NullInt64
	if tr.CommitPosition > 0 {
		gitReferenceHash = tr.GitReferenceHash
		commitPosition = spanner.NullInt64{Valid: true, Int64: tr.CommitPosition}
	}

	changelistHosts := make([]string, 0, len(tr.Changelists))
	changelistChanges := make([]int64, 0, len(tr.Changelists))
	changelistPatchsets := make([]int64, 0, len(tr.Changelists))
	for _, cl := range tr.Changelists {
		changelistHosts = append(changelistHosts, cl.Host)
		changelistChanges = append(changelistChanges, cl.Change)
		changelistPatchsets = append(changelistPatchsets, cl.Patchset)
	}

	isUnexpected := spanner.NullBool{Bool: tr.IsUnexpected, Valid: tr.IsUnexpected}

	exonerationReasons := tr.ExonerationReasons
	if len(exonerationReasons) == 0 {
		// Store absence of exonerations as a NULL value in the database
		// rather than an empty array. Backfilling the column is too
		// time consuming and NULLs use slightly less storage space.
		exonerationReasons = nil
	}

	// Specify values in a slice directly instead of
	// creating a map and using spanner.InsertOrUpdateMap.
	// Profiling revealed ~15% of all CPU cycles spent
	// ingesting test results were wasted generating a
	// map and converting it back to the slice
	// needed for a *spanner.Mutation using InsertOrUpdateMap.
	// Ingestion appears to be CPU bound at times.
	vals := []interface{}{
		tr.Project, tr.TestID, tr.PartitionTime, tr.VariantHash,
		tr.IngestedInvocationID, tr.RunIndex, tr.ResultIndex,
		isUnexpected, runDurationUsec, int64(tr.Status),
		spanutil.ToSpanner(exonerationReasons), tr.SubRealm, int64(tr.BuildStatus),
		presubmitRunMode, presubmitRunOwner,
		gitReferenceHash, commitPosition,
		changelistHosts, changelistChanges, changelistPatchsets,
	}
	return spanner.InsertOrUpdate("TestResults", TestResultSaveCols, vals)
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

// statement generates a spanner statement for the specified query template.
func (opts ReadTestHistoryOptions) statement(ctx context.Context, tmpl string, paginationParams []string) (spanner.Statement, error) {
	params := map[string]interface{}{
		"project":   opts.Project,
		"testId":    opts.TestID,
		"subRealms": opts.SubRealms,
		"limit":     opts.PageSize,

		// If the filter is unspecified, this param will be ignored during the
		// statement generation step.
		"hasUnsubmittedChanges": opts.SubmittedFilter == pb.SubmittedFilter_ONLY_UNSUBMITTED,

		// Verdict status enum values.
		"unexpected":          int(pb.TestVerdictStatus_UNEXPECTED),
		"unexpectedlySkipped": int(pb.TestVerdictStatus_UNEXPECTEDLY_SKIPPED),
		"flaky":               int(pb.TestVerdictStatus_FLAKY),
		"exonerated":          int(pb.TestVerdictStatus_EXONERATED),
		"expected":            int(pb.TestVerdictStatus_EXPECTED),

		// Test result status enum values.
		"skip": int(pb.TestResultStatus_SKIP),
		"pass": int(pb.TestResultStatus_PASS),
	}
	input := map[string]interface{}{
		"hasSubRealms":       len(opts.SubRealms) > 0,
		"hasLimit":           opts.PageSize > 0,
		"hasSubmittedFilter": opts.SubmittedFilter != pb.SubmittedFilter_SUBMITTED_FILTER_UNSPECIFIED,
		"pagination":         opts.PageToken != "",
		"params":             params,
	}

	if opts.TimeRange.GetEarliest() != nil {
		params["afterTime"] = opts.TimeRange.GetEarliest().AsTime()
	} else {
		params["afterTime"] = minSpannerTimestamp
	}
	if opts.TimeRange.GetLatest() != nil {
		params["beforeTime"] = opts.TimeRange.GetLatest().AsTime()
	} else {
		params["beforeTime"] = maxSpannerTimestamp
	}

	switch p := opts.VariantPredicate.GetPredicate().(type) {
	case *pb.VariantPredicate_Equals:
		input["hasVariantHash"] = true
		params["variantHash"] = pbutil.VariantHash(p.Equals)
	case *pb.VariantPredicate_Contains:
		if len(p.Contains.Def) > 0 {
			input["hasVariantKVs"] = true
			params["variantKVs"] = pbutil.VariantToStrings(p.Contains)
		}
	case *pb.VariantPredicate_HashEquals:
		input["hasVariantHash"] = true
		params["variantHash"] = p.HashEquals
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

	stmt, err := spanutil.GenerateStatement(testHistoryQueryTmpl, tmpl, input)
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
			&tv.PartitionTime,
			&tv.VariantHash,
			&tv.InvocationId,
			&status,
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

// TestVariantRealm represents a row in the TestVariantRealm table.
type TestVariantRealm struct {
	Project           string
	TestID            string
	VariantHash       string
	SubRealm          string
	Variant           *pb.Variant
	LastIngestionTime time.Time
}

// ReadTestVariantRealms read test variant realms from the TestVariantRealms
// table.
// Must be called in a spanner transactional context.
func ReadTestVariantRealms(ctx context.Context, keys spanner.KeySet, fn func(tvr *TestVariantRealm) error) error {
	var b spanutil.Buffer
	fields := []string{"Project", "TestId", "VariantHash", "SubRealm", "Variant", "LastIngestionTime"}
	return span.Read(ctx, "TestVariantRealms", keys, fields).Do(
		func(row *spanner.Row) error {
			tvr := &TestVariantRealm{}
			err := b.FromSpanner(
				row,
				&tvr.Project,
				&tvr.TestID,
				&tvr.VariantHash,
				&tvr.SubRealm,
				&tvr.Variant,
				&tvr.LastIngestionTime,
			)
			if err != nil {
				return err
			}
			return fn(tvr)
		})
}

// TestVariantRealmSaveCols is the set of columns written to in a test variant
// realm save. Allocated here once to avoid reallocating on every save.
var TestVariantRealmSaveCols = []string{
	"Project", "TestId", "VariantHash", "SubRealm",
	"Variant", "LastIngestionTime",
}

// SaveUnverified creates a mutation to save the test variant realm into
// the TestVariantRealms table. The test variant realm is not verified.
// Must be called in spanner RW transactional context.
func (tvr *TestVariantRealm) SaveUnverified() *spanner.Mutation {
	vals := []interface{}{
		tvr.Project, tvr.TestID, tvr.VariantHash, tvr.SubRealm,
		pbutil.VariantToStrings(tvr.Variant), tvr.LastIngestionTime,
	}
	return spanner.InsertOrUpdate("TestVariantRealms", TestVariantRealmSaveCols, vals)
}

// ReadVariantsOptions specifies options for ReadVariants().
type ReadVariantsOptions struct {
	SubRealms        []string
	VariantPredicate *pb.VariantPredicate
	PageSize         int
	PageToken        string
}

// parseQueryVariantsPageToken parses the positions from the page token.
func parseQueryVariantsPageToken(pageToken string) (afterHash string, err error) {
	tokens, err := pagination.ParseToken(pageToken)
	if err != nil {
		return "", err
	}

	if len(tokens) != 1 {
		return "", pagination.InvalidToken(errors.Reason("expected 1 components, got %d", len(tokens)).Err())
	}

	return tokens[0], nil
}

// ReadVariants reads all the variants of the specified test from the
// spanner database.
// Must be called in a spanner transactional context.
func ReadVariants(ctx context.Context, project, testID string, opts ReadVariantsOptions) (variants []*pb.QueryVariantsResponse_VariantInfo, nextPageToken string, err error) {
	paginationVariantHash := ""
	if opts.PageToken != "" {
		paginationVariantHash, err = parseQueryVariantsPageToken(opts.PageToken)
		if err != nil {
			return nil, "", err
		}
	}

	params := map[string]interface{}{
		"project":   project,
		"testId":    testID,
		"subRealms": opts.SubRealms,

		// Control pagination.
		"limit":                 opts.PageSize,
		"paginationVariantHash": paginationVariantHash,
	}
	input := map[string]interface{}{
		"hasSubRealms": len(opts.SubRealms) > 0,
		"hasLimit":     opts.PageSize > 0,
		"params":       params,
	}

	switch p := opts.VariantPredicate.GetPredicate().(type) {
	case *pb.VariantPredicate_Equals:
		input["hasVariantHash"] = true
		params["variantHash"] = pbutil.VariantHash(p.Equals)
	case *pb.VariantPredicate_Contains:
		if len(p.Contains.Def) > 0 {
			input["hasVariantKVs"] = true
			params["variantKVs"] = pbutil.VariantToStrings(p.Contains)
		}
	case *pb.VariantPredicate_HashEquals:
		input["hasVariantHash"] = true
		params["variantHash"] = p.HashEquals
	case nil:
		// No filter.
	default:
		panic(errors.Reason("unexpected variant predicate %q", opts.VariantPredicate).Err())
	}

	stmt, err := spanutil.GenerateStatement(variantsQueryTmpl, variantsQueryTmpl.Name(), input)
	if err != nil {
		return nil, "", err
	}
	stmt.Params = params

	var b spanutil.Buffer
	variants = make([]*pb.QueryVariantsResponse_VariantInfo, 0, opts.PageSize)
	err = span.Query(ctx, stmt).Do(func(row *spanner.Row) error {
		variant := &pb.QueryVariantsResponse_VariantInfo{}
		err := b.FromSpanner(
			row,
			&variant.VariantHash,
			&variant.Variant,
		)
		if err != nil {
			return err
		}
		variants = append(variants, variant)
		return nil
	})
	if err != nil {
		return nil, "", err
	}

	if opts.PageSize != 0 && len(variants) == opts.PageSize {
		lastVariant := variants[len(variants)-1]
		nextPageToken = pagination.Token(lastVariant.VariantHash)
	}
	return variants, nextPageToken, nil
}

var testHistoryQueryTmpl = template.Must(template.New("").Parse(`
	{{define "tvStatus"}}
		CASE
			WHEN ANY_VALUE(ExonerationReasons IS NOT NULL AND ARRAY_LENGTH(ExonerationReasons) > 0) THEN @exonerated
			-- Use COALESCE as IsUnexpected uses NULL to indicate false.
			WHEN LOGICAL_AND(NOT COALESCE(IsUnexpected, FALSE)) THEN @expected
			WHEN LOGICAL_AND(COALESCE(IsUnexpected, FALSE) AND Status = @skip) THEN @unexpectedlySkipped
			WHEN LOGICAL_AND(COALESCE(IsUnexpected, FALSE)) THEN @unexpected
			ELSE @flaky
		END TvStatus
	{{end}}

	{{define "testResultFilter"}}
		Project = @project
			AND TestId = @testId
			AND PartitionTime >= @afterTime
			AND PartitionTime < @beforeTime
			{{if .hasSubRealms}}
				AND SubRealm IN UNNEST(@subRealms)
			{{end}}
			{{if .hasVariantHash}}
				AND VariantHash = @variantHash
			{{end}}
			{{if .hasVariantKVs}}
				AND VariantHash IN (
					SELECT DISTINCT VariantHash
					FROM TestVariantRealms
					WHERE
						Project = @project
						AND TestId = @testId
						{{if .hasSubRealms}}
							AND SubRealm IN UNNEST(@subRealms)
						{{end}}
						AND (SELECT LOGICAL_AND(kv IN UNNEST(Variant)) FROM UNNEST(@variantKVs) kv)
				)
			{{end}}
			{{if .hasSubmittedFilter}}
				AND (ARRAY_LENGTH(ChangelistHosts) > 0) = @hasUnsubmittedChanges
			{{end}}
	{{end}}

	{{define "testHistoryQuery"}}
		SELECT
			PartitionTime,
			VariantHash,
			IngestedInvocationId,
			{{template "tvStatus" .}},
			CAST(AVG(IF(Status = @pass, RunDurationUsec, NULL)) AS INT64) AS PassedAvgDurationUsec,
		FROM TestResults
		WHERE
			{{template "testResultFilter" .}}
			{{if .pagination}}
				AND	(
					PartitionTime < TIMESTAMP(@paginationTime)
						OR (PartitionTime = TIMESTAMP(@paginationTime) AND VariantHash > @paginationVariantHash)
						OR (PartitionTime = TIMESTAMP(@paginationTime) AND VariantHash = @paginationVariantHash AND IngestedInvocationId > @paginationInvId)
				)
			{{end}}
		GROUP BY PartitionTime, VariantHash, IngestedInvocationId
		ORDER BY
			PartitionTime DESC,
			VariantHash ASC,
			IngestedInvocationId ASC
		{{if .hasLimit}}
			LIMIT @limit
		{{end}}
	{{end}}

	{{define "testHistoryStatsQuery"}}
		WITH verdicts AS (
			SELECT
				PartitionTime,
				VariantHash,
				IngestedInvocationId,
				{{template "tvStatus" .}},
				COUNTIF(Status = @pass AND RunDurationUsec IS NOT NULL) AS PassedWithDurationCount,
				SUM(IF(Status = @pass, RunDurationUsec, 0)) AS SumPassedDurationUsec,
			FROM TestResults
			WHERE
				{{template "testResultFilter" .}}
				{{if .pagination}}
					AND	PartitionTime < TIMESTAMP_ADD(TIMESTAMP(@paginationDate), INTERVAL 1 DAY)
				{{end}}
			GROUP BY PartitionTime, VariantHash, IngestedInvocationId
		)

		SELECT
			TIMESTAMP_TRUNC(PartitionTime, DAY, "UTC") AS PartitionDate,
			VariantHash,
			COUNTIF(TvStatus = @unexpected) AS UnexpectedCount,
			COUNTIF(TvStatus = @unexpectedlySkipped) AS UnexpectedlySkippedCount,
			COUNTIF(TvStatus = @flaky) AS FlakyCount,
			COUNTIF(TvStatus = @exonerated) AS ExoneratedCount,
			COUNTIF(TvStatus = @expected) AS ExpectedCount,
			CAST(SAFE_DIVIDE(SUM(SumPassedDurationUsec), SUM(PassedWithDurationCount)) AS INT64) AS PassedAvgDurationUsec,
		FROM verdicts
		GROUP BY PartitionDate, VariantHash
		{{if .pagination}}
			HAVING
				PartitionDate < TIMESTAMP(@paginationDate)
					OR (PartitionDate = TIMESTAMP(@paginationDate) AND VariantHash > @paginationVariantHash)
		{{end}}
		ORDER BY
			PartitionDate DESC,
			VariantHash ASC
		{{if .hasLimit}}
			LIMIT @limit
		{{end}}
	{{end}}
`))

var variantsQueryTmpl = template.Must(template.New("variantsQuery").Parse(`
	SELECT
		VariantHash,
		ANY_VALUE(Variant) as Variant,
	FROM TestVariantRealms
	WHERE
		Project = @project
			AND TestId = @testId
			{{if .hasSubRealms}}
				AND SubRealm IN UNNEST(@subRealms)
			{{end}}
			{{if .hasVariantHash}}
				AND VariantHash = @variantHash
			{{end}}
			{{if .hasVariantKVs}}
				AND (SELECT LOGICAL_AND(kv IN UNNEST(Variant)) FROM UNNEST(@variantKVs) kv)
			{{end}}
			AND VariantHash > @paginationVariantHash
	GROUP BY VariantHash
	ORDER BY VariantHash ASC
	{{if .hasLimit}}
		LIMIT @limit
	{{end}}
`))
