
CREATE TEMP FUNCTION dedupe(file_names Array<string>)
RETURNS ARRAY<STRING>
LANGUAGE js
AS r"""
  Array.from(new Set(file_names));
""";

MERGE INTO %s.%s.weekly_file_metrics AS T
USING (
  -- Get all the relevant summaries for the week
  WITH agged_files AS(
    SELECT
      DATE_TRUNC(date, WEEK(SUNDAY)) AS `date`,
      repo,
      component,
      node_name,
      ANY_VALUE(is_file) AS is_file,
      SUM(num_runs) AS num_runs,
      SUM(num_failures) AS num_failures,
      SUM(num_flake) AS num_flake,
      AVG(avg_runtime)	AS avg_runtime,
      SUM(total_runtime) AS total_runtime,
      ARRAY_CONCAT_AGG(child_file_summaries) all_child_file_summaries
    FROM %s.%s.file_metrics
    WHERE `date` BETWEEN
      -- The date range is inclusive so only go up to the Saturday
      DATE_TRUNC(DATE(@from_date), WEEK) AND
      DATE_ADD(DATE_TRUNC(DATE(@to_date), WEEK), INTERVAL 6 DAY)
    GROUP BY DATE_TRUNC(date, WEEK(SUNDAY)), component, node_name, repo
  )
  -- Combine the child file summaries used to aggregate with filter
  SELECT * EXCEPT(all_child_file_summaries),
    (
      SELECT ARRAY_AGG(file_summary)
      FROM (
        SELECT
          file_name,
          SUM(c.num_runs) AS num_runs,
          SUM(c.num_failures) AS num_failures,
          SUM(c.num_flake) AS num_flake,
          AVG(c.avg_runtime)	AS avg_runtime,
          SUM(c.total_runtime) AS total_runtime
        FROM UNNEST(all_child_file_summaries) c
        GROUP BY file_name
    ) AS file_summary) AS child_file_summaries
  FROM agged_files
  ) AS S
ON
  T.date = S.date
  AND T.node_name = S.node_name
  AND T.component = S.component
  AND (T.repo = S.repo OR (T.repo IS NULL AND S.repo IS NULL))
WHEN MATCHED THEN
  UPDATE SET
    num_runs = S.num_runs,
    num_failures = S.num_failures,
    num_flake = S.num_flake,
    avg_runtime = S.avg_runtime,
    total_runtime = S.total_runtime,
    child_file_summaries = S.child_file_summaries
WHEN NOT MATCHED THEN
  INSERT ROW
