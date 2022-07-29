// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/resultdb/rdbperms"
	"go.chromium.org/luci/server/auth/realms"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/pagination"
	"infra/appengine/weetbix/internal/perms"
	"infra/appengine/weetbix/internal/testresults"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

func init() {
	rdbperms.PermListTestResults.AddFlags(realms.UsedInQueryRealms)
	rdbperms.PermListTestExonerations.AddFlags(realms.UsedInQueryRealms)
}

var pageSizeLimiter = pagination.PageSizeLimiter{
	Default: 100,
	Max:     1000,
}

// testHistoryServer implements pb.TestHistoryServer.
type testHistoryServer struct {
}

// NewTestHistoryServer returns a new pb.TestHistoryServer.
func NewTestHistoryServer() pb.TestHistoryServer {
	return &pb.DecoratedTestHistory{
		Service:  &testHistoryServer{},
		Postlude: gRPCifyAndLogPostlude,
	}
}

// Retrieves test verdicts for a given test ID in a given project and in a given
// range of time.
func (*testHistoryServer) Query(ctx context.Context, req *pb.QueryTestHistoryRequest) (*pb.QueryTestHistoryResponse, error) {
	if err := validateQueryTestHistoryRequest(req); err != nil {
		return nil, invalidArgumentError(err)
	}

	subRealms, err := perms.QuerySubRealmsNonEmpty(ctx, req.Project, req.Predicate.SubRealm, nil, perms.ListTestResultsAndExonerations...)
	if err != nil {
		return nil, err
	}

	pageSize := int(pageSizeLimiter.Adjust(req.PageSize))
	opts := testresults.ReadTestHistoryOptions{
		Project:          req.Project,
		TestID:           req.TestId,
		SubRealms:        subRealms,
		VariantPredicate: req.Predicate.VariantPredicate,
		SubmittedFilter:  req.Predicate.SubmittedFilter,
		TimeRange:        req.Predicate.PartitionTimeRange,
		PageSize:         pageSize,
		PageToken:        req.PageToken,
	}

	verdicts, nextPageToken, err := testresults.ReadTestHistory(span.Single(ctx), opts)
	if err != nil {
		return nil, err
	}

	return &pb.QueryTestHistoryResponse{
		Verdicts:      verdicts,
		NextPageToken: nextPageToken,
	}, nil
}

func validateQueryTestHistoryRequest(req *pb.QueryTestHistoryRequest) error {
	switch {
	case req.GetProject() == "":
		return errors.Reason("project missing").Err()
	case req.GetTestId() == "":
		return errors.Reason("test_id missing").Err()
	}

	if err := pbutil.ValidateTestVerdictPredicate(req.GetPredicate()); err != nil {
		return errors.Annotate(err, "predicate").Err()
	}

	if err := pagination.ValidatePageSize(req.GetPageSize()); err != nil {
		return errors.Annotate(err, "page_size").Err()
	}

	return nil
}

// Retrieves a summary of test verdicts for a given test ID in a given project
// and in a given range of times.
func (*testHistoryServer) QueryStats(ctx context.Context, req *pb.QueryTestHistoryStatsRequest) (*pb.QueryTestHistoryStatsResponse, error) {
	if err := validateQueryTestHistoryStatsRequest(req); err != nil {
		return nil, invalidArgumentError(err)
	}

	subRealms, err := perms.QuerySubRealmsNonEmpty(ctx, req.Project, req.Predicate.SubRealm, nil, perms.ListTestResultsAndExonerations...)
	if err != nil {
		return nil, err
	}

	pageSize := int(pageSizeLimiter.Adjust(req.PageSize))
	opts := testresults.ReadTestHistoryOptions{
		Project:          req.Project,
		TestID:           req.TestId,
		SubRealms:        subRealms,
		VariantPredicate: req.Predicate.VariantPredicate,
		SubmittedFilter:  req.Predicate.SubmittedFilter,
		TimeRange:        req.Predicate.PartitionTimeRange,
		PageSize:         pageSize,
		PageToken:        req.PageToken,
	}

	groups, nextPageToken, err := testresults.ReadTestHistoryStats(span.Single(ctx), opts)
	if err != nil {
		return nil, err
	}

	return &pb.QueryTestHistoryStatsResponse{
		Groups:        groups,
		NextPageToken: nextPageToken,
	}, nil
}

