# Creating a new raw events table

Table definitions are stored in infra.git
`go/src/infra/tools/bqschemaupdater/rawevents` and added/updated with the
[bqschemaupdater](../go/src/infra/tools/bqschemaupdater/README.md) tool. You
need to be authenticated in the chrome-infra-events project. Check your
authentication status:

```
gcloud info
```

If you don't have the Cloud SDK installed, follow the [installation
directions](https://cloud.google.com/sdk/docs/quickstarts).

If you don't see: `Project: [chrome-infra-events]`, reach out to an
[editor](https://pantheon.corp.google.com/iam-admin/iam/project?project=chrome-infra-events&organizationId=433637338589)
to request access.

To create a new raw events table:

```
cd go/src/infra/tools/bqschemaupdater  # In infra.git
touch rawevents/<table-id>.json
# Reference tabledef/table_def.proto for message format
# DatasetID is "raw_events"
go build
./bqschemaupdater --dryrun rawevents/<table-id>.json
# Looks good? Create CL for review...
./bqshemaupdater rawevents/<table-id>.json  # Actually create the table
```
