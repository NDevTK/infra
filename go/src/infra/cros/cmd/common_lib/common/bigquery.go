// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/api/option"
)

const dataset = "analytics"
const resultsTable = "CTPV2Metrics"

const saProject = "chromeos-test-platform-data"

// tableProject := "chromeos-test-platform-data"
const saFile = "/creds/service_accounts/service-account-chromeos.json"

type BqData struct {
	Step string
}

// CtpAnalyticsBQClient will build the client for the CTP BQ tables, using the default CTP SA
func CtpAnalyticsBQClient(ctx context.Context) *bigquery.Client {
	c, err := bigquery.NewClient(ctx, saProject,
		option.WithCredentialsFile(saFile))
	if err != nil {
		logging.Infof(ctx, "Unable to make BQ client :%s", err)
		return nil
	}

	return c
}

// InsertRows will insert the CTP Analytics Data into the CTPv2Metrics Table.
func InsertRows(c *bigquery.Client, data []*BqData) error {
	ctx := context.Background()
	inserter := c.Dataset(dataset).Table(resultsTable).Inserter()
	if err := inserter.Put(ctx, data); err != nil {
		return err
	}

	logging.Infof(ctx, "Successfully inserted %v rows to %s", len(data), resultsTable)
	return nil
}
