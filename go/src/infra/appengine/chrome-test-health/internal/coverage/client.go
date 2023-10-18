// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package coverage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"go.chromium.org/luci/common/logging"

	"infra/appengine/chrome-test-health/api"
	"infra/appengine/chrome-test-health/datastorage"
	"infra/appengine/chrome-test-health/internal/coverage/entities"
)

var (
	ErrInternalServerError = errors.New("Internal Server Error")
)

// TODO(crbug.com/1474096) - Refactor the code here into Common and Infra layers
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
// configuration including the path. The path here can be a dir/file path like
// //foo/foo1/foo2/ or a component like C1>C2>C3.
func (c *Client) GetCoverageSummary(ctx context.Context, req *api.GetCoverageSummaryRequest) (*api.GetCoverageSummaryResponse, error) {
	host := req.GitilesHost
	project := req.GitilesProject
	ref := req.GitilesRef
	revision := req.GitilesRevision
	path := req.Path
	unitTestsOnly := req.UnitTestsOnly
	dataType := req.DataType
	bucket := req.Bucket
	builder := c.getModifedBuilder(req.Builder, &unitTestsOnly)

	// Fetch the SummaryCoverageReport entity for the given configuration.
	summary := entities.SummaryCoverageData{}
	err := summary.Get(ctx, c.coverageV1DsClient, host, project, ref, revision, dataType, path, bucket, builder)
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

	return &api.GetCoverageSummaryResponse{
		Summary: &coverageDetailsStruct,
	}, nil
}
