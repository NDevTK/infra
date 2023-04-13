# UFS developer guide

*This document is intended to be a reference for any developers looking to
modify or add functionality to UFS. It provides necessary details for most of
the use cases that we expect. Please contact chrome-fleet-automation@google.com
for any questions with regards to UFS*

[TOC]

[go/ufs-dev](http://go/ufs-dev)

## Testing UFS service locally

Run UFS locally on your workstation (provided you have permissions). Makefile
has a few builds that help you with this.
```
make dev
```
You might need to run
```
luci-auth login -scopes "https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email"
```
to setup local auth for the service.

You might need to run
```
gcloud auth application-default login
```
for PubSub permissions.

## Debugging UFS service using VSCode

If you're using VSCode, you can attach to your locally running service through
the debug menu. When prompted, find the `ufs-service` process and debug through
VSCode as usual. If you get an error, try following the prompt to modify your
[Yama](https://chromium.googlesource.com/chromium/src/+/HEAD/docs/linux/debugging.md#Allowing-attaching-to-foreign-processes)
settings and try again.

## Testing dumper service locally

Run dumper locally on your workstation.
```
make dev-dumper
```
You might need to run
```
gcloud auth application-default login
```
for BigQuery permissions.

## Running cron jobs locally
shivas can be used to trigger the cron jobs locally. The makefile in shivas
source creates a local version `dev-shivas`. This can be used to trigger cron on
the local instance of `dumper`.
```
dev-shivas admin cron <cron-job-name>
```

## Adding a new field

Adding a new field to a proto includes some subtleties and may involve changes
outside this directory.

### Logging

Proto entity updates are logged at the controller layer in
[change.go](app/controller/change.go). Make sure any new fields get added there.

### BigQuery

If the proto is used in BigQuery, the BigQuery table schemas should be updated
once after the proto is merged to dev and once after a UFS push to prod. To do
so, run the [setup\_bigquery.sh](api/v1/models/bigquery/setup_bigquery.sh)
script.

In addition, any PLX scripts that require this field should get updated within
[google3](http://google3/configs/monitoring/chromeos_infra_monitoring/lab_platform/plx/).

### shivas

If the field needs shivas support, update the appropriate files in the shivas
repo. Notably, this may require UFS changes that aren't directly used by the
service, such as [app/util/input.go](app/util/input.go)

### ChromeOS Swarming for DUTs

ChromeOS Swarming fetches DUT info via a shivas command.

If the new field **is not** used to schedule tests,
[shivas](../cmd/shivas/internal/ufs/cmds/bot/internal-print-bot-info.go) should
be updated to expose the relevant field.

If the new field is used to schedule tests, shivas does not need any updates.
Instead, the skylab swarming library will need to be updated in two areas:

*   [SchedulableLabels proto](../libs/skylab/inventory/device.proto)
*   [A new converter/reverter for SchedulableLabels <-> swarming dimensions](../libs/skylab/inventory/swarming/)

In addition, another UFS util file,
[app/util/osutil/exporting\_adapter.go](app/util/osutil/exporting_adapter.go),
will need to include the new field when constructing a SchedulableLabels struct
for the corresponding entity.

Any skylab proto changes will require regeneration of code via `go generate`
within both the skylab library and UFS directories.
