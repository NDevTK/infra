// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/appstatus"
	"go.chromium.org/luci/resultdb/rdbperms"
	"go.chromium.org/luci/server/auth/realms"
	"go.chromium.org/luci/server/span"
	"google.golang.org/grpc/codes"

	"infra/appengine/weetbix/internal/pagination"
	"infra/appengine/weetbix/internal/testverdicts"
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
	opts := testverdicts.ReadTestHistoryOptions{
		Project:          req.GetProject(),
		TestID:           req.GetTestId(),
		SubRealms:        []string{req.GetPredicate().GetSubRealm()},
		VariantPredicate: req.GetPredicate().GetVariantPredicate(),
		SubmittedFilter:  req.GetPredicate().GetSubmittedFilter(),
		TimeRange:        req.GetPredicate().GetPartitionTimeRange(),
		PageSize:         pageSize,
		PageToken:        req.GetPageToken(),
	}

	verdicts, nextPageToken, err := testverdicts.ReadTestHistory(span.Single(ctx), opts)
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

	pageSize := int(pageSizeLimiter.Adjust(req.GetPageSize()))
	opts := testverdicts.ReadTestHistoryOptions{
		Project:          req.GetProject(),
		TestID:           req.GetTestId(),
		SubRealms:        []string{req.GetPredicate().GetSubRealm()},
		VariantPredicate: req.GetPredicate().GetVariantPredicate(),
		SubmittedFilter:  req.GetPredicate().GetSubmittedFilter(),
		TimeRange:        req.GetPredicate().GetPartitionTimeRange(),
		PageSize:         pageSize,
		PageToken:        req.GetPageToken(),
	}

	groups, nextPageToken, err := testverdicts.ReadTestHistoryStats(span.Single(ctx), opts)
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
	return nil, appstatus.Errorf(codes.Unimplemented, "method QueryVariantsRequest not implemented")
}
