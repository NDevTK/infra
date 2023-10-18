// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package coverage

import (
	"context"
	"errors"
	"reflect"
	"testing"

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

func TestGetProjectConfig(t *testing.T) {
	t.Parallel()
	Convey(`Should have valid "FinditConfig" entity`, t, func() {
		client := Client{}
		ctx := context.Background()
		Convey(`Invalid "CodeCoverageSettings" JSON`, func() {
			fakeFinditConfig := entities.FinditConfig{
				CodeCoverageSettings: []byte(""),
			}
			config, err := client.getProjectConfig(ctx, &fakeFinditConfig, "chromium")
			So(config, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
		})
		Convey(`Missing "default_postsubmit_report_config" property`, func() {
			fakeFinditConfig := entities.FinditConfig{
				CodeCoverageSettings: []byte("{}"),
			}
			config, err := client.getProjectConfig(ctx, &fakeFinditConfig, "chromium")
			So(config, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
		})
		Convey(`Missing project from "default_postsubmit_report_config" property`, func() {
			fakeFinditConfig := getMockFinditConfigWithoutAnyProject()
			config, err := client.getProjectConfig(ctx, fakeFinditConfig, "chromium")
			So(config, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
		})
		Convey(`Valid "FinditConfig" entity`, func() {
			fakeFinditConfig := getMockFinditConfig()
			var config *api.GetProjectDefaultConfigResponse
			config, err := client.getProjectConfig(ctx, fakeFinditConfig, "chromium")
			So(config, ShouldResemble, &api.GetProjectDefaultConfigResponse{
				Host:     "chromium.googlesource.com",
				Platform: "linux",
				Project:  "chromium/src",
				Ref:      "refs/heads/main",
			})
			So(err, ShouldBeNil)
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
		client.coverageV1DsClient = mockDataClient

		req := api.GetProjectDefaultConfigRequest{
			Project: "chromium",
		}
		res, err := client.GetProjectDefaultConfig(ctx, &req)
		So(err, ShouldBeNil)
		So(res.Host, ShouldEqual, "chromium.googlesource.com")
		So(res.Platform, ShouldEqual, "linux")
		So(res.Project, ShouldEqual, "chromium/src")
		So(res.Ref, ShouldEqual, "refs/heads/main")
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
			DataType:        "dirs",
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
