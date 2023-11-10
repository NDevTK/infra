
CREATE TEMP FUNCTION dedupe(file_names Array<string>)
RETURNS ARRAY<STRING>
LANGUAGE js
AS r"""
  Array.from(new Set(file_names));
""";

MERGE INTO {project}.{dataset}.weekly_file_metrics AS T
USING (
  -- Get all the relevant summaries for the week
  SELECT
    DATE_TRUNC(date, WEEK(SUNDAY)) AS `date`,
    repo,
    component,
    node_name,
    ANY_VALUE(is_file) AS is_file,
    SUM(num_runs) AS num_runs,
    SUM(num_failures) AS num_failures,
    SUM(num_flake) AS num_flake,
    SUM(total_runtime) AS total_runtime,
    -- Use weighted averages for aggregate quantiles
    SUM(avg_runtime * num_runs) / SUM(num_runs) avg_runtime,
    SUM(p50_runtime * num_runs) / SUM(num_runs) p50_runtime,
    SUM(p90_runtime * num_runs) / SUM(num_runs) p90_runtime,
  FROM {project}.{dataset}.daily_file_metrics
  WHERE `date` BETWEEN
    -- The date range is inclusive so only go up to the Saturday
    DATE_TRUNC(DATE(@from_date), WEEK) AND
    DATE_ADD(DATE_TRUNC(DATE(@to_date), WEEK), INTERVAL 6 DAY)
  GROUP BY DATE_TRUNC(date, WEEK(SUNDAY)), component, node_name, repo
  ) AS S
ON
  T.date = S.date
  AND T.date BETWEEN
    DATE_TRUNC(DATE(@from_date), WEEK) AND
    DATE_ADD(DATE_TRUNC(DATE(@to_date), WEEK), INTERVAL 6 DAY)
  AND T.node_name = S.node_name
  AND (T.component = S.component OR (T.component IS NULL AND S.component IS NULL))
  AND (T.repo = S.repo OR (T.repo IS NULL AND S.repo IS NULL))
WHEN MATCHED THEN
  UPDATE SET
    num_runs = S.num_runs,
    num_failures = S.num_failures,
    num_flake = S.num_flake,
    avg_runtime = S.avg_runtime,
    total_runtime = S.total_runtime
WHEN NOT MATCHED THEN
  INSERT (`date`, repo, component, node_name, is_file, num_runs, num_failures, num_flake, total_runtime, avg_runtime, p50_runtime, p90_runtime)
  VALUES (`date`, repo, component, node_name, is_file, num_runs, num_failures, num_flake, total_runtime, avg_runtime, p50_runtime, p90_runtime)
