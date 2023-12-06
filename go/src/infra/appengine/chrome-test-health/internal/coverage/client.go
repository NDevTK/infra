// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package coverage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"go.chromium.org/luci/common/logging"

	"infra/appengine/chrome-test-health/api"
	"infra/appengine/chrome-test-health/datastorage"
	"infra/appengine/chrome-test-health/internal/coverage/entities"
)

var (
	ErrInternalServerError = errors.New("Internal Server Error")
)

const (
	chromiumProject    = "chromium/src"
	chromiumServerHost = "chromium.googlesource.com"
	chromiumRef        = "refs/heads/main"
)

type Client struct {
	// Refers to findit's cloud project
	FinditCloudProject string
	// Refers to chrome-test-health's cloud project
	ChromeTestHealthCloudProject string
	// References to findit-for-me's datastore
	coverageV1DsClient datastorage.IDataClient
	// References to chrome-test-health's datastore
	coverageV2DsClient datastorage.IDataClient
}

func (c *Client) Init(ctx context.Context) error {
	covV1DsClient, err := datastorage.NewDataStoreClient(ctx, c.FinditCloudProject)
	if err != nil {
		logging.Errorf(ctx, "Error connecting to %s", c.FinditCloudProject)
		return ErrInternalServerError
	}
	c.coverageV1DsClient = covV1DsClient

	covV2DsClient, err := datastorage.NewDataStoreClient(ctx, c.ChromeTestHealthCloudProject)
	if err != nil {
		logging.Errorf(ctx, "Error connecting to %s", c.ChromeTestHealthCloudProject)
		return ErrInternalServerError
	}
	c.coverageV2DsClient = covV2DsClient

	return nil
}

type CoveragePerDate struct {
	date    string
	covered float64
	total   float64
}

// getProjectConfig extracts out the code coverage default settings from the
// fetched FinditConfig entity for the given project.
func (c *Client) getProjectConfig(
	ctx context.Context,
	finditConfig *entities.FinditConfig,
	project string,
	config *api.GetProjectDefaultConfigResponse,
) error {
	code_cov_settings := map[string]interface{}{}
	err := json.Unmarshal(finditConfig.CodeCoverageSettings, &code_cov_settings)
	if err != nil {
		logging.Errorf(ctx, "Failed to unmarshall CodeCoverageSettings: %s", err)
		return ErrInternalServerError
	}

	defaultPostSubmitConfig := code_cov_settings["default_postsubmit_report_config"]
	if defaultPostSubmitConfig == nil {
		logging.Errorf(ctx, "Missing default_postsubmit_report_config from FinditConfig")
		return ErrInternalServerError
	}

	projectConfig := defaultPostSubmitConfig.(map[string]interface{})[project]
	if projectConfig == nil {
		logging.Errorf(ctx, "Missing config for project %s", project)
		return ErrInternalServerError
	}
	projectConfig = projectConfig.(map[string]interface{})

	config.GitilesHost = projectConfig.(map[string]interface{})["host"].(string)
	config.GitilesProject = projectConfig.(map[string]interface{})["project"].(string)
	config.GitilesRef = projectConfig.(map[string]interface{})["ref"].(string)

	return nil
}

func (c *Client) getBuilderOptions(
	ctx context.Context,
	luciProject string,
	host string,
	project string,
	finditConfig *entities.FinditConfig,
	config *api.GetProjectDefaultConfigResponse,
) error {
	code_cov_settings := map[string]interface{}{}
	err := json.Unmarshal(finditConfig.CodeCoverageSettings, &code_cov_settings)
	if err != nil {
		logging.Errorf(ctx, "Failed to unmarshall CodeCoverageSettings: %s", err)
		return ErrInternalServerError
	}

	postsubmitPlatformInfoMap := code_cov_settings["postsubmit_platform_info_map"]
	if postsubmitPlatformInfoMap == nil {
		logging.Errorf(ctx, "Missing postsubmit_platform_info_map from FinditConfig")
		return ErrInternalServerError
	}

	platformListForProject := postsubmitPlatformInfoMap.(map[string]interface{})[luciProject]

	var builderConfigDetails []*api.BuilderConfig
	for platform, builderConfigDetail := range platformListForProject.(map[string]interface{}) {
		bucket := builderConfigDetail.(map[string]interface{})["bucket"].(string)
		builder := builderConfigDetail.(map[string]interface{})["builder"].(string)
		postsubmitReport := entities.PostsubmitReport{}
		err := postsubmitReport.Filter(ctx, c.coverageV1DsClient, project, host, bucket, builder)
		if err != nil {
			continue
		}

		builderConfigDetails = append(builderConfigDetails, &api.BuilderConfig{
			Platform:       platform,
			Bucket:         bucket,
			Builder:        builder,
			UiName:         builderConfigDetail.(map[string]interface{})["ui_name"].(string),
			LatestRevision: postsubmitReport.GitilesCommitRevision,
		})
	}

	config.BuilderConfig = builderConfigDetails
	return nil
}

func (c *Client) getModifedBuilder(builder string, unitTestsOnly *bool) string {
	if unitTestsOnly == nil || !(*unitTestsOnly) {
		return builder
	}
	return fmt.Sprintf("%s_unit", builder)
}

// GetProjectDefaultConfig fetches the latest version of FinditConfig from the
// datastore and returns the desired configuration extracted from the entity.
func (c *Client) GetProjectDefaultConfig(
	ctx context.Context,
	req *api.GetProjectDefaultConfigRequest,
) (*api.GetProjectDefaultConfigResponse, error) {
	project := req.LuciProject
	finditConfig := &entities.FinditConfig{}
	if err := finditConfig.Get(ctx, c.coverageV1DsClient); err != nil {
		logging.Errorf(ctx, "%s", err.Error())
		return nil, ErrInternalServerError
	}

	response := &api.GetProjectDefaultConfigResponse{}
	if err := c.getProjectConfig(ctx, finditConfig, project, response); err != nil {
		return nil, err
	}

	if err := c.getBuilderOptions(ctx, project, response.GitilesHost,
		response.GitilesProject, finditConfig, response); err != nil {
		return nil, err
	}

	return response, nil
}

