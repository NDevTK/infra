
CREATE TEMP FUNCTION dedupe(file_names Array<string>)
RETURNS ARRAY<STRING>
LANGUAGE js
AS r"""
  Array.from(new Set(file_names));
""";

MERGE INTO %s.%s.weekly_file_metrics AS T
USING (
  SELECT
    DATE_TRUNC(date, WEEK(SUNDAY)) AS `date`,
    component,
    repo,
    node_name,
    ANY_VALUE(is_file) AS is_file,
    SUM(num_runs) AS num_runs,
    SUM(num_failures) AS num_failures,
    SUM(num_flake) AS num_flake,
    AVG(avg_runtime)	AS avg_runtime,
    SUM(total_runtime) AS total_runtime,
    dedupe(ARRAY_CONCAT_AGG(file_names)) AS file_names,
  FROM %s.%s.file_metrics
  WHERE `date` BETWEEN
    -- The date range is inclusive so only go up to the Saturday
    DATE_TRUNC(DATE(@from_date), WEEK) AND
    DATE_ADD(DATE_TRUNC(DATE(@to_date), WEEK), INTERVAL 6 DAY)
  GROUP BY DATE_TRUNC(date, WEEK(SUNDAY)), component, node_name, repo
  ) AS S
ON
  T.date = S.date
  AND T.node_name = S.node_name
  AND (T.repo = S.repo OR (T.repo IS NULL AND S.repo IS NULL))
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
