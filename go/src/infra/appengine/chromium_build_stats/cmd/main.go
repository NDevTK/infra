package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/bq"

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

	// ninjalogFile

	f, err := os.Open(*ninjalogFile)
	panicIfErr(err)
	defer f.Close()

	nlog, err := ninjalog.Parse(*ninjalogFile, f)
	panicIfErr(err)

	client, err := bigquery.NewClient(ctx, "chromium-build-stats-staging")
	panicIfErr(err)

	up := bq.NewUploader(ctx, client, "ninjalog", "test")
	ninjaproto := ninjalog.ToProto(nlog)
	fmt.Printf("%d\n", len(ninjaproto.LogEntries))
	ninjaproto.LogEntries = ninjaproto.LogEntries[:40000]
	err = up.Put(ctx, ninjaproto)
	panicIfErr(err)

	fmt.Printf("put %d entries\n", len(ninjaproto.LogEntries))
}
