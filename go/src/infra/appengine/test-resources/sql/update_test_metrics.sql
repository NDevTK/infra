MERGE INTO %s.test_results.test_metrics AS T
USING (
  WITH
    raw_results_tables AS (
      SELECT
        CAST(REGEXP_EXTRACT(exported.id, r'build-(\d+)') AS INT64) AS build_id,
        test_metadata.name AS test_name,
        partition_time,
        test_id,
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
      FROM chrome-luci-data.chromium.try_test_results as tr
      WHERE DATE(partition_time) BETWEEN @from_date AND @to_date
    ), tests AS (
      SELECT
        DATE(tr.partition_time) AS date,
        test_name,
        test_id,
        repo,
        file_name,
        `project`,
        bucket,
        builder,
        test_suite,
        target_platform,
        variant_hash,
        ANY_VALUE(component) AS component,
        COUNT(*) AS num_runs,
        COUNTIF(NOT tr.expected AND NOT tr.exonerated) AS num_failures,
        AVG(tr.duration) AS avg_runtime,
        SUM(tr.duration) AS total_runtime,
      FROM
        raw_results_tables AS tr
      GROUP BY
        date, test_name, test_id, repo, file_name, `project`, bucket, builder, test_suite,
        target_platform, variant_hash
    ), attempt_builds AS (
      SELECT
        DATE(partition_time) date,
        b.id AS build_id,
        ps.change,
        ps.earliest_equivalent_patchset AS patchset,
        partition_time AS ps_approx_timestamp,
      FROM `commit-queue`.chromium.attempts a, a.gerrit_changes ps, a.builds b
      WHERE DATE(partition_time) BETWEEN @from_date AND @to_date
    ), patchset_flakes AS (
      SELECT
        date,
        tr.test_id,
        tr.variant_hash,
        LOGICAL_OR(expected) AND LOGICAL_OR(NOT expected) AS undetected_flake,
        LOGICAL_OR(exonerated) AS detected_flake,
      FROM raw_results_tables AS tr
      INNER JOIN attempt_builds AS b
        USING (build_id)
      GROUP BY change, patchset, tr.variant_hash, tr.test_id, date
    ), flakes AS (
      SELECT
        date,
        test_id,
        variant_hash,
        COUNTIF(undetected_flake OR detected_flake) AS num_flake,
      FROM patchset_flakes
      GROUP BY variant_hash, test_id, date
    ), variant_summaries AS (
      SELECT
        t.date,
        t.test_id,
        t.test_name,
        t.repo,
        t.file_name,
        t.`project`,
        t.bucket,
        t.component,
        t.num_runs,
        t.num_failures,
        f.num_flake,
        t.avg_runtime,
        t.total_runtime,
        t.variant_hash,
        t.builder,
        t.test_suite,
        t.target_platform,
      FROM tests AS t
      INNER JOIN flakes AS f
        USING (variant_hash, test_id, date)
    )
  SELECT
    v.date,
    v.test_id,
    ANY_VALUE(v.test_name) AS test_name,
    ANY_VALUE(v.repo) AS repo,
    ANY_VALUE(v.file_name) AS file_name,
    ANY_VALUE(v.`project`) AS `project`,
    ANY_VALUE(v.bucket) AS bucket,
    ANY_VALUE(v.component) AS component,
    # Test level metrics
    SUM(num_runs) AS num_runs,
    SUM(num_failures) AS num_failures,
    SUM(num_flake) AS num_flake,
    SUM(avg_runtime) AS avg_runtime,
    SUM(total_runtime) AS total_runtime,
    ARRAY_AGG(STRUCT(
      v.variant_hash AS variant_hash,
      v.target_platform AS target_platform,
      v.builder AS builder,
      v.test_suite AS test_suite,
      v.num_runs AS num_runs,
      v.num_failures AS num_failures,
      v.num_flake AS num_flake,
      v.avg_runtime AS avg_runtime,
      v.total_runtime AS total_runtime
    )) AS variant_summaries
  FROM variant_summaries v
  GROUP BY date, test_id
  ) AS S
ON
  T.date = S.date
  AND T.test_name = S.test_name
  AND T.test_id = S.test_id
  AND T.repo = S.repo
  AND T.file_name = S.file_name
  AND T.`project` = S.`project`
  AND T.bucket = S.bucket
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
