// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/clock/testclock"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"
	"go.chromium.org/luci/server/secrets"
	"go.chromium.org/luci/server/secrets/testsecrets"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"

	"infra/appengine/weetbix/internal/testresults"
	"infra/appengine/weetbix/internal/testutil"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

func TestTestVariantsServer(t *testing.T) {
	Convey("Given a projects server", t, func() {
		ctx := testutil.SpannerTestContext(t)

		// For user identification.
		ctx = authtest.MockAuthConfig(ctx)
		ctx = auth.WithState(ctx, &authtest.FakeState{
			Identity:       "user:someone@example.com",
			IdentityGroups: []string{"weetbix-access"},
		})
		ctx = secrets.Use(ctx, &testsecrets.Store{})

		// Provides datastore implementation needed for project config.
		ctx = memory.Use(ctx)
		server := NewTestVariantsServer()

		Convey("Unauthorised requests are rejected", func() {
			ctx = auth.WithState(ctx, &authtest.FakeState{
				Identity: "user:someone@example.com",
				// Not a member of weetbix-access.
				IdentityGroups: []string{"other-group"},
			})

			// Make some request (the request should not matter, as
			// a common decorator is used for all requests.)
			request := &pb.QueryTestVariantFailureRateRequest{}

			response, err := server.QueryFailureRate(ctx, request)
			st, _ := grpcStatus.FromError(err)
			So(st.Code(), ShouldEqual, codes.PermissionDenied)
			So(st.Message(), ShouldEqual, "not a member of weetbix-access")
			So(response, ShouldBeNil)
		})
		Convey("QueryFailureRate", func() {
			Convey("Valid input", func() {
				// March 11, 2022 is a Friday.
				referenceTime := time.Date(2022, time.March, 11, 12, 0, 0, 0, time.UTC)
				err := testresults.CreateQueryFailureRateTestData(ctx, referenceTime)
				So(err, ShouldBeNil)

				// Perform the query on following the Monday, at 6 hours earlier.
				queryTime := time.Date(2022, time.March, 14, 6, 0, 0, 0, time.UTC)
				ctx, _ := testclock.UseTime(ctx, queryTime)

				project, tvs := testresults.QueryFailureRateSampleRequest()
				request := &pb.QueryTestVariantFailureRateRequest{
					Project:      project,
					TestVariants: tvs,
				}

				response, err := server.QueryFailureRate(ctx, request)
				st, _ := grpcStatus.FromError(err)
				So(st.Code(), ShouldEqual, codes.OK)

				expectedResult := testresults.QueryFailureRateSampleResponse(referenceTime)
				So(response, ShouldResembleProto, &pb.QueryTestVariantFailureRateResponse{
					TestVariants: expectedResult,
				})
			})
			Convey("Invalid input", func() {
				// This checks at least one case of invalid input is detected, sufficient to verify
				// validation is invoked.
				// Exhaustive checking of request validation is performed in TestValidateQueryRateRequest.
				request := &pb.QueryTestVariantFailureRateRequest{
					Project: "",
					TestVariants: []*pb.TestVariantIdentifier{
						{
							TestId: "my_test",
						},
					},
				}

				response, err := server.QueryFailureRate(ctx, request)
				st, _ := grpcStatus.FromError(err)
				So(st.Code(), ShouldEqual, codes.InvalidArgument)
				So(st.Message(), ShouldEqual, `project missing`)
				So(response, ShouldBeNil)
			})
		})
	})
}

