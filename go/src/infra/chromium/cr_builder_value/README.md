# Builder value

This directory contains source code for services that identify low-value
Chrome and Chromium builders. The value of a builder is calculated based on the
following factors:
- Frequency: Builders that are run rarely are less valuable.
- Failure ratio: Builders that fail more often are less valuable.
- Google Analytics page views: Builders with little to none Google Analytics
page views per build failure are less valuable.
- Success ratio: For builders that test code, a success ratio that is close to
100% might indicate that the builder is not providing much added value.
- Operating cost: Builders that are more expensive to run are less valuable.
Operating cost of a builder is calculated based on the estimated number of bots
used to run the builder.

This directory currently includes the following app(s):
-   Generate: An app for retrieving the list of Chrome and Chromium builders.
It then pushes this list to a BigQuery table so that the builders dashboard can
read it.

# How to launch the app?

luci-auth context -- go run . generate