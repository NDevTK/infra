// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"infra/appengine/chrome-test-health/api"
	"testing"

	"cloud.google.com/go/civil"
	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"google.golang.org/grpc/codes"
)

type clientMock struct {
	lastListReq     *api.ListComponentsRequest
	lastFetchReq    *api.FetchTestMetricsRequest
	lastFetchDirReq *api.FetchDirectoryMetricsRequest
}

type coverageClientMock struct {
	lastGetProjectDefaultConfigReq *api.GetProjectDefaultConfigRequest
}

func (cm *clientMock) UpdateSummary(_ context.Context, fromDate civil.Date, toDate civil.Date) error {
	return nil
}

func (cm *clientMock) ListComponents(ctx context.Context, req *api.ListComponentsRequest) (*api.ListComponentsResponse, error) {
	cm.lastListReq = req
	return &api.ListComponentsResponse{}, nil
}

func (cm *clientMock) FetchMetrics(ctx context.Context, req *api.FetchTestMetricsRequest) (*api.FetchTestMetricsResponse, error) {
	cm.lastFetchReq = req
	return &api.FetchTestMetricsResponse{}, nil
}

func (cm *clientMock) FetchDirectoryMetrics(ctx context.Context, req *api.FetchDirectoryMetricsRequest) (*api.FetchDirectoryMetricsResponse, error) {
	cm.lastFetchDirReq = req
	return &api.FetchDirectoryMetricsResponse{}, nil
}

func (cm *coverageClientMock) GetProjectDefaultConfig(ctx context.Context, req *api.GetProjectDefaultConfigRequest) (*api.GetProjectDefaultConfigResponse, error) {
	cm.lastGetProjectDefaultConfigReq = req
	return &api.GetProjectDefaultConfigResponse{}, nil
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
		Convey("Missing from date", func() {
			request := &api.UpdateMetricsTableRequest{
				ToDate: "2023-01-01",
			}
			resp, err := srv.UpdateMetricsTable(ctx, request)

			So(err, ShouldErrLike, "from_date")
			So(resp, ShouldBeNil)
		})
		Convey("Missing to date", func() {
			request := &api.UpdateMetricsTableRequest{
				FromDate: "2023-01-01",
			}
			resp, err := srv.UpdateMetricsTable(ctx, request)

			So(err, ShouldErrLike, "to_date")
			So(resp, ShouldBeNil)
		})
	})

}

func TestListComponents(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	Convey("ListComponents", t, func() {
		mock := &clientMock{}

		srv := &testResourcesServer{
			Client: mock,
		}
		request := &api.ListComponentsRequest{}
		srv.ListComponents(ctx, request)

		So(request, ShouldResemble, mock.lastListReq)
	})
}

