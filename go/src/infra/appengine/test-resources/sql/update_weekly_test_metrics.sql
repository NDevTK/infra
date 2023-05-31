MERGE INTO %s.%s.weekly_test_metrics AS T
USING (
  WITH aggregated_variants AS (
    SELECT
      DATE_TRUNC(m.date, WEEK(SUNDAY)) date,
      m.test_id,
      ANY_VALUE(m.test_name) AS test_name,
      m.repo AS repo,
      ANY_VALUE(m.file_name) AS file_name,
      ANY_VALUE(m.component) AS component,
      # Test level metrics
      SUM(m.num_runs) AS num_runs,
      SUM(m.num_failures) AS num_failures,
      SUM(m.num_flake) AS num_flake,
      SUM(m.avg_runtime) AS avg_runtime,
      SUM(m.total_runtime) AS total_runtime,
      ARRAY_CONCAT_AGG(m.variant_summaries) AS aggregate_variants,
    FROM %s.%s.test_metrics m
    WHERE m.date BETWEEN
      -- The date range is inclusive so only go up to the Saturday
      DATE_TRUNC(DATE(@from_date), WEEK) AND
      DATE_ADD(DATE_TRUNC(DATE(@to_date), WEEK), INTERVAL 6 DAY)
    GROUP BY DATE_TRUNC(m.date, WEEK(SUNDAY)), test_id, repo
  )

  SELECT a.* EXCEPT (aggregate_variants),
    (
      SELECT ARRAY_AGG(variant_summary) FROM (
        SELECT
          STRUCT(
            v.variant_hash AS variant_hash,
            v.`project` AS `project`,
            v.bucket AS bucket,
            ANY_VALUE(v.target_platform) AS target_platform,
            ANY_VALUE(v.builder) AS builder,
            ANY_VALUE(v.test_suite) AS test_suite,
            SUM(v.num_runs) AS num_runs,
            SUM(v.num_failures) AS num_failures,
            SUM(v.num_flake) AS num_flake,
            AVG(v.avg_runtime) AS avg_runtime,
            SUM(v.total_runtime) AS total_runtime
          ) AS variant_summary
        FROM a.aggregate_variants v
        -- bucket and project are not part of the variant hash
        GROUP BY v.variant_hash, v.bucket, v.`project`
        )) AS variant_summaries
  FROM aggregated_variants a

  ) AS S
ON
  T.date = S.date
  AND T.test_name = S.test_name
  AND T.test_id = S.test_id
  AND T.repo = S.repo
  AND T.component = S.component
WHEN MATCHED THEN
  UPDATE SET
    num_runs = S.num_runs,
    num_failures = S.num_failures,
    num_flake = S.num_flake,
    avg_runtime = S.avg_runtime,
    total_runtime = S.total_runtime,
    variant_summaries = S.variant_summaries
WHEN NOT MATCHED THEN
  INSERT ROW
