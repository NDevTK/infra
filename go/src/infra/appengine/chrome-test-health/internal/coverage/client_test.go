// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package coverage

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	. "github.com/smartystreets/goconvey/convey"

	"infra/appengine/chrome-test-health/api"
	"infra/appengine/chrome-test-health/datastorage"
	"infra/appengine/chrome-test-health/datastorage/mocks"
	"infra/appengine/chrome-test-health/internal/coverage/entities"

	mock "github.com/stretchr/testify/mock"
)

var (
	ErrInsufficientArgs = errors.New("insufficent arguments")
	ErrConnection       = errors.New("connection error")
	ErrInvalidKey       = errors.New("invalid key")
	ErrInvalidType      = errors.New("invalid type")
	ErrEntityNotFound   = errors.New("entity not found")
	ErrInternal         = errors.New("internal error")
)

func getMockFinditConfigRoot() *entities.FinditConfigRoot {
	return &entities.FinditConfigRoot{
		Key:     datastore.IDKey("FinditConfigRoot", 1, nil),
		Current: 218,
	}
}

func getMockFinditConfig() *entities.FinditConfig {
	mockSettingsData := []byte(`
{
	"default_postsubmit_report_config": {
		"chromium": {
			"project": "chromium/src",
			"platform": "linux",
			"host": "chromium.googlesource.com",
			"ref": "refs/heads/main"
		}
	},
	"postsubmit_platform_info_map": {
		"chromium": {
			"linux": {
				"bucket": "ci",
				"builder": "linux-code-coverage",
				"coverage_tool": "clang",
				"ui_name": "Linux (C/C++)"
			},
			"android-java": {
				"bucket": "ci",
				"builder": "android-code-coverage",
				"coverage_tool": "jacoco",
				"ui_name": "Android (Java)"
			}
		}
	}
}`)

	parent := getMockFinditConfigRoot()
	return &entities.FinditConfig{
		Key:                  datastore.IDKey("FinditConfig", int64(parent.Current), parent.Key),
		CodeCoverageSettings: mockSettingsData,
	}
}

func getMockFinditConfigWithoutAnyProject() *entities.FinditConfig {
	fakeSettingsData := []byte(`
{
	"default_postsubmit_report_config": {}
}`)

	return &entities.FinditConfig{
		CodeCoverageSettings: fakeSettingsData,
	}
}

func getMockPostsubmitReport() []*entities.PostsubmitReport {
	return []*entities.PostsubmitReport{
		{
			GitilesCommitProject:    "chromium/src",
			GitilesCommitServerHost: "chromium.googlesource.com",
			Bucket:                  "ci",
			Builder:                 "linux-code-coverage",
			GitilesCommitRevision:   "12345",
		},
		{
			GitilesCommitProject:    "chromium/src",
			GitilesCommitServerHost: "chromium.googlesource.com",
			Bucket:                  "ci",
			Builder:                 "andr-code-coverage",
			GitilesCommitRevision:   "23456",
		},
		{
			GitilesCommitProject:    "chromium/src",
			GitilesCommitServerHost: "chromium.googlesource.com",
			Bucket:                  "ci",
			Builder:                 "linux-code-coverage",
			GitilesCommitRevision:   "03d4e64771cbc97f3ca5e4bbe85490d7cf909a0a",
			CommitTimestamp:         time.Date(2023, 11, 17, 20, 34, 58, 0, time.UTC),
		},
	}
}

func getMockSummaryData() *entities.SummaryCoverageData {
	mockKey := "chromium.googlesource.com$chromium/src$refs/heads/main" +
		"$03d4e64771cbc97f3ca5e4bbe85490d7cf909a0a$dirs$//$ci$linux-code-coverage$0"
	mockSummaryData, _ := compressString(`
{
	"dirs": [
		{
			"name": "a/",
			"path": "//a/",
			"summaries": [
				{
					"covered": 59,
					"name": "line",
					"total": 200
				}
			]
		}
	],
	"files": [
		{
			"name": "file.cc",
			"path": "//file.cc",
			"summaries": [
				{
					"covered": 64,
					"name": "line",
					"total": 100
				}
			]
		}
	],
	"path": "//",
	"summaries": [
		{
			"covered": 123,
			"name": "line",
			"total": 300
		}
	]
}`)

	return &entities.SummaryCoverageData{
		Key:  datastore.NameKey("SummaryCoverageData", mockKey, nil),
		Data: mockSummaryData,
	}
}

