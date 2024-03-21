# BotsRegulator
BotsRegulator(BR) is a Cloud Run service with a single cron job flow. BR does not accept incoming request except from this cron job. Periodically, BR retrieves specific UFS DUTs and update a specific GCE Provider config with these DUTs.

To pass a service account use `-service-account-json` flag.

## Context
[go/botsregulator](http://go/botsregulator)
[go/cloudbots-gce](http://go/cloudbots-gce)
[go/cloudbots](http://go/cloudbots)

## Dev
gcp project: bots-regulator-dev

## Production
gcp project: bots-regulator-prod

## Deployment
Deployment process can be found at [data/cloud-run/projects/bots-regulator](https://source.corp.google.com/h/chromium/infra/infra_superproject/+/main:data/cloud-run/projects/bots-regulator/).