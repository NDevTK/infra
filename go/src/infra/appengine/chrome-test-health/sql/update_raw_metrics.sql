MERGE INTO {project}.{dataset}.raw_metrics AS T
USING (
  WITH
    raw_results_tables AS (
      SELECT
        CAST(REGEXP_EXTRACT(exported.id, r'build-(\d+)') AS INT64) AS build_id,
        -- Remove any parameterized portion of the name. These consistently end
        -- in a backlash with a number (e.g. foo/0, bar/1, etc.)
        REGEXP_REPLACE(test_metadata.name, '/[0-9]+$', '') AS test_name,
        partition_time,
        -- Remove any parameterized portion of the test_id. These end in /0
        -- or /All.0
        REGEXP_REPLACE(test_id, r'[/\.][0-9]+$', '') AS test_id,
        test_metadata.location.repo AS repo,
        test_metadata.location.file_name AS file_name,
        SPLIT(exported.realm, ':')[SAFE_OFFSET(0)] AS `project`,
        SPLIT(exported.realm, ':')[SAFE_OFFSET(1)] AS bucket,
        (SELECT v.value FROM tr.variant AS v WHERE v.key = 'builder' LIMIT 1) AS builder,
        (SELECT v.value FROM tr.variant AS v WHERE v.key = 'test_suite' LIMIT 1) AS test_suite,
        (SELECT t.value FROM tr.tags AS t WHERE t.key = 'target_platform' LIMIT 1) AS target_platform,
        variant_hash,
        (SELECT t.value FROM tr.tags t WHERE t.key = 'monorail_component' LIMIT 1) AS component,
        expected,
        exonerated,
        duration,
      FROM {chromium_try_rdb_table} as tr
      WHERE
        DATE(partition_time, "PST8PDT") BETWEEN @from_date AND @to_date
        AND tr.status != "SKIP"
      UNION ALL
      SELECT
        CAST(REGEXP_EXTRACT(exported.id, r'build-(\d+)') AS INT64) AS build_id,
        REGEXP_REPLACE(test_metadata.name, '/[0-9]+$', '') AS test_name,
        partition_time,
        REGEXP_REPLACE(test_id, r'[/\.][0-9]+$', '') AS test_id,
        test_metadata.location.repo AS repo,
        test_metadata.location.file_name AS file_name,
        SPLIT(exported.realm, ':')[SAFE_OFFSET(0)] AS `project`,
        SPLIT(exported.realm, ':')[SAFE_OFFSET(1)] AS bucket,
        (SELECT v.value FROM tr.variant AS v WHERE v.key = 'builder' LIMIT 1) AS builder,
        (SELECT v.value FROM tr.variant AS v WHERE v.key = 'test_suite' LIMIT 1) AS test_suite,
        (SELECT t.value FROM tr.tags AS t WHERE t.key = 'target_platform' LIMIT 1) AS target_platform,
        variant_hash,
        (SELECT t.value FROM tr.tags t WHERE t.key = 'monorail_component' LIMIT 1) AS component,
        expected,
        exonerated,
        duration,
      FROM {chromium_ci_rdb_table} as tr
      WHERE
        DATE(partition_time, "PST8PDT") BETWEEN @from_date AND @to_date
        AND tr.status != "SKIP"
    ), tests AS (
      SELECT
        EXTRACT(DATE FROM partition_time AT TIME ZONE "PST8PDT") AS `date`,
        test_id,
        repo,
        `project`,
        bucket,
        builder,
        test_suite,
        target_platform,
        variant_hash,
        ANY_VALUE(IFNULL(test_name, test_id)) AS test_name,
        ANY_VALUE(file_name) AS file_name,
        IFNULL(ANY_VALUE(component), "Unknown") AS component,
        COUNT(*) AS num_runs,
        COUNTIF(NOT tr.expected AND NOT tr.exonerated) AS num_failures,
        AVG(tr.duration) AS avg_runtime,
        SUM(tr.duration) AS total_runtime,
        APPROX_QUANTILES(tr.duration, 1000) AS runtime_quantiles
      FROM
        raw_results_tables AS tr
      GROUP BY
        `date`, test_id, repo, `project`, bucket, builder, test_suite,
        target_platform, variant_hash
    ), attempt_builds AS (
      SELECT
        EXTRACT(DATE FROM start_time AT TIME ZONE "PST8PDT") AS `date`,
        b.id AS build_id,
        ps.change,
        ps.earliest_equivalent_patchset AS patchset,
      FROM {attempts_table} a, a.gerrit_changes ps, a.builds b
      WHERE DATE(start_time, "PST8PDT") BETWEEN @from_date AND @to_date
    ), patchset_flakes AS (
      SELECT
        `date`,
        tr.test_id,
        tr.variant_hash,
        LOGICAL_OR(expected) AND LOGICAL_OR(NOT expected) AS undetected_flake,
        LOGICAL_OR(exonerated) AS detected_flake,
      FROM raw_results_tables AS tr
      INNER JOIN attempt_builds AS b
        USING (build_id)
      GROUP BY change, patchset, tr.variant_hash, tr.test_id, `date`
    ), flakes AS (
      SELECT
        `date`,
        test_id,
        variant_hash,
        COUNTIF(undetected_flake OR detected_flake) AS num_flake,
      FROM patchset_flakes
      GROUP BY variant_hash, test_id, `date`
    )
    SELECT
      t.date,

      -- test level info
      t.test_id,
      t.test_name,
      t.repo,
      t.file_name,
      t.component,

      -- variant level info
      t.variant_hash,
      t.`project`,
      t.bucket,
      t.target_platform,
      t.builder,
      t.test_suite,

      -- metrics
      t.num_runs,
      t.num_failures,
      f.num_flake,
      t.avg_runtime,
      t.total_runtime,
      t.runtime_quantiles[500] p50_runtime,
      t.runtime_quantiles[900] p90_runtime,
    FROM tests AS t
    LEFT JOIN flakes AS f
      USING (variant_hash, test_id, date)
  ) AS S
ON
  T.date = S.date
  AND T.date BETWEEN @from_date AND @to_date
  AND T.test_id = S.test_id
  AND (T.repo = S.repo OR (T.repo IS NULL AND S.repo IS NULL))
  AND (T.component = S.component OR (T.component IS NULL AND S.component IS NULL))
  AND (T.variant_hash = S.variant_hash OR (T.variant_hash IS NULL AND S.variant_hash IS NULL))
  AND (T.`project` = S.`project` OR (T.`project` IS NULL AND S.`project` IS NULL))
  AND (T.bucket = S.bucket OR (T.bucket IS NULL AND S.bucket IS NULL))
  AND (T.target_platform = S.target_platform OR (T.target_platform IS NULL AND S.target_platform IS NULL))
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
  INSERT (`date`, test_id, test_name, file_name, repo, component, variant_hash, `project`, bucket, target_platform, builder, test_suite, num_runs, num_failures, num_flake, total_runtime, avg_runtime, p50_runtime, p90_runtime)
  VALUES (`date`, test_id, test_name, file_name, repo, component, variant_hash, `project`, bucket, target_platform, builder, test_suite, num_runs, num_failures, num_flake, total_runtime, avg_runtime, p50_runtime, p90_runtime)