func getMockSummaryDataByComponent() []*entities.SummaryCoverageData {
	res := []*entities.SummaryCoverageData{}

	mockKey := "chromium.googlesource.com$chromium/src$refs/heads/main" +
		"$03d4e64771cbc97f3ca5e4bbe85490d7cf909a0a$components$C1$ci$linux-code-coverage$0"
	mockSummaryData, _ := compressString(`
{
	"dirs": [],
	"files": [],
	"path": "C1",
	"summaries": [
		{
			"covered": 59,
			"name": "line",
			"total": 200
		}
	]
}`)
	res = append(res, &entities.SummaryCoverageData{
		Key:  datastore.NameKey("SummaryCoverageData", mockKey, nil),
		Data: mockSummaryData,
	})

	mockKey = "chromium.googlesource.com$chromium/src$refs/heads/main" +
		"$03d4e64771cbc97f3ca5e4bbe85490d7cf909a0a$components$C2>C3$ci$linux-code-coverage$0"
	mockSummaryData, _ = compressString(`
{
	"dirs": [],
	"files": [],
	"path": "C2>C3",
	"summaries": []
}`)
	res = append(res, &entities.SummaryCoverageData{
		Key:  datastore.NameKey("SummaryCoverageData", mockKey, nil),
		Data: mockSummaryData,
	})

	return res
}

func TestGetProjectConfig(t *testing.T) {
	t.Parallel()
	Convey(`Should have valid "FinditConfig" entity`, t, func() {
		client := Client{}
		ctx := context.Background()
		Convey(`Invalid "CodeCoverageSettings" JSON`, func() {
			fakeFinditConfig := entities.FinditConfig{
				CodeCoverageSettings: []byte(""),
			}
			config := &api.GetProjectDefaultConfigResponse{}
			err := client.getProjectConfig(ctx, &fakeFinditConfig, "chromium", config)
			So(config, ShouldResemble, &api.GetProjectDefaultConfigResponse{})
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
		})
		Convey(`Missing "default_postsubmit_report_config" property`, func() {
			fakeFinditConfig := entities.FinditConfig{
				CodeCoverageSettings: []byte("{}"),
			}
			config := &api.GetProjectDefaultConfigResponse{}
			err := client.getProjectConfig(ctx, &fakeFinditConfig, "chromium", config)
			So(config, ShouldResemble, &api.GetProjectDefaultConfigResponse{})
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
		})
		Convey(`Missing project from "default_postsubmit_report_config" property`, func() {
			fakeFinditConfig := getMockFinditConfigWithoutAnyProject()
			config := &api.GetProjectDefaultConfigResponse{}
			err := client.getProjectConfig(ctx, fakeFinditConfig, "chromium", config)
			So(config, ShouldResemble, &api.GetProjectDefaultConfigResponse{})
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
		})
		Convey(`Valid "FinditConfig" entity`, func() {
			fakeFinditConfig := getMockFinditConfig()
			config := &api.GetProjectDefaultConfigResponse{}
			err := client.getProjectConfig(ctx, fakeFinditConfig, "chromium", config)
			So(config, ShouldResemble, &api.GetProjectDefaultConfigResponse{
				GitilesHost:    "chromium.googlesource.com",
				GitilesProject: "chromium/src",
				GitilesRef:     "refs/heads/main",
			})
			So(err, ShouldBeNil)
		})
	})
}

