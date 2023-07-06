-- This query updates the 'Test Case Count' metric in the cq_builder_metrics_week
-- table. It uses an interval of the last 8 days so that on Mondays, it will update stats for the full
-- previous week as well as stats for the current week.
-- This query is meant to be almost identical to the one in cq_builder_metrics_day_test_counts.sql.

-- The lines below are used by the deploy tool.
--name: Populate cq_builder_metrics_week slow test metrics
--schedule: every 8 hours synchronized

DECLARE start_date DATE DEFAULT DATE_SUB(CURRENT_DATE('PST8PDT'), INTERVAL 8 DAY);
-- This isn't really needed, but useful to have around when doing backfills
-- The end_date is exclusive, so if it's on a Monday then that week won't be
-- included. This is why we add a day to it.
DECLARE end_date DATE DEFAULT DATE_ADD(CURRENT_DATE('PST8PDT'), INTERVAL 1 DAY);
DECLARE start_ts TIMESTAMP DEFAULT TIMESTAMP(start_date, 'PST8PDT');
DECLARE end_ts TIMESTAMP DEFAULT TIMESTAMP(end_date, 'PST8PDT');

-- Merge statement
MERGE INTO
  `chrome-trooper-analytics.metrics.cq_builder_metrics_week` AS T
USING
  (
  WITH test_ids AS (
      SELECT
        DATE(partition_time) `date`,
        MAX(partition_time) max_result_time,
        COUNT(DISTINCT tr.test_id) test_count,
        (SELECT v.value FROM tr.variant v WHERE v.key = 'builder') builder,
        (SELECT v.value FROM tr.variant v WHERE v.key = 'test_suite') test_suite,
      FROM chrome-luci-data.chrome.try_test_results tr
      WHERE tr.parent.realm = "chrome:try"
        AND tr.partition_time >= start_ts
        AND tr.partition_time < end_ts
      GROUP BY builder, test_suite, `date`
      UNION ALL
      SELECT
        DATE(partition_time) `date`,
        MAX(partition_time) max_result_time,
        COUNT(DISTINCT tr.test_id) test_count,
        (SELECT v.value FROM tr.variant v WHERE v.key = 'builder') builder,
        (SELECT v.value FROM tr.variant v WHERE v.key = 'test_suite') test_suite,
      FROM chrome-luci-data.chrome.try_test_results tr
      WHERE tr.parent.realm = "chromium:try"
        AND tr.partition_time >= start_ts
        AND tr.partition_time < end_ts
      GROUP BY builder, test_suite, `date`
    )
  SELECT
    `date`,
    'Test Case Count' AS metric,
    builder,
    -- Use the latest result as our dirty flag so we don't have to waste merging with the build table
    MAX(max_result_time) AS max_result_time,
    ARRAY_AGG(
      STRUCT(test_suite AS label, CAST(test_count AS NUMERIC) AS value)
      ORDER BY test_suite
    ) AS value_agg,
  FROM test_ids
  GROUP BY `date`, builder
  ) S
ON
  T.date = S.date AND T.metric = S.metric AND T.builder = S.builder
  WHEN MATCHED AND T.checkpoint != string(S.max_result_time, "UTC") THEN
    UPDATE SET value_agg = S.value_agg, checkpoint = string(S.max_result_time, "UTC"), last_updated = current_timestamp()
  WHEN NOT MATCHED THEN
    INSERT (date, metric, builder, value_agg, last_updated, checkpoint)
    VALUES (date, metric, builder, value_agg, current_timestamp(), string(max_result_time, "UTC"));
