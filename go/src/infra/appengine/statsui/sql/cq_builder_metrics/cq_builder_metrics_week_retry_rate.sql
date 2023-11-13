-- This query updates the retry rates metric in the
-- cq_builder_metrics_week table. It uses an interval of the last 2 days so that
-- there is some redundancy if the job fails This query is meant to be almost
-- identical to the one in cq_builder_metrics_day_retry_rate.sql.

-- The lines below are used by the deploy tool.
--name: Populate cq_builder_metrics_week retry rates
--schedule: every 8 hours synchronized

DECLARE start_date DATE DEFAULT DATE_SUB(CURRENT_DATE('PST8PDT'), INTERVAL 8 DAY);
-- This isn't really needed, but useful to have around when doing backfills
-- The end_date is exclusive, which is why we add a day to the current date.
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
      b.infra.swarming.task_id
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
    deduped_tasks AS (
      SELECT
        b.date,
        b.id AS build_id,
        t.task_id,
        MAX(b.start_time) AS start_time,
        any_value(b.builder) builder,
        (
          SELECT
            SPLIT(tag, ':')[offset(1)]
          FROM UNNEST(t.request.tags) tag
          WHERE STARTS_WITH(tag, 'test_suite:')
        ) AS test_suite,
        request.name LIKE '%retry shards with patch%' AS retry,
        request.name LIKE '%without patch%' AS without_patch,
      FROM
        builds b,
        `chromium-swarm.swarming.task_results_summary` t
      WHERE
        t.end_time >= TIMESTAMP_SUB(start_ts, INTERVAL 1 DAY)
        AND t.end_time < TIMESTAMP_ADD(end_ts, INTERVAL 1 DAY)
        AND b.task_id = t.request.parent_task_id
        -- Exclude compilator tasks
        AND t.request.name not like '%-compilator-%'
        AND t.end_time IS NOT NULL
        AND t.start_time IS NOT NULL
      GROUP BY
        b.date, b.id, t.task_id, test_suite, request.name
      HAVING test_suite IS NOT NULL
    ),
    -- Combine the swarming task times by the test suite for each build
    builder_suite_retries AS (
      SELECT
        build_id,
        test_suite,
        ANY_VALUE(b.date) date,
        ANY_VALUE(b.builder) builder,
        LOGICAL_OR(retry) retried,
        LOGICAL_OR(without_patch) retried_without_patch,
        MAX(start_time) AS start_time
      FROM
        deduped_tasks b
      GROUP BY build_id, test_suite
    ),
    builder_suite_retry_ratio AS (
      SELECT
        b.date,
        b.builder,
        COUNTIF(b.retried)/COUNT(*) retry_with_patch_ratio,
        COUNTIF(b.retried_without_patch)/COUNT(*) retried_without_patch,
        b.test_suite,
        MAX(start_time) AS start_time
      FROM builder_suite_retries b
      WHERE
        b.test_suite is not null
        -- This makes sure that we only include full weeks.  For example, if the start_date is a
        -- Wednesday, the week will start on Monday, which will fall outside of the allowed range.
        AND b.date >= start_date AND b.date < end_date
      GROUP BY b.date, b.builder, b.test_suite
      HAVING COUNT(*) > 50
    )
  SELECT
    date,
    'Retry With Patch Rate' AS metric,
    builder,
    MAX(start_time) AS max_builder_start_time,
    ARRAY_AGG(
      STRUCT(test_suite AS label, CAST(retry_with_patch_ratio AS NUMERIC) AS value)
      ORDER BY test_suite
    ) AS value_agg,
  FROM builder_suite_retry_ratio
  GROUP BY date, builder
  UNION ALL
  SELECT
    date,
    'Retry Without Patch Rate' AS metric,
    builder,
    MAX(start_time) AS max_builder_start_time,
    ARRAY_AGG(
      STRUCT(test_suite AS label, CAST(retried_without_patch AS NUMERIC) AS value)
      ORDER BY test_suite
    ) AS value_agg,
  FROM builder_suite_retry_ratio
  GROUP BY date, builder
  ) S
ON
  T.date = S.date AND T.metric = S.metric AND T.builder = S.builder
  WHEN MATCHED AND T.checkpoint != string(S.max_builder_start_time, "UTC") THEN
    UPDATE SET value_agg = S.value_agg, checkpoint = string(S.max_builder_start_time, "UTC"), last_updated = current_timestamp()
  WHEN NOT MATCHED THEN
    INSERT (date, metric, builder, value_agg, last_updated, checkpoint)
    VALUES (date, metric, builder, value_agg, current_timestamp(), string(max_builder_start_time, "UTC"));
