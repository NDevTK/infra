// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/appstatus"
	"go.chromium.org/luci/resultdb/rdbperms"
	"go.chromium.org/luci/server/auth/realms"
	"go.chromium.org/luci/server/span"
	"google.golang.org/grpc/codes"

	"infra/appengine/weetbix/internal/pagination"
	"infra/appengine/weetbix/internal/testresults"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
	"infra/appengine/weetbix/utils"
)

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

	if req.GetPredicate().GetSubRealm() == "" {
		return nil, appstatus.Errorf(codes.Unimplemented, "multi-realm test history not implemented")
	}
	realm := req.GetProject() + ":" + req.GetPredicate().GetSubRealm()

	requiredPerms := []realms.Permission{rdbperms.PermListTestResults, rdbperms.PermListTestExonerations}
	if err := utils.HasPermissions(ctx, requiredPerms, realm, nil); err != nil {
		return nil, err
	}

	pageSize := int(pageSizeLimiter.Adjust(req.GetPageSize()))
	opts := testresults.ReadTestHistoryOptions{
		Project:          req.GetProject(),
		TestID:           req.GetTestId(),
		SubRealms:        []string{req.GetPredicate().GetSubRealm()},
		VariantPredicate: req.GetPredicate().GetVariantPredicate(),
		SubmittedFilter:  req.GetPredicate().GetSubmittedFilter(),
		TimeRange:        req.GetPredicate().GetPartitionTimeRange(),
		PageSize:         pageSize,
		PageToken:        req.GetPageToken(),
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

	if req.GetPredicate().GetSubRealm() == "" {
		return nil, appstatus.Errorf(codes.Unimplemented, "multi-realm test history not implemented")
	}
	realm := req.GetProject() + ":" + req.GetPredicate().GetSubRealm()

	requiredPerms := []realms.Permission{rdbperms.PermListTestResults, rdbperms.PermListTestExonerations}
	if err := utils.HasPermissions(ctx, requiredPerms, realm, nil); err != nil {
		return nil, err
	}

	logging.Infof(ctx, "project: %s test_id: %s sub_realm: %s", req.GetProject(), req.GetTestId(), req.GetPredicate().GetSubRealm())

	pageSize := int(pageSizeLimiter.Adjust(req.GetPageSize()))
	opts := testresults.ReadTestHistoryOptions{
		Project:          req.GetProject(),
		TestID:           req.GetTestId(),
		SubRealms:        []string{req.GetPredicate().GetSubRealm()},
		VariantPredicate: req.GetPredicate().GetVariantPredicate(),
		SubmittedFilter:  req.GetPredicate().GetSubmittedFilter(),
		TimeRange:        req.GetPredicate().GetPartitionTimeRange(),
		PageSize:         pageSize,
		PageToken:        req.GetPageToken(),
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

	if req.GetSubRealm() == "" {
		return nil, appstatus.Errorf(codes.Unimplemented, "multi-realm test history not implemented")
	}
	realm := req.GetProject() + ":" + req.GetSubRealm()

	requiredPerms := []realms.Permission{rdbperms.PermListTestResults}
	if err := utils.HasPermissions(ctx, requiredPerms, realm, nil); err != nil {
		return nil, err
	}

	pageSize := int(pageSizeLimiter.Adjust(req.GetPageSize()))
	opts := testresults.ReadVariantsOptions{
		SubRealms:        []string{req.GetSubRealm()},
		VariantPredicate: req.GetVariantPredicate(),
		PageSize:         pageSize,
		PageToken:        req.GetPageToken(),
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

	if req.GetSubRealm() == "" {
		return nil, appstatus.Errorf(codes.Unimplemented, "multi-realm test history not implemented")
	}
	realm := req.GetProject() + ":" + req.GetSubRealm()

	requiredPerms := []realms.Permission{rdbperms.PermListTestResults}
	if err := utils.HasPermissions(ctx, requiredPerms, realm, nil); err != nil {
		return nil, err
	}

	pageSize := int(pageSizeLimiter.Adjust(req.GetPageSize()))
	opts := testresults.QueryTestsOptions{
		SubRealms: []string{req.GetSubRealm()},
		PageSize:  pageSize,
		PageToken: req.GetPageToken(),
	}

	testIDs, nextPageToken, err := testresults.QueryTests(span.Single(ctx), req.GetProject(), req.GetTestIdSubstring(), opts)
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