func TestFetchMetrics(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("FetchMetrics", t, func() {
		mock := &clientMock{}

		srv := &testResourcesServer{
			Client: mock,
		}
		Convey("Valid request", func() {
			request := &api.FetchTestMetricsRequest{
				Components: []string{"some>component"},
				Period:     api.Period_DAY,
				Dates:      []string{"2023-01-01"},
				Metrics:    []api.MetricType{api.MetricType_NUM_RUNS},
				Filter:     "filter:this",
				PageOffset: 1,
				PageSize:   10,
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
		Convey("Missing dates", func() {
			request := &api.FetchTestMetricsRequest{
				Components: []string{"some>component"},
				Period:     api.Period_DAY,
				Metrics:    []api.MetricType{api.MetricType_NUM_RUNS},
				Filter:     "filter:this",
				PageOffset: 1,
				PageSize:   10,
				Sort: &api.SortBy{
					Metric:    api.SortType_SORT_NUM_RUNS,
					Ascending: true,
				},
			}
			resp, err := srv.FetchTestMetrics(ctx, request)

			So(err, ShouldErrLike, "dates")
			So(resp, ShouldBeNil)
		})
		Convey("Missing metrics", func() {
			request := &api.FetchTestMetricsRequest{
				Components: []string{"some>component"},
				Period:     api.Period_DAY,
				Dates:      []string{"2023-01-01"},
				Filter:     "filter:this",
				PageOffset: 1,
				PageSize:   10,
				Sort: &api.SortBy{
					Metric:    api.SortType_SORT_NUM_RUNS,
					Ascending: true,
				},
			}
			resp, err := srv.FetchTestMetrics(ctx, request)

			So(err, ShouldErrLike, "metrics")
			So(resp, ShouldBeNil)
		})
	})
}

func TestFetchFileMetrics(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("FetchFileMetrics", t, func() {
		mock := &clientMock{}

		srv := &testResourcesServer{
			Client: mock,
		}
		Convey("Valid request", func() {
			request := &api.FetchDirectoryMetricsRequest{
				Components: []string{"some>component"},
				Period:     api.Period_DAY,
				Dates:      []string{"2023-01-01"},
				Metrics:    []api.MetricType{api.MetricType_NUM_RUNS},
				Filter:     "filter:this",
				ParentIds:  []string{"/"},
				Sort: &api.SortBy{
					Metric:    api.SortType_SORT_NUM_RUNS,
					Ascending: true,
				},
			}
			resp, err := srv.FetchDirectoryMetrics(ctx, request)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(mock.lastFetchDirReq, ShouldResemble, request)
		})
		Convey("Missing dates", func() {
			request := &api.FetchDirectoryMetricsRequest{
				Components: []string{"some>component"},
				Period:     api.Period_DAY,
				Metrics:    []api.MetricType{api.MetricType_NUM_RUNS},
				Filter:     "filter:this",
				ParentIds:  []string{"/"},
				Sort: &api.SortBy{
					Metric:    api.SortType_SORT_NUM_RUNS,
					Ascending: true,
				},
			}
			resp, err := srv.FetchDirectoryMetrics(ctx, request)

			So(err, ShouldErrLike, "dates")
			So(resp, ShouldBeNil)
		})
		Convey("Missing parentId", func() {
			request := &api.FetchDirectoryMetricsRequest{
				Components: []string{"some>component"},
				Period:     api.Period_DAY,
				Dates:      []string{"2023-01-01"},
				Metrics:    []api.MetricType{api.MetricType_NUM_RUNS},
				Filter:     "filter:this",
				Sort: &api.SortBy{
					Metric:    api.SortType_SORT_NUM_RUNS,
					Ascending: true,
				},
			}
			resp, err := srv.FetchDirectoryMetrics(ctx, request)

			So(err, ShouldErrLike, "parent_id")
			So(resp, ShouldBeNil)
		})
		Convey("Missing metrics", func() {
			request := &api.FetchDirectoryMetricsRequest{
				Components: []string{"some>component"},
				Period:     api.Period_DAY,
				Dates:      []string{"2023-01-01"},
				Filter:     "filter:this",
				ParentIds:  []string{"/"},
				Sort: &api.SortBy{
					Metric:    api.SortType_SORT_NUM_RUNS,
					Ascending: true,
				},
			}
			resp, err := srv.FetchDirectoryMetrics(ctx, request)

			So(err, ShouldErrLike, "metrics")
			So(resp, ShouldBeNil)
		})
	})
}

func TestGetProjectDefaultConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("GetProjectDefaultConfig", t, func() {
		mock := &coverageClientMock{}

		srv := &coverageServer{
			Client: mock,
		}
		Convey("Valid request", func() {
			request := &api.GetProjectDefaultConfigRequest{
				Project: "chromium",
			}
			resp, err := srv.GetProjectDefaultConfig(ctx, request)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(mock.lastGetProjectDefaultConfigReq, ShouldResemble, request)
		})
		Convey("Invalid argument Project", func() {
			request := &api.GetProjectDefaultConfigRequest{
				Project: "chromium src",
			}
			resp, err := srv.GetProjectDefaultConfig(ctx, request)

			So(err, ShouldErrLike, "Argument Project is invalid")
			So(err, ShouldHaveAppStatus, codes.InvalidArgument)
			So(resp, ShouldBeNil)
		})
		Convey("Missing project", func() {
			request := &api.GetProjectDefaultConfigRequest{}
			resp, err := srv.GetProjectDefaultConfig(ctx, request)

			So(err, ShouldErrLike, "project")
			So(resp, ShouldBeNil)
		})
	})
}
