-- This query updates the 'Test Case Count' metric in the cq_builder_metrics_day
-- table. It uses an interval of the last 2 days so that there is some redundancy if the job fails
-- Test Case Count is the distinct number of tests ran in the test suite each day.
-- This query is meant to be almost identical to the one in cq_builder_metrics_week_test_counts.sql.

-- The lines below are used by the deploy tool.
--name: Populate cq_builder_metrics_day test case count
--schedule: every 4 hours synchronized

DECLARE start_date DATE DEFAULT DATE_SUB(CURRENT_DATE('PST8PDT'), INTERVAL 2 DAY);
-- This isn't really needed, but useful to have around when doing backfills
-- The end_date is exclusive, which is why we add a day to the current date.
DECLARE end_date DATE DEFAULT DATE_ADD(CURRENT_DATE('PST8PDT'), INTERVAL 1 DAY);
DECLARE start_ts TIMESTAMP DEFAULT TIMESTAMP(start_date, 'PST8PDT');
DECLARE end_ts TIMESTAMP DEFAULT TIMESTAMP(end_date, 'PST8PDT');

-- Merge statement
MERGE INTO
  `chrome-trooper-analytics.metrics.cq_builder_metrics_day` AS T
USING
  (
  WITH test_ids AS (
    SELECT
      EXTRACT(DATE FROM partition_time AT TIME ZONE "PST8PDT") AS `date`,
      MAX(partition_time) max_result_time,
      COUNT(DISTINCT tr.test_id) test_count,
      (SELECT v.value FROM tr.variant v WHERE v.key = 'builder') builder,
      (SELECT v.value FROM tr.variant v WHERE v.key = 'test_suite') test_suite,
    FROM chrome-luci-data.chromium.try_test_results tr
    WHERE tr.parent.realm = "chromium:try"
      AND tr.partition_time >= start_ts
      AND tr.partition_time < end_ts
      AND NOT tr.status = "SKIP"
    GROUP BY builder, test_suite, `date`, exported.id
  ), test_id_per_build AS (
    SELECT
      builder,
      test_suite,
      `date`,
      MAX(max_result_time) AS max_result_time,
      -- There isn't likely much value in other quantiles, this is just to filter
      -- days where the tests are getting added but not landed (avoid the P100)
      APPROX_QUANTILES(test_count, 100)[50] test_count,
    FROM test_ids
    GROUP BY builder, test_suite, `date`
  )
  SELECT
    t.date,
    'Test Case Count' AS metric,
    t.builder,
    -- Use the latest result as our dirty flag so we don't have to waste merging with the build table
    MAX(t.max_result_time) AS max_builder_start_time,
    ARRAY_AGG(
      STRUCT(t.test_suite AS label, CAST(t.test_count AS NUMERIC) AS value)
      ORDER BY t.test_suite
    ) AS value_agg,
  FROM
    test_id_per_build t,
    `chrome-trooper-analytics.metrics.cq_builders` cq
  WHERE
    t.builder IS NOT NULL AND t.test_suite IS NOT NULL
    AND t.builder = cq.builder
  GROUP BY t.date, t.builder
  ) S
ON
  T.date = S.date AND T.metric = S.metric AND T.builder = S.builder
  WHEN MATCHED AND T.checkpoint != string(S.max_builder_start_time, "UTC") THEN
    UPDATE SET value_agg = S.value_agg, checkpoint = string(S.max_builder_start_time, "UTC"), last_updated = current_timestamp()
  WHEN NOT MATCHED THEN
    INSERT (date, metric, builder, value_agg, last_updated, checkpoint)
    VALUES (date, metric, builder, value_agg, current_timestamp(), string(max_builder_start_time, "UTC"));
