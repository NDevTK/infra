-- This query updates the 'Shard Expirations' metrics in the cq_builder_metrics_week
-- table. It uses an interval of the last 8 days so that on Mondays, it will update stats for the full
-- previous week as well as stats for the current week.
-- This query is meant to be almost identical to the one in cq_builder_metrics_day_shard_expirations.sql.

-- The lines below are used by the deploy tool.
--name: Populate cq_builder_metrics_week Shard Expirations
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
  WITH builds AS (
    SELECT
      DATE_TRUNC(EXTRACT(DATE FROM b.start_time AT TIME ZONE "PST8PDT"), WEEK(MONDAY)) AS `date`,
      b.id,
      b.builder.builder,
      b.start_time,
      b.infra.backend.task.id.id as task_id
    FROM
      `cr-buildbucket.chromium.builds` b,
      `chrome-trooper-analytics.metrics.cq_builders` cq
    WHERE
      -- As we bucket the build using start_date, we need to include any builds
      -- that were created on the previous day.
      b.create_time >= TIMESTAMP_SUB(start_ts, INTERVAL 1 DAY)
      AND b.create_time < TIMESTAMP_ADD(end_ts, INTERVAL 8 DAY)
      AND b.builder.bucket = 'try'
      AND b.builder.project = 'chromium'
      AND b.builder.builder = cq.builder
    ),
    test_tasks AS (
    SELECT
      b.id,
      ANY_VALUE(b.date) date,
      ANY_VALUE(b.builder) builder,
      ANY_VALUE(b.start_time) builder_start_time,
      (
        SELECT SUBSTR(i, 12) FROM t.request.tags i WHERE i LIKE 'test_suite:%'
      ) test_name,
      ANY_VALUE(state) state,
      MAX(TIMESTAMP_DIFF(t.start_time, t.create_time, SECOND)) max_shard_pending,
      MAX(TIMESTAMP_DIFF(t.end_time, t.start_time, SECOND)) max_shard_runtime
    FROM
      `chromium-swarm.swarming.task_results_summary` t,
      builds b
    WHERE
      b.date >= start_date AND b.date < end_date
      AND t.end_time >= TIMESTAMP_SUB(start_ts, INTERVAL 1 DAY)
      AND t.end_time < TIMESTAMP_ADD(end_ts, INTERVAL 1 DAY)
      AND t.request.parent_task_id = b.task_id
    GROUP BY id, test_name
    ),
    test_metrics AS (
    SELECT
      t.date,
      t.builder,
      MAX(t.builder_start_time) max_builder_start_time,
      t.test_name,
      countif(state = 'EXPIRED') expired_shards,
    FROM test_tasks t
    WHERE t.test_name is not null
    GROUP BY t.date, t.builder, t.test_name
    )
  SELECT
    date,
    'Shard Expirations' AS metric,
    builder,
    MAX(max_builder_start_time) AS max_builder_start_time,
    ARRAY_AGG(
      STRUCT(test_name AS label, CAST(expired_shards AS NUMERIC) AS value)
      ORDER BY test_name
    ) AS value_agg,
  FROM test_metrics
  GROUP BY date, builder
  ) S
ON
  T.date = S.date AND T.metric = S.metric AND T.builder = S.builder
  WHEN MATCHED AND T.checkpoint != string(S.max_builder_start_time, "UTC") THEN
    UPDATE SET value_agg = S.value_agg, checkpoint = string(S.max_builder_start_time, "UTC"), last_updated = current_timestamp()
  WHEN NOT MATCHED THEN
    INSERT (date, metric, builder, value_agg, last_updated, checkpoint)
    VALUES (date, metric, builder, value_agg, current_timestamp(), string(max_builder_start_time, "UTC"));