func TestGetBuilderOptions(t *testing.T) {
	Convey(`Should get builder configurations`, t, func() {
		client := Client{}
		ctx := context.Background()

		Convey(`Invalid "CodeCoverageSettings" JSON`, func() {
			mockFinditConfig := &entities.FinditConfig{
				CodeCoverageSettings: []byte(""),
			}
			config := &api.GetProjectDefaultConfigResponse{}
			err := client.getBuilderOptions(ctx, "chromium", "chromium.googlesource.com",
				"chromium", mockFinditConfig, config)
			So(config, ShouldResemble, &api.GetProjectDefaultConfigResponse{})
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
		})

		Convey(`Missing "postsubmit_platform_info_map" property`, func() {
			mockFinditConfig := &entities.FinditConfig{
				CodeCoverageSettings: []byte("{}"),
			}
			config := &api.GetProjectDefaultConfigResponse{}
			err := client.getBuilderOptions(ctx, "chromium", "chromium.googlesource.com",
				"chromium", mockFinditConfig, config)
			So(config, ShouldResemble, &api.GetProjectDefaultConfigResponse{})
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
		})

		Convey(`FinditConfig has platform options`, func() {
			postsubmitReports := getMockPostsubmitReport()
			mockDataClient := mocks.NewIDataClient(t)
			mockDataClient.On(
				"Query",
				mock.AnythingOfType("backgroundCtx"),
				mock.Anything,
				"PostsubmitReport",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				func(c context.Context, result interface{}, dataType string, queryFilters []datastorage.QueryFilter, order interface{}, limit int, options ...interface{}) error {
					for _, rep := range postsubmitReports {
						if queryFilters[2].Value == rep.Bucket && queryFilters[3].Value == rep.Builder {
							res := reflect.ValueOf(result).Elem()
							res.Set(reflect.Append(res, reflect.ValueOf(rep).Elem()))
							return nil
						}
					}
					return nil
				},
			)
			client.coverageV1DsClient = mockDataClient

			finditConfig := getMockFinditConfig()
			config := &api.GetProjectDefaultConfigResponse{}
			err := client.getBuilderOptions(ctx, "chromium", "chromium.googlesource.com",
				"chromium/src", finditConfig, config)
			So(err, ShouldBeNil)
			So(config.BuilderConfig, ShouldHaveLength, 1)
		})
	})

}

func TestGetModifiedBuilder(t *testing.T) {
	t.Parallel()
	Convey(`Should be able to modify builder based on field unitTestsOnly`, t, func() {
		client := Client{}
		Convey(`Field unitTestsOnly is set to true`, func() {
			unitTestsOnly := true
			modifiedBuilder := client.getModifedBuilder("builder", &unitTestsOnly)
			So(modifiedBuilder, ShouldEqual, "builder_unit")
		})
		Convey(`Field unitTestsOnly is set to false`, func() {
			unitTestsOnly := false
			modifiedBuilder := client.getModifedBuilder("builder", &unitTestsOnly)
			So(modifiedBuilder, ShouldEqual, "builder")
		})
		Convey(`Field unitTestsOnly is not provided`, func() {
			modifiedBuilder := client.getModifedBuilder("builder", nil)
			So(modifiedBuilder, ShouldEqual, "builder")
		})
	})
}

func TestGetProjectDefaultConfig(t *testing.T) {
	t.Parallel()
	Convey(`Should get project's default configuration`, t, func() {
		client := Client{}
		ctx := context.Background()

		finditConfigRoot := getMockFinditConfigRoot()
		finditConfig := getMockFinditConfig()
		postsubmitReports := getMockPostsubmitReport()

		mockDataClient := mocks.NewIDataClient(t)
		mockDataClient.On(
			"Query",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"FinditConfigRoot",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(c context.Context, result interface{}, dataType string, queryFilters []datastorage.QueryFilter, order interface{}, limit int, options ...interface{}) error {
				res := reflect.ValueOf(result).Elem()
				res.Set(reflect.Append(res, reflect.ValueOf(finditConfigRoot).Elem()))
				return nil
			},
		)

		mockDataClient.On(
			"Get",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"FinditConfig",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(ctx context.Context, result interface{}, dataType string, key interface{}, options ...interface{}) error {
				res := reflect.ValueOf(result).Elem()
				res.Set(reflect.ValueOf(finditConfig).Elem())
				return nil
			},
		)

		mockDataClient.On(
			"Query",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"PostsubmitReport",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(c context.Context, result interface{}, dataType string, queryFilters []datastorage.QueryFilter, order interface{}, limit int, options ...interface{}) error {
				for _, rep := range postsubmitReports {
					if queryFilters[2].Value == rep.Bucket && queryFilters[3].Value == rep.Builder {
						res := reflect.ValueOf(result).Elem()
						res.Set(reflect.Append(res, reflect.ValueOf(rep).Elem()))
						return nil
					}
				}
				return nil
			},
		)
		client.coverageV1DsClient = mockDataClient

		req := api.GetProjectDefaultConfigRequest{
			LuciProject: "chromium",
		}
		res, err := client.GetProjectDefaultConfig(ctx, &req)
		So(err, ShouldBeNil)
		So(res.GitilesHost, ShouldEqual, "chromium.googlesource.com")
		So(res.GitilesProject, ShouldEqual, "chromium/src")
		So(res.GitilesRef, ShouldEqual, "refs/heads/main")
		So(res.BuilderConfig, ShouldHaveLength, 1)
	})
}

