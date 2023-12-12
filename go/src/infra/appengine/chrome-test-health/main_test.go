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
	lastGetProjectDefaultConfigReq           *api.GetProjectDefaultConfigRequest
	lastGetCoverageSummaryReq                *api.GetCoverageSummaryRequest
	lastGetAbsoluteCoverageDataOneYearReq    *api.GetAbsoluteCoverageDataOneYearRequest
	lastGetIncrementalCoverageDataOneYearReq *api.GetIncrementalCoverageDataOneYearRequest
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

func (ccm *coverageClientMock) GetProjectDefaultConfig(ctx context.Context, req *api.GetProjectDefaultConfigRequest) (*api.GetProjectDefaultConfigResponse, error) {
	ccm.lastGetProjectDefaultConfigReq = req
	return &api.GetProjectDefaultConfigResponse{}, nil
}

func (ccm *coverageClientMock) GetCoverageSummary(ctx context.Context, req *api.GetCoverageSummaryRequest) (*api.GetCoverageSummaryResponse, error) {
	ccm.lastGetCoverageSummaryReq = req
	return &api.GetCoverageSummaryResponse{}, nil
}

func (ccm *coverageClientMock) GetAbsoluteCoverageDataOneYear(
	ctx context.Context,
	req *api.GetAbsoluteCoverageDataOneYearRequest,
) (*api.GetAbsoluteCoverageDataOneYearResponse, error) {
	ccm.lastGetAbsoluteCoverageDataOneYearReq = req
	return &api.GetAbsoluteCoverageDataOneYearResponse{}, nil
}

func (ccm *coverageClientMock) GetIncrementalCoverageDataOneYear(
	ctx context.Context,
	req *api.GetIncrementalCoverageDataOneYearRequest,
) (*api.GetIncrementalCoverageDataOneYearResponse, error) {
	ccm.lastGetIncrementalCoverageDataOneYearReq = req
	return &api.GetIncrementalCoverageDataOneYearResponse{}, nil
}

func TestValidatePresence(t *testing.T) {
	t.Parallel()

	Convey("Validate Presence", t, func() {
		Convey("Should be false for empty string", func() {
			isPresent := validatePresence("   ")
			So(isPresent, ShouldBeFalse)
		})
		Convey("Should be false for nil", func() {
			isPresent := validatePresence(nil)
			So(isPresent, ShouldBeFalse)
		})
		Convey("Should be true", func() {
			isPresent := validatePresence("test")
			So(isPresent, ShouldBeTrue)
		})
	})
}

func TestValidateFormat(t *testing.T) {
	t.Parallel()

	Convey("Validate Format", t, func() {
		Convey("Should be false", func() {
			isValidFormat := validateFormat("test4", "^(test1|test2|test3)$")
			So(isValidFormat, ShouldBeFalse)
		})
		Convey("Should be true", func() {
			isValidFormat := validateFormat("test1", "^(test1|test2|test3)$")
			So(isValidFormat, ShouldBeTrue)
		})
	})
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

func TestGetCoverageSummary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("GetCoverageSummary", t, func() {
		mock := &coverageClientMock{}
		srv := &coverageServer{
			Client: mock,
		}
		request := &api.GetCoverageSummaryRequest{
			GitilesHost:     "chromium.googlesource.com",
			GitilesProject:  "chromium/src",
			GitilesRef:      "refs/heads/main",
			GitilesRevision: "03d4e64771cbc97f3ca5e4bbe85490d7cf909a0a",
			UnitTestsOnly:   false,
			Path:            "//chrome/browser/display_capture/",
			Bucket:          "ci",
			Builder:         "linux-code-coverage",
		}
		Convey("Valid request", func() {
			resp, err := srv.GetCoverageSummary(ctx, request)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(mock.lastGetCoverageSummaryReq, ShouldResemble, request)
		})
		Convey("Missing gitiles host", func() {
			req := request
			req.GitilesHost = ""
			resp, err := srv.GetCoverageSummary(ctx, request)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Gitiles Host is a required argument")
			So(resp, ShouldBeNil)
		})
		Convey("Missing gitiles project", func() {
			req := request
			req.GitilesProject = ""
			resp, err := srv.GetCoverageSummary(ctx, request)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Gitiles Project is a required argument")
			So(resp, ShouldBeNil)
		})
		Convey("Missing gitiles ref", func() {
			req := request
			req.GitilesRef = ""
			resp, err := srv.GetCoverageSummary(ctx, request)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Gitiles Ref is a required argument")
			So(resp, ShouldBeNil)
		})
		Convey("Missing gitiles revision", func() {
			req := request
			req.GitilesRevision = ""
			resp, err := srv.GetCoverageSummary(ctx, request)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Gitiles Revision is a required argument")
			So(resp, ShouldBeNil)
		})
		Convey("Missing gitiles both path and components", func() {
			req := request
			req.Path = ""
			req.Components = []string{}
			resp, err := srv.GetCoverageSummary(ctx, request)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Either path or components should be specified")
			So(resp, ShouldBeNil)
		})
		Convey("Both path and components specified", func() {
			req := request
			req.Path = "//"
			req.Components = []string{"C1", "C2"}
			resp, err := srv.GetCoverageSummary(ctx, request)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Either path or components should be specified not both")
			So(resp, ShouldBeNil)
		})
		Convey("Invalid Builder", func() {
			req := request
			req.Builder = ""
			resp, err := srv.GetCoverageSummary(ctx, request)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Builder is a required argument")
			So(resp, ShouldBeNil)

			req.Builder = "linux-code-coverage&123"
			resp, err = srv.GetCoverageSummary(ctx, request)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Builder is not provided in required format")
			So(resp, ShouldBeNil)
		})
		Convey("Invalid Bucket", func() {
			req := request
			req.Bucket = ""
			resp, err := srv.GetCoverageSummary(ctx, request)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Bucket is a required argument")
			So(resp, ShouldBeNil)

			req.Bucket = "ci#121"
			resp, err = srv.GetCoverageSummary(ctx, request)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Bucket is not provided in required format")
			So(resp, ShouldBeNil)
		})
	})
}

