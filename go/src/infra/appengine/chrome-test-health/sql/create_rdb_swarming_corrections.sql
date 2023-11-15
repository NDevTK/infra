CREATE OR REPLACE TABLE `APP_ID`.DATASET.rdb_swarming_correction (
  `date` DATE OPTIONS (description = 'The day for which the correction factors apply'),
  builder STRING OPTIONS (description = 'The builder running the suites'),
  test_suite STRING OPTIONS (description = 'The test suite ran on the builder'),
  target_platform string OPTIONS (description = 'The platform of the builder/suite'),
  actual_cores FLOAT64 OPTIONS (description = 'The average number of cores used by swarming when actively running the test suite'),
  rdb_duration FLOAT64 OPTIONS (description = 'Average of the sum of rdb results from the builder/test suite reported to RDB'),
  swarming_duration FLOAT64 OPTIONS (description = 'Average time spent running the test according to swarming'),
  swarming_correction FLOAT64 OPTIONS (description = 'The amount rdb durations needs to be multiplied by to achieve the real swarming time spent'),
  core_correction FLOAT64 OPTIONS (description = 'The amount rdb durations needs to be multiplied by to achieve the real swarming time spent per core'),
  )
PARTITION BY `date`