func TestGetCoverageSummary(t *testing.T) {
	t.Parallel()
	Convey(`Should get summary data`, t, func() {
		client := Client{}
		ctx := context.Background()

		summaryData := getMockSummaryData()

		mockDataClient := mocks.NewIDataClient(t)
		mockDataClient.On(
			"Get",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"SummaryCoverageData",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(ctx context.Context, result interface{}, dataType string, key interface{}, options ...interface{}) error {
				if key.(string) != summaryData.Key.Name {
					return ErrEntityNotFound
				}

				res := reflect.ValueOf(result).Elem()
				res.Set(reflect.ValueOf(summaryData).Elem())
				return nil
			},
		)
		client.coverageV1DsClient = mockDataClient

		req := api.GetCoverageSummaryRequest{
			GitilesHost:     "chromium.googlesource.com",
			GitilesProject:  "chromium/src",
			GitilesRef:      "refs/heads/main",
			GitilesRevision: "03d4e64771cbc97f3ca5e4bbe85490d7cf909a0a",
			Path:            "//",
			UnitTestsOnly:   false,
			Bucket:          "ci",
			Builder:         "linux-code-coverage",
		}
		Convey(`with valid params`, func() {
			res, err := client.GetCoverageSummary(ctx, &req)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.Summary, ShouldNotBeEmpty)
		})
		Convey(`with no matching entity in datastore`, func() {
			req.Bucket = "random"
			res, err := client.GetCoverageSummary(ctx, &req)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
			So(res, ShouldBeNil)
		})
		Convey(`with malformed data`, func() {
			summaryData.Data, _ = compressString("{")
			mockDataClient.On(
				"Get",
				mock.AnythingOfType("backgroundCtx"),
				mock.Anything,
				"SummaryCoverageData",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				func(ctx context.Context, result interface{}, dataType string, key interface{}, options ...interface{}) error {
					if key.(string) != summaryData.Key.Name {
						return ErrEntityNotFound
					}

					res := reflect.ValueOf(result).Elem()
					res.Set(reflect.ValueOf(summaryData).Elem())
					return nil
				},
			)

			res, err := client.GetCoverageSummary(ctx, &req)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
			So(res, ShouldBeNil)
		})
	})
}