func TestGetAbsoluteCoverageDataOneYear(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mock := &coverageClientMock{}
	srv := &coverageServer{
		Client: mock,
	}
	request := &api.GetAbsoluteCoverageDataOneYearRequest{
		Paths:         []string{"//p1/p2/"},
		Components:    []string{"C1", "C2"},
		UnitTestsOnly: true,
		Bucket:        "ci",
		Builder:       "linux-code-coverage",
	}

	Convey("Should pass", t, func() {
		req := request
		_, err := srv.GetAbsoluteCoverageDataOneYear(ctx, req)
		So(err, ShouldBeNil)
	})

	Convey("Should fail", t, func() {
		Convey("Missing required params", func() {
			Convey("Missing both paths and components", func() {
				req := request
				req.Paths = []string{}
				req.Components = []string{}
				resp, err := srv.GetAbsoluteCoverageDataOneYear(ctx, req)
				So(err, ShouldNotBeNil)
				So(err, ShouldErrLike, "Either paths or components should be specified")
				So(resp, ShouldBeNil)
			})
			Convey("Missing bucket", func() {
				req := request
				req.Bucket = ""
				resp, err := srv.GetAbsoluteCoverageDataOneYear(ctx, req)
				So(err, ShouldNotBeNil)
				So(err, ShouldErrLike, "Bucket is a required argument")
				So(resp, ShouldBeNil)
			})
			Convey("Missing builder", func() {
				req := request
				req.Builder = ""
				resp, err := srv.GetAbsoluteCoverageDataOneYear(ctx, req)
				So(err, ShouldNotBeNil)
				So(err, ShouldErrLike, "Builder is a required argument")
				So(resp, ShouldBeNil)
			})
		})

		Convey("Invalid params", func() {
			Convey("Invalid Builder", func() {
				req := request
				req.Builder = "a___$$$b"
				resp, err := srv.GetAbsoluteCoverageDataOneYear(ctx, req)
				So(err, ShouldNotBeNil)
				So(err, ShouldErrLike, "Builder is not provided in required format")
				So(resp, ShouldBeNil)
			})
			Convey("Invalid Bucket", func() {
				req := request
				req.Builder = "linux-code-coverage"
				req.Bucket = "a___$$$b"
				resp, err := srv.GetAbsoluteCoverageDataOneYear(ctx, req)
				So(err, ShouldNotBeNil)
				So(err, ShouldErrLike, "Bucket is not provided in required format")
				So(resp, ShouldBeNil)
			})
		})
	})
}

func TestGetIncrementalCoverageDataOneYear(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mock := &coverageClientMock{}
	srv := &coverageServer{
		Client: mock,
	}
	request := &api.GetIncrementalCoverageDataOneYearRequest{
		Paths: []string{"//p1/p2/"},
	}

	Convey("Should fail", t, func() {
		Convey("Missing paths", func() {
			req := request
			req.Paths = []string{}
			resp, err := srv.GetIncrementalCoverageDataOneYear(ctx, req)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Paths should be specified")
			So(resp, ShouldBeNil)
		})

		Convey("Path not relative to project root", func() {
			req := request
			req.Paths = []string{"/a/b/"}
			resp, err := srv.GetIncrementalCoverageDataOneYear(ctx, req)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Path /a/b/ is not relative to root, it should start with //")
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
				LuciProject: "chromium",
			}
			resp, err := srv.GetProjectDefaultConfig(ctx, request)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(mock.lastGetProjectDefaultConfigReq, ShouldResemble, request)
		})
		Convey("Invalid argument Project", func() {
			request := &api.GetProjectDefaultConfigRequest{
				LuciProject: "chromium src",
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

func TestPathRelativeToRoot(t *testing.T) {
	t.Parallel()

	Convey("Should be true when path starts with //", t, func() {
		isRel := pathRelativeToRoot("//a/b/")
		So(isRel, ShouldBeTrue)
	})

	Convey("Should be false when path doesn't start with //", t, func() {
		isRel := pathRelativeToRoot("/a/b/")
		So(isRel, ShouldBeFalse)
	})
}
