CREATE OR REPLACE VIEW `APP_ID.PROJECT_NAME.failing_steps`
AS
/*
Failing steps table.
Each row represents a step that has failed in the most recent run of the
given builder (bucket, project etc).
As the status of the build system changes, so should the contents of this
view.
*/
WITH
  latest_builds AS (
  SELECT
    project,
    bucket,
    builder,
    buildergroup,
    ARRAY_AGG(b
    ORDER BY
      # Latest, meaning sort by commit position if it exists, otherwise by the build id or number.
      b.output_commit.position DESC, id, number DESC
    LIMIT
      1)[
  OFFSET
    (0)] latest
  FROM
    `sheriff-o-matic.materialized.buildbucket_completed_builds_prod` AS b
  WHERE
    PROJECT_FILTER_CONDITIONS
  GROUP BY
    1,
    2,
    3,
    4),
  recent_tests AS (
    SELECT
    project,
    SUBSTR(r.ingested_invocation_id, 7) AS build_id,
    r.test_id,
    r.realm,
    r.variant_hash,
    r.source_ref_hash,
    ANY_VALUE(COALESCE((SELECT value FROM UNNEST(r.tags) WHERE key = "test_name"), r.test_id)) AS test_name,
    ANY_VALUE((SELECT value FROM UNNEST(r.tags) WHERE key = "step_name" limit 1)) as step_name,
    -- we prefix 'rules' algorithms with 'a' and others with 'b' so that MIN chooses clusters in order of [rules, reason, testname].
    CONCAT(ANY_VALUE(project), '/', SUBSTR(MIN(CONCAT(
        IF
          (STARTS_WITH(cluster_algorithm, 'rule'),
            'a',
            'b'), cluster_algorithm, '/', cluster_id)), 2)) AS cluster_name
  FROM (
      SELECT
        project,
        cluster_algorithm,
        cluster_id,
        test_result_system,
        test_result_id,
        DATE(partition_time) as partition_time,
        ARRAY_AGG(STRUCT(
          ingested_invocation_id,
          test_id,
          is_ingested_invocation_blocked,
          exonerations,
          is_included,
          tags,
          realm,
          variant_hash,
          source_ref_hash
        ) ORDER BY last_updated DESC LIMIT 1)[OFFSET(0)] as r
      FROM `luci-analysis.internal.clustered_failures` cf
      WHERE
        partition_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 1 DAY)
        AND project in ("chrome", "chromeos", "chromium", "turquoise", "fuchsia") -- TODO: should only query the LUCI Analysis projects relevant to the tree rather than all of them.
      GROUP BY project, cluster_algorithm, cluster_id, test_result_system, test_result_id, DATE(partition_time)
      )
  WHERE r.is_included
    AND r.is_ingested_invocation_blocked
    AND ARRAY_LENGTH(r.exonerations) = 0
  GROUP BY
    project,
    r.ingested_invocation_id,
    r.test_id,
    r.realm,
    r.variant_hash,
    r.source_ref_hash),
changepoint_analysis_unexpected_realtime AS(
    SELECT
      project, test_id, variant_hash, ref_hash,
      ARRAY_AGG(merged ORDER BY version DESC LIMIT 1)[OFFSET(0)] AS row
    FROM (
        SELECT cp.*
        FROM luci-analysis.internal.test_variant_segments cp
        WHERE cp.project IN ("chrome", "chromeos", "chromium", "turquoise", "fuchsia") AND has_recent_unexpected_results = 1
        UNION ALL
        SELECT cp.*
        FROM luci-analysis.internal.test_variant_segment_updates cp
        WHERE cp.project IN ("chrome", "chromeos", "chromium", "turquoise", "fuchsia") AND has_recent_unexpected_results = 1
      ) merged
    GROUP BY project, test_id, variant_hash, ref_hash
),
recent_tests_with_changepoint_analysis as (
  SELECT tr.*, cp.row.segments
  FROM recent_tests tr
  LEFT JOIN changepoint_analysis_unexpected_realtime cp
      ON  tr.project = cp.project
      AND tr.test_id = cp.test_id
      AND tr.variant_hash = cp.variant_hash
      AND tr.source_ref_hash = cp.ref_hash
)
SELECT
  b.project,
  bucket,
  builder,
  latest.number,
  b.latest.id build_id,
  latest.buildergroup as buildergroup,
  s.name step,
  ANY_VALUE(b.latest.sheriff_rotations) as sheriff_rotations,
  ANY_VALUE(b.latest.status) status,
  ANY_VALUE(b.latest.critical) critical,
  ANY_VALUE(b.latest.output_commit) output_commit,
  ANY_VALUE(b.latest.input_commit) input_commit,
  FARM_FINGERPRINT(STRING_AGG(tr.test_name, "\n"
    ORDER BY
      tr.test_name)) AS test_names_fp,
  -- TODO: remove test_names_trunc field once tests_trunc field is used in production.
  STRING_AGG(tr.test_name, "\n"
    ORDER BY
      tr.test_name
    LIMIT
      40) AS test_names_trunc,
  ARRAY_AGG(STRUCT(
    tr.test_name as TestName,
    tr.test_id as TestID,
    tr.realm as Realm,
    tr.variant_hash as VariantHash,
    tr.cluster_name as ClusterName,
    tr.source_ref_hash as RefHash,
    tr.segments as Segments)
    ORDER BY
      tr.test_name
    LIMIT
      40) AS tests_trunc,
  COUNT(tr.test_name) AS num_tests
FROM
  latest_builds b,
  b.latest.steps s
LEFT OUTER JOIN
  recent_tests_with_changepoint_analysis tr
ON
  SAFE_CAST(tr.build_id AS int64) = b.latest.id
  AND (tr.step_name IS NULL OR tr.step_name = s.name)
WHERE
  (b.latest.status = 'FAILURE' AND s.status = 'FAILURE')
  OR
  (
    b.latest.status = 'INFRA_FAILURE'
    AND (s.status = 'INFRA_FAILURE' OR s.status = 'CANCELED')
  )
GROUP BY
  1,
  2,
  3,
  4,
  5,
  6,
  7
