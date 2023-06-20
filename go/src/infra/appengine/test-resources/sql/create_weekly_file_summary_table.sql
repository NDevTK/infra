CREATE OR REPLACE TABLE APP_ID.DATASET.weekly_file_metrics (
  `date` DATE OPTIONS (description = 'The week being summarized starting on a Sunday'),
  repo STRING OPTIONS (description = 'Repo from which the file summaries apply'),
  component STRING OPTIONS (description = 'Component rollup. Only files that apply to this component are rolled up here.'),
  node_name STRING OPTIONS (description = 'Either the directory name or file name rooted with // or / for the root node'),
  is_file BOOL OPTIONS (description = 'Whether the row is a summary of a directory or file'),
  num_runs INTEGER OPTIONS (description = 'How many times a test was run in the file or folder'),
  num_failures INTEGER OPTIONS (description = 'How often a test failed in the file or folder'),
  num_flake INTEGER OPTIONS (description = 'How often a test provided conflicting statuses for an equivalent patchset in the file or folder'),
  avg_runtime FLOAT64 OPTIONS (description = 'The summed average runtime for all tests in the file or directory'),
  total_runtime FLOAT64 OPTIONS (description = 'The total time spent running tests in this file or directory'),
  child_file_summaries ARRAY<STRUCT<
    file_name STRING OPTIONS (description = 'File name rooted with //'),
    num_runs INTEGER OPTIONS (description = 'How many times a test was run in the file'),
    num_failures INTEGER OPTIONS (description = 'How often a test failed in the file'),
    num_flake INTEGER OPTIONS (description = 'How often a test provided conflicting statuses for an equivalent patchset in the file'),
    avg_runtime FLOAT64 OPTIONS (description = 'The summed average runtime for all tests in the file'),
    total_runtime FLOAT64 OPTIONS (description = 'The total time spent running tests in this file')
    >> OPTIONS (description = 'All the file summaries for this directory')
  )
PARTITION BY `date`
CLUSTER BY component
OPTIONS (partition_expiration_days = 540);
