# Infra Stats UI

Project to display stats for infrastructure performance

## Backend

Running the server:

```sh
# Needs to be run the first time to set up BigQuery credentials
# This may also occasionally expire and need to be refreshed.
gcloud auth application-default login
go run main.go
```

This will set up the backend server running on port `8800`

## Frontend

Running the frontend:

```sh
cd frontend
npm install
npm start
```

This will set up the frontend client running on port `3000` with an automatic
proxy to the backend server running on `8800`.  To view the UI, go to
[localhost:3000](http://localhost:3000)

Formatting:

```sh
npm run fix
```

## Deployment

```sh
./deploy.sh
```

See the latest version at [https://chrome-infra-stats.googleplex.com/](https://chrome-infra-stats.googleplex.com/)

## Adding a new metric

### Create the update query
Create a new query in sql/cq_builder_metrics. A daily and weekly query should
both be created that insert rows into
chrome-trooper-analytics.metrics.cq_builder_metrics_day and
chrome-trooper-analytics.metrics.cq_builder_metrics_week respectively. Note the
comment at the top like:

```
-- The lines below are used by the deploy tool.
--name: Populate cq_builder_metrics_day slow test metrics
--schedule: every 4 hours synchronized
```

These lines are used to schedule how often the query will run when the
query gets deployed as a scheduled query in chrome-trooper-analytics. This
should typically be 4 hours for daily queries and 8 hours for weekly to keep
them in sync with the other scheduled query cadence

### Add the metric to the frontend

The metric should be added with the same name as the query to
frontend/src/features/dataSources/dataSourcesSlice.ts

### Get the CL reviewed

Before deploying the query should be reviewed since the deployment is done
manually

### Deploy the scheduled queries and backend

Follow instructions in sql/README.md to deploy the scheduled query. The
front and backend instructions are in the above Deployment section

### Backfill

Because the queries are and new queries should be based on independent, bucketed
days, a new backfill can be done by deleting existing data with:

```
DELETE FROM chrome-trooper-analytics.metrics.cq_builder_metrics_day WHERE metric IN ('METRIC TO REMOVE')
```

Make sure to not forget to remove the weekly metrics as well from
chrome-trooper-analytics.metrics.cq_builder_metrics_week. The back fill is then
normally done by modifying the query with a new start/end time and manually
running in pantheon

## Add a new alerts

The [alerts] are currently managed in PLX under the chromium-cq-metrics project
The code for which is kept in [piper]

### Unpiper the alerts

Follow the [README] in piper to un-piper the alerts. This will allow you to
manually modify the alerts or add new ones 

### Create/modify a new rule

After the project has been un-pipered the [alerts] should be modifiable. Follow
the dialogues to configure the alert however you want. The existing ones use
autofocus to manage alerting thresholds which has worked well so far.

### Serialize the alert

Again follow the [README] to update cbi_stats_alerts. This will serialize the
alerts in your current workspace and allow you to create a CL to make the
change permanent

### Land the change and deploy

After getting the CL reviewed run the :deploy target in [piper]. This will
disable manual editing and deploy the current configuration to plx

## [Roadmap](ROADMAP.md)

[alerts]: https://plx-alerts.corp.google.com/project/chromium-cq-metrics/alerts
[piper]: https://source.corp.google.com/piper///depot/google3/chrome/ops/browser_infra/plx/cbi_stats/alerts/
[README]: https://source.corp.google.com/piper///depot/google3/chrome/ops/browser_infra/plx/cbi_stats/alerts/README.md