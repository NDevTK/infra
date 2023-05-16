// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"infra/appengine/test-resources/api"
	"testing"

	"cloud.google.com/go/civil"
	. "github.com/smartystreets/goconvey/convey"
)

type clientMock struct {
	lastFetchReq *api.FetchTestMetricsRequest
}

func (cm *clientMock) UpdateSummary(_ context.Context, fromDate civil.Date, toDate civil.Date) error {
	return nil
}

func (cm *clientMock) FetchMetrics(ctx context.Context, req *api.FetchTestMetricsRequest) (*api.FetchTestMetricsResponse, error) {
	cm.lastFetchReq = req
	return &api.FetchTestMetricsResponse{}, nil
}

func TestUpdateDailySummary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	Convey("DailySummary", t, func() {
		mock := &clientMock{}

		srv := &testResourcesServer{
			Client: mock,
		}
		Convey("Valid request", func() {
			request := &api.UpdateMetricsTableRequest{
				FromDate: "2023-01-01",
				ToDate:   "2023-01-02",
			}
			resp, err := srv.UpdateMetricsTable(ctx, request)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
		})
		Convey("Bad from date request", func() {
			request := &api.UpdateMetricsTableRequest{
				FromDate: "asdf",
				ToDate:   "2023-01-02",
			}
			resp, err := srv.UpdateMetricsTable(ctx, request)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
		})
		Convey("Bad to date request", func() {
			request := &api.UpdateMetricsTableRequest{
				FromDate: "2023-01-01",
				ToDate:   "asdf",
			}
			resp, err := srv.UpdateMetricsTable(ctx, request)

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
		})
	})

}

func TestFetchMetrics(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Valid request", t, func() {
		mock := &clientMock{}

		srv := &testResourcesServer{
			Client: mock,
		}
		request := &api.FetchTestMetricsRequest{
			Component: "some>component",
			Period:    api.Period_DAY,
			Dates:     []string{"2023-01-01"},
			Metrics:   []api.MetricType{api.MetricType_NUM_RUNS},
			Filter:    "filter:this",
			Page:      1,
			PageSize:  10,
			Sort: &api.SortBy{
				Metric:    api.SortType_SORT_NUM_RUNS,
				Ascending: true,
			},
		}
		resp, err := srv.FetchTestMetrics(ctx, request)

		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		So(mock.lastFetchReq, ShouldResemble, request)
	})
}
