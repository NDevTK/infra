// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testmetrics

import (
	"infra/appengine/chrome-test-health/api"
	"testing"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	. "github.com/smartystreets/goconvey/convey"
)

func ShouldContainParameter(actual any, expected ...any) string {
	expectedParameter := expected[0].(bigquery.QueryParameter)
	for _, parameter := range actual.([]bigquery.QueryParameter) {
		actualParameter := bigquery.QueryParameter(parameter)
		if actualParameter.Name == expectedParameter.Name {
			return ShouldResemble(actualParameter, expectedParameter)
		}
	}
	return "Parameter not found in the actual"
}

func TestCreateFetchMetricsQuery(t *testing.T) {
	t.Parallel()

	Convey("createFetchMetricsQuery", t, func() {
		client := Client{
			ProjectId: "chrome-test-health-project",
			DataSet:   "normal-dataset",
		}
		request := &api.FetchTestMetricsRequest{
			Components: []string{
				"Blink",
			},
			Dates: []string{
				"2023-07-12",
			},
			Period: api.Period_DAY,
			Metrics: []api.MetricType{
				api.MetricType_NUM_RUNS,
			},
			PageOffset: 0,
			PageSize:   10,
			Sort: &api.SortBy{
				Metric:    api.SortType_SORT_NAME,
				Ascending: true,
			},
		}
		Convey("Valid unfiltered request", func() {
			query, err := client.createFetchMetricsQuery(request)

			So(err, ShouldBeNil)
			So(query, ShouldNotBeNil)
			So(query.QueryConfig.Q, ShouldResemble, `
WITH base AS (
	SELECT
		m.date,
		m.test_id,
		ANY_VALUE(m.test_name) AS test_name,
		ANY_VALUE(m.file_name) AS file_name,
		SUM(num_runs) AS num_runs,
		ARRAY_AGG(STRUCT(
			builder AS builder,
			bucket AS bucket,
			test_suite AS test_suite,
			num_runs
			)
		) AS variants
	FROM
		chrome-test-health-project.normal-dataset.daily_test_metrics AS m
	WHERE
		DATE(date) IN UNNEST(@dates)
		AND component IN UNNEST(@components)
	GROUP BY date, test_id
	ORDER BY test_id ASC
	LIMIT @page_size OFFSET @page_offset
)
SELECT
	* EXCEPT (variants),
	(SELECT ARRAY_AGG(v ORDER BY test_id ASC) FROM UNNEST(variants) v) AS variants
FROM base`)
		})

		Convey("Valid filtered request", func() {
			request.Filter = "linux-rel blink_python_tests"
			query, err := client.createFetchMetricsQuery(request)

			So(err, ShouldBeNil)
			So(query, ShouldNotBeNil)
			So(query.QueryConfig.Q, ShouldResemble, `
WITH base AS (
	SELECT
		m.date,
		m.test_id,
		ANY_VALUE(m.test_name) AS test_name,
		ANY_VALUE(m.file_name) AS file_name,
		SUM(num_runs) AS num_runs,
		ARRAY_AGG(STRUCT(
			builder AS builder,
			bucket AS bucket,
			test_suite AS test_suite,
			num_runs
			)
		) AS variants
	FROM
		chrome-test-health-project.normal-dataset.daily_test_metrics AS m
	WHERE
		DATE(date) IN UNNEST(@dates)
		AND component IN UNNEST(@components)
		AND REGEXP_CONTAINS(CONCAT('id:', test_id, ' ', 'name:', IFNULL(test_name, ''), ' ', 'file:', IFNULL(file_name, ''), ' ', 'bucket:', IFNULL(bucket, ''), '/', IFNULL(builder, ''), 'builder:', IFNULL(builder, ''), ' ', 'test_suite:', IFNULL(test_suite, '')), @filter0)
		AND REGEXP_CONTAINS(CONCAT('id:', test_id, ' ', 'name:', IFNULL(test_name, ''), ' ', 'file:', IFNULL(file_name, ''), ' ', 'bucket:', IFNULL(bucket, ''), '/', IFNULL(builder, ''), 'builder:', IFNULL(builder, ''), ' ', 'test_suite:', IFNULL(test_suite, '')), @filter1)
	GROUP BY date, test_id
	ORDER BY test_id ASC
	LIMIT @page_size OFFSET @page_offset
)
SELECT
	* EXCEPT (variants),
	(SELECT ARRAY_AGG(v ORDER BY test_id ASC) FROM UNNEST(variants) v) AS variants
FROM base`)
		})

		Convey("No component request", func() {
			request.Components = []string{}
			query, err := client.createFetchMetricsQuery(request)

			So(err, ShouldBeNil)
			So(query, ShouldNotBeNil)
			So(query.QueryConfig.Q, ShouldResemble, `
WITH base AS (
	SELECT
		m.date,
		m.test_id,
		ANY_VALUE(m.test_name) AS test_name,
		ANY_VALUE(m.file_name) AS file_name,
		SUM(num_runs) AS num_runs,
		ARRAY_AGG(STRUCT(
			builder AS builder,
			bucket AS bucket,
			test_suite AS test_suite,
			num_runs
			)
		) AS variants
	FROM
		chrome-test-health-project.normal-dataset.daily_test_metrics AS m
	WHERE
		DATE(date) IN UNNEST(@dates)
	GROUP BY date, test_id
	ORDER BY test_id ASC
	LIMIT @page_size OFFSET @page_offset
)
SELECT
	* EXCEPT (variants),
	(SELECT ARRAY_AGG(v ORDER BY test_id ASC) FROM UNNEST(variants) v) AS variants
FROM base`)
		})

		Convey("Valid filename filtered request", func() {
			request.Filter = "linux-rel blink_python_tests"
			request.FileNames = []string{"filename.html"}
			query, err := client.createFetchMetricsQuery(request)

			So(err, ShouldBeNil)
			So(query, ShouldNotBeNil)
			So(query.QueryConfig.Q, ShouldResemble, `
WITH base AS (
	SELECT
		m.date,
		m.test_id,
		ANY_VALUE(m.test_name) AS test_name,
		ANY_VALUE(m.file_name) AS file_name,
		SUM(num_runs) AS num_runs,
		ARRAY_AGG(STRUCT(
			builder AS builder,
			bucket AS bucket,
			test_suite AS test_suite,
			num_runs
			)
		) AS variants
	FROM
		chrome-test-health-project.normal-dataset.daily_test_metrics AS m
	WHERE
		DATE(date) IN UNNEST(@dates)
		AND component IN UNNEST(@components)
		AND file_name IN UNNEST(@file_names)
		AND REGEXP_CONTAINS(CONCAT('id:', test_id, ' ', 'name:', IFNULL(test_name, ''), ' ', 'file:', IFNULL(file_name, ''), ' ', 'bucket:', IFNULL(bucket, ''), '/', IFNULL(builder, ''), 'builder:', IFNULL(builder, ''), ' ', 'test_suite:', IFNULL(test_suite, '')), @filter0)
		AND REGEXP_CONTAINS(CONCAT('id:', test_id, ' ', 'name:', IFNULL(test_name, ''), ' ', 'file:', IFNULL(file_name, ''), ' ', 'bucket:', IFNULL(bucket, ''), '/', IFNULL(builder, ''), 'builder:', IFNULL(builder, ''), ' ', 'test_suite:', IFNULL(test_suite, '')), @filter1)
	GROUP BY date, test_id
	ORDER BY test_id ASC
	LIMIT @page_size OFFSET @page_offset
)
SELECT
	* EXCEPT (variants),
	(SELECT ARRAY_AGG(v ORDER BY test_id ASC) FROM UNNEST(variants) v) AS variants
FROM base`)
		})

		Convey("Valid filtered multi-day request", func() {
			request.Filter = "linux-rel blink_python_tests"
			request.Dates = append(request.Dates, "2023-07-13")
			query, err := client.createFetchMetricsQuery(request)

			So(err, ShouldBeNil)
			So(query, ShouldNotBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "sort_date",
				Value: "2023-07-12",
			})
			So(query.QueryConfig.Q, ShouldResemble, `
WITH tests AS (
	SELECT
		m.date,
		m.test_id,
		ANY_VALUE(m.test_name) AS test_name,
		ANY_VALUE(m.file_name) AS file_name,
		SUM(num_runs) AS num_runs,
		ARRAY_AGG(STRUCT(
			builder AS builder,
			bucket AS bucket,
			test_suite AS test_suite,
			num_runs
			) ORDER BY test_id ASC
		) AS variants
	FROM
		chrome-test-health-project.normal-dataset.daily_test_metrics AS m
	WHERE
		DATE(date) IN UNNEST(@dates)
		AND component IN UNNEST(@components)
		AND REGEXP_CONTAINS(CONCAT('id:', test_id, ' ', 'name:', IFNULL(test_name, ''), ' ', 'file:', IFNULL(file_name, ''), ' ', 'bucket:', IFNULL(bucket, ''), '/', IFNULL(builder, ''), 'builder:', IFNULL(builder, ''), ' ', 'test_suite:', IFNULL(test_suite, '')), @filter0)
		AND REGEXP_CONTAINS(CONCAT('id:', test_id, ' ', 'name:', IFNULL(test_name, ''), ' ', 'file:', IFNULL(file_name, ''), ' ', 'bucket:', IFNULL(bucket, ''), '/', IFNULL(builder, ''), 'builder:', IFNULL(builder, ''), ' ', 'test_suite:', IFNULL(test_suite, '')), @filter1)
	GROUP BY m.date, m.test_id
), sorted_day AS (
	SELECT
		test_id,
		test_id AS rank
	FROM tests
	WHERE date = @sort_date
	ORDER BY test_id ASC
	LIMIT @page_size OFFSET @page_offset
)
SELECT t.*
FROM sorted_day AS s FULL OUTER JOIN tests AS t USING(test_id)
ORDER BY rank IS NULL, rank ASC`)
		})

		Convey("Valid no component multi-day request", func() {
			request.Filter = "linux-rel blink_python_tests"
			request.Dates = append(request.Dates, "2023-07-13")
			request.Components = []string{}
			query, err := client.createFetchMetricsQuery(request)

			So(err, ShouldBeNil)
			So(query, ShouldNotBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "sort_date",
				Value: "2023-07-12",
			})
			So(query.QueryConfig.Q, ShouldResemble, `
WITH tests AS (
	SELECT
		m.date,
		m.test_id,
		ANY_VALUE(m.test_name) AS test_name,
		ANY_VALUE(m.file_name) AS file_name,
		SUM(num_runs) AS num_runs,
		ARRAY_AGG(STRUCT(
			builder AS builder,
			bucket AS bucket,
			test_suite AS test_suite,
			num_runs
			) ORDER BY test_id ASC
		) AS variants
	FROM
		chrome-test-health-project.normal-dataset.daily_test_metrics AS m
	WHERE
		DATE(date) IN UNNEST(@dates)
		AND REGEXP_CONTAINS(CONCAT('id:', test_id, ' ', 'name:', IFNULL(test_name, ''), ' ', 'file:', IFNULL(file_name, ''), ' ', 'bucket:', IFNULL(bucket, ''), '/', IFNULL(builder, ''), 'builder:', IFNULL(builder, ''), ' ', 'test_suite:', IFNULL(test_suite, '')), @filter0)
		AND REGEXP_CONTAINS(CONCAT('id:', test_id, ' ', 'name:', IFNULL(test_name, ''), ' ', 'file:', IFNULL(file_name, ''), ' ', 'bucket:', IFNULL(bucket, ''), '/', IFNULL(builder, ''), 'builder:', IFNULL(builder, ''), ' ', 'test_suite:', IFNULL(test_suite, '')), @filter1)
	GROUP BY m.date, m.test_id
), sorted_day AS (
	SELECT
		test_id,
		test_id AS rank
	FROM tests
	WHERE date = @sort_date
	ORDER BY test_id ASC
	LIMIT @page_size OFFSET @page_offset
)
SELECT t.*
FROM sorted_day AS s FULL OUTER JOIN tests AS t USING(test_id)
ORDER BY rank IS NULL, rank ASC`)
		})

		Convey("Valid sorted multi-day request", func() {
			request.Dates = append(request.Dates, "2023-07-13")
			request.Sort.SortDate = "2023-07-13"
			request.Sort.Ascending = false
			query, err := client.createFetchMetricsQuery(request)

			So(err, ShouldBeNil)
			So(query, ShouldNotBeNil)
			So(query.QueryConfig.Q, ShouldResemble, `
WITH tests AS (
	SELECT
		m.date,
		m.test_id,
		ANY_VALUE(m.test_name) AS test_name,
		ANY_VALUE(m.file_name) AS file_name,
		SUM(num_runs) AS num_runs,
		ARRAY_AGG(STRUCT(
			builder AS builder,
			bucket AS bucket,
			test_suite AS test_suite,
			num_runs
			) ORDER BY test_id DESC
		) AS variants
	FROM
		chrome-test-health-project.normal-dataset.daily_test_metrics AS m
	WHERE
		DATE(date) IN UNNEST(@dates)
		AND component IN UNNEST(@components)
	GROUP BY m.date, m.test_id
), sorted_day AS (
	SELECT
		test_id,
		test_id AS rank
	FROM tests
	WHERE date = @sort_date
	ORDER BY test_id DESC
	LIMIT @page_size OFFSET @page_offset
)
SELECT t.*
FROM sorted_day AS s FULL OUTER JOIN tests AS t USING(test_id)
ORDER BY rank IS NULL, rank DESC`)
		})

		Convey("Parameterized args", func() {
			query, err := client.createFetchMetricsQuery(request)

			So(err, ShouldBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "components",
				Value: []string{"Blink"},
			})
		})

		Convey("Parameterized page args", func() {
			request.PageSize = 10
			request.PageOffset = 5
			query, err := client.createFetchMetricsQuery(request)

			So(err, ShouldBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "page_size",
				Value: int64(10 + 1),
			})
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "page_offset",
				Value: int64(5),
			})
		})

		Convey("Parameterized filter arg", func() {
			request.Filter = "linux-rel blink_python_tests"

			query, err := client.createFetchMetricsQuery(request)

			So(err, ShouldBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "filter0",
				Value: "linux-rel",
			})
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "filter1",
				Value: "blink_python_tests",
			})
		})

		Convey("Parameterized dates arg", func() {
			request.Dates = []string{
				"2023-07-12",
				"2023-07-13",
			}

			query, err := client.createFetchMetricsQuery(request)

			So(err, ShouldBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name: "dates",
				Value: []civil.Date{
					{
						Year:  2023,
						Month: 7,
						Day:   12,
					},
					{
						Year:  2023,
						Month: 7,
						Day:   13,
					},
				},
			})
		})

		Convey("Partially defined sort returns error", func() {
			request.Sort = &api.SortBy{Metric: 99}

			_, err := client.createFetchMetricsQuery(request)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestCreateUnfilteredDirectoryQuery(t *testing.T) {
	t.Parallel()

	Convey("createFetchMetricsQuery", t, func() {
		client := Client{
			ProjectId: "chrome-test-health-project",
			DataSet:   "normal-dataset",
		}
		request := &api.FetchDirectoryMetricsRequest{
			ParentIds: []string{"/"},
			Components: []string{
				"Blink",
			},
			Dates: []string{
				"2023-07-12",
			},
			Period: api.Period_DAY,
			Metrics: []api.MetricType{
				api.MetricType_NUM_RUNS,
			},
			Sort: &api.SortBy{
				Metric:    api.SortType_SORT_NAME,
				Ascending: true,
			},
		}

		Convey("Valid no component request", func() {
			request.Components = []string{}
			query, err := client.createDirectoryQuery(request)

			So(err, ShouldBeNil)
			So(query, ShouldNotBeNil)
			So(query.QueryConfig.Q, ShouldResemble, `
SELECT
	date,
	node_name,
	ARRAY_REVERSE(SPLIT(node_name, '/'))[SAFE_OFFSET(0)] AS display_name,
	ANY_VALUE(is_file) AS is_file,
	SUM(num_runs) AS num_runs,
FROM chrome-test-health-project.normal-dataset.daily_file_metrics, UNNEST(@parents) AS parent
WHERE
	((STARTS_WITH(node_name, parent || "/")
	-- The child folders and files can't have a / after the parent's name
	AND REGEXP_CONTAINS(SUBSTR(node_name, LENGTH(parent) + 2), "^[^/]*$"))
	OR (parent = '' AND NOT STARTS_WITH(node_name, "/")))
	AND DATE(date) IN UNNEST(@dates)
GROUP BY date, node_name
ORDER BY is_file, node_name ASC`)
		})

		Convey("Valid unfiltered request", func() {
			query, err := client.createDirectoryQuery(request)

			So(err, ShouldBeNil)
			So(query, ShouldNotBeNil)
			So(query.QueryConfig.Q, ShouldResemble, `
SELECT
	date,
	node_name,
	ARRAY_REVERSE(SPLIT(node_name, '/'))[SAFE_OFFSET(0)] AS display_name,
	ANY_VALUE(is_file) AS is_file,
	SUM(num_runs) AS num_runs,
FROM chrome-test-health-project.normal-dataset.daily_file_metrics, UNNEST(@parents) AS parent
WHERE
	((STARTS_WITH(node_name, parent || "/")
	-- The child folders and files can't have a / after the parent's name
	AND REGEXP_CONTAINS(SUBSTR(node_name, LENGTH(parent) + 2), "^[^/]*$"))
	OR (parent = '' AND NOT STARTS_WITH(node_name, "/")))
	AND DATE(date) IN UNNEST(@dates)
		AND component IN UNNEST(@components)
GROUP BY date, node_name
ORDER BY is_file, node_name ASC`)
		})

		Convey("Valid unfiltered multi-day request", func() {
			request.Dates = append(request.Dates, "2023-07-13")
			request.Sort.SortDate = "2023-07-13"
			request.Sort.Ascending = false
			query, err := client.createDirectoryQuery(request)

			So(err, ShouldBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "sort_date",
				Value: "2023-07-13",
			})
			So(query, ShouldNotBeNil)
			So(query.QueryConfig.Q, ShouldResemble, `
WITH nodes AS(
	SELECT
		date,
		node_name,
		ARRAY_REVERSE(SPLIT(node_name, '/'))[SAFE_OFFSET(0)] AS display_name,
		ANY_VALUE(is_file) AS is_file,
		SUM(num_runs) AS num_runs,
	FROM chrome-test-health-project.normal-dataset.daily_file_metrics, UNNEST(@parents) AS parent
	WHERE
		((STARTS_WITH(node_name, parent || "/")
		-- The child folders and files can't have a / after the parent's name
		AND REGEXP_CONTAINS(SUBSTR(node_name, LENGTH(parent) + 2), "^[^/]*$"))
		OR (parent = '' AND NOT STARTS_WITH(node_name, "/")))
		AND DATE(date) IN UNNEST(@dates)
		AND component IN UNNEST(@components)
	GROUP BY date, node_name
), sorted_day AS (
	SELECT
		node_name,
		node_name AS rank
	FROM nodes
	WHERE date = @sort_date
)
SELECT t.*
FROM nodes AS t FULL OUTER JOIN sorted_day AS s USING(node_name)
ORDER BY is_file, s.rank IS NULL, s.rank DESC`)
		})

		Convey("Parameterized args", func() {
			query, err := client.createDirectoryQuery(request)

			So(err, ShouldBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "components",
				Value: []string{"Blink"},
			})
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "parents",
				Value: []string{"/"},
			})
		})

		Convey("Parameterized dates arg", func() {
			request.Dates = []string{
				"2023-07-12",
				"2023-07-13",
			}

			query, err := client.createDirectoryQuery(request)

			So(err, ShouldBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name: "dates",
				Value: []civil.Date{
					{
						Year:  2023,
						Month: 7,
						Day:   12,
					},
					{
						Year:  2023,
						Month: 7,
						Day:   13,
					},
				},
			})
		})
	})
}