func TestGetCoverageSummaryForComponents(t *testing.T) {
	t.Parallel()
	Convey(`Should get summary data by components`, t, func() {
		client := Client{}
		ctx := context.Background()

		summaryData := getMockSummaryDataByComponent()

		mockDataClient := mocks.NewIDataClient(t)
		mockDataClient.On(
			"Get",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"SummaryCoverageData",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(ctx context.Context, result interface{}, dataType string, key interface{}, options ...interface{}) error {
				for _, sum := range summaryData {
					if key.(string) == sum.Key.Name {
						res := reflect.ValueOf(result).Elem()
						res.Set(reflect.ValueOf(sum).Elem())
						return nil
					}
				}
				return nil
			},
		)
		client.coverageV1DsClient = mockDataClient

		req := api.GetCoverageSummaryRequest{
			GitilesHost:     "chromium.googlesource.com",
			GitilesProject:  "chromium/src",
			GitilesRef:      "refs/heads/main",
			GitilesRevision: "03d4e64771cbc97f3ca5e4bbe85490d7cf909a0a",
			Components:      []string{"C1", "C2>C3"},
			UnitTestsOnly:   false,
			Bucket:          "ci",
			Builder:         "linux-code-coverage",
		}

		res, err := client.GetCoverageSummary(ctx, &req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		So(res.Summary, ShouldHaveLength, 2)
	})
}

func TestGetCoverageReportsForLastYear(t *testing.T) {
	t.Parallel()
	client := Client{}
	ctx := context.Background()

	Convey("Should return reports", t, func() {
		postsubmitReports := getMockPostsubmitReport()
		mockDataClient := mocks.NewIDataClient(t)
		mockDataClient.On(
			"Query",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"PostsubmitReport",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(c context.Context, result interface{}, dataType string, queryFilters []datastorage.QueryFilter, order interface{}, limit int, options ...interface{}) error {
				for _, rep := range postsubmitReports {
					if queryFilters[2].Value == rep.Bucket && queryFilters[3].Value == rep.Builder {
						res := reflect.ValueOf(result).Elem()
						res.Set(reflect.Append(res, reflect.ValueOf(rep).Elem()))
						return nil
					}
				}
				return nil
			},
		)
		client.coverageV1DsClient = mockDataClient

		reports, err := client.getCoverageReportsForLastYear(ctx, "ci", "linux-code-coverage")
		So(err, ShouldBeNil)
		So(reports, ShouldHaveLength, 1)
		expectedReports := []entities.PostsubmitReport{
			{
				GitilesCommitProject:    "chromium/src",
				GitilesCommitServerHost: "chromium.googlesource.com",
				Bucket:                  "ci",
				Builder:                 "linux-code-coverage",
				GitilesCommitRevision:   "12345",
			},
		}
		So(reports, ShouldResemble, expectedReports)
	})

	Convey("Should error out with no matching index message", t, func() {
		mockDataClient := mocks.NewIDataClient(t)
		mockDataClient.On(
			"Query",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"PostsubmitReport",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(c context.Context, result interface{}, dataType string, queryFilters []datastorage.QueryFilter, order interface{}, limit int, options ...interface{}) error {
				return fmt.Errorf("PostsubmitReport: %s", "No matching indexes found")
			},
		)
		client.coverageV1DsClient = mockDataClient

		reports, err := client.getCoverageReportsForLastYear(ctx, "ci", "linux-code-coverage")
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrInternalServerError)
		So(reports, ShouldBeNil)
	})
}

func TestGetCoverageNumbersForPath(t *testing.T) {
	t.Parallel()
	client := Client{}
	ctx := context.Background()

	Convey("Should return coverage numbers per day for the path", t, func() {
		summaryData := getMockSummaryData()
		mockDataClient := mocks.NewIDataClient(t)
		mockDataClient.On(
			"Get",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"SummaryCoverageData",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(ctx context.Context, result interface{}, dataType string, key interface{}, options ...interface{}) error {
				if key.(string) != summaryData.Key.Name {
					return ErrEntityNotFound
				}

				res := reflect.ValueOf(result).Elem()
				res.Set(reflect.ValueOf(summaryData).Elem())
				return nil
			},
		)
		client.coverageV1DsClient = mockDataClient

		reports := []entities.PostsubmitReport{*getMockPostsubmitReport()[0]}
		t := time.Date(2009, 11, 17, 20, 34, 58, 0, time.UTC)
		reports[0].GitilesCommitRevision = "03d4e64771cbc97f3ca5e4bbe85490d7cf909a0a"
		reports[0].CommitTimestamp = t

		data := client.getCoverageNumbersForPath(ctx, reports, "//", "ci", "linux-code-coverage")
		So(data, ShouldHaveLength, 1)
		expectedData := []CoveragePerDate{
			{
				date:    "2009-11-17",
				covered: 123,
				total:   300,
			},
		}
		So(data, ShouldResemble, expectedData)
	})
}

