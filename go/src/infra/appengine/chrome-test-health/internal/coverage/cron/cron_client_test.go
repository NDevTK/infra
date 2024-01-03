// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cron

import (
	"context"
	"errors"
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
			Change:          1,
			Patchset:        2,
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
			Change:          1,
			Patchset:        1,
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

func TestSplitSinglePresubmitData(t *testing.T) {
	t.Parallel()
	client := CronClient{}

	Convey("Should split presubmit data if patchset is latest", t, func() {
		reports := getMockPresubmitData()
		maxPatchsetMap := map[int64]int64{1: 2}
		result := client.splitSinglePresubmitData(reports[0], maxPatchsetMap, false)
		expected := map[string]IncrementalCoverageData{
			"//":                   {CoveredFiles: 1, TotalFiles: 2, IsDir: true},
			"//dir1/":              {CoveredFiles: 1, TotalFiles: 2, IsDir: true},
			"//dir1/dir2/":         {CoveredFiles: 0, TotalFiles: 1, IsDir: true},
			"//dir1/dir2/file1.cc": {CoveredFiles: 0, TotalFiles: 1, IsDir: false},
			"//dir1/dir3/":         {CoveredFiles: 1, TotalFiles: 1, IsDir: true},
			"//dir1/dir3/file2.cc": {CoveredFiles: 1, TotalFiles: 1, IsDir: false},
		}
		So(result, ShouldResemble, expected)
	})

	Convey("Should return nil if patchset is not latest", t, func() {
		reports := getMockPresubmitData()
		maxPatchsetMap := map[int64]int64{1: 2}
		result := client.splitSinglePresubmitData(reports[1], maxPatchsetMap, false)
		So(result, ShouldBeNil)
	})
}

func TestGetMaxPatchsetToChangeMap(t *testing.T) {
	t.Parallel()
	client := CronClient{}

	Convey("Should return map", t, func() {
		Convey("With no reports", func() {
			reports := []entities.PresubmitCoverageData{}
			have := client.getMaxPatchsetToChangeMap(reports)
			want := map[int64]int64{}
			So(have, ShouldResemble, want)
		})

		Convey("With some reports", func() {
			mockRep := getMockPresubmitData()
			rep1 := *mockRep[0]
			rep2 := rep1
			rep2.Change = 2
			rep2.Patchset = 1
			rep3 := rep1
			rep3.Change = 1
			rep3.Patchset = 3
			reports := []entities.PresubmitCoverageData{rep1, rep2, rep3}
			have := client.getMaxPatchsetToChangeMap(reports)
			want := map[int64]int64{1: 3, 2: 1}
			So(have, ShouldResemble, want)
		})
	})
}

func TestGetDir(t *testing.T) {
	t.Parallel()

	Convey("Should return parent directory", t, func() {
		Convey("For a directory path", func() {
			parent := getDir("//a/b/")
			So(parent, ShouldEqual, "//a/")

			parent = getDir("//a/")
			So(parent, ShouldEqual, "//")
		})

		Convey("For a file path", func() {
			parent := getDir("//a/b/c.ext")
			So(parent, ShouldEqual, "//a/b/")
		})

		Convey("For root path", func() {
			parent := getDir("//")
			So(parent, ShouldEqual, "//")
		})
	})
}

func TestCreateCqSummaryDatat(t *testing.T) {
	t.Parallel()
	client := CronClient{}

	Convey("Create CQ summary coverage data", t, func() {
		Convey("Should pass", func() {
			mockDataClient := mocks.NewIDataClient(t)
			mockDataClient.On(
				"BatchPut",
				mock.AnythingOfType("backgroundCtx"),
				mock.Anything,
				mock.Anything,
			).Return(
				func(c context.Context, entities interface{}, keys interface{}) error {
					return nil
				},
			)
			client.coverageV2DsClient = mockDataClient

			mockData := map[string]IncrementalCoverageData{
				"//":   {CoveredFiles: 2, TotalFiles: 3, IsDir: true},
				"//a/": {CoveredFiles: 1, TotalFiles: 1, IsDir: true},
				"//b/": {CoveredFiles: 1, TotalFiles: 2, IsDir: true},
			}
			err := client.createCqSummaryData(context.Background(), time.Now(), int64(1234), int64(1), false, mockData)
			So(err, ShouldBeNil)
		})

		Convey("Should fail", func() {
			mockDataClient := mocks.NewIDataClient(t)
			mockDataClient.On(
				"BatchPut",
				mock.AnythingOfType("backgroundCtx"),
				mock.Anything,
				mock.Anything,
			).Return(
				func(c context.Context, entities interface{}, keys interface{}) error {
					return fmt.Errorf("Datastore: %s", "Error putting entities")
				},
			)
			client.coverageV2DsClient = mockDataClient
			err := client.createCqSummaryData(context.Background(), time.Now(), int64(1234), int64(1), false, map[string]IncrementalCoverageData{})
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Datastore: Error putting entities"))
		})
	})
}
