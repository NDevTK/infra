# To detect disabled tests within the past 1 day.
#
# Assumptions for the disabled tests in this query are:
# 1. A disabled test should have 'SKIP' in run.expected
WITH
  last_runs AS (
  SELECT
    path AS test_name,
    step_name,
    buildbot_info.builder_name,
    ARRAY_AGG( STRUCT(build_id,
        run)
    ORDER BY
      start_time DESC
    LIMIT
      1)[
  OFFSET
    (0)] latest_run
  FROM
    `test-results-hrd.events.test_results` AS tr
  WHERE
  # According to https://cloud.google.com/bigquery/docs/partitioned-tables,
  # _PARTITIONTIME is always the start of each day, so to make sure all data
  # within the past 1 day is included, _PARTITIONTIME needs to be greater
  # than or equal to the timestamp of 2 days ago.
    _PARTITIONTIME >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 2 day)
  GROUP BY
    1,
    2,
    3)
SELECT
  test_name,
  step_name,
  builder_name,
  latest_run.build_id,
  latest_run.run.bugs
FROM
  last_runs
WHERE
  'SKIP' IN UNNEST(latest_run.run.expected)
