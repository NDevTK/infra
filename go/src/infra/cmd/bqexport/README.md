bqexport is a `go:generate` utility to generate Go structs based on BigQuery
schema protobufs compatible with [bqschemaupdater](../../tools/bqschemaupdater).
These structs can then be used in combination with
[eventupload](../../libs/eventupload) to send events to BigQuery.

# Usage

 Create a [TableDef protobuf](../../libs/bqschema/tabledef/table_def.proto).
 Once your protobuf has been reviewed and committed, run the bqexport command:
 `bqexport --help`. Don't have bqexport in your path? Try this inside the [infra
 Go env](../../../../env.py).