func TestGetCoverageNumbersForComponent(t *testing.T) {
	t.Parallel()
	client := Client{}
	ctx := context.Background()

	Convey("Should return coverage numbers per day for the component", t, func() {
		summaryData := getMockSummaryDataByComponent()
		mockDataClient := mocks.NewIDataClient(t)
		mockDataClient.On(
			"Get",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"SummaryCoverageData",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(ctx context.Context, result interface{}, dataType string, key interface{}, options ...interface{}) error {
				for _, sum := range summaryData {
					if key.(string) == sum.Key.Name {
						res := reflect.ValueOf(result).Elem()
						res.Set(reflect.ValueOf(sum).Elem())
						return nil
					}
				}
				return nil
			},
		)
		client.coverageV1DsClient = mockDataClient

		reports := []entities.PostsubmitReport{*getMockPostsubmitReport()[0]}
		t := time.Date(2009, 11, 17, 20, 34, 58, 0, time.UTC)
		reports[0].GitilesCommitRevision = "03d4e64771cbc97f3ca5e4bbe85490d7cf909a0a"
		reports[0].CommitTimestamp = t

		data := client.getCoverageNumbersForComponent(ctx, reports, "C1", "ci", "linux-code-coverage")
		So(data, ShouldHaveLength, 1)
		expectedData := []CoveragePerDate{
			{
				date:    "2009-11-17",
				covered: 59,
				total:   200,
			},
		}
		So(data, ShouldResemble, expectedData)
	})
}

