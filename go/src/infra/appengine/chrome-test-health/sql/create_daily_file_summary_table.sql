CREATE OR REPLACE TABLE APP_ID.DATASET.daily_file_metrics (
  `date` DATE OPTIONS (description = 'The day being summarized'),
  repo STRING OPTIONS (description = 'Repo from which the file summaries apply'),
  component STRING OPTIONS (description = 'Component rollup. Only files that apply to this component are rolled up here.'),
  node_name STRING OPTIONS (description = 'Either the directory name or file name rooted with // or / for the root node'),
  is_file BOOL OPTIONS (description = 'Whether the row is a summary of a directory or file'),
  num_runs INTEGER OPTIONS (description = 'How many times a test was run in the file or folder'),
  num_failures INTEGER OPTIONS (description = 'How often a test failed in the file or folder'),
  num_flake INTEGER OPTIONS (description = 'How often a test provided conflicting statuses for an equivalent patchset in the file or folder'),
  total_runtime FLOAT64 OPTIONS (description = 'The total time spent running tests in this file or directory'),
  corrected_swarming_runtime FLOAT64 OPTIONS (description = 'total_runtime corrected to match swarming time'),
  corrected_core_runtime FLOAT64 OPTIONS (description = 'total_runtime corrected to match swarming time of all cores used'),
  avg_bots FLOAT64 OPTIONS (description = 'corrected_swarming_runtime normalized on the time at which it was measured'),
  avg_cores FLOAT64 OPTIONS (description = 'corrected_core_runtime normalized on the time at which it was measured'),
  avg_runtime FLOAT64 OPTIONS (description = 'The summed average runtime for all tests in the file or directory'),
  p50_runtime FLOAT64 OPTIONS (description = 'The p50 runtime for the test from that day in this file or directory'),
  p90_runtime FLOAT64 OPTIONS (description = 'The p90 runtime for the test from that day in this file or directory'),
  )
PARTITION BY `date`
CLUSTER BY component;