// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"

	"infra/cr_builder_health/healthpb"
)

type bbClientMock struct {
	setHealthCalls int
}

var existantBuilder = "existant-builder"
var nonExistantBuilder = "non-existant-builder"

func (c *bbClientMock) SetBuilderHealth(ctx context.Context, in *buildbucketpb.SetBuilderHealthRequest, opts ...grpc.CallOption) (*buildbucketpb.SetBuilderHealthResponse, error) {
	c.setHealthCalls += 1
	result := &buildbucketpb.SetBuilderHealthResponse{
		Responses: []*buildbucketpb.SetBuilderHealthResponse_Response{},
	}

	for _, req := range in.Health {
		if req.Id.Builder == existantBuilder {
			result.Responses = append(result.Responses, &buildbucketpb.SetBuilderHealthResponse_Response{
				Response: &buildbucketpb.SetBuilderHealthResponse_Response_Result{},
			})
		} else {
			result.Responses = append(result.Responses, &buildbucketpb.SetBuilderHealthResponse_Response{
				Response: &buildbucketpb.SetBuilderHealthResponse_Response_Error{
					Error: &status.Status{
						Code:    400,
						Message: "Invalid builder name",
					},
				},
			})
		}
	}

	return result, nil
}

func TestIsWeekend(t *testing.T) {
	t.Parallel()
	Convey("Test isWeekend function", t, func() {
		date1 := civil.Date{
			Year:  2023,
			Month: time.December,
			Day:   1,
		}
		date2 := civil.Date{
			Year:  2023,
			Month: time.December,
			Day:   2,
		}
		date3 := civil.Date{
			Year:  2023,
			Month: time.December,
			Day:   3,
		}
		date4 := civil.Date{
			Year:  2023,
			Month: time.December,
			Day:   4,
		}
		So(isWeekend(date1), ShouldEqual, false)
		So(isWeekend(date2), ShouldEqual, true)
		So(isWeekend(date3), ShouldEqual, true)
		So(isWeekend(date4), ShouldEqual, false)
	})
}

func TestBuilderID(t *testing.T) {
	t.Parallel()
	Convey("Test BuilderID function", t, func() {
		So(builderID("chromium", "ci", "builder1"), ShouldEqual, "chromium/ci/builder1")
		So(builderID("chrome", "try", "builder2"), ShouldEqual, "chrome/try/builder2")
	})
}

func TestCalculateIndicators(t *testing.T) {
	t.Parallel()

	Convey("(Slightly) healthy builders", t, func() {
		ctx := context.Background()
		rowsWithHealthScores := []Row{{
			Bucket:      "bucket",
			Builder:     existantBuilder,
			HealthScore: UNHEALTHY_SCORE,
			Date: civil.Date{
				Year:  2023,
				Month: time.December,
				Day:   5,
			},
		}, {
			Bucket:      "bucket",
			Builder:     existantBuilder,
			HealthScore: 8,
			Date: civil.Date{
				Year:  2023,
				Month: time.December,
				Day:   4,
			},
		}, {
			Bucket:      "bucket",
			Builder:     existantBuilder,
			HealthScore: HEALTHY_SCORE,
			Date: civil.Date{
				Year:  2023,
				Month: time.December,
				Day:   3,
			},
		}}

		input := healthpb.InputParams{
			Date: timestamppb.New(time.Date(2023, 12, 6, 0, 0, 0, 0, time.UTC)),
		}

		var srcConfig = SrcConfig{
			BucketSpecs: map[string]BuilderSpecs{
				"bucket": {
					"existant-builder": BuilderSpec{
						ProblemSpecs: []ProblemSpec{
							{
								Name:  "Unhealthy",
								Score: UNHEALTHY_SCORE,
								Thresholds: Thresholds{
									FailRate: AverageThresholds{Average: 0.2},
								},
							},
						},
					},
				},
			},
		}

		rowsWithIndicators, err := calculateIndicators(ctx, &input, rowsWithHealthScores, srcConfig)
		So(err, ShouldBeNil)
		So(len(rowsWithIndicators), ShouldEqual, 1)

		// As 2023/12/03, being a Sunday, is excluded from the health score calculation, the final health score should be 8
		So(rowsWithIndicators[0].HealthScore, ShouldEqual, 8)
	},
	)
}

func TestGenerate(t *testing.T) {
	t.Parallel()

	Convey("RPC Buildbucket is called ok", t, func() {
		ctx := context.Background()
		client := &bbClientMock{}
		rows := []Row{{
			Bucket:  "bucket",
			Builder: existantBuilder,
			Metrics: []*Metric{
				{Type: "build_mins_p50", Value: 59},
				{Type: "build_mins_p95", Value: 119},
				{Type: "pending_mins_p50", Value: 59},
				{Type: "pending_mins_p95", Value: 119},
				{Type: "fail_rate", Value: 0.05},
				{Type: "infra_fail_rate", Value: 0},
			},
		}}
		err := rpcBuildbucket(ctx, rows, client)
		So(client.setHealthCalls, ShouldEqual, 1)
		So(ctx.Err(), ShouldBeNil)
		So(err, ShouldBeNil)
	})

	Convey("RPC Buildbucket is called error", t, func() {
		ctx := context.Background()
		client := &bbClientMock{}
		rows := []Row{
			{
				Bucket:  "bucket",
				Builder: nonExistantBuilder,
				Metrics: []*Metric{
					{Type: "build_mins_p50", Value: 59},
					{Type: "build_mins_p95", Value: 119},
					{Type: "pending_mins_p50", Value: 59},
					{Type: "pending_mins_p95", Value: 119},
					{Type: "fail_rate", Value: 0.05},
					{Type: "infra_fail_rate", Value: 0},
				},
			},
			{
				Bucket:  "bucket",
				Builder: existantBuilder,
				Metrics: []*Metric{
					{Type: "build_mins_p50", Value: 59},
					{Type: "build_mins_p95", Value: 119},
					{Type: "pending_mins_p50", Value: 59},
					{Type: "pending_mins_p95", Value: 119},
					{Type: "fail_rate", Value: 0.05},
					{Type: "infra_fail_rate", Value: 0},
				},
			},
		}
		err := rpcBuildbucket(ctx, rows, client)
		So(client.setHealthCalls, ShouldEqual, 1)
		So(ctx.Err(), ShouldBeNil)
		So(err, ShouldBeNil)
	})
}
