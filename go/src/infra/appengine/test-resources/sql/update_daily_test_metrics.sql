MERGE INTO %s.%s.daily_test_metrics AS T
USING (
  SELECT
    t.date,

    -- test level info
    t.test_id,
    ANY_VALUE(t.test_name) AS test_name,
    t.repo,
    ANY_VALUE(t.file_name) AS file_name,
    t.component,

    -- variant level info
    t.builder,
    t.test_suite,

    -- metrics
    SUM(num_runs) num_runs,
    SUM(num_failures) num_failures,
    SUM(num_flake) num_flake,
    SUM(total_runtime) total_runtime,
    -- Weighted averages
    SUM(avg_runtime * num_runs) / SUM(num_runs) avg_runtime,
    SUM(p50_runtime * num_runs) / SUM(num_runs) p50_runtime,
    SUM(p90_runtime * num_runs) / SUM(num_runs) p90_runtime,
  FROM %s.%s.raw_metrics AS t
  GROUP BY `date`, test_id, repo, component, builder, test_suite
  ) AS S
ON
  T.date = S.date
  AND T.test_id = S.test_id
  AND (T.repo = S.repo OR (T.repo IS NULL AND S.repo IS NULL))
  AND (T.component = S.component OR (T.component IS NULL AND S.component IS NULL))
  AND (T.builder = S.builder OR (T.builder IS NULL AND S.builder IS NULL))
  AND (T.test_suite = S.test_suite OR (T.test_suite IS NULL AND S.test_suite IS NULL))
WHEN MATCHED THEN
  UPDATE SET
    test_name = S.test_name,
    file_name = S.file_name,
    component = S.component,
    num_runs = S.num_runs,
    num_failures = S.num_failures,
    num_flake = S.num_flake,
    avg_runtime = S.avg_runtime,
    total_runtime = S.total_runtime
WHEN NOT MATCHED THEN
  INSERT ROW
