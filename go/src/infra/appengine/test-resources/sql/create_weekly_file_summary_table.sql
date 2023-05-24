CREATE OR REPLACE TABLE %s.test_results.weekly_file_metrics (
  date DATE,
  component STRING,
  node_name STRING,
  is_file BOOL,
  num_runs INTEGER,
  num_failures INTEGER,
  num_flake INTEGER,
  avg_runtime FLOAT64,
  total_runtime FLOAT64,
  )
PARTITION BY `date`
CLUSTER BY component
OPTIONS (partition_expiration_days = 540);
