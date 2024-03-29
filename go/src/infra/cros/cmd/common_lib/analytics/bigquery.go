// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package analytics

import (
	"context"
	"infra/cros/cmd/ctpv2/data"
	"strconv"
	"strings"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/api/option"
)

const dataset = "analytics"
const resultsTable = "CTPV2Metrics"
const taskResultsTable = "CTPV2TaskMetrics"

const saProject = "chromeos-test-platform-data"

// tableProject := "chromeos-test-platform-data"
const saFile = "/creds/service_accounts/service-account-chromeos.json"

const Start = "START"
const Success = "SUCCESS"
const Fail = "FAIL"
const Panic = "PANIC"

type BqData struct {
	SuiteName     string
	AnalyticsName string
	BBID          string
	Build         string
	Step          string
	Freeform      string
	Pool          string
	Status        string
	Duration      float32
}

type TaskData struct {
	SuiteName     string
	AnalyticsName string
	BBID          string
	Build         string
	Step          string
	Freeform      string
	Pool          string
	Status        string
	Duration      float32
	DisplayName   string
	TrTaskID      string
	SchedukeID    string
	Board         string
	Model         string
	Deps          []string
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

// InsertCTPMetrics will insert the CTP Analytics Data into the CTPv2Metrics Table.
func InsertCTPMetrics(c *bigquery.Client, data []*BqData) error {
	ctx := context.Background()
	inserter := c.Dataset(dataset).Table(resultsTable).Inserter()
	if err := inserter.Put(ctx, data); err != nil {
		return err
	}

	logging.Infof(ctx, "Successfully inserted %v rows to %s", len(data), resultsTable)
	return nil
}

// InsertCTPMetrics will insert the CTP Analytics Data into the CTPv2Metrics Table.
func InsertCTPTaskMetrics(c *bigquery.Client, data []*TaskData) error {
	ctx := context.Background()
	inserter := c.Dataset(dataset).Table(taskResultsTable).Inserter()
	if err := inserter.Put(ctx, data); err != nil {
		return err
	}

	logging.Infof(ctx, "Successfully inserted %v rows to %s", len(data), resultsTable)
	return nil
}

// SoftInsertStep insert a step info to BQ. Do not fail on errors.
func SoftInsertStep(ctx context.Context, BQClient *bigquery.Client, data *BqData) {
	var rows []*BqData
	rows = append(rows, data)

	if BQClient != nil {
		err := InsertCTPMetrics(BQClient, rows)
		if err != nil {
			logging.Infof(ctx, "ERROR DURING BQ WITE: %s", err)

		}
	}
}

// SoftInsertStepWCtp2Req insert a step info to BQl built from the CTPv2Request req. Do not fail on errors.
func SoftInsertStepWCtp2Req(ctx context.Context, BQClient *bigquery.Client, data *BqData, req *api.CTPv2Request, build *build.State) {
	if req != nil && len(req.GetRequests()) > 0 {
		data = buildDataFromCTPReq(data, req)
	}
	data = addBBID(data, build)

	var rows []*BqData
	rows = append(rows, data)

	if BQClient != nil {
		err := InsertCTPMetrics(BQClient, rows)
		if err != nil {
			logging.Infof(ctx, "ERROR DURING BQ WITE: %s", err)

		}
	} else {
		logging.Infof(ctx, "Skipped BQ write as no client provided.")

	}
}

// SoftInsertStepWTrReq insert a step info to BQ built from the Trreq. Do not fail on errors.
func SoftInsertStepWTrReq(ctx context.Context, BQClient *bigquery.Client, data *TaskData, req *data.TrRequest, suiteInfo *api.SuiteInfo, build *build.State) {
	if req == nil {
		return
	}
	// BBID          string
	// Build         string
	// Freeform      string
	// Pool          string
	// DisplayName   string
	// TrTaskID      string
	// SchedukeID    string
	// Board         string
	// Model         string
	// Deps          []string
	// data.SuiteName = req.Req

	var rows []*TaskData
	data.SuiteName = suiteInfo.GetSuiteRequest().GetTestSuite().GetName()
	data.AnalyticsName = suiteInfo.GetSuiteRequest().GetAnalyticsName()
	rows = append(rows, data)

	if BQClient != nil {
		err := InsertCTPTaskMetrics(BQClient, rows)
		if err != nil {
			logging.Infof(ctx, "ERROR DURING BQ WITE: %s", err)

		}
	}
}

// buildDataFromInternal will append the CTP Request details to the to be inserted BQ row.
func buildDataFromCTPReq(data *BqData, req *api.CTPv2Request) *BqData {
	poolName := req.GetRequests()[0].GetPool()
	aName := req.GetRequests()[0].GetSuiteRequest().GetAnalyticsName()
	sName := req.GetRequests()[0].GetSuiteRequest().GetTestSuite().GetName()
	b := req.GetRequests()[0].GetScheduleTargets()[0].GetTargets()[0].GetSwTarget().GetLegacySw().GetGcsPath()
	if aName != "" {
		data.AnalyticsName = aName
	}
	if sName != "" {
		data.SuiteName = sName
	}
	if b != "" {
		data.Build = buildFromGcs(b)
	}
	if poolName != "" {
		data.Pool = poolName
	}
	return data
}

// SoftInsertStepWInternalPlan insert a step info to BQl built from the InternalTestPlan req. Do not fail on errors.
func SoftInsertStepWInternalPlan(ctx context.Context, BQClient *bigquery.Client, data *BqData, req *api.InternalTestplan, build *build.State) {
	if req != nil {
		data = buildDataFromInternalTP(data, req)
	}
	data = addBBID(data, build)

	var rows []*BqData
	rows = append(rows, data)

	if BQClient != nil {
		err := InsertCTPMetrics(BQClient, rows)
		if err != nil {
			logging.Infof(ctx, "ERROR DURING BQ WITE: %s", err)

		}
	}
}

// gs://chromeos-image-archive/dedede-release/R124-15815.0.0 --> R124-15815.0.0
func buildFromGcs(build string) string {
	f := strings.Split(build, "/")
	return f[len(f)-1]

}

// buildDataFromInternal will append the BBID to the to be inserted BQ row.
func addBBID(data *BqData, build *build.State) *BqData {
	if build == nil {
		return data
	}
	swarmingID := strconv.FormatInt(build.Build().GetId(), 10)
	if swarmingID != "" {
		data.BBID = swarmingID
	}
	return data
}

// buildDataFromInternal will append the internal test plan details to the to be inserted BQ row.
func buildDataFromInternalTP(data *BqData, req *api.InternalTestplan) *BqData {
	aName := req.GetSuiteInfo().GetSuiteRequest().GetAnalyticsName()
	poolName := req.GetSuiteInfo().GetSuiteMetadata().GetPool()
	sName := req.GetSuiteInfo().GetSuiteRequest().GetTestSuite().GetName()
	if len(req.GetSuiteInfo().GetSuiteMetadata().GetTargetRequirements()) > 0 {
		b := req.GetSuiteInfo().GetSuiteMetadata().GetTargetRequirements()[0].GetSwRequirement().GetGcsPath()
		if b != "" {
			data.Build = buildFromGcs(b)
		}

	}
	if aName != "" {
		data.AnalyticsName = aName
	}
	if sName != "" {
		data.SuiteName = sName
	}
	if poolName != "" {
		data.Pool = poolName
	}
	return data
}
