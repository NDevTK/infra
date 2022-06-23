// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"context"
	"regexp"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/config"
	"infra/appengine/weetbix/internal/testresults"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

var variantHashRe = regexp.MustCompile("^[0-9a-f]{16}$")

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
		Project:      req.Project,
		TestVariants: req.TestVariants,
		AsAtTime:     now,
	}
	ctx, cancel := span.ReadOnlyTransaction(ctx)
	defer cancel()
	response, err := testresults.QueryFailureRate(ctx, opts)
	if err != nil {
		return nil, err
	}
	return response, nil
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
		var variantHash string
		if tv.VariantHash != "" {
			if !variantHashRe.MatchString(tv.VariantHash) {
				return errors.Reason("test_variants[%v]: variant_hash is not valid", i).Err()
			}
			variantHash = tv.VariantHash
		}

		// Variant may be nil as not all tests have variants.
		if tv.Variant != nil || tv.VariantHash == "" {
			calculatedHash := pbutil.VariantHash(tv.Variant)
			if tv.VariantHash != "" && calculatedHash != tv.VariantHash {
				return errors.Reason("test_variants[%v]: variant and variant_hash mismatch, variant hashed to %s, expected %s", i, calculatedHash, tv.VariantHash).Err()
			}
			variantHash = calculatedHash
		}

		key := testVariant{testID: tv.TestId, variantHash: variantHash}
		if _, ok := uniqueTestVariants[key]; ok {
			return errors.Reason("test_variants[%v]: already requested in the same request", i).Err()
		}
		uniqueTestVariants[key] = struct{}{}
	}
	return nil
}
