CREATE OR REPLACE TABLE APP_ID.DATASET.test_metrics (
  date DATE,
  test_id STRING,
  test_name STRING,
  repo STRING,
  file_name STRING,
  component STRING,
  # Test id summaries
  num_runs INTEGER,
  num_failures INTEGER,
  num_flake INTEGER,
  avg_runtime FLOAT64,
  total_runtime FLOAT64,
  # Variants breakdown
  variant_summaries ARRAY<STRUCT<
    variant_hash STRING,
    `project` STRING,
    bucket STRING,
    target_platform STRING,
    builder STRING,
    test_suite STRING,
    num_runs INTEGER,
    num_failures INTEGER,
    num_flake INTEGER,
    avg_runtime FLOAT64,
    total_runtime FLOAT64>>)
PARTITION BY `date`
CLUSTER BY component, file_name, test_id
OPTIONS (partition_expiration_days = 540);
