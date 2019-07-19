# To detect disabled tests within the past 1 day.
#
# Assumptions for the disabled tests in this query are:
# 1. A disabled test should have 'SKIP' in run.expected
SELECT
  DISTINCT path AS test_name,
  step_name,
  buildbot_info.builder_name,
  bug
FROM
  `test-results-hrd.events.test_results`
LEFT JOIN
    UNNEST(run.bugs) AS bug
WHERE
  # According to https://cloud.google.com/bigquery/docs/partitioned-tables,
  # _PARTITIONTIME is always the start of each day, so to make sure all data
  # within the past 1 day is included, _PARTITIONTIME needs to be greater
  # than or equal to the timestamp of 2 days ago.
  _PARTITIONTIME >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 2 day)
  AND 'SKIP' IN UNNEST(run.expected)
