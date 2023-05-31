CREATE OR REPLACE TABLE APP_ID.DATASET.file_metrics (
  `date` DATE,
  repo STRING,
  component STRING,
  node_name STRING,
  is_file BOOL,
  num_runs INTEGER,
  num_failures INTEGER,
  num_flake INTEGER,
  avg_runtime FLOAT64,
  total_runtime FLOAT64,
  file_names ARRAY<STRING>,
  )
PARTITION BY `date`
CLUSTER BY component
OPTIONS (partition_expiration_days = 540);