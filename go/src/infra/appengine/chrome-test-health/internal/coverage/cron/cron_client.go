// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cron

import (
	"context"
	"errors"

	"go.chromium.org/luci/common/logging"

	"infra/appengine/chrome-test-health/datastorage"
)

var (
	ErrInternalServerError = errors.New("internal Server Error")
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
		return nil, ErrInternalServerError
	}
	c.coverageV1DsClient = covV1DsClient

	covV2DsClient, err := datastorage.NewDataStoreClient(ctx, chromeTestHealthCloudProject)
	if err != nil {
		logging.Errorf(ctx, "Error connecting to %s", chromeTestHealthCloudProject)
		return nil, ErrInternalServerError
	}
	c.coverageV2DsClient = covV2DsClient

	return &c, nil
}

// UpdatePresubmitData TO_BE_IMPLEMENTED
func (c *CronClient) UpdatePresubmitData(ctx context.Context) error {
	return nil
}
