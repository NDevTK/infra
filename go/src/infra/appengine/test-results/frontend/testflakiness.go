// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/luci/gae/service/memcache"
	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/common/retry"
	"github.com/luci/luci-go/server/router"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"golang.org/x/sync/errgroup"
	bigquery "google.golang.org/api/bigquery/v2"
	"google.golang.org/appengine"
)

const teamsQuery = `
  SELECT
    layout_test_team
  FROM
    plx.google.chrome_infra.flaky_tests_with_layout_team_dir_info.all
  GROUP BY
    layout_test_team;`

const dirsQuery = `
  SELECT
    layout_test_dir
  FROM
    plx.google.chrome_infra.flaky_tests_with_layout_team_dir_info.all
  GROUP BY
    layout_test_dir;`

const suitesQuery = `
  SELECT
    if(regexp_contains(test_name, r'^org\..*#.*$'),
       regexp_extract(test_name, r'^(org\..*)#.*$'),
       regexp_extract(test_name, r'^([^\.\/]+)\.[^\/]+(?:\/[^\.]+)?$')) as suite
  FROM
    plx.google.chrome_infra.flaky_tests_with_layout_team_dir_info.all
  GROUP BY
    suite
  HAVING
    suite is not Null;`

const flakesQuery = `
  SELECT
    test_name,
    normalized_step_name,
    total_flaky_failures,
    total_tries,
    flakiness
  FROM
    plx.google.chrome_infra.flaky_tests_with_layout_team_dir_info.all
  %s
  ORDER BY
    flakiness DESC
  LIMIT
    1000;`

const bqProjectID = "test-results-hrd"

// Different kinds of groups to be used for Group struct below.
const (
	// Represents group of tests owned a team with specified name.
	TeamKind = "team"

	// Represents groups of tests in a diretory in the source code with path
	// specified as name.
	DirKind = "dir"

	// Represents group of tests not owned by any team.
	UnknownTeamKind = "unknown-team"

	// Represents group of tests for which we do not know the containing
	// directory.
	UnknownDirKind = "unknown-dir"

	// Represents a group of tests belonging to the test suite with specified
	// name.
	TestSuiteKind = "test-suite"

	// Represents a group of tests containing substring specified as name.
	SearchKind = "search"

	// Represents group of all tests. Practically will show 1000 most flaky tests.
	AllKind = "all"
)

const (
	flakinessGroupsKey       = "flakinessGroups"
	flakinessDataKeyTemplate = "flakinessData-%s-%s"
)

// Flakiness represents infromation about a single flaky test.
type Flakiness struct {
	TestName           string  `json:"test_name"`
	Flakiness          float64 `json:"flakiness"`
	NormalizedStepName string  `json:"normalized_step_name"`
	TotalFlakyFailures uint64  `json:"total_flaky_failures"`
	TotalTries         uint64  `json:"total_tries"`
}

// Group represents infromation about flakiness of a group of tests.
type Group struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

func writeError(ctx *router.Context, err error, funcName string, msg string) {
	reason := fmt.Sprintf("%s: %s", funcName, msg)
	if err == nil {
		err = errors.New(reason)
	}

	errors.Log(ctx.Context, errors.Annotate(err).Reason(reason).Err())
	http.Error(ctx.Writer, "Internal Server Error", http.StatusInternalServerError)
}

func writeResponse(ctx *router.Context, funcName string, data interface{}) {
	res, err := json.Marshal(data)
	if err != nil {
		writeError(ctx, err, funcName, "failed to marshal JSON")
		return
	}

	if _, err = ctx.Writer.Write(res); err != nil {
		writeError(ctx, err, funcName, "failed to write HTTP response")
	}
}

