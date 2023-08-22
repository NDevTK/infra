-- Update raw table
MERGE INTO {project}.{dataset}.raw_metrics AS T
USING (
  SELECT
    `date`,
    -- test level info
    test_id,
    test_name,
    repo,
    file_name,
    component,
    -- variant level info
    variant_hash,
    `project`,
    bucket,
    target_platform,
    builder,
    test_suite,
    -- Total runtime in seconds divided by the seconds in the time window
    total_runtime / (LEAST(
        TIMESTAMP_DIFF(CURRENT_TIMESTAMP(), TIMESTAMP(`date`, "PST8PDT"), SECOND),
        60 * 60 * 24
        )) AS avg_cores
  FROM {project}.{dataset}.raw_metrics
  WHERE `date` BETWEEN @from_date AND @to_date
  ) AS S
ON
  T.date = S.date
  AND T.date BETWEEN @from_date AND @to_date
  AND T.test_id = S.test_id
  AND (T.repo = S.repo OR (T.repo IS NULL AND S.repo IS NULL))
  AND (T.component = S.component OR (T.component IS NULL AND S.component IS NULL))
  AND (T.variant_hash = S.variant_hash OR (T.variant_hash IS NULL AND S.variant_hash IS NULL))
  AND (T.`project` = S.`project` OR (T.`project` IS NULL AND S.`project` IS NULL))
  AND (T.bucket = S.bucket OR (T.bucket IS NULL AND S.bucket IS NULL))
  AND (T.target_platform = S.target_platform OR (T.target_platform IS NULL AND S.target_platform IS NULL))
  AND (T.builder = S.builder OR (T.builder IS NULL AND S.builder IS NULL))
  AND (T.test_suite = S.test_suite OR (T.test_suite IS NULL AND S.test_suite IS NULL))
WHEN MATCHED THEN
  UPDATE SET
    avg_cores = S.avg_cores;

-- Update daily test metric
MERGE INTO {project}.{dataset}.daily_test_metrics AS T
USING (
  SELECT
    `date`,
    -- test level info
    test_id,
    test_name,
    repo,
    file_name,
    component,
    -- variant level info
    builder,
    bucket,
    test_suite,
    -- Total runtime in seconds divided by the seconds in the time window
    total_runtime / (LEAST(
        TIMESTAMP_DIFF(CURRENT_TIMESTAMP(), TIMESTAMP(`date`, "PST8PDT"), SECOND),
        60 * 60 * 24
        )) AS avg_cores
  FROM {project}.{dataset}.daily_test_metrics
  WHERE `date` BETWEEN @from_date AND @to_date
  ) AS S
ON
  T.date = S.date
  AND T.date BETWEEN @from_date AND @to_date
  AND T.test_id = S.test_id
  AND (T.repo = S.repo OR (T.repo IS NULL AND S.repo IS NULL))
  AND (T.component = S.component OR (T.component IS NULL AND S.component IS NULL))
  AND (T.builder = S.builder OR (T.builder IS NULL AND S.builder IS NULL))
  AND (T.bucket = S.bucket OR (T.bucket IS NULL AND S.bucket IS NULL))
  AND (T.test_suite = S.test_suite OR (T.test_suite IS NULL AND S.test_suite IS NULL))
WHEN MATCHED THEN
  UPDATE SET
    avg_cores = S.avg_cores;

-- Update weekly test metric
MERGE INTO {project}.{dataset}.weekly_test_metrics AS T
USING (
  SELECT
    `date`,
    -- test level info
    test_id,
    test_name,
    repo,
    file_name,
    component,
    -- variant level info
    builder,
    bucket,
    test_suite,
    -- Total runtime in seconds divided by the seconds in the time window
    total_runtime / (LEAST(
        TIMESTAMP_DIFF(CURRENT_TIMESTAMP(), TIMESTAMP(DATE_TRUNC(`date`, WEEK), "PST8PDT"), SECOND),
        60 * 60 * 24 * 7
        )) AS avg_cores
  FROM {project}.{dataset}.weekly_test_metrics
  WHERE `date` BETWEEN
    DATE_TRUNC(DATE(@from_date), WEEK) AND
    DATE_ADD(DATE_TRUNC(DATE(@to_date), WEEK), INTERVAL 6 DAY)
  ) AS S
ON
  T.date = S.date
  AND T.date BETWEEN
    DATE_TRUNC(DATE(@from_date), WEEK) AND
    DATE_ADD(DATE_TRUNC(DATE(@to_date), WEEK), INTERVAL 6 DAY)
  AND T.test_id = S.test_id
  AND (T.repo = S.repo OR (T.repo IS NULL AND S.repo IS NULL))
  AND (T.component = S.component OR (T.component IS NULL AND S.component IS NULL))
  AND (T.builder = S.builder OR (T.builder IS NULL AND S.builder IS NULL))
  AND (T.bucket = S.bucket OR (T.bucket IS NULL AND S.bucket IS NULL))
  AND (T.test_suite = S.test_suite OR (T.test_suite IS NULL AND S.test_suite IS NULL))
WHEN MATCHED THEN
  UPDATE SET
    avg_cores = S.avg_cores;

-- Update daily file metric
MERGE INTO {project}.{dataset}.daily_file_metrics AS T
USING (
  SELECT
    `date`,
    repo,
    component,
    node_name,
    -- Total runtime in seconds divided by the seconds in the time window
    total_runtime / (LEAST(
        TIMESTAMP_DIFF(CURRENT_TIMESTAMP(), TIMESTAMP(`date`, "PST8PDT"), SECOND),
        60 * 60 * 24
        )) AS avg_cores
  FROM {project}.{dataset}.daily_file_metrics
  WHERE `date` BETWEEN @from_date AND @to_date
  ) AS S
ON
  T.date = S.date
  AND T.date BETWEEN @from_date AND @to_date
  AND T.node_name = S.node_name
  AND (T.component = S.component OR (T.component IS NULL AND S.component IS NULL))
  AND (T.repo = S.repo OR (T.repo IS NULL AND S.repo IS NULL))
WHEN MATCHED THEN
  UPDATE SET
    avg_cores = S.avg_cores;

-- Update weekly file metric
MERGE INTO {project}.{dataset}.weekly_file_metrics AS T
USING (
  SELECT
    `date`,
    repo,
    component,
    node_name,
    -- Total runtime in seconds divided by the seconds in the time window
    total_runtime / (LEAST(
        TIMESTAMP_DIFF(CURRENT_TIMESTAMP(), TIMESTAMP(DATE_TRUNC(`date`, WEEK), "PST8PDT"), SECOND),
        60 * 60 * 24 * 7
        )) AS avg_cores
  FROM {project}.{dataset}.weekly_file_metrics
  WHERE `date` BETWEEN
    DATE_TRUNC(DATE(@from_date), WEEK) AND
    DATE_ADD(DATE_TRUNC(DATE(@to_date), WEEK), INTERVAL 6 DAY)
  ) AS S
ON
  T.date = S.date
  AND T.date BETWEEN
    DATE_TRUNC(DATE(@from_date), WEEK) AND
    DATE_ADD(DATE_TRUNC(DATE(@to_date), WEEK), INTERVAL 6 DAY)
  AND T.node_name = S.node_name
  AND (T.component = S.component OR (T.component IS NULL AND S.component IS NULL))
  AND (T.repo = S.repo OR (T.repo IS NULL AND S.repo IS NULL))
WHEN MATCHED THEN
  UPDATE SET
    avg_cores = S.avg_cores;