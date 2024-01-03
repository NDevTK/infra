// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cron

import (
	"context"
	"time"

	"go.chromium.org/luci/common/logging"

	"infra/appengine/chrome-test-health/datastorage"
	"infra/appengine/chrome-test-health/internal/coverage"
	"infra/appengine/chrome-test-health/internal/coverage/entities"
)

const (
	chromiumHost = "chromium-review.googlesource.com"
)

type CronClient struct {
	// Refers to Findit's cloud project
	FinditCloudProject string
	// Refers to Chrome-test-health's cloud project
	ChromeTestHealthCloudProject string
	// References to Findit-for-me's datastore
	coverageV1DsClient datastorage.IDataClient
	// References to Chrome-test-health's datastore
	coverageV2DsClient datastorage.IDataClient
}

func NewClient(ctx context.Context, finditCloudProject string, chromeTestHealthCloudProject string) (*CronClient, error) {
	c := CronClient{
		FinditCloudProject:           finditCloudProject,
		ChromeTestHealthCloudProject: chromeTestHealthCloudProject,
	}

	covV1DsClient, err := datastorage.NewDataStoreClient(ctx, finditCloudProject)
	if err != nil {
		logging.Errorf(ctx, "Error connecting to %s", finditCloudProject)
		return nil, coverage.ErrInternalServerError
	}
	c.coverageV1DsClient = covV1DsClient

	covV2DsClient, err := datastorage.NewDataStoreClient(ctx, chromeTestHealthCloudProject)
	if err != nil {
		logging.Errorf(ctx, "Error connecting to %s", chromeTestHealthCloudProject)
		return nil, coverage.ErrInternalServerError
	}
	c.coverageV2DsClient = covV2DsClient

	return &c, nil
}

// UpdatePresubmitData TO_BE_IMPLEMENTED
func (c *CronClient) UpdatePresubmitData(ctx context.Context) error {
	return nil
}

// getPresubmitReportsOneDay gets the Presubmit Coverage Reports from the
// datastore for the last 24 hours
func (c *CronClient) getPresubmitReportsOneDay(
	ctx context.Context,
) ([]entities.PresubmitCoverageData, error) {
	records := []entities.PresubmitCoverageData{}
	queryFilters := []datastorage.QueryFilter{
		{Field: "cl_patchset.server_host", Operator: "=", Value: chromiumHost},
		{Field: "update_timestamp", Operator: ">", Value: time.Now().Add(-time.Hour * 24)},
	}

	if err := c.coverageV1DsClient.Query(
		ctx, &records, "PresubmitCoverageData",
		queryFilters, nil, 0,
	); err != nil {
		logging.Errorf(ctx, "PresubmitCoverageData: %w", err)
		return nil, coverage.ErrInternalServerError
	}

	return records, nil
}

// getMaxPatchsetToChangeMap returns a map. The key is the CL number and value is
// the latest patchset number for which we have presubmit reports
func (c *CronClient) getMaxPatchsetToChangeMap(
	presubmitData []entities.PresubmitCoverageData,
) map[int64]int64 {
	highestPatchsetMap := make(map[int64]int64)
	for _, data := range presubmitData {
		change := data.Change
		patchset := data.Patchset
		if _, ok := highestPatchsetMap[change]; !ok {
			highestPatchsetMap[change] = patchset
		}

		if patchset > highestPatchsetMap[change] {
			highestPatchsetMap[change] = patchset
		}
	}
	return highestPatchsetMap
}