func getFlakinessData(ctx context.Context, bq *bigquery.Service, group Group) ([]Flakiness, error) {
	var filter string
	switch group.Kind {
	case TeamKind:
		// TODO(sergiyb): Change this when we have a way to detect which team owns a
		// given test (other than layout test).
		filter = "WHERE layout_test_team = @groupname"
	case DirKind:
		filter = "WHERE layout_test_dir = @groupname"
	case UnknownTeamKind:
		filter = "WHERE layout_test_team is Null"
	case UnknownDirKind:
		filter = "WHERE layout_test_dir is Null"
	case TestSuiteKind:
		filter = "WHERE regexp_contains(test_name, concat('^', @groupname, '[.#]'))"
	case SearchKind:
		filter = "WHERE strpos(test_name, @groupname) != 0"
	case AllKind:
		filter = ""
	default:
		return nil, errors.New("unknown group kind " + group.Kind)
	}

	queryParams := []*bigquery.QueryParameter{
		{
			Name:           "groupname",
			ParameterType:  &bigquery.QueryParameterType{Type: "STRING"},
			ParameterValue: &bigquery.QueryParameterValue{Value: group.Name},
		},
	}

	rows, err := executeBQQuery(
		ctx, bq, fmt.Sprintf(flakesQuery, filter), queryParams)
	if err != nil {
		return nil, errors.Annotate(err).Reason("failed to execute query").Err()
	}

	data := make([]Flakiness, 0, len(rows))
	for _, row := range rows {
		name, ok := row.F[0].V.(string)
		if !ok {
			return nil, errors.New("query returned non-string test name column")
		}

		normalizedStepName, ok := row.F[1].V.(string)
		if !ok {
			return nil, errors.New("query returned non-string value for normalized_step_name column")
		}

		totalFlakyFailuresStr, ok := row.F[2].V.(string)
		if !ok {
			return nil, errors.New("query returned non-string value for total_flaky_failures column")
		}

		totalFlakyFailures, err := strconv.ParseUint(totalFlakyFailuresStr, 10, 64)
		if err != nil {
			return nil, errors.Annotate(err).Reason("Failed to convert total_flaky_failures value to uint64").Err()
		}

		totalTriesStr, ok := row.F[3].V.(string)
		if !ok {
			return nil, errors.New("query returned non-string value for total_tries column")
		}

		totalTries, err := strconv.ParseUint(totalTriesStr, 10, 64)
		if err != nil {
			return nil, errors.Annotate(err).Reason("Failed to convert total_tries value to uint64").Err()
		}

		flakinessStr, ok := row.F[4].V.(string)
		if !ok {
			return nil, errors.New("query returned non-string value for string column")
		}

		flakiness, err := strconv.ParseFloat(flakinessStr, 64)
		if err != nil {
			return nil, errors.Annotate(err).Reason("Failed to convert flakiness value to float64").Err()
		}

		// TODO(sergiyb): Add number of false rejections per test.
		data = append(data, Flakiness{
			TestName:           name,
			NormalizedStepName: normalizedStepName,
			Flakiness:          flakiness,
			TotalTries:         totalTries,
			TotalFlakyFailures: totalFlakyFailures,
		})
	}

	return data, nil
}

func testFlakinessHandler(ctx *router.Context) {
	// TODO(sergiyb): Add a layer of caching results using memcache.
	name := ctx.Request.FormValue("groupName")
	kind := ctx.Request.FormValue("groupKind")

	if kind == "" {
		writeError(ctx, nil, "testFlakinessHandler", "missing groupKind parameter")
		return
	}

	if name == "" {
		writeError(ctx, nil, "testFlakinessHandler", "missing groupName parameter")
		return
	}

	// Check if we have recent results in memcache.
	var data []Flakiness
	key := fmt.Sprintf(flakinessDataKeyTemplate, kind, name)
	memcacheItem, err := memcache.GetKey(ctx.Context, key)
	if err == nil {
		if err = json.Unmarshal(memcacheItem.Value(), &data); err != nil {
			logging.Fields{logging.ErrorKey: err, "item": memcacheItem}.Warningf(
				ctx.Context, "Failed to unmarshal cached results as JSON")
		} else {
			writeResponse(ctx, "testFlakinessHandler", data)
			return
		}
	}

	aeCtx := appengine.NewContext(ctx.Request)
	bq, err := createBQService(aeCtx)
	if err != nil {
		writeError(ctx, err, "testFlakinessHandler", "failed create BigQuery client")
		return
	}

	data, err = getFlakinessData(aeCtx, bq, Group{Name: name, Kind: kind})
	if err != nil {
		writeError(ctx, err, "testFlakinessHandler", "failed to get flakiness data")
		return
	}

	// Store results in memcache for 1 hour.
	if dataStr, err := json.Marshal(data); err == nil {
		memcacheItem.SetValue(dataStr).SetExpiration(time.Hour)
		if err = memcache.Set(ctx.Context, memcacheItem); err != nil {
			logging.WithError(err).Warningf(
				ctx.Context, "Failed to store query results in memcache")
		}
	} else {
		logging.WithError(err).Warningf(
			ctx.Context, "Failed to marshal query results as JSON: %#v", data)
	}

	writeResponse(ctx, "testFlakinessHandler", data)
}

func createBQService(aeCtx context.Context) (*bigquery.Service, error) {
	hc, err := google.DefaultClient(aeCtx, bigquery.BigqueryScope)
	if err != nil {
		return nil, errors.Annotate(err).Reason("failed to create http client").Err()
	}

	bq, err := bigquery.New(hc)
	if err != nil {
		return nil, errors.Annotate(err).Reason("failed to create service object").Err()
	}

	return bq, nil
}

