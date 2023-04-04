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
	lastUpdateReq *api.UpdateMetricsTableRequest
}

func (cm *clientMock) UpdateSummary(_ context.Context, req *api.UpdateMetricsTableRequest) (*api.UpdateMetricsTableResponse, error) {
	cm.lastUpdateReq = req
	return &api.UpdateMetricsTableResponse{}, nil
}

func (cm *clientMock) UpdateDateSummary(context.Context, civil.Date) error {
	panic("must not be called")
}

func TestUpdateDailySummary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Valid request", t, func() {
		mock := &clientMock{}

		srv := &testResourcesServer{
			Client: mock,
		}
		request := &api.UpdateMetricsTableRequest{
			FromDate: "2023-01-01",
			ToDate:   "2023-01-02",
		}
		resp, err := srv.UpdateMetricsTable(ctx, request)

		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		So(mock.lastUpdateReq, ShouldResemble, request)
	})
}
