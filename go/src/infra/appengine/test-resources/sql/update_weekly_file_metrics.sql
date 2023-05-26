
CREATE TEMP FUNCTION dedupe(file_names Array<string>)
RETURNS ARRAY<STRING>
LANGUAGE js
AS r"""
  Array.from(new Set(file_names));
""";

MERGE INTO %s.test_results.weekly_file_metrics AS T
USING (
  SELECT
    DATE_TRUNC(date, WEEK(SUNDAY)) AS date,
    component,
    node_name,
    ANY_VALUE(is_file) AS is_file,
    SUM(num_runs) AS num_runs,
    SUM(num_failures) AS num_failures,
    SUM(num_flake) AS num_flake,
    AVG(avg_runtime)	AS avg_runtime,
    SUM(total_runtime) AS total_runtime,
    dedupe(ARRAY_CONCAT_AGG(file_names)) AS file_names,
  FROM %s.test_results.file_metrics
  WHERE date BETWEEN
    -- The date range is inclusive so only go up to the Saturday
    DATE_TRUNC(DATE(@from_date), WEEK) AND
    DATE_ADD(DATE_TRUNC(DATE(@to_date), WEEK), INTERVAL 6 DAY)
  GROUP BY DATE_TRUNC(date, WEEK(SUNDAY)), component, node_name
  ) AS S
ON
  T.date = S.date
  AND T.component = S.component
  AND T.node_name = S.node_name
WHEN MATCHED THEN
  UPDATE SET
    num_runs = S.num_runs,
    num_failures = S.num_failures,
    num_flake = S.num_flake,
    avg_runtime = S.avg_runtime,
    total_runtime = S.total_runtime,
    file_names = S.file_names
WHEN NOT MATCHED THEN
  INSERT ROW
