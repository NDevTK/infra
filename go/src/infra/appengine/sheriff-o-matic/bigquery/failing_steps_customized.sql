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
    SUBSTR(ingested_invocation_id, 7) AS build_id,
    realm,
    variant_hash,
    ANY_VALUE(COALESCE((SELECT value FROM UNNEST(tags) WHERE key = "test_name"), test_id)) AS test_name,
    ANY_VALUE((SELECT value FROM UNNEST(tags) WHERE key = "step_name" limit 1)) as step_name,
    -- we prefix 'rules' algorithms with 'a' and others with 'b' so that MIN chooses clusters in order of [rules, reason, testname].
    SUBSTR(MIN(CONCAT(
        IF
          (STARTS_WITH(cluster_algorithm, 'rule'),
            'a',
            'b'), cluster_algorithm, '/', cluster_id)), 2) AS cluster_names
  FROM ((
      SELECT
        ingested_invocation_id,
        test_id,
        cluster_algorithm,
        cluster_id,
        partition_time,
        is_ingested_invocation_blocked,
        exoneration_status,
        is_included,
        tags,
        realm,
        variant_hash
      FROM      
        `chops-weetbix.chromium.clustered_failures`)      
    UNION ALL (
      SELECT
        ingested_invocation_id,
        test_id,
        cluster_algorithm,
        cluster_id,
        partition_time,
        is_ingested_invocation_blocked,
        exoneration_status,
        is_included,
        tags,
        realm,
        variant_hash
      FROM
        `chops-weetbix.chrome.clustered_failures`)
    )
  WHERE
    partition_time > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 1 DAY)
    AND is_ingested_invocation_blocked
    AND exoneration_status = 'NOT_EXONERATED'
    AND is_included
  GROUP BY
    ingested_invocation_id,
    test_id,
    realm,
    variant_hash)
SELECT
  project,
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
  STRING_AGG(tr.test_name, "\n"
    ORDER BY
      tr.test_name
    LIMIT
      40) AS test_names_trunc,
  STRING_AGG(tr.realm, "\n"
    ORDER BY
      tr.test_name
    LIMIT
      40) AS test_realms_trunc,
  STRING_AGG(tr.variant_hash, "\n"
    ORDER BY
      tr.test_name
    LIMIT
      40) AS test_variant_hashes_trunc,
  ARRAY_AGG(COALESCE(cluster_names, '')
    ORDER BY
      tr.test_name
    LIMIT
      40) AS cluster_names,
  COUNT(tr.test_name) AS num_tests
FROM
  latest_builds b,
  b.latest.steps s
LEFT OUTER JOIN
  recent_tests tr
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