// GetCoverageSummary fetches the code coverage metrics/percentages for the specified
// configuration including the path or component list.
// The path param here can be a dir/file path like //foo/foo1/foo2/.
// Components param should be be a list of monorail components like ["C1>C2", "C3"]
// This endpoint accepts either path or component not both.
func (c *Client) GetCoverageSummary(ctx context.Context, req *api.GetCoverageSummaryRequest) (*api.GetCoverageSummaryResponse, error) {
	host := req.GitilesHost
	project := req.GitilesProject
	ref := req.GitilesRef
	revision := req.GitilesRevision
	path := req.Path
	components := req.Components
	unitTestsOnly := req.UnitTestsOnly
	bucket := req.Bucket
	builder := c.getModifedBuilder(req.Builder, &unitTestsOnly)

	var dataType string
	var nodes []string
	if path != "" {
		dataType = "dirs"
		nodes = append(nodes, path)
	} else {
		dataType = "components"
		nodes = components
	}

	var combinedSummary []*structpb.Struct = [](*structpb.Struct){}
	for _, node := range nodes {
		// Fetch the SummaryCoverageReport entity for the given configuration.
		summary := entities.SummaryCoverageData{}
		err := summary.Get(ctx, c.coverageV1DsClient, host, project, ref, revision, dataType, node, bucket, builder)
		if err != nil {
			logging.Errorf(ctx, "Error fetching SummaryCoverageData: %s", err)
			return nil, ErrInternalServerError
		}

		// Take the Data field out of the summary. The data field is compressed using
		// zlib, the following code decompresses that data and puts it into a struct.
		coverageDetailsStruct := structpb.Struct{}
		err = getStructFromCompressedData(summary.Data, &coverageDetailsStruct)
		if err != nil {
			logging.Errorf(ctx, "Unable to decompress the data: %s", err)
			return nil, ErrInternalServerError
		}

		combinedSummary = append(combinedSummary, &coverageDetailsStruct)
	}

	return &api.GetCoverageSummaryResponse{
		Summary: combinedSummary,
	}, nil
}

// getCoverageReportsForLastYear fetches the absolute code coverage reports
// for the last 365 days. These reports are specific to builder configuration
// which is supplied to this function as builder and bucket.
func (c *Client) getCoverageReportsForLastYear(
	ctx context.Context,
	bucket string,
	builder string,
) ([]entities.PostsubmitReport, error) {
	records := []entities.PostsubmitReport{}
	queryFilters := []datastorage.QueryFilter{
		{Field: "gitiles_commit.project", Operator: "=", Value: chromiumProject},
		{Field: "gitiles_commit.server_host", Operator: "=", Value: chromiumServerHost},
		{Field: "bucket", Operator: "=", Value: bucket},
		{Field: "builder", Operator: "=", Value: builder},
		{Field: "visible", Operator: "=", Value: true},
		{Field: "modifier_id", Operator: "=", Value: 0},
		{Field: "commit_timestamp", Operator: ">", Value: time.Now().Add(-time.Hour * 24 * 365)},
	}

	if err := c.coverageV1DsClient.Query(
		ctx, &records, "PostsubmitReport",
		queryFilters, "-commit_timestamp", 0,
	); err != nil {
		logging.Errorf(ctx, "PostsubmitReport: %w", err)
		return nil, ErrInternalServerError
	}

	return records, nil
}

// getCoverageNumbersForPath fetches absolute code coverage numbers for a given
// path/component for a given set of builder config & commit hashes. It returns
// per date numbers.
func (c *Client) getCoverageNumbersForPath(
	ctx context.Context,
	reports []entities.PostsubmitReport,
	path string,
	bucket string,
	builder string,
	isComponent bool,
) []CoveragePerDate {
	dataType := "dirs"
	if isComponent {
		dataType = "components"
	}

	coverageNumbers := []CoveragePerDate{}
	for _, report := range reports {
		summary := entities.SummaryCoverageData{}
		err := summary.Get(
			ctx, c.coverageV1DsClient,
			chromiumServerHost,
			chromiumProject,
			chromiumRef,
			report.GitilesCommitRevision,
			dataType,
			path,
			bucket,
			builder,
		)
		if err != nil {
			continue
		}
		coverageDetailsStruct := structpb.Struct{}
		err = getStructFromCompressedData(summary.Data, &coverageDetailsStruct)
		if err != nil {
			continue
		}
		metrics := coverageDetailsStruct.AsMap()
		for _, metric := range metrics["summaries"].([]interface{}) {
			metricMap := metric.(map[string]interface{})
			if metricMap["name"] == "line" {
				covNumber := CoveragePerDate{
					date:    report.CommitTimestamp.Format("2006-01-02"),
					covered: metricMap["covered"].(float64),
					total:   metricMap["total"].(float64),
				}
				coverageNumbers = append(coverageNumbers, covNumber)
			}
		}
	}
	return coverageNumbers
}

// GetAbsoluteCoverageDataOneYear TO_BE_IMPLEMENTED
func (c *Client) GetAbsoluteCoverageDataOneYear(
	ctx context.Context,
	req *api.GetAbsoluteCoverageDataOneYearRequest,
) (*api.GetAbsoluteCoverageDataOneYearResponse, error) {
	return nil, nil
}
