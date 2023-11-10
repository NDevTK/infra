MERGE INTO {project}.{dataset}.rdb_swarming_correction AS T
USING (
  WITH swarming_times AS (
    SELECT DISTINCT
      task_id, -- only used for deduping the swarming table
      DATE(end_time, "PST8PDT") AS `date`,
      timestamp_diff(end_time, start_time, SECOND) AS duration, -- Includes overhead
      -- The last digit is 0 for the shard
      SUBSTR(request.task_id, 0, LENGTH(request.task_id) - 1) AS task_join,
      -- The actual cores on the bot that ran the test suite
      CAST((SELECT d.values FROM t.bot.dimensions AS d WHERE d.key = 'cores' LIMIT 1)[SAFE_OFFSET(0)] AS INT64) AS actual_cores,
    FROM
      {swarming_tasks_table} AS t
    WHERE
      -- The swarming task may have finished the day after the individual test ran
      DATE(end_time, "PST8PDT") BETWEEN @from_date AND DATE_ADD(@to_date, INTERVAL 1 DAY)
      -- Ignore swarming deduped shards
      AND `state` != 'DEDUPED'
  ),

  test_results AS (
    SELECT
        ANY_VALUE((SELECT v.value FROM tr.variant AS v WHERE v.key = 'builder' LIMIT 1)) AS builder,
        ANY_VALUE((SELECT v.value FROM tr.variant AS v WHERE v.key = 'test_suite' LIMIT 1)) AS test_suite,
        ANY_VALUE((SELECT t.value FROM tr.tags AS t WHERE t.key = 'target_platform' LIMIT 1)) AS target_platform,
        COUNT(DISTINCT test_id) AS test_count,
        COUNT(*) AS num_runs,
        -- The prefix needs to be removed and last digit is 1 for the rdb
        REGEXP_EXTRACT(parent.id, r'([^-]+)1$') AS parent_task_join,
        SUM(duration) AS duration,
    FROM
      {chromium_try_rdb_table} AS tr
    WHERE
      DATE(partition_time, "PST8PDT") BETWEEN @from_date AND @to_date
      AND tr.status != "SKIP"
      GROUP BY parent.id
    UNION ALL
    SELECT
        ANY_VALUE((SELECT v.value FROM tr.variant AS v WHERE v.key = 'builder' LIMIT 1)) AS builder,
        ANY_VALUE((SELECT v.value FROM tr.variant AS v WHERE v.key = 'test_suite' LIMIT 1)) AS test_suite,
        ANY_VALUE((SELECT t.value FROM tr.tags AS t WHERE t.key = 'target_platform' LIMIT 1)) AS target_platform,
        COUNT(DISTINCT test_id) AS test_count,
        COUNT(*) AS num_runs,
        -- The prefix needs to be removed and last digit is 1 for the rdb
        REGEXP_EXTRACT(parent.id, r'([^-]+)1$') AS parent_task_join,
        SUM(duration) AS duration,
    FROM
      {chromium_ci_rdb_table} AS tr
    WHERE
      DATE(partition_time, "PST8PDT") BETWEEN @from_date AND @to_date
      AND tr.status != "SKIP"
      GROUP BY parent.id
  ),

  rdb_swarming_join AS (
    SELECT
      `date`,
      builder,
      test_suite,
      target_platform,
      actual_cores,
      tr.duration AS rdb_duration,
      s.duration AS swarming_duration,
      tr.duration / s.duration AS estimated_cores,
      test_count,
      num_runs,
    FROM swarming_times AS s INNER JOIN test_results AS tr
      ON s.task_join = tr.parent_task_join
    WHERE
      -- Both durations can be null. Both of these are obviously incorrect and dont help
      tr.duration IS NOT NULL AND s.duration IS NOT NULL
      -- RDB durations can be 0 which is unhelpful and clearly incorrect
      AND tr.duration > 0
    ORDER BY estimated_cores ASC
  )

  SELECT
    `date`,
    builder,
    test_suite,
    target_platform,
    AVG(actual_cores) AS actual_cores,
    AVG(rdb_duration) AS rdb_duration,
    AVG(swarming_duration) AS swarming_duration,
    -- The correction required to arrive at swarming time spent with a bot
    AVG(swarming_duration / rdb_duration) AS swarming_correction,
    -- The correction required to arrive at swarming time spent with a core of the bot
    AVG((swarming_duration * actual_cores) / rdb_duration) AS core_correction,
  FROM rdb_swarming_join
  GROUP BY builder, test_suite, target_platform, `date`
  ) AS S
ON
  T.date = S.date
  AND T.builder = S.builder
  AND T.test_suite = S.test_suite
  AND T.target_platform = S.target_platform
WHEN MATCHED THEN
  UPDATE SET
    actual_cores = S.actual_cores,
    rdb_duration = S.rdb_duration,
    swarming_duration = S.swarming_duration,
    swarming_correction = S.swarming_correction,
    core_correction = S.core_correction
WHEN NOT MATCHED THEN
  INSERT ROW
