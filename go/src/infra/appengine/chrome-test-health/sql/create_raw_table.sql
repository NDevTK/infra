CREATE OR REPLACE TABLE APP_ID.DATASET.raw_metrics (
  date DATE OPTIONS (description = 'The day being summarized'),

  -- test level info
  test_id STRING OPTIONS (description = 'Unique test id'),
  test_name STRING OPTIONS (description = 'More display friendly name of the test'),
  repo STRING OPTIONS (description = 'Repo from which the test summaries apply'),
  file_name STRING OPTIONS (description = 'File in which the test resides'),
  component STRING OPTIONS (description = 'Component which owns the test'),

  -- detailed variant level info
  variant_hash STRING OPTIONS (description = 'Hash that identifies the variant'),
  `project` STRING OPTIONS (description = 'The project (e.g. milestone) being summarized'),
  bucket STRING OPTIONS (description = 'The bucket (e.g. "try", "ci") being summarized'),
  target_platform STRING OPTIONS (description = 'The platform being run (e.g. linux, mac, win, etc.)'),
  builder STRING OPTIONS (description = 'The breakdown of the builder being run'),
  test_suite STRING OPTIONS (description = 'Test suite in which the test is being run'),

  -- metrics
  num_runs INTEGER OPTIONS (description = 'How many times a test was run'),
  num_failures INTEGER OPTIONS (description = 'How often a test failed'),
  num_flake INTEGER OPTIONS (description = 'How often a test provided conflicting statuses for an equivalent patchset'),
  total_runtime FLOAT64 OPTIONS (description = 'The total time spent running this test throughout the day'),
  avg_cores FLOAT64 OPTIONS (description = 'The total time normalized on time'),
  avg_runtime FLOAT64 OPTIONS (description = 'The average runtime for a single run of the test'),
  p50_runtime FLOAT64 OPTIONS (description = 'The p50 runtime for the test from that day. Aggregated as an average of the p50 runtimes of variants that ran this test'),
  p90_runtime FLOAT64 OPTIONS (description = 'The p90 runtime for the test from that day. Aggregated as an average of the p90 runtimes of variants that ran this test'),
  )
PARTITION BY `date`
CLUSTER BY component, test_id, builder, test_suite;