func validateQueryTestHistoryStatsRequest(req *pb.QueryTestHistoryStatsRequest) error {
	switch {
	case req.GetProject() == "":
		return errors.Reason("project missing").Err()
	case req.GetTestId() == "":
		return errors.Reason("test_id missing").Err()
	}

	if err := pbutil.ValidateTestVerdictPredicate(req.GetPredicate()); err != nil {
		return errors.Annotate(err, "predicate").Err()
	}

	if err := pagination.ValidatePageSize(req.GetPageSize()); err != nil {
		return errors.Annotate(err, "page_size").Err()
	}

	return nil
}

// Retrieves variants for a given test ID in a given project that were recorded
// in the past 90 days.
func (*testHistoryServer) QueryVariants(ctx context.Context, req *pb.QueryVariantsRequest) (*pb.QueryVariantsResponse, error) {
	if err := validateQueryVariantsRequest(req); err != nil {
		return nil, invalidArgumentError(err)
	}

	subRealms, err := perms.QuerySubRealmsNonEmpty(ctx, req.Project, req.SubRealm, nil, rdbperms.PermListTestResults)
	if err != nil {
		return nil, err
	}

	pageSize := int(pageSizeLimiter.Adjust(req.PageSize))
	opts := testresults.ReadVariantsOptions{
		SubRealms:        subRealms,
		VariantPredicate: req.VariantPredicate,
		PageSize:         pageSize,
		PageToken:        req.PageToken,
	}

	variants, nextPageToken, err := testresults.ReadVariants(span.Single(ctx), req.GetProject(), req.GetTestId(), opts)
	if err != nil {
		return nil, err
	}

	return &pb.QueryVariantsResponse{
		Variants:      variants,
		NextPageToken: nextPageToken,
	}, nil
}

func validateQueryVariantsRequest(req *pb.QueryVariantsRequest) error {
	switch {
	case req.GetProject() == "":
		return errors.Reason("project missing").Err()
	case req.GetTestId() == "":
		return errors.Reason("test_id missing").Err()
	}

	if err := pagination.ValidatePageSize(req.GetPageSize()); err != nil {
		return errors.Annotate(err, "page_size").Err()
	}

	if req.GetVariantPredicate() != nil {
		if err := pbutil.ValidateVariantPredicate(req.GetVariantPredicate()); err != nil {
			return errors.Annotate(err, "predicate").Err()
		}
	}

	return nil
}

// QueryTests finds all test IDs that contain the given substring in a given
// project that were recorded in the past 90 days.
func (*testHistoryServer) QueryTests(ctx context.Context, req *pb.QueryTestsRequest) (*pb.QueryTestsResponse, error) {
	if err := validateQueryTestsRequest(req); err != nil {
		return nil, invalidArgumentError(err)
	}

	subRealms, err := perms.QuerySubRealmsNonEmpty(ctx, req.Project, req.SubRealm, nil, rdbperms.PermListTestResults)
	if err != nil {
		return nil, err
	}

	pageSize := int(pageSizeLimiter.Adjust(req.PageSize))
	opts := testresults.QueryTestsOptions{
		SubRealms: subRealms,
		PageSize:  pageSize,
		PageToken: req.GetPageToken(),
	}

	testIDs, nextPageToken, err := testresults.QueryTests(span.Single(ctx), req.Project, req.TestIdSubstring, opts)
	if err != nil {
		return nil, err
	}

	return &pb.QueryTestsResponse{
		TestIds:       testIDs,
		NextPageToken: nextPageToken,
	}, nil
}

func validateQueryTestsRequest(req *pb.QueryTestsRequest) error {
	switch {
	case req.GetProject() == "":
		return errors.Reason("project missing").Err()
	case req.GetTestIdSubstring() == "":
		return errors.Reason("test_id_substring missing").Err()
	}

	if err := pagination.ValidatePageSize(req.GetPageSize()); err != nil {
		return errors.Annotate(err, "page_size").Err()
	}

	return nil
}
