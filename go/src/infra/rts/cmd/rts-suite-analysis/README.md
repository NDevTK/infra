# RTS suite analysis CLI

This CLI makes use of the evaluation built into the RTS framework to analyze
the expected results of removing a specific test suite from a specific builder.

1. Fetch a sample of rejections, e.g. for the previous month of all builders:
   ```bash
   go run ./cmd/rts-suite-analysis fetch-rejections \
     -from 2022-08-01 -to 2021-09-01 \
     -out samples.rej
   ```
1. Fetch a sample of test durations, e.g. for for the past week of all builders.
   ```bash
   go run ./cmd/rts-suite-analysis fetch-durations \
     -from 2022-08-25 -to 2021-09-01 \
     -out samples.dur
   ```
   Note: Duration data is expensive to collect/analyze. To reduce this cost only
   a fraction of the applicable durations will be collected. By default this
   optional -frac flag hold .1 (10%) over the given time period and savings
   are estimated based this fraction (the durations don't need to be over the
   same time period as rejections to function as this is an estimate, as long
   as the average test in the suite hasn't changed dramatically the estimate
   should still be accurate)
1. Run the analysis on the rejections and durations which you created from -out
   in the previous steps, note the ChangeRecall/Savings:
   ```bash
   go run ./cmd/rts-suite-analysis analyze \
     -rejections samples.rej \
     -durations samples.dur \
     -builder browser_tests \
     -testSuite linux_chromium_asan_rel_ng
   ```
   Output will look somthing like:
   ```ChangeRecall | Savings
   ----------------------
    99.78%      |  14.45% 
   100.00%      |   0.00% 
   
   based on 6878 rejections, 8066798 test failures, 913690h51m52.584049s testing time
   ```
   The ChangeRecall being the expected number of failed CLs in the given time
   period that would still have failed if the test suite was not run on the
   builder (successful CLs do not factor into this number). The savings is how
   much estimated time would have been saved for that tradeoff in recall. The
   100% entry can be ignored for this CLI. Note the rejections as this should
   match the number of rejections collected in the fetch-rejections step and
   can help confirm the values are reasonable

Rejection and duration data can take a long time to collect (you will likely not
want to go over a months worth of data as the results tend to stay consistent
while the collection time can be very long) but are reuseable for testing
different builder/test suite combinations and only the last step needs to be
repeated for different tests.
