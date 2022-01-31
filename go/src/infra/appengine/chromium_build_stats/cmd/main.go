package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"cloud.google.com/go/bigquery/storage/managedwriter"
	"cloud.google.com/go/bigquery/storage/managedwriter/adapt"
	"google.golang.org/protobuf/proto"

	"infra/appengine/chromium_build_stats/ninjalog"
)

var ninjalogFile = flag.String("ninjalog", "", "Path to ninja_log file to upload.")

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	ctx := context.Background()

	client, err := managedwriter.NewClient(ctx, "chromium-build-stats-staging") // option.WithGRPCDialOption(grpc.WithCompressor(grpc.NewGZIPCompressor()))

	panicIfErr(err)
	defer client.Close()

	// ninjalogFile
	f, err := os.Open(*ninjalogFile)
	panicIfErr(err)
	defer f.Close()

	nlog, err := ninjalog.Parse(*ninjalogFile, f)
	panicIfErr(err)

	// client, err := bigquery.NewClient(ctx, "chromium-build-stats-staging")
	// panicIfErr(err)

	// up := bq.NewUploader(ctx, client, "ninjalog", "test")
	ninjaproto := ninjalog.ToProto(nlog)
	fmt.Printf("%d\n", len(ninjaproto.LogEntries))
	ninjaproto.LogEntries = ninjaproto.LogEntries

	descriptorProto, err := adapt.NormalizeDescriptor(ninjaproto.ProtoReflect().Descriptor())
	panicIfErr(err)

	tableName := fmt.Sprintf("projects/%s/datasets/%s/tables/%s", "chromium-build-stats-staging", "ninjalog", "test")

	managedStream, err := client.NewManagedStream(ctx,
		managedwriter.WithDestinationTable(tableName),
		managedwriter.WithSchemaDescriptor(descriptorProto))
	panicIfErr(err)

	fmt.Println("hoge")

	ninjaproto.LogEntries = ninjaproto.LogEntries[:70000]

	// Encode the messages into binary format.
	b, err := proto.Marshal(ninjaproto)
	panicIfErr(err)

	fmt.Printf("size: %d\n", len(b))

	// Send the rows to the service, and specify an offset for managing deduplication.
	result, err := managedStream.AppendRows(ctx, [][]byte{b})

	fmt.Println("append")

	// Block until the write is complete and return the result.
	_, err = result.GetResult(ctx)
	panicIfErr(err)

	fmt.Println("results")

	// err = up.Put(ctx, ninjaproto)
	// panicIfErr(err)

	fmt.Printf("put %d entries\n", len(ninjaproto.LogEntries))
}