func TestFailureRateQueryAfterTime(t *testing.T) {
	Convey("failureRateQueryAfterTime", t, func() {
		// Expect failureRateQueryAfterTime to go back in time just far enough that 24 workday hours
		// are between the returned time and now.
		Convey("Monday", func() {
			// Given an input on a Monday (e.g. 14th of March 2022), expect
			// failureRateQueryAfterTime to return the corresponding time
			// on the previous Friday.

			now := time.Date(2022, time.March, 14, 23, 59, 59, 999999999, time.UTC)
			afterTime := failureRateQueryAfterTime(now)
			So(afterTime, ShouldEqual, time.Date(2022, time.March, 11, 23, 59, 59, 999999999, time.UTC))

			now = time.Date(2022, time.March, 14, 0, 0, 0, 0, time.UTC)
			afterTime = failureRateQueryAfterTime(now)
			So(afterTime, ShouldEqual, time.Date(2022, time.March, 11, 0, 0, 0, 0, time.UTC))
		})
		Convey("Sunday", func() {
			// Given a time on a Sunday (e.g. 13th of March 2022), expect
			// failureRateQueryAfterTime to return the start of the previous
			// Friday.
			startOfFriday := time.Date(2022, time.March, 11, 0, 0, 0, 0, time.UTC)

			now := time.Date(2022, time.March, 13, 23, 59, 59, 999999999, time.UTC)
			afterTime := failureRateQueryAfterTime(now)
			So(afterTime, ShouldEqual, startOfFriday)

			now = time.Date(2022, time.March, 13, 0, 0, 0, 0, time.UTC)
			afterTime = failureRateQueryAfterTime(now)
			So(afterTime, ShouldEqual, startOfFriday)
		})
		Convey("Saturday", func() {
			// Given a time on a Saturday (e.g. 12th of March 2022), expect
			// failureRateQueryAfterTime to return the start of the previous
			// Friday.
			startOfFriday := time.Date(2022, time.March, 11, 0, 0, 0, 0, time.UTC)

			now := time.Date(2022, time.March, 12, 23, 59, 59, 999999999, time.UTC)
			afterTime := failureRateQueryAfterTime(now)
			So(afterTime, ShouldEqual, startOfFriday)

			now = time.Date(2022, time.March, 12, 0, 0, 0, 0, time.UTC)
			afterTime = failureRateQueryAfterTime(now)
			So(afterTime, ShouldEqual, startOfFriday)
		})
		Convey("Tuesday to Friday", func() {
			// Given an input on a Tuesday (e.g. 15th of March 2022), expect
			// failureRateQueryAfterTime to return the corresponding time
			// the previous day.
			now := time.Date(2022, time.March, 15, 1, 2, 3, 4, time.UTC)
			afterTime := failureRateQueryAfterTime(now)
			So(afterTime, ShouldEqual, time.Date(2022, time.March, 14, 1, 2, 3, 4, time.UTC))

			// Given an input on a Friday (e.g. 18th of March 2022), expect
			// failureRateQueryAfterTime to return the corresponding time
			// the previous day.
			now = time.Date(2022, time.March, 18, 1, 2, 3, 4, time.UTC)
			afterTime = failureRateQueryAfterTime(now)
			So(afterTime, ShouldEqual, time.Date(2022, time.March, 17, 1, 2, 3, 4, time.UTC))
		})
	})
}

func TestValidateQueryFailureRateRequest(t *testing.T) {
	Convey("ValidateQueryFailureRateRequest", t, func() {
		req := &pb.QueryTestVariantFailureRateRequest{
			Project: "project",
			TestVariants: []*pb.TestVariantIdentifier{
				{
					TestId: "my_test",
					// Variant is optional as not all tests have variants.
				},
				{
					TestId:  "my_test2",
					Variant: pbutil.Variant("key1", "val1", "key2", "val2"),
				},
			},
		}

		Convey("valid", func() {
			err := validateQueryTestVariantFailureRateRequest(req)
			So(err, ShouldBeNil)
		})

		Convey("no project", func() {
			req.Project = ""
			err := validateQueryTestVariantFailureRateRequest(req)
			So(err, ShouldErrLike, "project missing")
		})

		Convey("invalid project", func() {
			req.Project = ":"
			err := validateQueryTestVariantFailureRateRequest(req)
			So(err, ShouldErrLike, `project is invalid, expected [a-z0-9\-]{1,40}`)
		})

		Convey("no test variants", func() {
			req.TestVariants = nil
			err := validateQueryTestVariantFailureRateRequest(req)
			So(err, ShouldErrLike, `test_variants missing`)
		})

		Convey("too many test variants", func() {
			req.TestVariants = make([]*pb.TestVariantIdentifier, 0, 101)
			for i := 0; i < 101; i++ {
				req.TestVariants = append(req.TestVariants, &pb.TestVariantIdentifier{
					TestId: fmt.Sprintf("test_id%v", i),
				})
			}
			err := validateQueryTestVariantFailureRateRequest(req)
			So(err, ShouldErrLike, `no more than 100 may be queried at a time`)
		})

		Convey("no test id", func() {
			req.TestVariants[1].TestId = ""
			err := validateQueryTestVariantFailureRateRequest(req)
			So(err, ShouldErrLike, `test_variants[1]: test_id missing`)
		})

		Convey("duplicate test variants", func() {
			req.TestVariants = []*pb.TestVariantIdentifier{
				{
					TestId:  "my_test",
					Variant: pbutil.Variant("key1", "val1", "key2", "val2"),
				},
				{
					TestId:  "my_test",
					Variant: pbutil.Variant("key1", "val1", "key2", "val2"),
				},
			}
			err := validateQueryTestVariantFailureRateRequest(req)
			So(err, ShouldErrLike, `test_variants[1]: already requested in the same request`)
		})
	})
}
