// Copyright 2022 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"github.com/maruel/subcommands"
	"os"

	"github.com/golang/protobuf/jsonpb"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/errors"
	sinkpb "go.chromium.org/luci/resultdb/sink/proto/v1"
)

func cmdNative() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: `native [flags] TEST_CMD [TEST_ARG]...`,
		ShortDesc: "Batch upload results of native sinkpb.TestResult format to ResultSink",
		LongDesc: text.Doc(`
			Runs the test command and waits for it to finish. The result file should be a jsonl, where
			each line is a json string of sinkpb.TestResult message. Native adapter uploads them directly
			to ResultDB via ResultSink.
		`),
		CommandRun: func() subcommands.CommandRun {
			r := &nativeRun{}
			r.baseRun.RegisterGlobalFlags()
			return r
		},
	}
}

type nativeRun struct {
	baseRun
}

func (r *nativeRun) Run(a subcommands.Application, args []string, env subcommands.Env) (ret int) {
	if err := r.baseRun.validate(); err != nil {
		return r.done(err)
	}

	ctx := cli.GetContext(a, r, env)
	return r.run(ctx, args, r.generateTestResults)
}

// generateTestResults converts test results from results file to sinkpb.TestResult.
func (r *nativeRun) generateTestResults(ctx context.Context, _ []byte) ([]*sinkpb.TestResult, error) {
	// Get results.
	f, err := os.Open(r.resultFile)
	if err != nil {
		return nil, errors.Annotate(err, "open result file").Err()
	}
	defer f.Close()

	trs := make([]*sinkpb.TestResult, 0)
	decoder := json.NewDecoder(f)
	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: true}
	for decoder.More() {
		var m sinkpb.TestResult
		if err := unmarshaler.UnmarshalNext(decoder, &m); err != nil {
			return nil, errors.Annotate(err, "failed to transform jsonl to pb").Err()
		}
		trs = append(trs, &m)
	}
	return trs, nil
}
