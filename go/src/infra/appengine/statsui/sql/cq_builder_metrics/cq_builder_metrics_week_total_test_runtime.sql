-- This query updates the total test runetime metric in the
-- cq_builder_metrics_week table. It uses an interval of the last 8 days so that
-- on Mondays, it will update stats for the full previous week as well as stats
-- for the current week.
-- This query is meant to be almost identical to the one in
-- cq_builder_metrics_day_total_test_runtime.sql.

-- The lines below are used by the deploy tool.
--name: Populate cq_builder_metrics_week total test runtime
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
      b.start_time,
      id,
      b.builder.builder builder,
      b.infra.swarming.task_id,
    FROM
      `cr-buildbucket.chromium.builds` b,
      `chrome-trooper-analytics.metrics.cq_builders` cq
    WHERE
      -- Include builds that were created on Sunday but started on Monday
      b.create_time >= TIMESTAMP_SUB(start_ts, INTERVAL 1 DAY)
      -- If the end date is on a Tuesday, make sure to include all builds for that week.
      AND b.create_time < TIMESTAMP_ADD(end_ts, INTERVAL 8 DAY)
      AND b.builder.bucket = 'try'
      AND b.builder.project = 'chromium'
      AND b.builder.builder = cq.builder
    ),
    -- This table is needed to dedupe swarming task IDs, as sometimes there are
    -- duplicate rows in task_results_summary
    tasks AS (
    SELECT
      b.date,
      b.id build_id,
      t.task_id,
      any_value(b.builder) builder,
      MAX(b.start_time) max_start_time,
      MAX(timestamp_diff(t.end_time, t.start_time, SECOND)) test_duration_sec,
    FROM
      builds b,
      `chromium-swarm.swarming.task_results_summary` t
    WHERE
      -- Include builds that were created on Sunday but started on Monday
      t.end_time >= TIMESTAMP_SUB(start_ts, INTERVAL 1 DAY)
      -- If the end date is on a Tuesday, make sure to include all builds for that week.
      AND t.end_time < TIMESTAMP_ADD(end_ts, INTERVAL 8 DAY)
      and b.task_id = t.request.parent_task_id
      -- Exclude compilator tasks
      and t.request.name not like '%-compilator-%'
    GROUP BY
      b.date, b.id, t.task_id
    ),
    -- Sum up the test task durations for each build
    build_stats AS (
    SELECT
      t.date,
      t.builder,
      t.build_id,
      MAX(t.max_start_time) max_start_time,
      SUM(t.test_duration_sec) total_test_duration_sec,
    FROM tasks t
    WHERE t.date >= start_date AND t.date < end_date
    GROUP BY t.date, t.builder, t.build_id
    ),
    -- Calculate quantiles for each builder
    builder_stats AS (
    SELECT
      s.date date,
      s.builder,
      COUNT(*) cnt,
      -- To inspect date boundaries
      MIN(s.max_start_time) min_start_time,
      -- To inspect date boundaries
      MAX(s.max_start_time) max_start_time,
      APPROX_QUANTILES(s.total_test_duration_sec, 100) test_runtime
    FROM build_stats s
    GROUP BY s.date, s.builder
    )
  SELECT date, 'P50 Total Test Runtime' AS metric, builder, max_start_time, test_runtime[OFFSET(50)] AS value FROM builder_stats
  UNION ALL SELECT date, 'P90 Total Test Runtime' AS metric, builder, max_start_time, test_runtime[OFFSET(90)] AS value FROM builder_stats
  ) S
ON
  T.date = S.date AND T.metric = S.metric AND T.builder = S.builder
  WHEN MATCHED AND T.checkpoint != string(S.max_start_time, "UTC") THEN
    UPDATE SET value = S.value, checkpoint = string(S.max_start_time, "UTC"), last_updated = current_timestamp()
  WHEN NOT MATCHED THEN
    INSERT (date, metric, builder, value, last_updated, checkpoint)
    VALUES (date, metric, builder, value, current_timestamp(), string(max_start_time, "UTC"));