func executeBQQuery(ctx context.Context, bq *bigquery.Service, query string, params []*bigquery.QueryParameter) ([]*bigquery.TableRow, error) {
	logging.Infof(ctx, "Executing query `%s` with params %#v", query, params)

	useLegacySQL := false
	request := bq.Jobs.Query(bqProjectID, &bigquery.QueryRequest{
		TimeoutMs:       30 * 60 * 1000, // 30 minutes
		Query:           query,
		UseLegacySql:    &useLegacySQL,
		QueryParameters: params,
	})

	var response *bigquery.QueryResponse
	err := retry.Retry(ctx, retry.Default, func() error {
		var err error
		response, err = request.Do()
		return err
	}, nil)

	if err != nil {
		return nil, errors.Annotate(err).Reason("failed to execute query").Err()
	}

	jobID := response.JobReference.JobId
	if !response.JobComplete {
		return nil, errors.New("timed out while executing BQ query")
	}

	rows := make([]*bigquery.TableRow, 0, response.TotalRows)
	rows = append(rows, response.Rows...)
	pageToken := response.PageToken

	// Get additional results if any.
	for pageToken != "" {
		resultsRequest := bq.Jobs.GetQueryResults(bqProjectID, jobID)
		resultsRequest.PageToken(pageToken)
		var resultsResponse *bigquery.GetQueryResultsResponse
		err := retry.Retry(ctx, retry.Default, func() error {
			var err error
			resultsResponse, err = resultsRequest.Do()
			return err
		}, nil)

		if err != nil {
			return nil, errors.Annotate(err).Reason("failed to retrive additional results").Err()
		}

		rows = append(rows, resultsResponse.Rows...)
		pageToken = resultsResponse.PageToken
	}

	return rows, nil
}

func getGroupsForQuery(ctx context.Context, bq *bigquery.Service, query, kind, nilKind string) ([]Group, error) {
	rows, err := executeBQQuery(ctx, bq, query, nil)
	if err != nil {
		return nil, errors.Annotate(err).Reason("failed to execute query").Err()
	}

	var groups []Group
	for _, row := range rows {
		value := row.F[0].V
		switch value := value.(type) {
		case string:
			groups = append(groups, Group{Name: value, Kind: kind})
		case nil:
			if nilKind == "" {
				return nil, errors.New("Unexpected NULL value for a group")
			}
			groups = append(groups, Group{Name: nilKind, Kind: nilKind})
		default:
			return nil, errors.New("query returned non-string non-nil value")
		}
	}

	return groups, nil
}

func getFlakinessGroups(ctx context.Context, bq *bigquery.Service) ([]Group, error) {
	var teamGroups, dirGroups, suiteGroups []Group
	var g errgroup.Group

	g.Go(func() (err error) {
		teamGroups, err = getGroupsForQuery(ctx, bq, teamsQuery, TeamKind, UnknownTeamKind)
		return
	})

	g.Go(func() (err error) {
		dirGroups, err = getGroupsForQuery(ctx, bq, dirsQuery, DirKind, UnknownDirKind)
		return
	})

	g.Go(func() (err error) {
		suiteGroups, err = getGroupsForQuery(ctx, bq, suitesQuery, TestSuiteKind, "")
		return
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return append(teamGroups, append(dirGroups, suiteGroups...)...), nil
}

func testFlakinessGroupsHandler(ctx *router.Context) {
	// Check if we have recent results in memcache.
	var groups []Group
	memcacheItem, err := memcache.GetKey(ctx.Context, flakinessGroupsKey)
	if err == nil {
		if err = json.Unmarshal(memcacheItem.Value(), &groups); err != nil {
			logging.Fields{logging.ErrorKey: err, "item": memcacheItem}.Warningf(
				ctx.Context, "Failed to unmarshal cached results as JSON")
		} else {
			writeResponse(ctx, "testFlakinessGroupsHandler", groups)
			return
		}
	}

	aeCtx := appengine.WithContext(ctx.Context, ctx.Request)
	bq, err := createBQService(aeCtx)
	if err != nil {
		writeError(ctx, err, "testFlakinessGroupsHandler", "failed create BigQuery client")
		return
	}

	groups, err = getFlakinessGroups(aeCtx, bq)
	if err != nil {
		writeError(ctx, err, "testFlakinessGroupsHandler", "failed to get flakiness groups")
		return
	}

	// Store results in memcache for 1 hour.
	if groupsStr, err := json.Marshal(groups); err == nil {
		memcacheItem.SetValue(groupsStr).SetExpiration(time.Hour)
		if err = memcache.Set(ctx.Context, memcacheItem); err != nil {
			logging.WithError(err).Warningf(
				ctx.Context, "Failed to store query results in memcache")
		}
	} else {
		logging.WithError(err).Warningf(
			ctx.Context, "Failed to marshal query results as JSON: %#v", groups)
	}

	writeResponse(ctx, "testFlakinessGroupsHandler", groups)
}