func TestCreateFilteredDirectoryQuery(t *testing.T) {
	t.Parallel()

	Convey("createFetchMetricsQuery", t, func() {
		client := Client{
			ProjectId: "chrome-test-health-project",
			DataSet:   "normal-dataset",
		}
		request := &api.FetchDirectoryMetricsRequest{
			ParentIds: []string{"/"},
			Components: []string{
				"Blink",
			},
			Dates: []string{
				"2023-07-12",
			},
			Period: api.Period_DAY,
			Metrics: []api.MetricType{
				api.MetricType_NUM_RUNS,
			},
			Sort: &api.SortBy{
				Metric:    api.SortType_SORT_NAME,
				Ascending: true,
			},
			Filter: "linux-rel",
		}

		Convey("Valid unfiltered request", func() {
			query, err := client.createDirectoryQuery(request)

			So(err, ShouldBeNil)
			So(query, ShouldNotBeNil)
			So(query.QueryConfig.Q, ShouldResemble, `
WITH
test_summaries AS (
	SELECT
		file_name AS node_name,
		date,
		component AS test_component,
		--metrics
		SUM(num_runs) AS num_runs,
	FROM chrome-test-health-project.normal-dataset.daily_test_metrics
	WHERE
		date IN UNNEST(@dates)
		AND file_name IS NOT NULL
		AND component IN UNNEST(@components)
		-- Apply the requested filter
		AND REGEXP_CONTAINS(CONCAT('id:', test_id, ' ', 'name:', IFNULL(test_name, ''), ' ', 'file:', IFNULL(file_name, ''), ' ', 'bucket:', IFNULL(bucket, ''), '/', IFNULL(builder, ''), 'builder:', IFNULL(builder, ''), ' ', 'test_suite:', IFNULL(test_suite, '')), @filter0)
	GROUP BY file_name, date, test_id, component
)
SELECT
	f.date,
	f.node_name,
	ARRAY_REVERSE(SPLIT(f.node_name, '/'))[SAFE_OFFSET(0)] AS display_name,
	ANY_VALUE(is_file) AS is_file,
	-- metrics
	SUM(t.num_runs) AS num_runs,
FROM chrome-test-health-project.normal-dataset.daily_file_metrics AS f, UNNEST(@parents) AS parent
JOIN test_summaries t ON
	f.date = t.date
	AND t.test_component = f.component
	AND STARTS_WITH(t.node_name, f.node_name)
WHERE
	((STARTS_WITH(f.node_name, parent || "/")
	-- The child folders and files can't have a / after the parent's name
	AND REGEXP_CONTAINS(SUBSTR(f.node_name, LENGTH(parent) + 2), "^[^/]*$"))
	OR (parent = '' AND NOT STARTS_WITH(f.node_name, "/")))
	AND DATE(f.date) IN UNNEST(@dates)
		AND component IN UNNEST(@components)
GROUP BY date, node_name
ORDER BY is_file, node_name ASC`)
		})

		Convey("Valid no component request", func() {
			request.Components = []string{}
			query, err := client.createDirectoryQuery(request)

			So(err, ShouldBeNil)
			So(query, ShouldNotBeNil)
			So(query.QueryConfig.Q, ShouldResemble, `
WITH
test_summaries AS (
	SELECT
		file_name AS node_name,
		date,
		component AS test_component,
		--metrics
		SUM(num_runs) AS num_runs,
	FROM chrome-test-health-project.normal-dataset.daily_test_metrics
	WHERE
		date IN UNNEST(@dates)
		AND file_name IS NOT NULL
		-- Apply the requested filter
		AND REGEXP_CONTAINS(CONCAT('id:', test_id, ' ', 'name:', IFNULL(test_name, ''), ' ', 'file:', IFNULL(file_name, ''), ' ', 'bucket:', IFNULL(bucket, ''), '/', IFNULL(builder, ''), 'builder:', IFNULL(builder, ''), ' ', 'test_suite:', IFNULL(test_suite, '')), @filter0)
	GROUP BY file_name, date, test_id, component
)
SELECT
	f.date,
	f.node_name,
	ARRAY_REVERSE(SPLIT(f.node_name, '/'))[SAFE_OFFSET(0)] AS display_name,
	ANY_VALUE(is_file) AS is_file,
	-- metrics
	SUM(t.num_runs) AS num_runs,
FROM chrome-test-health-project.normal-dataset.daily_file_metrics AS f, UNNEST(@parents) AS parent
JOIN test_summaries t ON
	f.date = t.date
	AND t.test_component = f.component
	AND STARTS_WITH(t.node_name, f.node_name)
WHERE
	((STARTS_WITH(f.node_name, parent || "/")
	-- The child folders and files can't have a / after the parent's name
	AND REGEXP_CONTAINS(SUBSTR(f.node_name, LENGTH(parent) + 2), "^[^/]*$"))
	OR (parent = '' AND NOT STARTS_WITH(f.node_name, "/")))
	AND DATE(f.date) IN UNNEST(@dates)
GROUP BY date, node_name
ORDER BY is_file, node_name ASC`)
		})

		Convey("Valid unfiltered multi-day request", func() {
			request.Dates = []string{
				"2023-07-12",
				"2023-07-13",
			}
			request.Sort.SortDate = "2023-07-13"
			request.Sort.Ascending = false

			query, err := client.createDirectoryQuery(request)

			So(err, ShouldBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "sort_date",
				Value: "2023-07-13",
			})
			So(query, ShouldNotBeNil)
			So(query.QueryConfig.Q, ShouldResemble, `
WITH
test_summaries AS (
	SELECT
		file_name AS node_name,
		date,
		component AS test_component,
		--metrics
		SUM(num_runs) AS num_runs,
	FROM chrome-test-health-project.normal-dataset.daily_test_metrics
	WHERE
		date IN UNNEST(@dates)
		AND file_name IS NOT NULL
		AND component IN UNNEST(@components)
		-- Apply the requested filter
		AND REGEXP_CONTAINS(CONCAT('id:', test_id, ' ', 'name:', IFNULL(test_name, ''), ' ', 'file:', IFNULL(file_name, ''), ' ', 'bucket:', IFNULL(bucket, ''), '/', IFNULL(builder, ''), 'builder:', IFNULL(builder, ''), ' ', 'test_suite:', IFNULL(test_suite, '')), @filter0)
	GROUP BY file_name, date, test_id, test_component
), node_summaries AS (
	SELECT
		f.date,
		f.node_name,
		ARRAY_REVERSE(SPLIT(f.node_name, '/'))[SAFE_OFFSET(0)] AS display_name,
		ANY_VALUE(is_file) AS is_file,
		-- Metrics
		SUM(t.num_runs) AS num_runs,
	FROM chrome-test-health-project.normal-dataset.daily_file_metrics AS f, UNNEST(@parents) AS parent
	JOIN test_summaries t ON
		f.date = t.date
		AND f.component = t.test_component
		AND STARTS_WITH(t.node_name, f.node_name)
	WHERE
		((STARTS_WITH(f.node_name, parent || "/")
		-- The child folders and files can't have a / after the parent's name
		AND REGEXP_CONTAINS(SUBSTR(f.node_name, LENGTH(parent) + 2), "^[^/]*$"))
		OR (parent = '' AND NOT STARTS_WITH(f.node_name, "/")))
		AND DATE(f.date) IN UNNEST(@dates)
		AND component IN UNNEST(@components)
	GROUP BY date, node_name
), sorted_day AS (
	SELECT
		node_name,
		node_name AS rank
	FROM node_summaries
	WHERE date = @sort_date
)

SELECT node_summaries.*
FROM node_summaries FULL OUTER JOIN sorted_day USING(node_name)
ORDER BY is_file, rank IS NULL, rank DESC`)
		})

		Convey("Parameterized args", func() {
			query, err := client.createDirectoryQuery(request)

			So(err, ShouldBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "components",
				Value: []string{"Blink"},
			})
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "parents",
				Value: []string{"/"},
			})
		})

		Convey("Parameterized dates arg", func() {
			request.Dates = []string{
				"2023-07-12",
				"2023-07-13",
			}

			query, err := client.createDirectoryQuery(request)

			So(err, ShouldBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name: "dates",
				Value: []civil.Date{
					{
						Year:  2023,
						Month: 7,
						Day:   12,
					},
					{
						Year:  2023,
						Month: 7,
						Day:   13,
					},
				},
			})
		})

		Convey("Parameterized filter arg", func() {
			request.Filter = "linux-rel blink_python_tests"

			query, err := client.createDirectoryQuery(request)

			So(err, ShouldBeNil)
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "filter0",
				Value: "linux-rel",
			})
			So(query.Parameters, ShouldContainParameter, bigquery.QueryParameter{
				Name:  "filter1",
				Value: "blink_python_tests",
			})
		})

		Convey("Invalid sort metric returns error", func() {
			request.Sort = &api.SortBy{Metric: 99}

			_, err := client.createDirectoryQuery(request)

			So(err, ShouldNotBeNil)
		})
	})
}