func TestGetAbsoluteCoverageDataOneYear(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := Client{
		coverageV1DsClient: mocks.NewIDataClient(t),
	}

	Convey("Should pass", t, func() {
		postsubmitReports := getMockPostsubmitReport()
		summaryData := getMockSummaryData()
		summaryDataByComp := getMockSummaryDataByComponent()
		mockDataClient := mocks.NewIDataClient(t)
		mockDataClient.On(
			"Query",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"PostsubmitReport",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(c context.Context, result interface{}, dataType string, queryFilters []datastorage.QueryFilter, order interface{}, limit int, options ...interface{}) error {
				for _, rep := range postsubmitReports {
					if queryFilters[2].Value == rep.Bucket &&
						queryFilters[3].Value == rep.Builder &&
						rep.CommitTimestamp.After(queryFilters[6].Value.(time.Time)) {
						res := reflect.ValueOf(result).Elem()
						res.Set(reflect.Append(res, reflect.ValueOf(rep).Elem()))
						return nil
					}
				}
				return nil
			},
		)

		mockDataClient.On(
			"Get",
			mock.AnythingOfType("backgroundCtx"),
			mock.Anything,
			"SummaryCoverageData",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(
			func(ctx context.Context, result interface{}, dataType string, key interface{}, options ...interface{}) error {
				if strings.Contains(key.(string), "dirs") {
					if key.(string) != summaryData.Key.Name {
						return ErrEntityNotFound
					}

					res := reflect.ValueOf(result).Elem()
					res.Set(reflect.ValueOf(summaryData).Elem())
					return nil
				} else {
					for _, sum := range summaryDataByComp {
						if key.(string) == sum.Key.Name {
							res := reflect.ValueOf(result).Elem()
							res.Set(reflect.ValueOf(sum).Elem())
							return nil
						}
					}
					return nil
				}
			},
		)

		client.coverageV1DsClient = mockDataClient

		Convey("Valid", func() {
			req := &api.GetAbsoluteCoverageDataOneYearRequest{
				Paths:         []string{"//"},
				Components:    []string{"C1", "C2>C3"},
				UnitTestsOnly: false,
				Bucket:        "ci",
				Builder:       "linux-code-coverage",
			}
			res, err := client.GetAbsoluteCoverageDataOneYear(ctx, req)
			So(err, ShouldBeNil)
			So(len(res.Reports), ShouldBeGreaterThan, 0)
			expectedRes := &api.GetAbsoluteCoverageDataOneYearResponse{
				Reports: []*api.AbsoluteCoverage{{Date: "2023-11-17", LinesCovered: 182, TotalLines: 500}},
			}
			So(res, ShouldResemble, expectedRes)
		})
		Convey("No components with some paths", func() {
			req := &api.GetAbsoluteCoverageDataOneYearRequest{
				Paths:         []string{"//"},
				Components:    []string{},
				UnitTestsOnly: false,
				Bucket:        "ci",
				Builder:       "linux-code-coverage",
			}
			res, err := client.GetAbsoluteCoverageDataOneYear(ctx, req)
			So(err, ShouldBeNil)
			So(len(res.Reports), ShouldBeGreaterThan, 0)
			expectedRes := &api.GetAbsoluteCoverageDataOneYearResponse{
				Reports: []*api.AbsoluteCoverage{{Date: "2023-11-17", LinesCovered: 123, TotalLines: 300}},
			}
			So(res, ShouldResemble, expectedRes)
		})
		Convey("No paths with some components", func() {
			req := &api.GetAbsoluteCoverageDataOneYearRequest{
				Paths:         []string{},
				Components:    []string{"C1", "C2>C3"},
				UnitTestsOnly: false,
				Bucket:        "ci",
				Builder:       "linux-code-coverage",
			}
			res, err := client.GetAbsoluteCoverageDataOneYear(ctx, req)
			So(err, ShouldBeNil)
			So(len(res.Reports), ShouldBeGreaterThan, 0)
			expectedRes := &api.GetAbsoluteCoverageDataOneYearResponse{
				Reports: []*api.AbsoluteCoverage{{Date: "2023-11-17", LinesCovered: 59, TotalLines: 200}},
			}
			So(res, ShouldResemble, expectedRes)
		})
	})

	Convey("Should fail", t, func() {
		Convey("PostsubmitReport fetch error", func() {
			mockDataClient := mocks.NewIDataClient(t)
			mockDataClient.On(
				"Query",
				mock.AnythingOfType("backgroundCtx"),
				mock.Anything,
				"PostsubmitReport",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				func(c context.Context, result interface{}, dataType string, queryFilters []datastorage.QueryFilter, order interface{}, limit int, options ...interface{}) error {
					return fmt.Errorf("PostsubmitReport: %s", "entity not found")
				},
			)
			client.coverageV1DsClient = mockDataClient
			req := &api.GetAbsoluteCoverageDataOneYearRequest{
				Paths:         []string{"//"},
				Components:    []string{},
				UnitTestsOnly: false,
				Bucket:        "ci",
				Builder:       "linux-code-coverage",
			}
			res, err := client.GetAbsoluteCoverageDataOneYear(ctx, req)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
			So(res, ShouldBeNil)
		})
	})
}

func TestAggregateCoverageReports(t *testing.T) {
	t.Parallel()
	client := Client{}

	Convey("Should aggregate coverage numbers into the supplied map", t, func() {
		Convey("When data is complete", func() {
			existingMap := make(map[string]map[string]int64)
			existingMap["2023-11-20"] = map[string]int64{
				"covered": 123,
				"total":   200,
			}
			existingMap["2023-11-21"] = map[string]int64{
				"covered": 150,
				"total":   200,
			}

			data := []CoveragePerDate{
				{date: "2023-11-20", covered: 10, total: 20},
				{date: "2023-11-21", covered: 20, total: 30},
			}
			existingMap = client.aggregateCoverageReports(existingMap, data)

			So(existingMap["2023-11-20"], ShouldResemble, map[string]int64{"covered": 133, "total": 220})
			So(existingMap["2023-11-21"], ShouldResemble, map[string]int64{"covered": 170, "total": 230})
		})

		Convey("When data has missing dates", func() {
			existingMap := make(map[string]map[string]int64)
			existingMap["2023-11-20"] = map[string]int64{
				"covered": 123,
				"total":   200,
			}
			existingMap["2023-11-21"] = map[string]int64{
				"covered": 150,
				"total":   200,
			}

			data := []CoveragePerDate{
				{date: "2023-11-20", covered: 10, total: 20},
			}
			existingMap = client.aggregateCoverageReports(existingMap, data)

			So(existingMap["2023-11-20"], ShouldResemble, map[string]int64{"covered": 133, "total": 220})
			So(existingMap["2023-11-21"], ShouldResemble, map[string]int64{"covered": 150, "total": 200})
		})
	})
}
