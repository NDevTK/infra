// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package coverage

import (
	"context"
	"encoding/json"
	"errors"

	"cloud.google.com/go/datastore"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/api/iterator"

	"infra/appengine/chrome-test-health/api"
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
	coverageV1DsClient *datastore.Client
	// References to chrome-test-health's datastore
	coverageV2DsClient *datastore.Client
}

func (c *Client) Init(ctx context.Context) error {
	covV1DsClient, err := datastore.NewClient(ctx, c.FinditCloudProject)
	if err != nil {
		return err
	}
	c.coverageV1DsClient = covV1DsClient

	covV2DsClient, err := datastore.NewClient(ctx, c.ChromeTestHealthCloudProject)
	if err != nil {
		return err
	}
	c.coverageV2DsClient = covV2DsClient

	return nil
}

// getProjectConfig extracts out the code coverage default settings from the
// fetched FinditConfig entity for the given project.
func (c *Client) getProjectConfig(ctx context.Context, finditConfig *entities.FinditConfig, project string) (*api.GetProjectDefaultConfigResponse, error) {
	code_cov_settings := map[string]interface{}{}
	err := json.Unmarshal(finditConfig.CodeCoverageSettings, &code_cov_settings)
	if err != nil {
		logging.Errorf(ctx, "Failed to unmarshall CodeCoverageSettings: %s", err)
		return nil, ErrInternalServerError
	}

	defaultPostSubmitConfig := code_cov_settings["default_postsubmit_report_config"]
	if defaultPostSubmitConfig == nil {
		logging.Errorf(ctx, "Missing default_postsubmit_report_config from FinditConfig")
		return nil, ErrInternalServerError
	}

	projectConfig := defaultPostSubmitConfig.(map[string]interface{})[project]
	if projectConfig == nil {
		logging.Errorf(ctx, "Missing config for project %s", project)
		return nil, ErrInternalServerError
	}
	projectConfig = projectConfig.(map[string]interface{})

	return &api.GetProjectDefaultConfigResponse{
		Host:     projectConfig.(map[string]interface{})["host"].(string),
		Platform: projectConfig.(map[string]interface{})["platform"].(string),
		Project:  projectConfig.(map[string]interface{})["project"].(string),
		Ref:      projectConfig.(map[string]interface{})["ref"].(string),
	}, nil
}

// GetProjectDefaultConfig fetches the latest version of FinditConfig from the datastore and returns
// the desired configuration extacted from the entity.
func (c *Client) GetProjectDefaultConfig(ctx context.Context, req *api.GetProjectDefaultConfigRequest) (*api.GetProjectDefaultConfigResponse, error) {
	project := req.Project

	// Fetches the FinditConfigRoot which has the latest version
	// number of the FinditConfig entity.
	finditConfigRoot := entities.FinditConfigRoot{}
	query := datastore.NewQuery("FinditConfigRoot")
	finditConfigRootKey, err := c.coverageV1DsClient.Run(ctx, query).Next(&finditConfigRoot)
	if err != nil {
		if err == iterator.Done {
			logging.Errorf(ctx, "FinditConfigRoot doesn't exist: %s", err)
			return nil, ErrInternalServerError
		}
		logging.Errorf(ctx, "Error fetching FinditConfigRoot: %s", err)
		return nil, ErrInternalServerError
	}

	// Fetches the FinditConfig entity corresponding to the version number
	// fetched above.
	finditConfigKey := datastore.IDKey("FinditConfig", int64(finditConfigRoot.Current), finditConfigRootKey)
	finditConfig := entities.FinditConfig{}
	err = c.coverageV1DsClient.Get(ctx, finditConfigKey, &finditConfig)
	if err != nil {
		logging.Errorf(ctx, "Error fetching FinditConfig: %s", err)
		return nil, ErrInternalServerError
	}

	return c.getProjectConfig(ctx, &finditConfig, project)
}
