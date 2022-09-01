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
