// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cron

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	mock "github.com/stretchr/testify/mock"

	"infra/appengine/chrome-test-health/datastorage"
	"infra/appengine/chrome-test-health/datastorage/mocks"
	"infra/appengine/chrome-test-health/internal/coverage"
	"infra/appengine/chrome-test-health/internal/coverage/entities"
)

func getMockPresubmitData() []*entities.PresubmitCoverageData {
	return []*entities.PresubmitCoverageData{
		{
			ServerHost:      "chromium-review.googlesource.com",
			UpdateTimestamp: time.Now().Add(-time.Hour * 6),
			IncrementalPercentages: []entities.Cov{
				{
					Path:         "//dir1/dir2/file1.cc",
					CoveredLines: 0,
					TotalLines:   4,
				},
				{
					Path:         "//dir1/dir3/file2.cc",
					CoveredLines: 11,
					TotalLines:   11,
				},
			},
		},
		{
			ServerHost:      "chromium-review.googlesource.com",
			UpdateTimestamp: time.Now().Add(-time.Hour * 25),
			IncrementalPercentages: []entities.Cov{
				{
					Path:         "//dir1/dir2/file1.cc",
					CoveredLines: 0,
					TotalLines:   4,
				},
				{
					Path:         "//dir1/dir3/file2.cc",
					CoveredLines: 11,
					TotalLines:   11,
				},
			},
		},
	}
}

func TestGetPresubmitReportsForLastYear(t *testing.T) {
	t.Parallel()
	client := CronClient{}
	ctx := context.Background()

	Convey("Should return presubmit data for last day", t, func() {
		reports := getMockPresubmitData()
		mockDataClient := mocks.NewIDataClient(t)
		mockDataClient.On(
			"Query",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"PresubmitCoverageData",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(c context.Context, result interface{}, dataType string, queryFilters []datastorage.QueryFilter, order interface{}, limit int, options ...interface{}) error {
				for _, rep := range reports {
					matchesHost := queryFilters[0].Value == rep.ServerHost
					t := queryFilters[1].Value.(time.Time)
					matchesTime := t.Before(rep.UpdateTimestamp)
					if matchesHost && matchesTime {
						res := reflect.ValueOf(result).Elem()
						res.Set(reflect.Append(res, reflect.ValueOf(rep).Elem()))
						return nil
					}
				}
				return nil
			},
		)
		client.coverageV1DsClient = mockDataClient

		data, err := client.getPresubmitReportsOneDay(ctx)
		So(err, ShouldBeNil)
		So(data, ShouldHaveLength, 1)
		expectedData := []entities.PresubmitCoverageData{
			*reports[0],
		}
		So(data, ShouldResemble, expectedData)
	})

	Convey("Should error out with no matching index message", t, func() {
		mockDataClient := mocks.NewIDataClient(t)
		mockDataClient.On(
			"Query",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"PresubmitCoverageData",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(c context.Context, result interface{}, dataType string, queryFilters []datastorage.QueryFilter, order interface{}, limit int, options ...interface{}) error {
				return fmt.Errorf("PresubmitCoverageData: %s", "No matching indexes found")
			},
		)
		client.coverageV1DsClient = mockDataClient

		reports, err := client.getPresubmitReportsOneDay(ctx)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, coverage.ErrInternalServerError)
		So(reports, ShouldBeNil)
	})
}
