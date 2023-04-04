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
        (SELECT t.value FROM tr.tags t WHERE t.key = 'monorail_component' LIMIT 1) AS team,
        expected,
        exonerated,
        duration,
      FROM chrome-luci-data.chromium.try_test_results as tr
      WHERE DATE(partition_time) = @run_date
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
        ANY_VALUE(team) AS team,
        COUNT(*) AS run_count,
        COUNTIF(tr.expected AND NOT tr.exonerated) AS fail_count,
        AVG(tr.duration) AS avg_duration,
        SUM(tr.duration) AS total_duration,
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
      WHERE DATE(partition_time) = @run_date
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
        COUNTIF(undetected_flake OR detected_flake) AS flake_count,
      FROM patchset_flakes
      GROUP BY variant_hash, test_id, date
    )
  SELECT
    t.date,
    t.test_id,
    t.variant_hash,
    t.test_name,
    t.repo,
    t.file_name,
    t.`project`,
    t.bucket,
    t.builder,
    t.test_suite,
    t.target_platform,
    t.team,
    t.run_count,
    t.fail_count,
    f.flake_count,
    t.avg_duration,
    t.total_duration,
  FROM tests AS t
  INNER JOIN flakes AS f
    USING (variant_hash, test_id, date)
  ) AS S
ON
  T.date = S.date
  AND T.test_name = S.test_name
  AND T.test_id = S.test_id
  AND T.variant_hash = S.variant_hash
  AND T.repo = S.repo
  AND T.file_name = S.file_name
  AND T.`project` = S.`project`
  AND T.bucket = S.bucket
  AND T.builder = S.builder
  AND T.test_suite = S.test_suite
  AND T.target_platform = S.target_platform
  AND T.team = S.team
WHEN MATCHED THEN
  UPDATE SET
    run_count = S.run_count,
    fail_count = S.fail_count,
    flake_count = S.flake_count,
    avg_duration = S.avg_duration,
    total_duration = S.total_duration
WHEN NOT MATCHED THEN
  INSERT ROW