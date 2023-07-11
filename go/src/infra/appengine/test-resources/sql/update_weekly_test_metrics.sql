MERGE INTO %s.%s.weekly_test_metrics AS T
USING (
  SELECT
    DATE_TRUNC(m.date, WEEK(SUNDAY)) date,
    m.test_id,
    ANY_VALUE(m.test_name) AS test_name,
    m.repo AS repo,
    ANY_VALUE(m.file_name) AS file_name,
    ANY_VALUE(m.component) AS component,
    builder,
    test_suite,
    -- Test level metrics
    SUM(m.num_runs) AS num_runs,
    SUM(m.num_failures) AS num_failures,
    SUM(m.num_flake) AS num_flake,
    SUM(m.total_runtime) AS total_runtime,
    -- Weighted averages
    SUM(avg_runtime * num_runs) / SUM(num_runs) avg_runtime,
    SUM(p50_runtime * num_runs) / SUM(num_runs) p50_runtime,
    SUM(p90_runtime * num_runs) / SUM(num_runs) p90_runtime,
  FROM %s.%s.daily_test_metrics m
  WHERE m.date BETWEEN
    -- The date range is inclusive so only go up to the Saturday
    DATE_TRUNC(DATE(@from_date), WEEK) AND
    DATE_ADD(DATE_TRUNC(DATE(@to_date), WEEK), INTERVAL 6 DAY)
  GROUP BY DATE_TRUNC(m.date, WEEK(SUNDAY)), test_id, repo, builder, test_suite
  ) AS S
ON
  T.date = S.date
  AND T.test_id = S.test_id
  AND (T.repo = S.repo OR (T.repo IS NULL AND S.repo IS NULL))
  AND (T.builder = S.builder OR (T.builder IS NULL AND S.builder IS NULL))
  AND (T.test_suite = S.repo OR (T.test_suite IS NULL AND S.test_suite IS NULL))
WHEN MATCHED THEN
  UPDATE SET
    num_runs = S.num_runs,
    num_failures = S.num_failures,
    num_flake = S.num_flake,
    avg_runtime = S.avg_runtime,
    total_runtime = S.total_runtime
WHEN NOT MATCHED THEN
  INSERT ROW
