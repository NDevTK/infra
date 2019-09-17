# Metrics

Tricium analyzer to check that metrics requests are correctly formatted and contain all of the necessary information.

Currently, this analyzer can be used for [UMA histogram](https://chromium.googlesource.com/chromium/src.git/+/HEAD/tools/metrics/histograms/README.md) submissions.

For each histogram, this analyzer checks:
1. There is more than 1 owner
2. The first owner is an individual, not a team

Note: We assume that histograms.xml has been autoformatted (i.e. by running `python pretty_print.py` in the histograms directory)
