// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"context"
	"time"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/config"
	"infra/appengine/weetbix/internal/testresults"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

// testVariantsServer implements pb.TestVariantServer.
type testVariantsServer struct {
}

// NewTestVariantsServer returns a new pb.TestVariantServer.
func NewTestVariantsServer() pb.TestVariantsServer {
	return &pb.DecoratedTestVariants{
		Prelude:  checkAllowedPrelude,
		Service:  &testVariantsServer{},
		Postlude: gRPCifyAndLogPostlude,
	}
}

// QueryFailureRate queries the failure rate of specified test variants,
// returning signals indicating if the test variant is flaky and/or
// deterministically failing.
func (*testVariantsServer) QueryFailureRate(ctx context.Context, req *pb.QueryTestVariantFailureRateRequest) (*pb.QueryTestVariantFailureRateResponse, error) {
	now := clock.Now(ctx)
	if err := validateQueryTestVariantFailureRateRequest(req); err != nil {
		return nil, invalidArgumentError(err)
	}

	opts := testresults.QueryFailureRateOptions{
		Project:            req.Project,
		TestVariants:       req.TestVariants,
		AfterPartitionTime: failureRateQueryAfterTime(now),
	}
	results, err := testresults.QueryFailureRate(span.Single(ctx), opts)
	if err != nil {
		return nil, err
	}
	return &pb.QueryTestVariantFailureRateResponse{
		TestVariants: results,
	}, nil
}

func validateQueryTestVariantFailureRateRequest(req *pb.QueryTestVariantFailureRateRequest) error {
	// MaxTestVariants is the maximum number of test variants to be queried in one request.
	const MaxTestVariants = 100

	if req.Project == "" {
		return errors.Reason("project missing").Err()
	}
	if !config.ProjectRe.MatchString(req.Project) {
		return errors.Reason("project is invalid, expected %s", config.ProjectRePattern).Err()
	}
	if len(req.TestVariants) == 0 {
		return errors.Reason("test_variants missing").Err()
	}
	if len(req.TestVariants) > MaxTestVariants {
		return errors.Reason("test_variants: no more than %v may be queried at a time", MaxTestVariants).Err()
	}
	type testVariant struct {
		testID      string
		variantHash string
	}
	uniqueTestVariants := make(map[testVariant]struct{})
	for i, tv := range req.TestVariants {
		if tv.GetTestId() == "" {
			return errors.Reason("test_variants[%v]: test_id missing", i).Err()
		}
		// Variant may be nil as not all tests have variants.

		key := testVariant{testID: tv.TestId, variantHash: pbutil.VariantHash(tv.Variant)}
		if _, ok := uniqueTestVariants[key]; ok {
			return errors.Reason("test_variants[%v]: already requested in the same request", i).Err()
		}
		uniqueTestVariants[key] = struct{}{}
	}
	return nil
}

// failureRateQueryAfterTime identifies the "after partition time" to use for
// the failure rate query. It calculates the start time of an interval
// ending at now, such that the interval includes exactly 24 hours of weekday
// data (in UTC).
//
// Rationale:
// Many projects see reduced testing activity on weekends, as fewer CLs are
// submitted. To try and increase the sample size of statistics returned
// on these days, we extend the time interval queried to still capture 24
// hours worth of weekday data.
func failureRateQueryAfterTime(now time.Time) time.Time {
	now = now.In(time.UTC)
	var startTime time.Time
	switch now.Weekday() {
	case time.Saturday:
		// Take us back to Saturday at 0:00.
		startTime = now.Truncate(24 * time.Hour)
		// Now take us back to Friday at 0:00.
		startTime = startTime.Add(-24 * time.Hour)
	case time.Sunday:
		// Take us back to Sunday at 0:00.
		startTime = now.Truncate(24 * time.Hour)
		// Now take us back to Friday at 0:00.
		startTime = startTime.Add(-2 * 24 * time.Hour)
	case time.Monday:
		// Take take us back to the same time on
		// Friday.
		startTime = now.Add(-3 * 24 * time.Hour)
	default:
		startTime = now.Add(-24 * time.Hour)
	}
	return startTime
}
