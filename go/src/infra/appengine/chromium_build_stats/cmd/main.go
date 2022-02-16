package main

import (
	_ "embed"
	"flag"
	"os"
	"time"

	"github.com/linkedin/goavro/v2"

	"infra/appengine/chromium_build_stats/ninjalog"
)

var ninjalogFile = flag.String("ninjalog", "", "Path to ninja_log file to upload.")

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

//go:embed ninjalog.json
var avroSchema string

func main() {
	flag.Parse()

	codec, err := goavro.NewCodec(avroSchema)
	panicIfErr(err)

	ninjalogf, err := os.Open(*ninjalogFile)
	panicIfErr(err)
	defer func() {
		panicIfErr(ninjalogf.Close())
	}()

	nl, err := ninjalog.Parse("ninjalog", ninjalogf)
	panicIfErr(err)
	nproto := ninjalog.ToProto(nl)

	// Convert native Go form to binary Avro data
	var data = map[string]interface{}{
		"build_id":   nproto.BuildId,
		"targets":    nproto.Targets,
		"jobs":       nproto.Jobs,
		"os":         nproto.Os.String(),
		"cpu_core":   nproto.CpuCore,
		"created_at": time.Now().UnixNano() / 1000,
	}

	var build_configs []map[string]interface{}
	for _, kv := range nproto.BuildConfigs {
		build_configs = append(build_configs, map[string]interface{}{
			"key":   kv.Key,
			"value": kv.Value,
		})
	}

	data["build_configs"] = build_configs

	var log_entries []map[string]interface{}
	for _, LogEntry := range nproto.LogEntries {
		log_entries = append(log_entries, map[string]interface{}{
			"outputs":               LogEntry.Outputs,
			"start_duration_sec":    LogEntry.StartDurationSec,
			"end_duration_sec":      LogEntry.EndDurationSec,
			"weighted_duration_sec": LogEntry.WeightedDurationSec,
		})
	}

	data["log_entries"] = log_entries

	// _, err = codec.BinaryFromNative(nil, data)
	// panicIfErr(err)

	f, err := os.Create("test.avro")
	panicIfErr(err)
	defer func() {
		panicIfErr(f.Close())
	}()

	wr, err := goavro.NewOCFWriter(goavro.OCFConfig{
		W:     f,
		Codec: codec,
	})
	panicIfErr(err)

	panicIfErr(wr.Append([]interface{}{data}))

	// panicIfErr(os.WriteFile("test.avro", binary, 0o666))
}
