package main

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, "chromium-build-stats-staging")
	panicIfErr(err)

	t := client.Dataset("ninjalog_test").Table("test_table")
	md, err := t.Metadata(ctx)
	panicIfErr(err)

	fmt.Printf("%#+v\n", md)

	md.Schema = bigquery.Schema{
		{
			Name: "build_id",
			Type: bigquery.IntegerFieldType,
		},
		{
			Name:     "targets",
			Type:     bigquery.StringFieldType,
			Repeated: true,
		},
		{
			Name: "step_name",
			Type: bigquery.StringFieldType,
		},
		{
			Name: "jobs",
			Type: bigquery.IntegerFieldType,
		},
		{
			Name: "os",
			Type: bigquery.StringFieldType,
		},
		{
			Name: "cpu_core",
			Type: bigquery.IntegerFieldType,
		},
		{
			Name: "build_configs",
			Type: bigquery.RecordFieldType,
			Schema: bigquery.Schema{
				{
					Name: "key",
					Type: bigquery.StringFieldType,
				},
				{
					Name: "value",
					Type: bigquery.StringFieldType,
				},
			},
			Repeated: true,
		},
		{
			Name: "log_entries",
			Type: bigquery.RecordFieldType,
			Schema: bigquery.Schema{
				{
					Name:     "outputs",
					Type:     bigquery.StringFieldType,
					Repeated: true,
				},
				{
					Name: "start_duration_sec",
					Type: bigquery.FloatFieldType,
				},
				{
					Name: "end_duration_sec",
					Type: bigquery.FloatFieldType,
				},
				{
					Name: "weighted_duration_sec",
					Type: bigquery.FloatFieldType,
				},
			},
			Repeated: true,
		},
		{
			Name: "created_at",
			Type: bigquery.TimestampFieldType,
		},
	}
	md.TimePartitioning = &bigquery.TimePartitioning{
		Type:                   bigquery.HourPartitioningType,
		RequirePartitionFilter: true,
		Field:                  "created_at",
		Expiration:             540 * 24 * time.Hour,
	}

	md2, err := t.Update(ctx,
		bigquery.TableMetadataToUpdate{
			Name: "new " + md.Name,
			// RequirePartitionFilter: true,
			Schema:         md.Schema,
			ExpirationTime: bigquery.NeverExpire,
			// TimePartitioning: md.TimePartitioning,
		},
		md.ETag)
	panicIfErr(err)

	fmt.Printf("%#+v\n", md2)
}
